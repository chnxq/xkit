package codegen

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	codegentemplate "github.com/chnxq/xkit/internal/codegen/template"
	"github.com/chnxq/xkit/internal/config"
	"github.com/chnxq/xkit/internal/entschema"
)

type frontendMetaTemplateData struct {
	templateBase
	I18nPrefix           string
	DefaultSortField     string
	DefaultSortDirection string
	ProviderBaseName     string
	FilterItems          []frontendFilterItem
	Columns              []frontendColumn
	HasDialogForm        bool
	DialogFields         []frontendDialogField
}

type frontendFilterItem struct {
	FieldName      string
	Component      string
	LabelKey       string
	PlaceholderKey string
	ExtraProps     string
}

type frontendColumn struct {
	Field        string
	TitleExpr    string
	Formatter    string
	SlotsDefault string
	TreeNode     bool
	Sortable     bool
	Width        int
}

type frontendDialogField struct {
	FieldName  string
	Component  string
	LabelKey   string
	ExtraProps string
}

var frontendExampleLangsDirForTest func(*Runner) (string, error)

type frontendFieldRuntime struct {
	FieldName string
	Relation  *frontendFieldRelationRuntime
	Enum      *frontendFieldEnumRuntime
}

type frontendFieldRelationRuntime struct {
	ResourceField string
	PlaceholderKey string
}

type frontendFieldEnumRuntime struct {
	ResourceKey string
	Values      []string
}

func (r *Runner) generateFrontendMetaFiles() (Result, error) {
	plans, err := r.plans()
	if err != nil {
		return Result{}, err
	}

	baseDir := r.frontendGeneratedMetaRoot()
	if baseDir == "" {
		return Result{}, fmt.Errorf("resolve frontend meta root failed")
	}

	var result Result

	if r.isModuleMode() {
		if err := r.removeModuleFrontendSharedOutputs(baseDir); err != nil {
			return result, err
		}
	} else {
		configContent, err := renderAnyTemplate(codegentemplate.FrontendViewConfig, r.templateBase())
		if err != nil {
			return result, err
		}
		if err := r.writeGeneratedFile(filepath.Join(baseDir, "config.ts"), configContent, &result); err != nil {
			return result, err
		}
	}

	var zhEntries []i18nEntry
	var enEntries []i18nEntry
	for _, plan := range plans {
		if plan.Resource.Frontend == nil || strings.TrimSpace(plan.Resource.Frontend.ViewPath) == "" {
			continue
		}

		metaData := r.frontendMetaData(plan)
		metaContent, err := renderAnyTemplate(codegentemplate.FrontendViewMeta, metaData)
		if err != nil {
			return result, err
		}
		metaRelPath := filepath.FromSlash(plan.Resource.Frontend.ViewPath + ".meta.ts")
		if r.isModuleMode() {
			metaRelPath = filepath.Base(filepath.FromSlash(plan.Resource.Frontend.ViewPath)) + ".meta.ts"
		}
		metaPath := filepath.Join(baseDir, metaRelPath)
		if err := r.writeGeneratedFile(metaPath, metaContent, &result); err != nil {
			return result, err
		}

		zhEntries = append(zhEntries, r.frontendI18nEntries(plan, "zh")...)
		enEntries = append(enEntries, r.frontendI18nEntries(plan, "en")...)
	}

	if r.isModuleMode() {
		if err := r.copyFrontendGeneratedLangs(baseDir, &result); err != nil {
			return result, err
		}
		return result, nil
	}

	if err := r.writeFrontendI18nFile(filepath.Join(baseDir, "page_i18n.zh-CN.json"), codegentemplate.FrontendPageI18nZH, zhEntries, &result); err != nil {
		return result, err
	}
	if err := r.writeFrontendI18nFile(filepath.Join(baseDir, "page_i18n.en-US.json"), codegentemplate.FrontendPageI18nEN, enEntries, &result); err != nil {
		return result, err
	}
	if err := r.copyFrontendGeneratedLangs(baseDir, &result); err != nil {
		return result, err
	}

	return result, nil
}

func (r *Runner) isModuleMode() bool {
	return strings.TrimSpace(r.options.ModuleName) != ""
}

