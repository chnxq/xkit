package codegen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"github.com/chnxq/xkit/internal/binding"
	codegentemplate "github.com/chnxq/xkit/internal/codegen/template"
	"github.com/chnxq/xkit/internal/config"
	"github.com/chnxq/xkit/internal/entschema"
	"github.com/chnxq/xkit/internal/project"
)

func (r *Runner) sharedModuleDir() string {
	return filepath.Join(r.project.Root, "shared", "modulex")
}

func (r *Runner) sharedModuleImport() string {
	return filepath.ToSlash(filepath.Join(r.project.Module, "shared", "modulex"))
}

func (r *Runner) normalizeConfiguredImportPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}

	path = strings.ReplaceAll(path, "{{module}}", r.layout.ModuleImport)
	path = filepath.ToSlash(path)

	if !r.isModuleMode() {
		return path
	}

	projectModule := strings.TrimSuffix(filepath.ToSlash(r.project.Module), "/")
	layoutModule := strings.TrimSuffix(filepath.ToSlash(r.layout.ModuleImport), "/")

	switch {
	case path == projectModule:
		return layoutModule
	case strings.HasPrefix(path, projectModule+"/api/gen/"):
		return strings.Replace(path, projectModule+"/api/gen/", layoutModule+"/api/gen/", 1)
	case strings.HasPrefix(path, projectModule+"/data/"):
		return strings.Replace(path, projectModule+"/data/", layoutModule+"/data/", 1)
	case strings.HasPrefix(path, projectModule+"/service/"):
		return strings.Replace(path, projectModule+"/service/", layoutModule+"/service/", 1)
	case strings.HasPrefix(path, projectModule+"/server/"):
		return strings.Replace(path, projectModule+"/server/", layoutModule+"/server/", 1)
	case strings.HasPrefix(path, projectModule+"/bootstrap/"):
		return strings.Replace(path, projectModule+"/bootstrap/", layoutModule+"/bootstrap/", 1)
	case strings.HasPrefix(path, projectModule+"/shared/modulex"):
		return r.sharedModuleImport()
	}

	return path
}

func (r *Runner) ensureModuleSharedExtFile(result *Result, needsIdentity bool) error {
	if !r.isModuleMode() {
		return nil
	}
	content, err := renderTemplate(codegentemplate.ModuleSharedExt, struct {
		templateBase
		NeedsIdentity bool
	}{
		templateBase:  r.templateBase(),
		NeedsIdentity: needsIdentity,
	})
	if err != nil {
		return err
	}
	path := filepath.Join(r.sharedModuleDir(), "module_shared_ext.go")
	return r.writeFile(path, content, result, true)
}

func resolveTypeScriptRoot(projectRoot, configured string) (string, error) {
	configured = strings.TrimSpace(configured)
	if configured == "" {
		return filepath.Join(filepath.Dir(projectRoot), filepath.Base(projectRoot)+"-ui"), nil
	}
	if filepath.IsAbs(configured) {
		return filepath.Clean(configured), nil
	}
	return filepath.Clean(filepath.Join(filepath.Dir(projectRoot), configured)), nil
}

func buildTreeConfig(plan *config.TreeConfig) *treeConfigData {
	if plan == nil {
		return nil
	}
	listMethod := strings.TrimSpace(plan.ListMethod)
	if listMethod == "" {
		listMethod = "ListTree"
	}
	return &treeConfigData{
		ParentField:   strings.TrimSpace(plan.ParentField),
		PathField:     strings.TrimSpace(plan.PathField),
		ChildrenField: strings.TrimSpace(plan.ChildrenField),
		ListMethod:    listMethod,
	}
}

func refreshExtensionHeader(existing, generated []byte) []byte {
	generatedHeader, _ := splitLeadingCommentAndBody(generated)
	_, existingBody := splitLeadingCommentAndBody(existing)
	if len(existingBody) == 0 {
		return generated
	}
	return append(append([]byte{}, generatedHeader...), existingBody...)
}

