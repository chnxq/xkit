package codegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"

	codegentemplate "github.com/chnxq/xkit/internal/codegen/template"
	"github.com/chnxq/xkit/internal/config"
	"github.com/chnxq/xkit/internal/entschema"
)

type repoTemplateData struct {
	templateBase
	Imports                  []importSpec
	RepoName                 string
	RepoStructName           string
	ConstructorName          string
	EntityName               string
	EntOperationName         string
	ResourceName             string
	EntPackage               string
	PredicateType            string
	DTOType                  string
	IDType                   string
	DefaultListSortField     string
	DefaultListSortDirection string
	Methods                  []repoMethodData
	Fields                   []entschema.Field
	UsesEnumSetter           bool
	UsesTimeSetter           bool
	UsesFieldMaskHelper      bool
	UsesJSONDecoder          bool
	EnumHelperName           string
	TimeHelperName           string
	JSONHelperName           string
	FilterTimeHelperName     string
	FieldMaskHelperName      string
	Filters                  []filterData
	UsesFilters              bool
	UsesAuditFields          bool
	TenantScope              string
	NeedsTenantHelpers       bool
	UseSharedModule          bool
	Tree                     *treeConfigData
	Aggregates               []aggregateConfigData
}

type repoMethodData struct {
	Name            string
	Params          []namedType
	ResponseType    string
	Kind            string
	Body            string
	CustomHookName  string
	Setters         []setterData
	AuditSetters    []setterData
	AuditUsesNow    bool
	AuditUsesViewer bool
	ReturnsValue    bool
	IDExpr          string
	ExistField      string
	ViewMaskExpr    string
	ListItemsField  string
	ListTotalField  string
	CountField      string
	ExistsCases     []existsCaseData
	NilReturn       string
	ZeroReturn      string
	UsesFilters     bool
	PagingReqExpr   string
}

func (r *Runner) generateRepoFiles() (Result, error) {
	plans, err := r.plans()
	if err != nil {
		return Result{}, err
	}

	var result Result
	needsTenantHelpers := repoNeedsTenantHelpers(plans)
	if r.isModuleMode() {
		if err := r.removeObsoleteGeneratedFile(filepath.Join(r.internalDir("data", "repo"), "repo_shared_ext.go")); err != nil {
			return result, err
		}
		if err := r.ensureModuleSharedExtFile(&result, needsTenantHelpers); err != nil {
			return result, err
		}
	} else {
		sharedPath := filepath.Join(r.internalDir("data", "repo"), "repo_shared_ext.go")
		if r.hasLegacyRepoSharedHelpers() {
			if err := r.removeObsoleteGeneratedFile(sharedPath); err != nil {
				return result, err
			}
		} else {
			sharedContent, err := renderTemplate(codegentemplate.RepoSharedExt, struct {
				templateBase
				NeedsTenantHelpers bool
			}{
				templateBase:       r.templateBase(),
				NeedsTenantHelpers: needsTenantHelpers,
			})
			if err != nil {
				return result, err
			}
			if err := r.writeGeneratedFile(sharedPath, sharedContent, &result); err != nil {
				return result, err
			}
		}
	}

	for _, plan := range plans {
		if !plan.Resource.Generate.EffectiveRepoCRUD() {
			continue
		}

		content, err := r.renderRepoFile(plan)
		if err != nil {
			return result, err
		}

		repoPath := filepath.Join(
			r.internalDir("data", "repo"),
			plan.FileBase+"_repo.gen.go",
		)
		if err := r.writeGeneratedFile(repoPath, content, &result); err != nil {
			return result, err
		}

		extContent, err := renderTemplate(codegentemplate.RepoExt, struct {
			templateBase
			RepoName string
		}{
			templateBase: r.templateBase(),
			RepoName:     repoInterfaceName(plan),
		})
		if err != nil {
			return result, err
		}
		extPath := filepath.Join(
			r.internalDir("data", "repo"),
			plan.FileBase+"_repo_ext.go",
		)
		if err := r.writeExtensionFile(extPath, extContent, &result); err != nil {
			return result, err
		}
	}

	return result, nil
}