func (r *Runner) removeModuleFrontendSharedOutputs(baseDir string) error {
	paths := []string{
		filepath.Join(baseDir, "config.ts"),
		filepath.Join(baseDir, "page_i18n.zh-CN.json"),
		filepath.Join(baseDir, "page_i18n.en-US.json"),
	}
	for _, path := range paths {
		if err := r.removeObsoleteGeneratedFile(path); err != nil {
			return err
		}
	}
	dirs := []string{
		r.frontendGeneratedLangRoot(),
		r.frontendLegacyGeneratedRoot(),
		r.frontendLegacyHostViewRoot(),
		r.frontendLegacyGeneratedAPIRoot(),
	}
	for _, dir := range dirs {
		if strings.TrimSpace(dir) == "" {
			continue
		}
		if err := removeObsoleteGeneratedDir(dir, r.options.DryRun); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) frontendMetaData(plan resourcePlan) frontendMetaTemplateData {
	frontend := plan.Resource.Frontend
	i18nPrefix := frontend.I18nPrefix
	if strings.TrimSpace(i18nPrefix) == "" {
		i18nPrefix = "page." + lowerFirst(plan.ResourceField)
	}

	data := frontendMetaTemplateData{
		templateBase:         r.templateBase(),
		I18nPrefix:           i18nPrefix,
		DefaultSortField:     r.frontendDefaultSortField(plan),
		DefaultSortDirection: "DESC",
		ProviderBaseName:     filepath.Base(strings.TrimSpace(frontend.ViewPath)),
		FilterItems:          r.frontendFilterItems(plan),
		Columns:              r.frontendColumns(plan),
		HasDialogForm:        frontend.Form != nil && (frontend.Form.Enabled == nil || *frontend.Form.Enabled) && len(frontend.Form.Fields) > 0,
		DialogFields:         r.frontendDialogFields(plan),
	}
	return data
}

func (r *Runner) frontendDefaultSortField(plan resourcePlan) string {
	for _, field := range plan.Schema.Fields {
		if field.Name == "created_at" {
			return "created_at"
		}
	}
	return "id"
}

func (r *Runner) frontendFilterItems(plan resourcePlan) []frontendFilterItem {
	if plan.Resource.Frontend == nil || plan.Resource.Frontend.List == nil {
		return nil
	}
	filters := normalizedFrontendFilters(plan.Resource.Frontend.List.Filters)
	items := make([]frontendFilterItem, 0, len(filters))
	for _, filter := range filters {
		fieldName := filter.Field
		component := r.frontendFilterComponent(plan, filter)
		labelKey := r.frontendFilterLabelKey(plan, filter)
		item := frontendFilterItem{
			FieldName:      fieldName,
			Component:      component,
			LabelKey:       labelKey,
			PlaceholderKey: frontendPlaceholderKey(component, labelKey),
		}
		if component == "RangePicker" {
			item.ExtraProps = "          showTime: true,\n          valueFormat: 'YYYY-MM-DD HH:mm:ss',"
		} else if component == "Select" {
			item.ExtraProps = r.frontendEnumOptionsProps(plan, fieldName)
		}
		items = append(items, item)
	}
	return items
}

func (r *Runner) frontendColumns(plan resourcePlan) []frontendColumn {
	if plan.Resource.Frontend == nil || plan.Resource.Frontend.List == nil {
		return nil
	}
	columns := make([]frontendColumn, 0, len(plan.Resource.Frontend.List.Columns))
	for _, columnCfg := range plan.Resource.Frontend.List.Columns {
		field := columnCfg.Field
		runtime := r.frontendFieldRuntime(plan, field)
		width := columnCfg.Width
		if width <= 0 {
			width = frontendWidth(field)
		}
		titleKey := strings.TrimSpace(columnCfg.TitleKey)
		if titleKey == "" {
			titleKey = plan.Resource.Frontend.I18nPrefix + "." + field
		}
		slotDefault := strings.TrimSpace(columnCfg.Slot)
		if slotDefault == "" {
			if runtime.Relation != nil || runtime.Enum != nil {
				slotDefault = simpleFieldName(field)
			} else {
				slotDefault = frontendSlot(field)
			}
		}
		columns = append(columns, frontendColumn{
			Field:        field,
			TitleExpr:    "t('" + titleKey + "')",
			Formatter:    frontendFormatter(field),
			SlotsDefault: slotDefault,
			TreeNode:     columnCfg.TreeNode,
			Sortable:     field != "deviceInfo.userAgent",
			Width:        width,
		})
	}
	return columns
}

func (r *Runner) frontendDialogFields(plan resourcePlan) []frontendDialogField {
	if plan.Resource.Frontend == nil || plan.Resource.Frontend.Form == nil {
		return nil
	}
	fields := plan.Resource.Frontend.Form.Fields
	items := make([]frontendDialogField, 0, len(fields))
	for _, field := range fields {
		component := r.frontendDialogComponent(plan, field.Field)
		items = append(items, frontendDialogField{
			FieldName:  field.Field,
			Component:  component,
			LabelKey:   field.Field,
			ExtraProps: r.frontendDialogExtraProps(plan, field.Field),
		})
	}
	return items
}

func (r *Runner) frontendDialogExtraProps(plan resourcePlan, fieldName string) string {
	component := r.frontendDialogComponent(plan, fieldName)
	if component == "Select" {
		return r.frontendEnumOptionsProps(plan, fieldName)
	}
	return ""
}

func (r *Runner) frontendI18nEntries(plan resourcePlan, lang string) []i18nEntry {
	if plan.Resource.Frontend == nil {
		return nil
	}
	prefix := strings.TrimSpace(plan.Resource.Frontend.I18nPrefix)
	if prefix == "" {
		return nil
	}
	seen := map[string]struct{}{}
	var entries []i18nEntry
	add := func(suffix, value string) {
		key := prefix + "." + suffix
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		entries = append(entries, i18nEntry{Key: key, Value: value})
	}

	if plan.Resource.Frontend.List != nil {
		for _, filter := range normalizedFrontendFilters(plan.Resource.Frontend.List.Filters) {
			component := r.frontendFilterComponent(plan, filter)
			labelKey := r.frontendFilterLabelKey(plan, filter)
			schemaField := r.frontendSchemaField(plan, labelKey)
			enTitle, cnTitle := r.frontendFilterTitles(plan, filter, labelKey)
			add(labelKey, frontendI18nValue(lang, labelKey, enTitle, cnTitle, schemaField))
			if placeholderKey := frontendPlaceholderKey(component, labelKey); placeholderKey != "" {
				add(placeholderKey, frontendI18nPlaceholderValue(lang, labelKey, component, enTitle, cnTitle, schemaField))
			}
		}
		for _, columnCfg := range plan.Resource.Frontend.List.Columns {
			schemaField := r.frontendSchemaField(plan, columnCfg.Field)
			add(columnCfg.Field, frontendI18nValue(lang, columnCfg.Field, columnCfg.EN, columnCfg.CN, schemaField))
		}
	}
	if plan.Resource.Frontend.Form != nil {
		for _, field := range plan.Resource.Frontend.Form.Fields {
			schemaField := r.frontendSchemaField(plan, field.Field)
			add(field.Field, frontendI18nValue(lang, field.Field, field.EN, field.CN, schemaField))
		}
	}
	return entries
}

func (r *Runner) frontendFilterLabelKey(plan resourcePlan, filter config.FrontendFilter) string {
	field := filter.Field
	if strings.TrimSpace(filter.EN) != "" || strings.TrimSpace(filter.CN) != "" {
		return "filter" + frontendKeySuffix(simpleFieldName(field))
	}
	if columnCfg := r.frontendColumnConfig(plan, field); columnCfg != nil {
		return columnCfg.Field
	}
	return field
}

func (r *Runner) frontendFilterTitles(plan resourcePlan, filter config.FrontendFilter, labelKey string) (string, string) {
	if strings.TrimSpace(filter.EN) != "" || strings.TrimSpace(filter.CN) != "" {
		return filter.EN, filter.CN
	}
	columnCfg := r.frontendColumnConfig(plan, labelKey)
	if columnCfg != nil {
		return columnCfg.EN, columnCfg.CN
	}
	return "", ""
}

func (r *Runner) frontendColumnConfig(plan resourcePlan, field string) *config.FrontendColumn {
	if plan.Resource.Frontend == nil || plan.Resource.Frontend.List == nil {
		return nil
	}
	for i := range plan.Resource.Frontend.List.Columns {
		if plan.Resource.Frontend.List.Columns[i].Field == field {
			return &plan.Resource.Frontend.List.Columns[i]
		}
	}
	return nil
}

func (r *Runner) frontendSchemaField(plan resourcePlan, field string) *entschema.Field {
	target := frontendSnakeCase(simpleFieldName(field))
	for i := range plan.Schema.Fields {
		if plan.Schema.Fields[i].Name == target {
			return &plan.Schema.Fields[i]
		}
	}
	return nil
}

func (r *Runner) writeFrontendI18nFile(path, tmpl string, entries []i18nEntry, result *Result) error {
	_ = tmpl
	slices.SortFunc(entries, func(a, b i18nEntry) int {
		return strings.Compare(a.Key, b.Key)
	})
	messages := make(map[string]any, len(entries))
	for _, entry := range entries {
		setNestedI18nValue(messages, strings.Split(entry.Key, "."), entry.Value)
	}
	content, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return err
	}
	return r.writeGeneratedFile(path, append(content, '\n'), result)
}