func splitLeadingCommentAndBody(content []byte) ([]byte, []byte) {
	lines := bytes.SplitAfter(content, []byte("\n"))
	headerEnd := 0
	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			headerEnd += len(line)
			continue
		}
		if bytes.HasPrefix(trimmed, []byte("//")) {
			headerEnd += len(line)
			continue
		}
		break
	}
	return content[:headerEnd], content[headerEnd:]
}

func receiverMethodNames(dir, receiverName string, skipFiles map[string]struct{}) map[string]struct{} {
	out := make(map[string]struct{})
	entries, err := os.ReadDir(dir)
	if err != nil {
		return out
	}

	fileSet := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		if _, skip := skipFiles[entry.Name()]; skip {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		file, err := parser.ParseFile(fileSet, path, nil, parser.SkipObjectResolution)
		if err != nil {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || fn.Name == nil {
				continue
			}
			if receiverMatches(fn.Recv, receiverName) {
				out[fn.Name.Name] = struct{}{}
			}
		}
	}
	return out
}

func receiverMatches(recv *ast.FieldList, receiverName string) bool {
	if recv == nil || len(recv.List) == 0 {
		return false
	}
	switch expr := recv.List[0].Type.(type) {
	case *ast.StarExpr:
		if ident, ok := expr.X.(*ast.Ident); ok {
			return ident.Name == receiverName
		}
	case *ast.Ident:
		return expr.Name == receiverName
	}
	return false
}

func exprString(expr ast.Expr) string {
	if expr == nil {
		return ""
	}
	var builder strings.Builder
	if err := format.Node(&builder, token.NewFileSet(), expr); err != nil {
		return ""
	}
	return builder.String()
}

func repoSetters(fields []entschema.Field, dtoFields map[string]string, methodName, entPackage, enumHelperName, timeHelperName, fieldMaskHelperName, zeroReturn string) []setterData {
	setters := make([]setterData, 0, len(fields))
	for _, field := range fields {
		if field.Name == "id" {
			continue
		}
		if !supportsGeneratedSetterKind(field.Kind) {
			continue
		}
		if skipGeneratedSetter(field.Name) {
			continue
		}
		if methodName == "Update" && field.Immutable {
			continue
		}
		entName := toGoName(field.Name)
		dtoName := toPascal(field.Name)
		dtoType := strings.TrimSpace(dtoFields[dtoName])
		if len(dtoFields) > 0 {
			if dtoType == "" {
				continue
			}
		}
		directMethod := "Set" + entName
		directExpr := directSetterExpr(field.Kind, entPackage, entName, dtoName, dtoType, enumHelperName, timeHelperName)
		pre := ""
		if field.Kind == "JSON" && dtoType == "*string" {
			pre = jsonStringSetterPre(entName, dtoName, zeroReturn)
			directExpr = jsonStringSetterValueExpr(dtoName, entName)
		}
		preferNillable := prefersNillableSetter(field.Kind, dtoType)
		if !field.Optional && !field.Default {
			setters = append(setters, setterData{
				Method: directMethod,
				Expr:   directExpr,
				Kind:   field.Kind,
			})
			continue
		}
		if methodName == "Create" {
			if field.Default && !field.Optional {
				setters = append(setters, setterData{
					Method:    directMethod,
					Expr:      directExpr,
					Kind:      field.Kind,
					Condition: "req.Data." + dtoName + " != nil",
					Pre:       pre,
				})
				continue
			}
			if field.Nillable || preferNillable {
				setters = append(setters, setterData{
					Method: "SetNillable" + entName,
					Expr:   nillableSetterExpr(field.Kind, dtoName, dtoType, entPackage, entName, enumHelperName, timeHelperName),
					Kind:   field.Kind,
					Pre:    pre,
				})
				continue
			}
			setters = append(setters, setterData{
				Method:    directMethod,
				Expr:      directExpr,
				Kind:      field.Kind,
				Condition: "req.Data." + dtoName + " != nil",
				Pre:       pre,
			})
			continue
		}
		if field.Default && !field.Optional {
			setters = append(setters, setterData{
				Method:    directMethod,
				Expr:      directExpr,
				Kind:      field.Kind,
				Condition: "req.Data." + dtoName + " != nil",
				Pre:       pre,
			})
			continue
		}
		updateMethod := directMethod
		updateExpr := directExpr
		if field.Nillable || preferNillable {
			updateMethod = "SetNillable" + entName
			updateExpr = nillableSetterExpr(field.Kind, dtoName, dtoType, entPackage, entName, enumHelperName, timeHelperName)
		}
		setters = append(setters, setterData{
			Method:         updateMethod,
			Expr:           updateExpr,
			Kind:           field.Kind,
			Condition:      "req.Data." + dtoName + " != nil",
			ClearMethod:    "Clear" + entName,
			ClearCondition: optionalFieldClearCondition(fieldMaskHelperName, field.Name),
			Pre:            pre,
		})
	}
	return setters
}