func (r *Runner) hasLegacyRepoSharedHelpers() bool {
	repoRoot := r.internalDir("data", "repo")
	legacyFiles := []string{
		"list_sorting_ext.go",
		"tenant_scope_ext.go",
	}
	for _, name := range legacyFiles {
		if _, err := os.Stat(filepath.Join(repoRoot, name)); err == nil {
			return true
		}
	}
	return false
}

func repoNeedsTenantHelpers(plans []resourcePlan) bool {
	for _, plan := range plans {
		if strings.TrimSpace(plan.Resource.TenantScope) == "tenant_scoped" {
			return true
		}
	}
	return false
}

func repoUsesSharedModule(plan resourcePlan) bool {
	if strings.TrimSpace(plan.Resource.TenantScope) == "tenant_scoped" {
		return true
	}
	for _, method := range plan.Binding.Methods {
		if repoMethodKind(method.Name) == "list" && resourceOperationEnabled(plan.Resource, "list") {
			return true
		}
	}
	return false
}

func (r *Runner) renderRepoFile(plan resourcePlan) ([]byte, error) {
	entityName := strings.TrimSpace(plan.Resource.Entity)
	if entityName == "" {
		entityName = plan.Binding.ServiceName
		entityName = strings.TrimSuffix(entityName, "Service")
	}

	dtoAlias := plan.APIPackageAlias
	dtoImport := plan.Binding.ImportPath
	if strings.TrimSpace(plan.Resource.DTOImport) != "" && importExistsInProject(r.project, strings.TrimSpace(plan.Resource.DTOImport)) {
		dtoImport = strings.TrimSpace(plan.Resource.DTOImport)
		dtoAlias = apiAlias(dtoImport)
	}
	dtoName := entityName
	if strings.TrimSpace(plan.Resource.DTOType) != "" {
		dtoName = strings.TrimSpace(plan.Resource.DTOType)
	}
	dtoName = r.resolveGeneratedTypeName(dtoImport, dtoName)
	dtoType := dtoAlias + "." + dtoName
	repoName := strings.TrimSpace(plan.Resource.RepoInterface)
	if repoName == "" {
		repoName = entityName + "Repo"
	}

	imports := []importSpec{
		{Path: "context"},
		{Path: "fmt"},
		{Path: "github.com/chnxq/x-crud/entgo", Alias: "entCrud"},
		{Path: "github.com/chnxq/x-utils/copierutil"},
		{Path: "github.com/chnxq/x-utils/mapper"},
		{Path: "github.com/chnxq/xkitmod/log"},
		{Path: "github.com/chnxq/xkitpkg/app"},
		{Alias: "ent", Path: r.layout.EntImportRoot},
		{Path: filepath.ToSlash(filepath.Join(r.layout.EntImportRoot, strings.ToLower(entityName)))},
		{Path: filepath.ToSlash(filepath.Join(r.layout.EntImportRoot, "predicate"))},
		{Alias: dtoAlias, Path: dtoImport},
	}
	if r.isModuleMode() && repoUsesSharedModule(plan) {
		imports = append(imports, importSpec{Alias: "modulex", Path: r.sharedModuleImport()})
	}
	filters := repoFilters(plan.Schema.Fields, effectiveFilterAllow(plan))
	usesFilters := len(filters) > 0
	usesAuditFields := hasGeneratedAuditFields(plan.Schema.Fields)
	if usesFilters {
		imports = append(imports,
			importSpec{Path: "strconv"},
			importSpec{Path: "strings"},
			importSpec{Path: "time"},
			importSpec{Alias: "paginationv1", Path: "github.com/chnxq/x-crud/api/gen/pagination/v1"},
		)
	}
	if usesAuditFields {
		imports = append(imports,
			importSpec{Alias: "crudviewer", Path: "github.com/chnxq/x-crud/viewer"},
		)
	}
	if strings.TrimSpace(plan.Resource.TenantScope) == "tenant_scoped" {
		imports = append(imports,
			importSpec{Alias: dtoAlias, Path: dtoImport},
		)
	}
	imports = append(imports, r.repoConfiguredImports(plan)...)
	usedAliases := make(map[string]struct{})
	var methods []repoMethodData
	usesEnumSetter := false
	usesTimeSetter := false
	for _, method := range plan.Binding.Methods {
		if !isCRUDMethod(method.Name) && !isSupportedRepoSpecialMethod(plan.Resource, method.Name) && !hasConfiguredRepoMethod(plan.Resource, method.Name) {
			continue
		}
		kind := repoMethodKind(method.Name)
		if !resourceOperationEnabled(plan.Resource, kind) {
			continue
		}
		normalizedParams := normalizeTypeAliases(method.Params, plan.Binding.Imports, dtoImport, dtoAlias)
		normalizedResults := normalizeTypeAliases(method.Results, plan.Binding.Imports, dtoImport, dtoAlias)
		zeroReturnValue := zeroReturn(firstResult(normalizedResults))
		enumHelperName := lowerFirst(entityName) + "EnumPtrFromProto"
		timeHelperName := lowerFirst(entityName) + "TimePtrFromProto"
		maskHelperName := lowerFirst(entityName) + "FieldMaskContains"
		setters := repoSetters(plan.Schema.Fields, r.dtoFieldTypes(dtoImport, dtoName), method.Name, strings.ToLower(entityName), enumHelperName, timeHelperName, maskHelperName, zeroReturnValue)
		auditSetters := repoAuditSetters(plan.Schema.Fields, method.Name)
		usesEnumSetter = usesEnumSetter || settersUseKind(setters, "Enum")
		usesTimeSetter = usesTimeSetter || settersUseKind(setters, "Time")
		methodData := repoMethodData{
			Name:            method.Name,
			Params:          nameParams(normalizedParams),
			ResponseType:    firstResult(normalizedResults),
			Kind:            kind,
			Body:            r.repoMethodBody(plan, method.Name, nameParams(normalizedParams), firstResult(normalizedResults)),
			CustomHookName:  repoCustomHookName(kind, entityName, method.Name),
			Setters:         setters,
			AuditSetters:    auditSetters,
			AuditUsesNow:    auditSettersUseExpr(auditSetters, "now"),
			AuditUsesViewer: auditSettersUseExprContains(auditSetters, "viewer."),
			ReturnsValue:    firstResult(normalizedResults) == "*"+dtoType,
			IDExpr:          "req.GetId()",
			ExistField:      "Exist",
			ViewMaskExpr:    "req.GetViewMask()",
			ListItemsField:  "Items",
			ListTotalField:  "Total",
			CountField:      "Count",
			ExistsCases:     existsCases(dtoAlias, requestParamType(normalizedParams), plan.Resource.ExistsFields),
			NilReturn:       nilReturn(firstResult(normalizedResults)),
			ZeroReturn:      zeroReturnValue,
			UsesFilters:     usesFilters,
			PagingReqExpr:   pagingRequestExpr(nameParams(normalizedParams)),
		}
		methods = append(methods, methodData)

		for _, typeText := range append(slices.Clone(normalizedParams), normalizedResults...) {
			for _, alias := range aliasesInType(typeText) {
				usedAliases[alias] = struct{}{}
			}
		}
	}
	for alias := range usedAliases {
		path, ok := plan.Binding.Imports[alias]
		if !ok {
			continue
		}
		imports = append(imports, importSpec{Alias: alias, Path: path})
	}
	if usesTimeSetter {
		if !usesAuditFields {
			imports = append(imports, importSpec{Path: "time"})
		}
		imports = append(imports,
			importSpec{Alias: "timestamppb", Path: "google.golang.org/protobuf/types/known/timestamppb"},
		)
	}
	usesFieldMaskHelper := false
	for _, method := range methods {
		if settersNeedFieldMaskHelper(method.Setters) {
			usesFieldMaskHelper = true
			break
		}
	}
	usesJSONDecoder := false
	for _, method := range methods {
		if settersUsePre(method.Setters) {
			usesJSONDecoder = true
			break
		}
	}
	if usesJSONDecoder {
		imports = append(imports, importSpec{Path: "encoding/json"})
		if !usesFilters {
			imports = append(imports, importSpec{Path: "strings"})
		}
	}

	data := repoTemplateData{
		templateBase:             r.templateBase(),
		Imports:                  uniqueImports(imports),
		RepoName:                 repoName,
		RepoStructName:           lowerFirst(repoName),
		ConstructorName:          "New" + repoName,
		EntityName:               entityName,
		EntOperationName:         entOperationName(entityName),
		ResourceName:             plan.Resource.Name,
		EntPackage:               strings.ToLower(entityName),
		PredicateType:            entityName,
		DTOType:                  dtoType,
		IDType:                   idGoType(plan.Schema.Fields),
		DefaultListSortField:     defaultListSortField(plan),
		DefaultListSortDirection: defaultListSortDirection(plan),
		Fields:                   plan.Schema.Fields,
		UsesEnumSetter:           usesEnumSetter,
		UsesTimeSetter:           usesTimeSetter,
		UsesFieldMaskHelper:      usesFieldMaskHelper,
		UsesJSONDecoder:          usesJSONDecoder,
		EnumHelperName:           lowerFirst(entityName) + "EnumPtrFromProto",
		TimeHelperName:           lowerFirst(entityName) + "TimePtrFromProto",
		JSONHelperName:           lowerFirst(entityName) + "DecodeJSONString",
		FilterTimeHelperName:     lowerFirst(entityName) + "ParseFilterTime",
		FieldMaskHelperName:      lowerFirst(entityName) + "FieldMaskContains",
		Filters:                  filters,
		UsesFilters:              usesFilters,
		UsesAuditFields:          usesAuditFields,
		TenantScope:              strings.TrimSpace(plan.Resource.TenantScope),
		NeedsTenantHelpers:       strings.TrimSpace(plan.Resource.TenantScope) == "tenant_scoped",
		UseSharedModule:          r.isModuleMode(),
		Tree:                     buildTreeConfig(plan.Resource.Tree),
		Aggregates:               r.aggregateConfigs(plan),
	}
	data.Methods = methods

	return renderTemplate(codegentemplate.RepoFile, data)
}