func setNestedI18nValue(target map[string]any, path []string, value string) {
	if len(path) == 0 {
		return
	}
	current := target
	for _, segment := range path[:len(path)-1] {
		next, ok := current[segment]
		if !ok {
			child := map[string]any{}
			current[segment] = child
			current = child
			continue
		}
		child, ok := next.(map[string]any)
		if !ok {
			child = map[string]any{}
			current[segment] = child
		}
		current = child
	}
	current[path[len(path)-1]] = value
}

func (r *Runner) copyFrontendGeneratedLangs(baseDir string, result *Result) error {
	sourceDir, err := r.frontendExampleLangsDir()
	if err != nil {
		return err
	}
	langRoot := filepath.Join(baseDir, "langs")
	if r.isModuleMode() {
		langRoot = r.frontendGeneratedLangRoot()
		if langRoot == "" {
			langRoot = filepath.Join(baseDir, "langs")
		}
	}
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		langDir := filepath.Join(sourceDir, entry.Name())
		files, err := os.ReadDir(langDir)
		if err != nil {
			return err
		}
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
				continue
			}
			sourcePath := filepath.Join(langDir, file.Name())
			content, err := os.ReadFile(sourcePath)
			if err != nil {
				return err
			}
			targetPath := filepath.Join(langRoot, entry.Name(), file.Name())
			if err := r.writeGeneratedFile(targetPath, content, result); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Runner) frontendExampleLangsDir() (string, error) {
	if frontendExampleLangsDirForTest != nil {
		return frontendExampleLangsDirForTest(r)
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("resolve xkit source location failed")
	}
	examplesRoot := filepath.Join(filepath.Dir(file), "..", "..", "examples")
	if r.isModuleMode() {
		moduleName := strings.TrimSpace(r.options.ModuleName)
		if moduleName == "" {
			return "", fmt.Errorf("module name is required for module frontend langs")
		}
		return filepath.Join(examplesRoot, moduleName, "langs"), nil
	}
	return filepath.Join(examplesRoot, "admin", "langs"), nil
}

