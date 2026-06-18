package codegen

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/chnxq/xkit/internal/binding"
	codegentemplate "github.com/chnxq/xkit/internal/codegen/template"
	"github.com/chnxq/xkit/internal/config"
	"github.com/chnxq/xkit/internal/entschema"
	"github.com/chnxq/xkit/internal/project"
	xproto "github.com/chnxq/xkit/internal/proto"
)

type Options struct {
	DryRun  bool
	Version string
}

type generatedMeta struct {
	Version     string
	GeneratedAt string
}

type templateBase struct {
	Generated generatedMeta
	Module    string
}

type Result struct {
	Written []string
	Skipped []string
}

type Runner struct {
	project      project.Info
	config       config.Config
	protoIndex   map[string]xproto.Service
	bindingIndex map[string]binding.ServiceBinding
	schemaIndex  map[string]entschema.Schema
	options      Options
}

type resourcePlan struct {
	Resource        config.Resource
	Proto           xproto.Service
	Binding         binding.ServiceBinding
	ResourceField   string
	FileBase        string
	APIPackageAlias string
	Schema          entschema.Schema
}

type importSpec struct {
	Alias string
	Path  string
}

type serviceTemplateData struct {
	templateBase
	Imports         []importSpec
	StructName      string
	ConstructorName string
	APIPackageAlias string
	Embeds          []string
	Methods         []serviceMethodData
	HasRepo         bool
	RepoField       string
	RepoType        string
	ExtraRepos      []serviceRepoData
	ExtraFields     []serviceFieldData
	ResourceName    string
	Tree            *treeConfigData
	Aggregates      []aggregateConfigData
}

type serviceRepoData struct {
	Field string
	Type  string
}

type serviceFieldData struct {
	Field string
	Type  string
}

type serviceMethodData struct {
	Name           string
	Classification string
	Params         []namedType
	ResponseType   string
	Delegate       bool
	RepoField      string
	SuccessReturn  string
	Body           string
}

type namedType struct {
	Name string
	Type string
}

type registerTemplateData struct {
	templateBase
	Imports          []importSpec
	StructName       string
	RegisterFuncName string
	ServerType       string
	Services         []registerServiceData
}

type registerServiceData struct {
	FieldName     string
	Alias         string
	InterfaceName string
	RegisterFunc  string
}

type wireTemplateData struct {
	templateBase
	Imports      []importSpec
	Constructors []string
	Alias        string
	Layer        string
}

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
	Tree                     *treeConfigData
	Aggregates               []aggregateConfigData
}

type treeConfigData struct {
	ParentField   string
	PathField     string
	ChildrenField string
	ListMethod    string
}

type aggregateConfigData struct {
	Name            string
	RepoField       string
	RepoType        string
	ForeignKey      string
	ParentField     string
	CollectionField string
	CurrentField    string
	Primary         bool
}

type bootstrapTemplateData struct {
	templateBase
	Module            string
	ServiceName       string
	AppName           string
	ServerImport      string
	DataBootImport    string
	RepoImport        string
	ServiceImport     string
	RepoResources     []bootstrapResourceData
	ProviderResources []bootstrapResourceData
	ServerResources   []bootstrapResourceData
	HTTPResources     []bootstrapResourceData
	GRPCResources     []bootstrapResourceData
	GenerateGetAppCtx bool
}

type bootstrapResourceData struct {
	FieldName       string
	RepoVar         string
	RepoInterface   string
	RepoConstructor string
	HasRepo         bool
	ServiceVar      string
	ServiceName     string
	Constructor     string
	ServiceRepoVars []string
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
}

type existsCaseData struct {
	OneofType string
	Predicate string
	ValueExpr string
}

type setterData struct {
	Method         string
	Expr           string
	Kind           string
	Condition      string
	ClearMethod    string
	ClearCondition string
	Pre            string
}

type filterData struct {
	Field        string
	Predicate    string
	Kind         string
	CastType     string
	ParseBitSize string
	TimeField    string
}

type frontendMetaTemplateData struct {
	templateBase
	I18nPrefix           string
	DefaultSortField     string
	DefaultSortDirection string
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
	FieldName string
	Component string
	LabelKey  string
}

type i18nEntry struct {
	Key   string
	Value string
}

var aliasPattern = regexp.MustCompile(`\b([A-Za-z_][A-Za-z0-9_]*)\.`)

func New(info project.Info, cfg config.Config, options Options) (*Runner, error) {
	if cfg.Module != "" && cfg.Module != info.Module {
		return nil, fmt.Errorf("config module %q does not match target project module %q", cfg.Module, info.Module)
	}

	protoIndex, err := xproto.LoadServices(info.Root)
	if err != nil {
		return nil, err
	}

	bindingIndex, err := binding.Load(info.Root, info.Module)
	if err != nil {
		return nil, err
	}

	schemaIndex, err := entschema.Load(info.Root)
	if err != nil {
		return nil, err
	}

	return &Runner{
		project:      info,
		config:       cfg,
		protoIndex:   protoIndex,
		bindingIndex: bindingIndex,
		schemaIndex:  schemaIndex,
		options:      options,
	}, nil
}