func directSetterExpr(kind, entPackage, entName, dtoName, dtoType, enumHelperName, timeHelperName string) string {
	switch kind {
	case "Enum":
		return fmt.Sprintf("%s.%s(req.Data.Get%s().String())", entPackage, entName, dtoName)
	case "Time":
		return "req.Data.Get" + dtoName + "().AsTime()"
	default:
		return "req.Data.Get" + dtoName + "()"
	}
}

func nillableSetterExpr(kind, dtoName, dtoType, entPackage, entName, enumHelperName, timeHelperName string) string {
	switch kind {
	case "Enum":
		return fmt.Sprintf("%s[%s.%s](req.Data.%s)", enumHelperName, entPackage, entName, dtoName)
	case "Time":
		return timeHelperName + "(req.Data." + dtoName + ")"
	case "JSON":
		if dtoType == "*string" {
			return jsonStringSetterValueExpr(dtoName, entName)
		}
		return "req.Data." + dtoName
	default:
		return "req.Data." + dtoName
	}
}

func prefersNillableSetter(kind, dtoType string) bool {
	if !strings.HasPrefix(strings.TrimSpace(dtoType), "*") {
		return false
	}
	switch kind {
	case "Bool", "Uint", "Uint8", "Uint16", "Uint32", "Uint64", "Int", "Int8", "Int16", "Int32", "Int64", "Float", "Float32", "String", "Enum", "Time":
		return true
	default:
		return false
	}
}

func jsonStringSetterPre(entName, dtoName, zeroReturn string) string {
	valueName := jsonStringSetterValueExpr(dtoName, entName)
	baseName := strings.TrimSuffix(valueName, "Value")
	jsonName := baseName + "JSON"
	return strings.Join([]string{
		fmt.Sprintf("%s := strings.TrimSpace(req.Data.Get%s())", jsonName, dtoName),
		fmt.Sprintf("%s := map[string]any{}", valueName),
		fmt.Sprintf("if %s != \"\" {", jsonName),
		fmt.Sprintf("\tif err := json.Unmarshal([]byte(%s), &%s); err != nil {", jsonName, valueName),
		"\t\treturn " + zeroReturn + ", err",
		"\t}",
		"}",
	}, "\n")
}

func jsonStringSetterValueExpr(dtoName, entName string) string {
	baseName := lowerFirst(dtoName)
	baseName = strings.TrimSuffix(baseName, "Json")
	baseName = strings.TrimSuffix(baseName, "JSON")
	if baseName == "" {
		baseName = lowerFirst(entName)
	}
	return baseName + "Value"
}

func optionalFieldClearCondition(helperName, fieldName string) string {
	return fmt.Sprintf("req.GetUpdateMask() != nil && %s(req.GetUpdateMask().GetPaths(), %s)", helperName, quotedStringList(fieldMaskCandidates(fieldName)))
}

func fieldMaskCandidates(fieldName string) []string {
	candidates := []string{fieldName}
	camel := lowerFirst(toPascal(fieldName))
	if camel != fieldName {
		candidates = append(candidates, camel)
	}
	return candidates
}

func quotedStringList(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, fmt.Sprintf("%q", value))
	}
	return strings.Join(quoted, ", ")
}