func removeObsoleteGeneratedDir(path string, dryRun bool) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read obsolete generated dir %s: %w", path, err)
	}
	for _, entry := range entries {
		childPath := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			if err := removeObsoleteGeneratedDir(childPath, dryRun); err != nil {
				return err
			}
			continue
		}
		content, err := os.ReadFile(childPath)
		if err != nil {
			return fmt.Errorf("read obsolete generated file %s: %w", childPath, err)
		}
		if !bytes.Contains(content, []byte("Code generated by xkit. DO NOT EDIT.")) && !bytes.Contains(content, []byte("\"page.")) {
			continue
		}
		if dryRun {
			continue
		}
		if err := os.Remove(childPath); err != nil {
			return fmt.Errorf("remove obsolete generated file %s: %w", childPath, err)
		}
	}
	if dryRun {
		return nil
	}
	remaining, err := os.ReadDir(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("re-read obsolete generated dir %s: %w", path, err)
	}
	if len(remaining) == 0 {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove obsolete generated dir %s: %w", path, err)
		}
	}
	return nil
}

func normalizedFrontendFilters(filters []config.FrontendFilter) []config.FrontendFilter {
	items := make([]config.FrontendFilter, 0, len(filters))
	for _, filter := range filters {
		if strings.TrimSpace(filter.Field) == "" {
			continue
		}
		items = append(items, filter)
	}
	slices.SortFunc(items, func(a, b config.FrontendFilter) int {
		return strings.Compare(a.Field, b.Field)
	})
	return items
}