func (r *Runner) dtoFieldNames(importPath, typeName string) map[string]struct{} {
	out := make(map[string]struct{})
	for name := range r.dtoFieldTypes(importPath, typeName) {
		out[name] = struct{}{}
	}
	return out
}

func (r *Runner) dtoFieldTypes(importPath, typeName string) map[string]string {
	out := make(map[string]string)
	if strings.TrimSpace(importPath) == "" || strings.TrimSpace(typeName) == "" {
		return out
	}
	prefix := strings.TrimSuffix(r.project.Module, "/") + "/"
	if !strings.HasPrefix(importPath, prefix) {
		return out
	}
	rel := strings.TrimPrefix(importPath, prefix)
	dir := filepath.Join(r.project.Root, filepath.FromSlash(rel))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return out
	}
	fileSet := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		file, err := parser.ParseFile(fileSet, path, nil, parser.SkipObjectResolution)
		if err != nil {
			continue
		}
		for _, decl := range file.Decls {
			genericDecl, ok := decl.(*ast.GenDecl)
			if !ok || genericDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genericDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || typeSpec.Name.Name != typeName {
					continue
				}
				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}
				for _, field := range structType.Fields.List {
					fieldType := exprString(field.Type)
					if fieldType == "" {
						continue
					}
					for _, name := range field.Names {
						out[name.Name] = fieldType
					}
				}
			}
		}
	}
	return out
}