func (r *Runner) Generate(target string) (Result, error) {
	switch target {
	case "service":
		return r.generateServiceFiles()
	case "repo":
		return r.generateRepoFiles()
	case "register":
		return r.generateRegisterFiles()
	case "wire":
		return r.generateWireFiles()
	case "bootstrap":
		return r.generateBootstrapFiles()
	case "frontend-meta":
		return r.generateFrontendMetaFiles()
	case "all":
		return r.generateAll()
	default:
		return Result{}, fmt.Errorf("unknown generation target %q", target)
	}
}

func (r *Runner) generateAll() (Result, error) {
	var result Result
	parts := []string{"service", "repo", "register", "bootstrap", "frontend-meta"}
	for _, part := range parts {
		partResult, err := r.Generate(part)
		if err != nil {
			return result, err
		}
		result.Written = append(result.Written, partResult.Written...)
		result.Skipped = append(result.Skipped, partResult.Skipped...)
	}
	return result, nil
}

func (r *Runner) generateFrontendMetaFiles() (Result, error) {
	plans, err := r.plans()
	if err != nil {
		return Result{}, err
	}

	outputRoot := "web/admin"
	if r.config.Frontend != nil && strings.TrimSpace(r.config.Frontend.OutputRoot) != "" {
		outputRoot = strings.TrimSpace(r.config.Frontend.OutputRoot)
	}
	baseDir := filepath.Join(r.project.Root, filepath.FromSlash(outputRoot), "views", "generated", "admin")

	var result Result

	configContent, err := renderAnyTemplate(codegentemplate.FrontendViewConfig, r.templateBase())
	if err != nil {
		return result, err
	}
	if err := r.writeGeneratedFile(filepath.Join(baseDir, "config.ts"), configContent, &result); err != nil {
		return result, err
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
		metaPath := filepath.Join(baseDir, filepath.FromSlash(plan.Resource.Frontend.ViewPath+".meta.ts"))
		if err := r.writeGeneratedFile(metaPath, metaContent, &result); err != nil {
			return result, err
		}

		zhEntries = append(zhEntries, r.frontendI18nEntries(plan, "zh")...)
		enEntries = append(enEntries, r.frontendI18nEntries(plan, "en")...)
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

func (r *Runner) generateServiceFiles() (Result, error) {
	plans, err := r.plans()
	if err != nil {
		return Result{}, err
	}

	var result Result
	sharedContent, err := renderTemplate(codegentemplate.ServiceSharedExt, r.templateBase())
	if err != nil {
		return result, err
	}
	sharedPath := filepath.Join(r.internalDir("service"), "service_shared_ext.go")
	if err := r.writeExtensionFile(sharedPath, sharedContent, &result); err != nil {
		return result, err
	}

	for _, plan := range plans {
		if !plan.Resource.Generate.EffectiveServiceStub() {
			continue
		}

		content, err := r.renderServiceFile(plan)
		if err != nil {
			return result, err
		}

		servicePath := filepath.Join(
			r.internalDir("service"),
			plan.FileBase+"_service.gen.go",
		)
		if err := r.writeGeneratedFile(servicePath, content, &result); err != nil {
			return result, err
		}

		extContent, err := renderTemplate(codegentemplate.ServiceExt, struct {
			templateBase
			StructName string
		}{
			templateBase: r.templateBase(),
			StructName:   plan.Binding.ServiceName,
		})
		if err != nil {
			return result, err
		}

		extPath := filepath.Join(
			r.internalDir("service"),
			plan.FileBase+"_service_ext.go",
		)
		if err := r.writeExtensionFile(extPath, extContent, &result); err != nil {
			return result, err
		}
	}

	return result, nil
}

func (r *Runner) generateRepoFiles() (Result, error) {
	plans, err := r.plans()
	if err != nil {
		return Result{}, err
	}

	var result Result
	sharedContent, err := renderTemplate(codegentemplate.RepoSharedExt, r.templateBase())
	if err != nil {
		return result, err
	}
	sharedPath := filepath.Join(r.internalDir("data", "repo"), "repo_shared_ext.go")
	if err := r.writeExtensionFile(sharedPath, sharedContent, &result); err != nil {
		return result, err
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

func (r *Runner) generateRegisterFiles() (Result, error) {
	plans, err := r.plans()
	if err != nil {
		return Result{}, err
	}

	var httpServices []registerServiceData
	var grpcServices []registerServiceData
	httpImports := []importSpec{{Alias: "httptransport", Path: "github.com/chnxq/xkitpkg/transport/http"}}
	grpcImports := []importSpec{{Path: "google.golang.org/grpc"}}

	for _, plan := range plans {
		if plan.Resource.Generate.EffectiveRestRegister() {
			httpInterfaceName := bindingHTTPServerInterfaceName(plan.Binding)
			httpServices = append(httpServices, registerServiceData{
				FieldName:     plan.ResourceField,
				Alias:         plan.APIPackageAlias,
				InterfaceName: httpInterfaceName,
				RegisterFunc:  "Register" + strings.TrimSuffix(httpInterfaceName, "HTTPServer") + "HTTPServer",
			})
			httpImports = append(httpImports, importSpec{Alias: plan.APIPackageAlias, Path: plan.Binding.ImportPath})
		}

		if plan.Resource.Generate.EffectiveGRPCRegister() {
			grpcInterfaceName := bindingGRPCServerInterfaceName(plan.Binding)
			grpcServices = append(grpcServices, registerServiceData{
				FieldName:     plan.ResourceField,
				Alias:         plan.APIPackageAlias,
				InterfaceName: grpcInterfaceName,
				RegisterFunc:  "Register" + strings.TrimSuffix(grpcInterfaceName, "Server") + "Server",
			})
			grpcImports = append(grpcImports, importSpec{Alias: plan.APIPackageAlias, Path: plan.Binding.ImportPath})
		}
	}

	var result Result
	if len(httpServices) > 0 {
		httpContent, err := renderTemplate(codegentemplate.RegisterFile, registerTemplateData{
			templateBase:     r.templateBase(),
			Imports:          uniqueImports(httpImports),
			StructName:       "GeneratedHTTPServices",
			RegisterFuncName: "RegisterGeneratedHTTPServices",
			ServerType:       "*httptransport.Server",
			Services:         httpServices,
		})
		if err != nil {
			return result, err
		}

		httpPath := filepath.Join(
			r.internalDir("server"),
			"rest_register.gen.go",
		)
		if err := r.writeGeneratedFile(httpPath, httpContent, &result); err != nil {
			return result, err
		}
	}

	if len(grpcServices) > 0 {
		grpcContent, err := renderTemplate(codegentemplate.RegisterFile, registerTemplateData{
			templateBase:     r.templateBase(),
			Imports:          uniqueImports(grpcImports),
			StructName:       "GeneratedGRPCServices",
			RegisterFuncName: "RegisterGeneratedGRPCServices",
			ServerType:       "grpc.ServiceRegistrar",
			Services:         grpcServices,
		})
		if err != nil {
			return result, err
		}

		grpcPath := filepath.Join(
			r.internalDir("server"),
			"grpc_register.gen.go",
		)
		if err := r.writeGeneratedFile(grpcPath, grpcContent, &result); err != nil {
			return result, err
		}
	}

	return result, nil
}

func (r *Runner) generateWireFiles() (Result, error) {
	plans, err := r.plans()
	if err != nil {
		return Result{}, err
	}

	var serviceConstructors []string
	var dataConstructors []string
	for _, plan := range plans {
		if !plan.Resource.Generate.EffectiveWire() {
			continue
		}
		if plan.Resource.Generate.EffectiveServiceStub() {
			serviceConstructors = append(serviceConstructors, "New"+plan.Binding.ServiceName)
		}
		if plan.Resource.Generate.EffectiveRepoCRUD() {
			repoName := strings.TrimSpace(plan.Resource.RepoInterface)
			if repoName == "" {
				entityName := strings.TrimSpace(plan.Resource.Entity)
				if entityName == "" {
					entityName = strings.TrimSuffix(plan.Binding.ServiceName, "Service")
				}
				repoName = entityName + "Repo"
			}
			dataConstructors = append(dataConstructors, "New"+repoName)
		}
	}

	serviceImports := []importSpec{{Path: "github.com/google/wire"}}
	if len(serviceConstructors) > 0 {
		serviceImports = append(serviceImports, importSpec{
			Alias: "servicepkg",
			Path:  r.internalImport("service"),
		})
	}
	dataImports := []importSpec{{Path: "github.com/google/wire"}}
	if len(dataConstructors) > 0 {
		dataImports = append(dataImports, importSpec{
			Alias: "repopkg",
			Path:  r.internalImport("data", "repo"),
		})
	}

	serviceWireContent, err := renderTemplate(codegentemplate.WireFile, wireTemplateData{
		templateBase: r.templateBase(),
		Imports:      serviceImports,
		Constructors: serviceConstructors,
		Alias:        "servicepkg",
		Layer:        "service",
	})
	if err != nil {
		return Result{}, err
	}

	dataWireContent, err := renderTemplate(codegentemplate.WireFile, wireTemplateData{
		templateBase: r.templateBase(),
		Imports:      dataImports,
		Constructors: dataConstructors,
		Alias:        "repopkg",
		Layer:        "data repo",
	})
	if err != nil {
		return Result{}, err
	}

	var result Result
	serviceWirePath := filepath.Join(
		r.internalDir("service"),
		"providers",
		"wire_set.gen.go",
	)
	if err := r.writeGeneratedFile(serviceWirePath, serviceWireContent, &result); err != nil {
		return result, err
	}

	dataWirePath := filepath.Join(
		r.internalDir("data"),
		"providers",
		"wire_set.gen.go",
	)
	if err := r.writeGeneratedFile(dataWirePath, dataWireContent, &result); err != nil {
		return result, err
	}

	return result, nil
}

func (r *Runner) plans() ([]resourcePlan, error) {
	var plans []resourcePlan
	for _, resource := range r.config.Resources {
		protoService, ok := r.protoIndex[resource.ProtoService]
		if !ok {
			return nil, fmt.Errorf("proto service %q not found under api/protos", resource.ProtoService)
		}

		bindingService, ok := r.bindingIndex[resource.ProtoService]
		if !ok {
			return nil, fmt.Errorf("binding service %q not found under api/gen", resource.ProtoService)
		}

		entityName := strings.TrimSpace(resource.Entity)
		if entityName == "" {
			entityName = strings.TrimSuffix(bindingService.ServiceName, "Service")
		}
		schema, ok := r.schemaIndex[entityName]
		if !ok && resource.Generate.EffectiveRepoCRUD() {
			return nil, fmt.Errorf("ent schema %q not found under internal/data/ent/schema", entityName)
		}

		plans = append(plans, resourcePlan{
			Resource:        resource,
			Proto:           protoService,
			Binding:         bindingService,
			ResourceField:   toPascal(resource.Name),
			FileBase:        resource.Name,
			APIPackageAlias: apiAlias(bindingService.ImportPath),
			Schema:          schema,
		})
	}

	return plans, nil
}

func (r *Runner) internalDir(parts ...string) string {
	pathParts := append([]string{r.project.Root, "internal"}, parts...)
	return filepath.Join(pathParts...)
}

func (r *Runner) internalImport(parts ...string) string {
	pathParts := append([]string{r.project.Module, "internal"}, parts...)
	return filepath.ToSlash(filepath.Join(pathParts...))
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
		component := strings.TrimSpace(filter.Component)
		labelKey := r.frontendFilterLabelKey(plan, filter)
		item := frontendFilterItem{
			FieldName:      fieldName,
			Component:      component,
			LabelKey:       labelKey,
			PlaceholderKey: frontendPlaceholderKey(component, labelKey),
		}
		if component == "RangePicker" {
			item.ExtraProps = "          showTime: true,\n          valueFormat: 'YYYY-MM-DD HH:mm:ss',"
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
			slotDefault = frontendSlot(field)
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
		items = append(items, frontendDialogField{
			FieldName: field.Field,
			Component: frontendDialogComponent(field.Field),
			LabelKey:  field.Field,
		})
	}
	return items
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
			component := filter.Component
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
	slices.SortFunc(entries, func(a, b i18nEntry) int {
		return strings.Compare(a.Key, b.Key)
	})
	data := struct {
		Entries   []i18nEntry
		LastIndex int
	}{
		Entries:   entries,
		LastIndex: len(entries) - 1,
	}
	content, err := renderAnyTemplate(tmpl, data)
	if err != nil {
		return err
	}
	var formatted bytes.Buffer
	if err := json.Indent(&formatted, content, "", "  "); err == nil {
		content = formatted.Bytes()
	}
	return r.writeGeneratedFile(path, append(content, '\n'), result)
}

func (r *Runner) copyFrontendGeneratedLangs(baseDir string, result *Result) error {
	sourceDir, err := xkitExampleLangsDir()
	if err != nil {
		return err
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
			targetPath := filepath.Join(baseDir, "langs", entry.Name(), file.Name())
			if err := r.writeGeneratedFile(targetPath, content, result); err != nil {
				return err
			}
		}
	}
	return nil
}

func xkitExampleLangsDir() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("resolve xkit source location failed")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "examples", "admin", "langs"), nil
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
	case "status", "actionType", "riskLevel", "platformSummary", "geoLocationSummary", "loginMethod", "successSummary", "httpMethod", "type", "auditStatus", "name":
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

func (r *Runner) bootstrapResources(plans []resourcePlan) []bootstrapResourceData {
	resourceIndex := make(map[string]resourcePlan, len(plans))
	for _, plan := range plans {
		resourceIndex[plan.Resource.Name] = plan
	}

	resources := make([]bootstrapResourceData, 0, len(plans))
	for _, plan := range plans {
		if !plan.Resource.Generate.EffectiveServiceStub() {
			continue
		}
		repoName := strings.TrimSpace(plan.Resource.RepoInterface)
		hasRepo := repoName != "" && plan.Resource.Generate.EffectiveRepoCRUD()
		data := bootstrapResourceData{
			FieldName:       plan.ResourceField,
			RepoVar:         repoVarFromInterface(repoName),
			RepoInterface:   repoName,
			RepoConstructor: "New" + repoName,
			HasRepo:         hasRepo,
			ServiceVar:      lowerFirst(strings.TrimSuffix(plan.Binding.ServiceName, "Service")) + "Service",
			ServiceName:     plan.Binding.ServiceName,
			Constructor:     "New" + plan.Binding.ServiceName,
		}

		for _, serviceRepo := range r.serviceRepoConfigs(plan) {
			extraPlan, ok := findPlanByRepoInterface(resourceIndex, serviceRepo.Interface)
			if !ok {
				continue
			}
			extraRepoVar := repoVarFromInterface(strings.TrimSpace(extraPlan.Resource.RepoInterface))
			if !extraPlan.Resource.Generate.EffectiveRepoCRUD() {
				continue
			}
			if !slices.Contains(data.ServiceRepoVars, extraRepoVar) {
				data.ServiceRepoVars = append(data.ServiceRepoVars, extraRepoVar)
			}
		}

		resources = append(resources, data)
	}
	return resources
}

func (r *Runner) bootstrapRepoResources(plans []resourcePlan) []bootstrapResourceData {
	resourceIndex := make(map[string]resourcePlan, len(plans))
	for _, plan := range plans {
		resourceIndex[plan.Resource.Name] = plan
	}

	resources := make([]bootstrapResourceData, 0, len(plans))
	seen := make(map[string]struct{}, len(plans))
	for _, plan := range plans {
		repoName := strings.TrimSpace(plan.Resource.RepoInterface)
		if repoName == "" || !plan.Resource.Generate.EffectiveRepoCRUD() {
			continue
		}
		if _, ok := seen[repoName]; ok {
			continue
		}
		seen[repoName] = struct{}{}

		resources = append(resources, r.bootstrapRepoResource(plan))
		for _, serviceRepo := range r.serviceRepoConfigs(plan) {
			extraPlan, ok := findPlanByRepoInterface(resourceIndex, serviceRepo.Interface)
			if !ok {
				continue
			}
			extraRepoName := strings.TrimSpace(extraPlan.Resource.RepoInterface)
			if extraRepoName == "" || !extraPlan.Resource.Generate.EffectiveRepoCRUD() {
				continue
			}
			if _, ok := seen[extraRepoName]; ok {
				continue
			}
			seen[extraRepoName] = struct{}{}
			resources = append(resources, r.bootstrapRepoResource(extraPlan))
		}
	}
	return resources
}

func (r *Runner) bootstrapRepoResource(plan resourcePlan) bootstrapResourceData {
	repoName := strings.TrimSpace(plan.Resource.RepoInterface)
	return bootstrapResourceData{
		FieldName:       plan.ResourceField,
		RepoVar:         repoVarFromInterface(repoName),
		RepoInterface:   repoName,
		RepoConstructor: "New" + repoName,
	}
}

func (r *Runner) bootstrapProviderResources(plans []resourcePlan) ([]bootstrapResourceData, bool) {
	existingMethods := receiverMethodNames(
		filepath.Join(r.project.Root, "internal", "bootstrap"),
		"GeneratedData",
		map[string]struct{}{
			"generated_data_providers.gen.go": {},
		},
	)

	providerResources := make([]bootstrapResourceData, 0, len(plans))
	for _, resource := range r.bootstrapRepoResources(plans) {
		methodName := resource.FieldName + "RepoProvider"
		if _, exists := existingMethods[methodName]; exists {
			continue
		}
		providerResources = append(providerResources, resource)
	}

	_, hasGetAppCtx := existingMethods["GetAppCtx"]
	return providerResources, !hasGetAppCtx
}

func bootstrapRegisteredResources(resources []bootstrapResourceData, plans []resourcePlan, enabled func(config.GenerateFlags) bool) []bootstrapResourceData {
	registeredFields := make(map[string]struct{}, len(plans))
	for _, plan := range plans {
		if !plan.Resource.Generate.EffectiveServiceStub() || !enabled(plan.Resource.Generate) {
			continue
		}
		registeredFields[plan.ResourceField] = struct{}{}
	}

	out := make([]bootstrapResourceData, 0, len(resources))
	for _, resource := range resources {
		if _, ok := registeredFields[resource.FieldName]; ok {
			out = append(out, resource)
		}
	}
	return out
}

func findPlanByRepoInterface(resourceIndex map[string]resourcePlan, repoInterface string) (resourcePlan, bool) {
	for _, plan := range resourceIndex {
		if plan.Resource.RepoInterface == repoInterface {
			return plan, true
		}
	}
	return resourcePlan{}, false
}

func (r *Runner) generateBootstrapFiles() (Result, error) {
	plans, err := r.plans()
	if err != nil {
		return Result{}, err
	}

	serverResources := r.bootstrapResources(plans)
	providerResources, generateGetAppCtx := r.bootstrapProviderResources(plans)
	data := bootstrapTemplateData{
		templateBase:      r.templateBase(),
		Module:            r.project.Module,
		ServiceName:       r.config.Service,
		AppName:           r.project.Module,
		ServerImport:      r.internalImport("server"),
		DataBootImport:    r.internalImport("data", "bootstrap"),
		RepoImport:        r.internalImport("data", "repo"),
		ServiceImport:     r.internalImport("service"),
		RepoResources:     r.bootstrapRepoResources(plans),
		ProviderResources: providerResources,
		ServerResources:   serverResources,
		HTTPResources: bootstrapRegisteredResources(serverResources, plans, func(flags config.GenerateFlags) bool {
			return flags.EffectiveRestRegister()
		}),
		GRPCResources: bootstrapRegisteredResources(serverResources, plans, func(flags config.GenerateFlags) bool {
			return flags.EffectiveGRPCRegister()
		}),
		GenerateGetAppCtx: generateGetAppCtx,
	}

	if err := r.removeObsoleteGeneratedFile(filepath.Join(r.project.Root, "internal", "data", "bootstrap", "ent_client.go")); err != nil {
		return Result{}, err
	}
	files := []struct {
		path     string
		template string
	}{
		{path: filepath.Join(r.project.Root, "internal", "bootstrap", "generated_servers.gen.go"), template: codegentemplate.BootstrapGeneratedServers},
		{path: filepath.Join(r.project.Root, "internal", "bootstrap", "generated_data_providers.gen.go"), template: codegentemplate.BootstrapGeneratedDataProviders},
		{path: filepath.Join(r.project.Root, "internal", "data", "bootstrap", "ent_client.gen.go"), template: codegentemplate.BootstrapEntClient},
	}

	var result Result
	for _, file := range files {
		content, err := renderAnyTemplate(file.template, data)
		if err != nil {
			return result, err
		}
		if err := r.writeGeneratedFile(file.path, content, &result); err != nil {
			return result, err
		}
	}

	hooksContent, err := renderTemplate(codegentemplate.BootstrapHooksExt, r.templateBase())
	if err != nil {
		return result, err
	}
	hooksPath := filepath.Join(r.project.Root, "internal", "bootstrap", "generated_hooks_ext.go")
	if err := r.writeExtensionFile(hooksPath, hooksContent, &result); err != nil {
		return result, err
	}

	entHooksContent, err := renderAnyTemplate(codegentemplate.BootstrapEntClientExt, data)
	if err != nil {
		return result, err
	}
	entHooksPath := filepath.Join(r.project.Root, "internal", "data", "bootstrap", "ent_client_ext.go")
	if err := r.writeExtensionFile(entHooksPath, entHooksContent, &result); err != nil {
		return result, err
	}

	return result, nil
}

func (r *Runner) renderServiceFile(plan resourcePlan) ([]byte, error) {
	imports := []importSpec{
		{Alias: plan.APIPackageAlias, Path: plan.Binding.ImportPath},
	}

	repoName := strings.TrimSpace(plan.Resource.RepoInterface)
	if repoName == "" {
		entityName := strings.TrimSpace(plan.Resource.Entity)
		if entityName == "" {
			entityName = strings.TrimSuffix(plan.Binding.ServiceName, "Service")
		}
		repoName = entityName + "Repo"
	}
	repoField := lowerFirst(strings.TrimSuffix(repoName, "Repo")) + "Repo"
	repoType := "repo." + repoName
	hasRepo := repoName != "" && plan.Resource.Generate.EffectiveRepoCRUD()
	extraRepos := r.extraServiceRepos(plan)
	extraFields := r.extraServiceFields(plan)
	if hasRepo {
		imports = append(imports,
			importSpec{Path: "github.com/chnxq/xkitmod/log"},
			importSpec{Path: "github.com/chnxq/xkitpkg/app"},
			importSpec{Path: r.internalImport("data", "repo")},
		)
	}
	imports = append(imports, r.serviceConfiguredImports(plan)...)

	usedAliases := make(map[string]struct{})
	methods := make([]serviceMethodData, 0, len(plan.Binding.Methods))
	for _, method := range plan.Binding.Methods {
		for _, typeText := range append(slices.Clone(method.Params), method.Results...) {
			for _, alias := range aliasesInType(typeText) {
				usedAliases[alias] = struct{}{}
			}
		}

		kind := repoMethodKind(method.Name)
		delegate := hasRepo && isCRUDMethod(method.Name) && resourceOperationEnabled(plan.Resource, kind)
		classification := lookupClassification(plan.Proto.Methods, method.Name)
		responseType := firstResult(method.Results)
		params := nameParams(method.Params)
		methods = append(methods, serviceMethodData{
			Name:           method.Name,
			Classification: classification,
			Params:         params,
			ResponseType:   responseType,
			Delegate:       delegate,
			RepoField:      repoField,
			SuccessReturn:  serviceSuccessReturn(responseType),
			Body:           r.serviceMethodBody(plan, method.Name, params, responseType, repoField, hasRepo),
		})
	}

	for alias := range usedAliases {
		path, ok := plan.Binding.Imports[alias]
		if !ok {
			continue
		}
		imports = append(imports, importSpec{Alias: alias, Path: path})
	}

	data := serviceTemplateData{
		templateBase:    r.templateBase(),
		Imports:         uniqueImports(imports),
		StructName:      plan.Binding.ServiceName,
		ConstructorName: "New" + plan.Binding.ServiceName,
		APIPackageAlias: plan.APIPackageAlias,
		Embeds:          serviceEmbeds(plan.APIPackageAlias, plan.Binding),
		Methods:         methods,
		HasRepo:         hasRepo,
		RepoField:       repoField,
		RepoType:        repoType,
		ExtraRepos:      extraRepos,
		ExtraFields:     extraFields,
		ResourceName:    plan.Resource.Name,
		Tree:            buildTreeConfig(plan.Resource.Tree),
		Aggregates:      r.aggregateConfigs(plan),
	}

	return renderTemplate(codegentemplate.ServiceFile, data)
}

func (r *Runner) serviceMethodBody(plan resourcePlan, methodName string, params []namedType, responseType, repoField string, hasRepo bool) string {
	methodConfig, ok := plan.Resource.ServiceMethods[methodName]
	if !hasRepo || !ok || strings.TrimSpace(methodConfig.Body) == "" {
		return ""
	}
	body := strings.TrimRight(methodConfig.Body, "\r\n")
	replacements := map[string]string{
		"{{repoField}}":     repoField,
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

func (r *Runner) extraServiceRepos(plan resourcePlan) []serviceRepoData {
	var repos []serviceRepoData
	seen := make(map[string]struct{})
	for _, repoConfig := range r.serviceRepoConfigs(plan) {
		field := strings.TrimSpace(repoConfig.Field)
		typeName := strings.TrimSpace(repoConfig.Interface)
		if field == "" || typeName == "" {
			continue
		}
		key := field + ":" + typeName
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		repos = append(repos, serviceRepoData{Field: field, Type: "repo." + typeName})
	}
	return repos
}

func (r *Runner) extraServiceFields(plan resourcePlan) []serviceFieldData {
	fields := make([]serviceFieldData, 0, len(plan.Resource.ServiceFields))
	for _, field := range plan.Resource.ServiceFields {
		if strings.TrimSpace(field.Field) == "" || strings.TrimSpace(field.Type) == "" {
			continue
		}
		fields = append(fields, serviceFieldData{
			Field: strings.TrimSpace(field.Field),
			Type:  strings.TrimSpace(field.Type),
		})
	}
	return fields
}

func (r *Runner) serviceRepoConfigs(plan resourcePlan) []config.RepoConfig {
	repos := make([]config.RepoConfig, 0, len(plan.Resource.ServiceRepos)+len(plan.Resource.ServiceMethods)+len(plan.Resource.Aggregates))
	repos = append(repos, plan.Resource.ServiceRepos...)
	for _, aggregate := range plan.Resource.Aggregates {
		repoInterface := strings.TrimSpace(aggregate.RepoInterface)
		if repoInterface == "" && strings.TrimSpace(aggregate.Resource) != "" {
			for _, resource := range r.config.Resources {
				if resource.Name == aggregate.Resource && strings.TrimSpace(resource.RepoInterface) != "" {
					repoInterface = strings.TrimSpace(resource.RepoInterface)
					break
				}
			}
		}
		if repoInterface == "" {
			continue
		}
		field := lowerFirst(strings.TrimSuffix(repoInterface, "Repo")) + "Repo"
		if strings.TrimSpace(aggregate.Name) != "" {
			field = lowerFirst(strings.TrimSpace(aggregate.Name)) + "Repo"
		}
		repos = append(repos, config.RepoConfig{
			Field:     field,
			Interface: repoInterface,
		})
	}
	for _, methodConfig := range plan.Resource.ServiceMethods {
		repos = append(repos, methodConfig.Repos...)
	}
	return repos
}

func (r *Runner) serviceConfiguredImports(plan resourcePlan) []importSpec {
	var imports []importSpec
	for _, importConfig := range plan.Resource.ServiceImports {
		path := strings.TrimSpace(importConfig.Path)
		if path == "" {
			continue
		}
		path = strings.ReplaceAll(path, "{{module}}", r.project.Module)
		imports = append(imports, importSpec{Alias: strings.TrimSpace(importConfig.Alias), Path: filepath.ToSlash(path)})
	}
	for _, methodConfig := range plan.Resource.ServiceMethods {
		for _, importConfig := range methodConfig.Imports {
			path := strings.TrimSpace(importConfig.Path)
			if path == "" {
				continue
			}
			path = strings.ReplaceAll(path, "{{module}}", r.project.Module)
			imports = append(imports, importSpec{Alias: strings.TrimSpace(importConfig.Alias), Path: filepath.ToSlash(path)})
		}
	}
	return imports
}

func (r *Runner) repoConfiguredImports(plan resourcePlan) []importSpec {
	var imports []importSpec
	for _, methodConfig := range plan.Resource.RepoMethods {
		for _, importConfig := range methodConfig.Imports {
			path := strings.TrimSpace(importConfig.Path)
			if path == "" {
				continue
			}
			path = strings.ReplaceAll(path, "{{module}}", r.project.Module)
			imports = append(imports, importSpec{Alias: strings.TrimSpace(importConfig.Alias), Path: filepath.ToSlash(path)})
		}
	}
	return imports
}

func indentLines(text, prefix string) string {
	if text == "" {
		return ""
	}
	lines := strings.Split(text, "\n")
	for index, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lines[index] = prefix + line
	}
	return strings.Join(lines, "\n")
}

func isSupportedRepoSpecialMethod(resource config.Resource, methodName string) bool {
	return resource.Name == "user_credential" && methodName == "ResetCredential"
}

func hasConfiguredRepoMethod(resource config.Resource, methodName string) bool {
	methodConfig, ok := resource.RepoMethods[methodName]
	return ok && strings.TrimSpace(methodConfig.Body) != ""
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

func (r *Runner) renderRepoFile(plan resourcePlan) ([]byte, error) {
	entityName := strings.TrimSpace(plan.Resource.Entity)
	if entityName == "" {
		entityName = plan.Binding.ServiceName
		entityName = strings.TrimSuffix(entityName, "Service")
	}

	dtoAlias := plan.APIPackageAlias
	dtoImport := plan.Binding.ImportPath
	if strings.TrimSpace(plan.Resource.DTOImport) != "" {
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
		{Path: filepath.ToSlash(filepath.Join(r.project.Module, "internal", "data", "ent"))},
		{Path: filepath.ToSlash(filepath.Join(r.project.Module, "internal", "data", "ent", strings.ToLower(entityName)))},
		{Path: filepath.ToSlash(filepath.Join(r.project.Module, "internal", "data", "ent", "predicate"))},
		{Alias: dtoAlias, Path: dtoImport},
	}
	filters := repoFilters(plan.Schema.Fields, effectiveFilterAllow(plan))
	usesFilters := len(filters) > 0
	usesAuditFields := hasGeneratedAuditFields(plan.Schema.Fields)
	if usesFilters {
		imports = append(imports,
			importSpec{Path: "strconv"},
			importSpec{Path: "strings"},
			importSpec{Alias: "paginationv1", Path: "github.com/chnxq/x-crud/api/gen/pagination/v1"},
		)
	}
	if usesAuditFields {
		imports = append(imports,
			importSpec{Alias: "crudviewer", Path: "github.com/chnxq/x-crud/viewer"},
			importSpec{Path: "time"},
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
		Tree:                     buildTreeConfig(plan.Resource.Tree),
		Aggregates:               r.aggregateConfigs(plan),
	}
	data.Methods = methods

	return renderTemplate(codegentemplate.RepoFile, data)
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

func (r *Runner) aggregateConfigs(plan resourcePlan) []aggregateConfigData {
	if len(plan.Resource.Aggregates) == 0 {
		return nil
	}

	byName := make(map[string]config.Resource, len(r.config.Resources))
	for _, resource := range r.config.Resources {
		byName[resource.Name] = resource
	}

	aggregates := make([]aggregateConfigData, 0, len(plan.Resource.Aggregates))
	for _, item := range plan.Resource.Aggregates {
		repoInterface := strings.TrimSpace(item.RepoInterface)
		if repoInterface == "" && strings.TrimSpace(item.Resource) != "" {
			if resource, ok := byName[item.Resource]; ok {
				if resource.Generate.EffectiveRepoCRUD() {
					repoInterface = strings.TrimSpace(resource.RepoInterface)
				}
			}
		}
		if repoInterface == "" {
			continue
		}
		repoField := lowerFirst(strings.TrimSuffix(repoInterface, "Repo")) + "Repo"
		if strings.TrimSpace(item.Name) != "" {
			repoField = lowerFirst(strings.TrimSpace(item.Name)) + "Repo"
		}
		aggregates = append(aggregates, aggregateConfigData{
			Name:            strings.TrimSpace(item.Name),
			RepoField:       repoField,
			RepoType:        "repo." + repoInterface,
			ForeignKey:      strings.TrimSpace(item.ForeignKey),
			ParentField:     strings.TrimSpace(item.ParentField),
			CollectionField: strings.TrimSpace(item.CollectionField),
			CurrentField:    strings.TrimSpace(item.CurrentField),
			Primary:         item.Primary,
		})
	}
	return aggregates
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

func (r *Runner) writeGeneratedFile(path string, content []byte, result *Result) error {
	return r.writeFile(path, content, result, false)
}

func (r *Runner) removeObsoleteGeneratedFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read obsolete generated file %s: %w", path, err)
	}
	if !bytes.Contains(data, []byte("Code generated by xkit. DO NOT EDIT.")) {
		return nil
	}
	if r.options.DryRun {
		return nil
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove obsolete generated file %s: %w", path, err)
	}
	return nil
}

func (r *Runner) templateBase() templateBase {
	return templateBase{
		Generated: generatedMeta{
			Version:     r.version(),
			GeneratedAt: time.Now().Format("2006-01-02 15:04:05 MST"),
		},
		Module: r.project.Module,
	}
}

func (r *Runner) version() string {
	if strings.TrimSpace(r.options.Version) == "" {
		return "dev"
	}
	return r.options.Version
}

func (r *Runner) writeExtensionFile(path string, content []byte, result *Result) error {
	if r.options.DryRun {
		result.Written = append(result.Written, path)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory for %s: %w", path, err)
	}

	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read existing extension file %s: %w", path, err)
	}
	if errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(path, content, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
		result.Written = append(result.Written, path)
		return nil
	}

	if generatedContentEquivalent(existing, content) {
		result.Skipped = append(result.Skipped, path)
		return nil
	}

	merged := refreshExtensionHeader(existing, content)
	if generatedContentEquivalent(existing, merged) {
		result.Skipped = append(result.Skipped, path)
		return nil
	}
	if err := os.WriteFile(path, merged, 0o644); err != nil {
		return fmt.Errorf("write refreshed extension file %s: %w", path, err)
	}

	result.Written = append(result.Written, path)
	return nil
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

func (r *Runner) writeFile(path string, content []byte, result *Result, skipIfExists bool) error {
	if skipIfExists {
		if _, err := os.Stat(path); err == nil {
			result.Skipped = append(result.Skipped, path)
			return nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat %s: %w", path, err)
		}
	}

	if r.options.DryRun {
		result.Written = append(result.Written, path)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory for %s: %w", path, err)
	}
	existing, err := os.ReadFile(path)
	if err == nil {
		if generatedContentEquivalent(existing, content) {
			result.Skipped = append(result.Skipped, path)
			return nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read existing file %s: %w", path, err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	result.Written = append(result.Written, path)
	return nil
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
		if !field.Optional {
			setters = append(setters, setterData{
				Method: directMethod,
				Expr:   directExpr,
				Kind:   field.Kind,
			})
			continue
		}
		if methodName == "Create" {
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

func lookupClassification(methods []xproto.Method, name string) string {
	for _, method := range methods {
		if method.Name == name {
			return method.Classification
		}
	}
	return "special"
}