func frontendPlaceholderKey(component, field string) string {
	switch component {
	case "Input", "InputNumber":
		return "search" + frontendKeySuffix(simpleFieldName(field))
	case "Select":
		return "select" + frontendKeySuffix(simpleFieldName(field))
	case "RangePicker":
		return simpleFieldName(field) + "Range"
	default:
		return ""
	}
}

func frontendKeySuffix(field string) string {
	if field == "" {
		return ""
	}
	return strings.ToUpper(field[:1]) + field[1:]
}

func frontendFormatter(field string) string {
	if strings.HasSuffix(field, "At") {
		return "formatDateTime"
	}
	return ""
}

func frontendSlot(field string) string {
	switch field {
	case "status", "actionType", "riskLevel", "platformSummary", "geoLocationSummary", "loginMethod", "successSummary", "httpMethod", "type", "auditStatus":
		return simpleFieldName(field)
	default:
		return ""
	}
}

func frontendWidth(field string) int {
	switch field {
	case "createdAt":
		return 150
	case "status":
		return 110
	case "username":
		return 140
	case "actionType", "loginMethod", "type", "auditStatus":
		return 120
	case "riskLevel":
		return 120
	case "platformSummary":
		return 170
	case "geoLocationSummary":
		return 220
	case "ipAddress":
		return 140
	case "deviceInfo.userAgent":
		return 280
	case "successSummary", "statusCode":
		return 120
	case "httpMethod":
		return 100
	case "path":
		return 260
	case "latencyMs", "memberCount":
		return 100
	case "apiOperation":
		return 180
	case "name":
		return 260
	case "domain":
		return 180
	default:
		return 160
	}
}

func frontendDialogComponent(field string) string {
	switch {
	case field == "method", field == "scope", field == "status", field == "type", field == "auditStatus", field == "actionType", field == "riskLevel", field == "loginMethod", field == "taskType":
		return "Select"
	case field == "remark", field == "description", field == "content":
		return "Textarea"
	case strings.HasSuffix(field, "At"):
		return "DatePicker"
	case strings.HasSuffix(field, "Status"), field == "status", field == "type":
		return "Select"
	default:
		return "Input"
	}
}

func (r *Runner) frontendDialogComponent(plan resourcePlan, field string) string {
	if r.frontendFieldRuntime(plan, field).Relation != nil {
		return "Select"
	}
	return frontendDialogComponent(field)
}

func (r *Runner) frontendFilterComponent(plan resourcePlan, filter config.FrontendFilter) string {
	if r.frontendFieldRuntime(plan, filter.Field).Relation != nil {
		return "Select"
	}
	return strings.TrimSpace(filter.Component)
}

func (r *Runner) frontendFieldRuntime(plan resourcePlan, field string) frontendFieldRuntime {
	fieldName := simpleFieldName(field)
	runtime := frontendFieldRuntime{FieldName: fieldName}

	if relation := r.frontendRelationForField(plan, field); relation != nil {
		runtime.Relation = &frontendFieldRelationRuntime{
			ResourceField:  frontendResourceFieldName(relation.Resource),
			PlaceholderKey: "select" + frontendKeySuffix(fieldName),
		}
	}

	if enumValues, ok := r.frontendEnumValues(plan, field); ok && len(enumValues) > 0 {
		runtime.Enum = &frontendFieldEnumRuntime{
			ResourceKey: lowerFirst(plan.ResourceField),
			Values:      enumValues,
		}
	}

	return runtime
}

func (r *Runner) frontendRelationForField(plan resourcePlan, field string) *config.FrontendRelationSpec {
	if plan.Resource.Frontend == nil {
		return nil
	}
	normalizedField := simpleFieldName(field)
	if plan.Resource.Frontend.Form != nil {
		for i := range plan.Resource.Frontend.Form.Fields {
			item := &plan.Resource.Frontend.Form.Fields[i]
			if simpleFieldName(item.Field) == normalizedField && item.Relation != nil {
				return item.Relation
			}
		}
	}
	if plan.Resource.Frontend.List != nil {
		for i := range plan.Resource.Frontend.List.Columns {
			item := &plan.Resource.Frontend.List.Columns[i]
			if simpleFieldName(item.Field) == normalizedField && item.Relation != nil {
				return item.Relation
			}
		}
	}
	return nil
}