func defaultListSortField(plan resourcePlan) string {
	if isRecordLikeResource(plan) && hasField(plan.Schema.Fields, "created_at") {
		return "created_at"
	}
	if isRecentEntityResource(plan) {
		return "id"
	}
	for _, candidate := range []string{"sort_order", "order", "sortOrder"} {
		if hasField(plan.Schema.Fields, candidate) {
			return candidate
		}
	}
	return "id"
}

func defaultListSortDirection(plan resourcePlan) string {
	if isRecordLikeResource(plan) && hasField(plan.Schema.Fields, "created_at") {
		return "DESC"
	}
	if isRecentEntityResource(plan) {
		return "DESC"
	}
	return "ASC"
}

func hasField(fields []entschema.Field, name string) bool {
	for _, field := range fields {
		if field.Name == name {
			return true
		}
	}
	return false
}

func isRecordLikeResource(plan resourcePlan) bool {
	name := strings.ToLower(plan.Resource.Name)
	entity := strings.ToLower(plan.Resource.Entity)
	return strings.Contains(name, "log") || strings.Contains(entity, "log") || strings.Contains(name, "record") || strings.Contains(entity, "record")
}

func isRecentEntityResource(plan resourcePlan) bool {
	name := strings.ToLower(plan.Resource.Name)
	entity := strings.ToLower(plan.Resource.Entity)
	for _, keyword := range []string{"message", "task"} {
		if strings.Contains(name, keyword) || strings.Contains(entity, keyword) {
			if strings.Contains(name, "category") || strings.Contains(entity, "category") {
				return false
			}
			return true
		}
	}
	return false
}

func repoAuditSetters(fields []entschema.Field, methodName string) []setterData {
	setters := make([]setterData, 0, len(fields))
	for _, field := range fields {
		auditKind := auditFieldKind(field.Name)
		if auditKind == "" {
			continue
		}
		switch methodName {
		case "Create":
			if auditKind != "create_time" && auditKind != "create_user" {
				continue
			}
		case "Update":
			if auditKind != "update_time" && auditKind != "update_user" {
				continue
			}
		default:
			continue
		}
		if !supportsGeneratedSetterKind(field.Kind) {
			continue
		}

		method := "Set" + toGoName(field.Name)
		expr := ""
		switch auditKind {
		case "create_time", "update_time":
			expr = auditTimeExpr(field.Kind)
			if expr == "" {
				continue
			}
		case "create_user", "update_user":
			expr = auditUserExpr(field.Kind)
			if expr == "" {
				continue
			}
		default:
			continue
		}

		setters = append(setters, setterData{
			Method: method,
			Expr:   expr,
			Kind:   field.Kind,
		})
	}
	return setters
}

func auditTimeExpr(kind string) string {
	switch kind {
	case "Time":
		return "now"
	case "Int64":
		return "now.UnixMilli()"
	case "Uint64":
		return "uint64(now.UnixMilli())"
	default:
		return ""
	}
}

func hasGeneratedAuditFields(fields []entschema.Field) bool {
	for _, field := range fields {
		if auditFieldKind(field.Name) != "" {
			return true
		}
	}
	return false
}

func auditFieldKind(fieldName string) string {
	switch fieldName {
	case "created_at", "create_at", "create_time":
		return "create_time"
	case "updated_at", "update_at", "update_time":
		return "update_time"
	case "created_by", "create_by", "creator_id":
		return "create_user"
	case "updated_by", "update_by":
		return "update_user"
	default:
		return ""
	}
}

func auditUserExpr(kind string) string {
	switch kind {
	case "Uint":
		return "uint(viewer.UserID())"
	case "Uint8":
		return "uint8(viewer.UserID())"
	case "Uint16":
		return "uint16(viewer.UserID())"
	case "Uint32":
		return "uint32(viewer.UserID())"
	case "Uint64":
		return "viewer.UserID()"
	case "Int":
		return "int(viewer.UserID())"
	case "Int8":
		return "int8(viewer.UserID())"
	case "Int16":
		return "int16(viewer.UserID())"
	case "Int32":
		return "int32(viewer.UserID())"
	case "Int64":
		return "int64(viewer.UserID())"
	default:
		return ""
	}
}