func repoInterfaceName(plan resourcePlan) string {
	repoName := strings.TrimSpace(plan.Resource.RepoInterface)
	if repoName != "" {
		return repoName
	}
	entityName := strings.TrimSpace(plan.Resource.Entity)
	if entityName == "" {
		entityName = strings.TrimSuffix(plan.Binding.ServiceName, "Service")
	}
	return entityName + "Repo"
}

func (r *Runner) repoMethodBody(plan resourcePlan, methodName string, params []namedType, responseType string) string {
	if methodConfig, ok := plan.Resource.RepoMethods[methodName]; ok && strings.TrimSpace(methodConfig.Body) != "" {
		body := strings.TrimRight(methodConfig.Body, "\r\n")
		replacements := map[string]string{
			"{{successReturn}}": serviceSuccessReturn(responseType),
		}
		for _, param := range params {
			replacements["{{param."+param.Name+"}}"] = param.Name
			if param.Type == "context.Context" {
				replacements["{{ctx}}"] = param.Name
			}
		}
		for key, value := range replacements {
			body = strings.ReplaceAll(body, key, value)
		}
		return indentLines(body, "\t")
	}

	if !isSupportedRepoSpecialMethod(plan.Resource, methodName) {
		return ""
	}

	ctxParam := ""
	reqParam := ""
	for _, param := range params {
		if param.Type == "context.Context" {
			ctxParam = param.Name
		}
		if strings.HasSuffix(param.Type, "ResetCredentialRequest") {
			reqParam = param.Name
		}
	}
	if ctxParam == "" || reqParam == "" {
		return ""
	}

	successReturn := serviceSuccessReturn(responseType)
	return fmt.Sprintf(`	if %s == nil {
		return nil, fmt.Errorf("invalid parameter")
	}

	credential := %s.GetNewCredential()
	// TODO: hash or encrypt credential before storage when password crypto is available.
	if %s.GetNeedDecrypt() {
		// TODO: decrypt credential before hashing.
	}

	_, err := r.entClient.Client().UserCredential.Update().
		Where(
			usercredential.IdentityTypeEQ(usercredential.IdentityTypeUsername),
			usercredential.IdentifierEQ(%s.GetIdentifier()),
			usercredential.CredentialTypeEQ(usercredential.CredentialTypePasswordHash),
		).
		SetCredential(credential).
		Save(%s)
	if err != nil {
		r.log.Errorf("reset user credential failed: %%s", err.Error())
		return nil, err
	}

	return %s, nil`, reqParam, reqParam, reqParam, reqParam, ctxParam, successReturn)
}