func frontendResourceFieldName(resource string) string {
	parts := strings.Split(strings.TrimSpace(resource), "_")
	for i := range parts {
		part := strings.TrimSpace(parts[i])
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, "")
}

func (r *Runner) frontendEnumOptionsProps(plan resourcePlan, fieldName string) string {
	enumValues, ok := r.frontendEnumValues(plan, fieldName)
	if !ok || len(enumValues) == 0 {
		return ""
	}
	resourceKey := lowerFirst(plan.ResourceField)
	simpleField := simpleFieldName(fieldName)
	lines := make([]string, 0, len(enumValues)+1)
	lines = append(lines, "          options: [")
	for _, value := range enumValues {
		lines = append(lines, fmt.Sprintf("            { label: t('enum.%s.%s.%s'), value: '%s' },", resourceKey, simpleField, value, value))
	}
	lines = append(lines, "          ],")
	return strings.Join(lines, "\n")
}

func (r *Runner) frontendEnumValues(plan resourcePlan, fieldName string) ([]string, bool) {
	sourceDir, err := r.frontendExampleLangsDir()
	if err != nil {
		return nil, false
	}
	enumPath := filepath.Join(sourceDir, "en-US", "enum.json")
	content, err := os.ReadFile(enumPath)
	if err != nil {
		return nil, false
	}
	var payload map[string]any
	if err := json.Unmarshal(content, &payload); err != nil {
		return nil, false
	}
	resourceKey := lowerFirst(plan.ResourceField)
	resourceValue, ok := payload[resourceKey]
	if !ok {
		return nil, false
	}
	resourceMap, ok := resourceValue.(map[string]any)
	if !ok {
		return nil, false
	}
	fieldValue, ok := resourceMap[simpleFieldName(fieldName)]
	if !ok {
		return nil, false
	}
	fieldMap, ok := fieldValue.(map[string]any)
	if !ok {
		return nil, false
	}
	values := make([]string, 0, len(fieldMap))
	for key := range fieldMap {
		values = append(values, key)
	}
	slices.Sort(values)
	return values, true
}

func frontendI18nValue(lang, field, enTitle, cnTitle string, schemaField *entschema.Field) string {
	if lang == "en" && strings.TrimSpace(enTitle) != "" {
		return enTitle
	}
	if lang == "zh" && strings.TrimSpace(cnTitle) != "" {
		return cnTitle
	}
	if lang == "en" {
		if schemaField != nil && strings.TrimSpace(schemaField.Name) != "" {
			return splitSnakeWords(schemaField.Name)
		}
		return splitCamelWords(simpleFieldName(field))
	}
	if schemaField != nil && strings.TrimSpace(schemaField.Comment) != "" {
		return schemaField.Comment
	}
	return splitCamelWords(simpleFieldName(field))
}

func frontendI18nPlaceholderValue(lang, field, component, enTitle, cnTitle string, schemaField *entschema.Field) string {
	label := frontendI18nValue(lang, field, enTitle, cnTitle, schemaField)
	switch component {
	case "Select":
		if lang == "en" {
			return "Select " + label
		} else if lang == "zh" {
			return "选择 " + label
		}
		return label
	case "RangePicker":
		if lang == "en" {
			return "SE"
		} else if lang == "zh" {
			return "起至"
		}
		return label
	default:
		if lang == "en" {
			return "Search " + label
		}
		return label
	}
}

func simpleFieldName(field string) string {
	parts := strings.Split(field, ".")
	return parts[len(parts)-1]
}

func splitCamelWords(input string) string {
	var out []rune
	for i, r := range input {
		if i > 0 && r >= 'A' && r <= 'Z' {
			out = append(out, ' ')
		}
		out = append(out, r)
	}
	return strings.Title(string(out))
}

func splitSnakeWords(input string) string {
	parts := strings.Split(input, "_")
	for i := range parts {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, " ")
}

func frontendSnakeCase(input string) string {
	if input == "" {
		return ""
	}
	var out []rune
	for i, r := range input {
		if i > 0 && r >= 'A' && r <= 'Z' {
			out = append(out, '_')
		}
		out = append(out, r)
	}
	return strings.ToLower(string(out))
}