func supportsGeneratedSetterKind(kind string) bool {
	switch kind {
	case "String", "Enum", "Time", "Bool", "Uint", "Uint8", "Uint16", "Uint32", "Uint64", "Int", "Int8", "Int16", "Int32", "Int64", "Float", "Float32", "JSON":
		return true
	default:
		return false
	}
}

func settersNeedFieldMaskHelper(setters []setterData) bool {
	for _, setter := range setters {
		if setter.ClearCondition != "" {
			return true
		}
	}
	return false
}

func settersUsePre(setters []setterData) bool {
	for _, setter := range setters {
		if strings.TrimSpace(setter.Pre) != "" {
			return true
		}
	}
	return false
}

func skipGeneratedSetter(fieldName string) bool {
	if auditFieldKind(fieldName) != "" {
		return true
	}
	switch fieldName {
	case "deleted_at", "delete_at", "delete_time", "deleted_by", "delete_by":
		return true
	default:
		return false
	}
}

func settersUseKind(setters []setterData, kind string) bool {
	for _, setter := range setters {
		if setter.Kind == kind {
			return true
		}
	}
	return false
}

func auditSettersUseExpr(setters []setterData, expr string) bool {
	for _, setter := range setters {
		if setter.Expr == expr {
			return true
		}
	}
	return false
}

func auditSettersUseExprContains(setters []setterData, value string) bool {
	for _, setter := range setters {
		if strings.Contains(setter.Expr, value) {
			return true
		}
	}
	return false
}

func repoFilters(fields []entschema.Field, allowed []string) []filterData {
	if len(allowed) == 0 {
		return nil
	}
	fieldByName := make(map[string]entschema.Field, len(fields)+1)
	fieldByName["id"] = entschema.Field{Name: "id", Kind: "Uint32"}
	for _, field := range fields {
		fieldByName[field.Name] = field
	}

	filters := make([]filterData, 0, len(allowed))
	seen := make(map[string]struct{})
	for _, name := range allowed {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		field, ok := fieldByName[name]
		if !ok {
			continue
		}
		kind := filterKind(field.Kind)
		if kind == "" {
			continue
		}
		filters = append(filters, filterData{
			Field:        name,
			Predicate:    toGoName(name),
			Kind:         kind,
			CastType:     filterCastType(field.Kind),
			ParseBitSize: filterParseBitSize(field.Kind),
			TimeField:    name,
			SupportsIn:   kind == "Uint" || kind == "Int",
		})
		seen[name] = struct{}{}
	}
	return filters
}

func effectiveFilterAllow(plan resourcePlan) []string {
	allowed := append([]string{}, plan.Resource.Filters.Allow...)
	if isAutoAppendCreatedAtFilterResource(plan) && hasField(plan.Schema.Fields, "created_at") && !slices.Contains(allowed, "created_at") {
		allowed = append(allowed, "created_at")
	}
	return allowed
}

func isAutoAppendCreatedAtFilterResource(plan resourcePlan) bool {
	name := strings.ToLower(plan.Resource.Name)
	entity := strings.ToLower(plan.Resource.Entity)
	return strings.HasSuffix(name, "_log") || strings.HasSuffix(entity, "log")
}

func idGoType(fields []entschema.Field) string {
	for _, field := range fields {
		if field.Name == "id" {
			return fieldGoType(field.Kind)
		}
	}
	return "uint32"
}

func fieldGoType(kind string) string {
	switch kind {
	case "Uint":
		return "uint"
	case "Uint8":
		return "uint8"
	case "Uint16":
		return "uint16"
	case "Uint32":
		return "uint32"
	case "Uint64":
		return "uint64"
	case "Int":
		return "int"
	case "Int8":
		return "int8"
	case "Int16":
		return "int16"
	case "Int32":
		return "int32"
	case "Int64":
		return "int64"
	default:
		return "uint32"
	}
}

func filterCastType(kind string) string {
	switch kind {
	case "Uint", "Uint8", "Uint16", "Uint32", "Uint64", "Int", "Int8", "Int16", "Int32", "Int64":
		return fieldGoType(kind)
	default:
		return ""
	}
}