func (r *Runner) repoConfiguredImports(plan resourcePlan) []importSpec {
	var imports []importSpec
	for _, methodConfig := range plan.Resource.RepoMethods {
		for _, importConfig := range methodConfig.Imports {
			path := strings.TrimSpace(importConfig.Path)
			if path == "" {
				continue
			}
			path = r.normalizeConfiguredImportPath(path)
			imports = append(imports, importSpec{Alias: strings.TrimSpace(importConfig.Alias), Path: path})
		}
	}
	return imports
}

func (r *Runner) resolveGeneratedTypeName(importPath, desiredName string) string {
	desiredName = strings.TrimSpace(desiredName)
	if desiredName == "" {
		return ""
	}
	names := r.generatedTypeNames(importPath)
	if len(names) == 0 {
		return desiredName
	}
	for _, name := range names {
		if name == desiredName {
			return name
		}
	}
	for _, name := range names {
		if strings.EqualFold(name, desiredName) {
			return name
		}
	}
	return desiredName
}

func (r *Runner) generatedTypeNames(importPath string) []string {
	if strings.TrimSpace(importPath) == "" {
		return nil
	}
	prefix := strings.TrimSuffix(r.project.Module, "/") + "/"
	if !strings.HasPrefix(importPath, prefix) {
		return nil
	}
	rel := strings.TrimPrefix(importPath, prefix)
	dir := filepath.Join(r.project.Root, filepath.FromSlash(rel))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	fileSet := token.NewFileSet()
	var names []string
	seen := make(map[string]struct{})
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		file, err := parser.ParseFile(fileSet, path, nil, parser.SkipObjectResolution)
		if err != nil {
			continue
		}
		for _, decl := range file.Decls {
			genericDecl, ok := decl.(*ast.GenDecl)
			if !ok || genericDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genericDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || typeSpec.Name == nil {
					continue
				}
				name := strings.TrimSpace(typeSpec.Name.Name)
				if name == "" {
					continue
				}
				if _, ok := seen[name]; ok {
					continue
				}
				seen[name] = struct{}{}
				names = append(names, name)
			}
		}
	}
	return names
}

func repoCustomHookName(kind, entityName, methodName string) string {
	prefix := lowerFirst(entityName)
	switch kind {
	case "create":
		return prefix + "CustomCreate"
	case "update":
		return prefix + "CustomUpdate"
	case "delete":
		return prefix + "CustomDelete"
	case "get":
		return prefix + "EnrichGetDTO"
	case "list":
		return prefix + "EnrichListDTOs"
	default:
		if methodName == "" {
			return prefix + "Custom"
		}
		return prefix + "Custom" + methodName
	}
}

func isSupportedRepoSpecialMethod(resource config.Resource, methodName string) bool {
	return resource.Name == "user_credential" && methodName == "ResetCredential"
}

func hasConfiguredRepoMethod(resource config.Resource, methodName string) bool {
	methodConfig, ok := resource.RepoMethods[methodName]
	return ok && strings.TrimSpace(methodConfig.Body) != ""
}