func filterParseBitSize(kind string) string {
	switch kind {
	case "Uint8", "Int8":
		return "8"
	case "Uint16", "Int16":
		return "16"
	case "Uint32", "Int32":
		return "32"
	case "Uint64", "Int64":
		return "64"
	default:
		return "0"
	}
}

func filterKind(kind string) string {
	switch kind {
	case "String":
		return "String"
	case "Enum":
		return "Enum"
	case "Time":
		return "Time"
	case "Uint", "Uint8", "Uint16", "Uint32", "Uint64":
		return "Uint"
	case "Int", "Int8", "Int16", "Int32", "Int64":
		return "Int"
	default:
		return ""
	}
}

func sanitizeIdentifier(value string) string {
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, "-", "")
	if value == "" {
		return "pkg"
	}
	return value
}

func toPascal(value string) string {
	if value == "" {
		return ""
	}
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	for index, part := range parts {
		if part == "" {
			continue
		}
		parts[index] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
	}
	return strings.Join(parts, "")
}

func toGoName(value string) string {
	if value == "" {
		return ""
	}
	initialisms := map[string]string{
		"api":  "API",
		"http": "HTTP",
		"id":   "ID",
		"ip":   "IP",
		"guid": "GUID",
		"json": "JSON",
		"sql":  "SQL",
		"url":  "URL",
		"uri":  "URI",
	}
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	for index, part := range parts {
		if part == "" {
			continue
		}
		if replacement, ok := initialisms[strings.ToLower(part)]; ok {
			parts[index] = replacement
			continue
		}
		parts[index] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
	}
	return strings.Join(parts, "")
}

func entOperationName(entityName string) string {
	switch strings.ToLower(entityName) {
	case "api":
		return "API"
	case "http":
		return "HTTP"
	case "id":
		return "ID"
	case "ip":
		return "IP"
	case "url":
		return "URL"
	case "uri":
		return "URI"
	default:
		return entityName
	}
}

func lowerFirst(value string) string {
	if value == "" {
		return ""
	}
	return strings.ToLower(value[:1]) + value[1:]
}

func repoVarFromInterface(repoName string) string {
	repoName = strings.TrimSpace(repoName)
	if repoName == "" {
		return ""
	}
	return lowerFirst(strings.TrimSuffix(repoName, "Repo")) + "Repo"
}

func upperFirst(value string) string {
	if value == "" {
		return ""
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

func nameParams(types []string) []namedType {
	params := make([]namedType, 0, len(types))
	for index, typeText := range types {
		name := fmt.Sprintf("arg%d", index)
		if index == 0 && typeText == "context.Context" {
			name = "ctx"
		}
		if index == 1 {
			name = "req"
		}
		params = append(params, namedType{
			Name: name,
			Type: typeText,
		})
	}
	return params
}

func firstResult(results []string) string {
	if len(results) == 0 {
		return "any"
	}
	return results[0]
}

func generatedContentEquivalent(existing, generated []byte) bool {
	if bytes.Equal(existing, generated) {
		return true
	}
	return bytes.Equal(normalizeGeneratedHeader(existing), normalizeGeneratedHeader(generated))
}

func normalizeGeneratedHeader(content []byte) []byte {
	content = bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
	lines := bytes.SplitAfter(content, []byte("\n"))
	normalized := make([][]byte, 0, len(lines))
	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if bytes.HasPrefix(trimmed, []byte("// generated at:")) {
			normalized = append(normalized, []byte("// generated at: <normalized>\n"))
			continue
		}
		normalized = append(normalized, line)
	}
	return bytes.Join(normalized, nil)
}

func renderTemplate(source string, data any) ([]byte, error) {
	content, err := renderRawTemplate(source, data)
	if err != nil {
		return nil, err
	}

	formatted, err := format.Source(content)
	if err != nil {
		return nil, fmt.Errorf("format generated source: %w", err)
	}

	return formatted, nil
}

func renderAnyTemplate(source string, data any) ([]byte, error) {
	content, err := renderRawTemplate(source, data)
	if err != nil {
		return nil, err
	}
	if looksLikeGoSource(content) {
		formatted, err := format.Source(content)
		if err != nil {
			return nil, fmt.Errorf("format generated source: %w", err)
		}
		return formatted, nil
	}
	return content, nil
}

func renderRawTemplate(source string, data any) ([]byte, error) {
	tmpl, err := template.New("file").Funcs(template.FuncMap{
		"trimPointer": trimPointer,
		"upperFirst":  upperFirst,
	}).Parse(source)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}
	return buf.Bytes(), nil
}

func looksLikeGoSource(content []byte) bool {
	trimmed := strings.TrimSpace(string(content))
	return strings.Contains(trimmed, "\npackage ") || strings.HasPrefix(trimmed, "package ") || strings.Contains(trimmed, "\n\npackage ")
}

func uniqueImports(imports []importSpec) []importSpec {
	seen := make(map[string]struct{})
	seenPathWithoutAlias := make(map[string]struct{})
	var out []importSpec
	for _, spec := range imports {
		if spec.Alias == "" {
			seenPathWithoutAlias[spec.Path] = struct{}{}
		}
	}
	for _, spec := range imports {
		if spec.Alias != "" {
			if _, ok := seenPathWithoutAlias[spec.Path]; ok {
				continue
			}
		}
		key := spec.Alias + "|" + spec.Path
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, spec)
	}
	slices.SortFunc(out, func(a, b importSpec) int {
		if a.Path == b.Path {
			return strings.Compare(a.Alias, b.Alias)
		}
		return strings.Compare(a.Path, b.Path)
	})
	return out
}

func aliasesInType(typeText string) []string {
	matches := aliasPattern.FindAllStringSubmatch(typeText, -1)
	aliases := make([]string, 0, len(matches))
	seen := make(map[string]struct{})
	for _, match := range matches {
		if len(match) != 2 {
			continue
		}
		alias := match[1]
		if _, ok := seen[alias]; ok {
			continue
		}
		seen[alias] = struct{}{}
		aliases = append(aliases, alias)
	}
	return aliases
}

func apiAlias(importPath string) string {
	parts := strings.Split(importPath, "/")
	if len(parts) >= 2 && parts[len(parts)-1] == "v1" {
		return sanitizeIdentifier(parts[len(parts)-2] + "v1")
	}
	return sanitizeIdentifier(parts[len(parts)-1])
}

func serviceEmbeds(alias string, binding binding.ServiceBinding) []string {
	return []string{
		alias + "." + bindingHTTPServerInterfaceName(binding),
		alias + ".Unimplemented" + strings.TrimSuffix(bindingGRPCServerInterfaceName(binding), "Server") + "Server",
	}
}

func bindingGRPCServerInterfaceName(binding binding.ServiceBinding) string {
	if strings.TrimSpace(binding.InterfaceName) != "" {
		return binding.InterfaceName
	}
	return binding.ServiceName + "Server"
}

func bindingHTTPServerInterfaceName(binding binding.ServiceBinding) string {
	name := bindingGRPCServerInterfaceName(binding)
	if strings.HasSuffix(name, "Server") {
		return strings.TrimSuffix(name, "Server") + "HTTPServer"
	}
	return binding.ServiceName + "HTTPServer"
}

func isCRUDMethod(name string) bool {
	switch name {
	case "List", "Get", "Create", "Update", "Delete", "Count", "Exists":
		return true
	}
	return strings.HasSuffix(name, "Exists") || strings.HasPrefix(name, "Count")
}

func repoMethodKind(name string) string {
	switch {
	case name == "Create":
		return "create"
	case name == "Update":
		return "update"
	case name == "Delete":
		return "delete"
	case name == "Exists":
		return "exists"
	case strings.HasSuffix(name, "Exists"):
		return "query_exists"
	default:
		return strings.ToLower(name)
	}
}

func resourceOperationEnabled(resource config.Resource, kind string) bool {
	if len(resource.Operations) == 0 {
		return true
	}

	operation := kind
	if kind == "query_exists" {
		operation = "exists"
	}

	enabled, ok := resource.Operations[operation]
	return ok && enabled
}

func existsCases(alias, requestType string, fields []string) []existsCaseData {
	requestName := trimPointer(requestType)
	if requestName == "" || requestName == requestType {
		return nil
	}
	requestName = strings.TrimPrefix(requestName, alias+".")

	fieldNames := map[string]struct{}{}
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field != "" {
			fieldNames[field] = struct{}{}
		}
	}

	var cases []existsCaseData
	for fieldName := range fieldNames {
		protoName := toPascal(fieldName)
		goName := toGoName(fieldName)
		getterName := protoName
		predicateName := goName
		if fieldName == "id" {
			predicateName = "ID"
		}
		cases = append(cases, existsCaseData{
			OneofType: alias + "." + requestName + "_" + protoName,
			Predicate: predicateName + "EQ",
			ValueExpr: "req.Get" + getterName + "()",
		})
	}
	slices.SortFunc(cases, func(a, b existsCaseData) int {
		if strings.HasSuffix(a.OneofType, "_Id") {
			return -1
		}
		if strings.HasSuffix(b.OneofType, "_Id") {
			return 1
		}
		return strings.Compare(a.OneofType, b.OneofType)
	})
	return cases
}

func requestParamType(types []string) string {
	if len(types) >= 2 {
		return types[1]
	}
	if len(types) == 1 {
		return types[0]
	}
	return ""
}

func pagingRequestExpr(params []namedType) string {
	if len(params) < 2 {
		return "nil"
	}
	requestType := strings.TrimSpace(strings.TrimPrefix(params[1].Type, "*"))
	if strings.HasSuffix(requestType, ".PagingRequest") {
		return params[1].Name
	}
	return params[1].Name + ".GetPaging()"
}

func trimPointer(typeText string) string {
	return strings.TrimPrefix(typeText, "*")
}

func nilReturn(typeText string) string {
	if strings.HasPrefix(typeText, "*") || strings.HasPrefix(typeText, "[]") || typeText == "any" || typeText == "interface{}" {
		return "nil"
	}
	return zeroReturn(typeText)
}

func serviceSuccessReturn(typeText string) string {
	if strings.HasPrefix(typeText, "*") {
		return "&" + trimPointer(typeText) + "{}"
	}
	return zeroReturn(typeText)
}

func zeroReturn(typeText string) string {
	if strings.HasPrefix(typeText, "*") || strings.HasPrefix(typeText, "[]") || typeText == "any" || typeText == "interface{}" {
		return "nil"
	}
	switch typeText {
	case "bool":
		return "false"
	case "string":
		return "\"\""
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
		return "0"
	default:
		return typeText + "{}"
	}
}

func normalizeTypeAliases(types []string, imports map[string]string, targetImport, targetAlias string) []string {
	out := make([]string, 0, len(types))
	for _, typeText := range types {
		normalized := typeText
		for alias, importPath := range imports {
			if importPath != targetImport || alias == targetAlias {
				continue
			}
			normalized = strings.ReplaceAll(normalized, alias+".", targetAlias+".")
		}
		if !strings.Contains(normalized, ".") && targetAlias != "" && looksLikeGeneratedType(normalized) {
			normalized = addTypeAlias(normalized, targetAlias)
		}
		out = append(out, normalized)
	}
	return out
}

func importExistsInProject(info project.Info, importPath string) bool {
	importPath = strings.TrimSpace(importPath)
	if importPath == "" {
		return false
	}
	prefix := strings.TrimSuffix(info.Module, "/") + "/"
	if !strings.HasPrefix(importPath, prefix) {
		return false
	}
	rel := strings.TrimPrefix(importPath, prefix)
	path := filepath.Join(info.Root, filepath.FromSlash(rel))
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}

func looksLikeGeneratedType(typeText string) bool {
	name := trimPointer(typeText)
	return strings.HasSuffix(name, "Request") || strings.HasSuffix(name, "Response") || name == "UserCredential"
}

func addTypeAlias(typeText, alias string) string {
	if strings.HasPrefix(typeText, "*") {
		return "*" + alias + "." + strings.TrimPrefix(typeText, "*")
	}
	return alias + "." + typeText
}
