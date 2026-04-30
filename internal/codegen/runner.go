package codegen

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
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
	ResourceName    string
}

type serviceRepoData struct {
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
	Imports          []importSpec
	RepoName         string
	RepoStructName   string
	ConstructorName  string
	EntityName       string
	EntOperationName string
	ResourceName     string
	EntPackage       string
	PredicateType    string
	DTOType          string
	IDType           string
	Methods          []repoMethodData
	Fields           []entschema.Field
	UsesEnumSetter   bool
	UsesTimeSetter   bool
	EnumHelperName   string
	TimeHelperName   string
	Filters          []filterData
	UsesFilters      bool
	UsesAuditFields  bool
}

type bootstrapTemplateData struct {
	templateBase
	Module          string
	ServiceName     string
	AppName         string
	ServerImport    string
	DataBootImport  string
	RepoImport      string
	ServiceImport   string
	RepoResources   []bootstrapResourceData
	ServerResources []bootstrapResourceData
	HTTPResources   []bootstrapResourceData
	GRPCResources   []bootstrapResourceData
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
	ExtraRepoVars   []string
}

type repoMethodData struct {
	Name            string
	Params          []namedType
	ResponseType    string
	Kind            string
	Body            string
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
	Method string
	Expr   string
	Kind   string
}

type filterData struct {
	Field        string
	Predicate    string
	Kind         string
	CastType     string
	ParseBitSize string
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
	case "all":
		return r.generateAll()
	default:
		return Result{}, fmt.Errorf("unknown generation target %q", target)
	}
}

func (r *Runner) generateAll() (Result, error) {
	var result Result
	parts := []string{"service", "repo", "register", "bootstrap"}
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

func (r *Runner) generateServiceFiles() (Result, error) {
	plans, err := r.plans()
	if err != nil {
		return Result{}, err
	}

	var result Result
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

		extContent, err := renderTemplate(codegentemplate.ServiceExt, map[string]string{
			"StructName": plan.Binding.ServiceName,
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

		extContent, err := renderTemplate(codegentemplate.RepoExt, map[string]string{
			"RepoName": repoInterfaceName(plan),
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
			httpServices = append(httpServices, registerServiceData{
				FieldName:     plan.ResourceField,
				Alias:         plan.APIPackageAlias,
				InterfaceName: plan.Binding.ServiceName + "HTTPServer",
				RegisterFunc:  "Register" + plan.Binding.ServiceName + "HTTPServer",
			})
			httpImports = append(httpImports, importSpec{Alias: plan.APIPackageAlias, Path: plan.Binding.ImportPath})
		}

		if plan.Resource.Generate.EffectiveGRPCRegister() {
			grpcServices = append(grpcServices, registerServiceData{
				FieldName:     plan.ResourceField,
				Alias:         plan.APIPackageAlias,
				InterfaceName: plan.Binding.ServiceName + "Server",
				RegisterFunc:  "Register" + plan.Binding.ServiceName + "Server",
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
		entityName := strings.TrimSuffix(plan.Binding.ServiceName, "Service")
		data := bootstrapResourceData{
			FieldName:       plan.ResourceField,
			RepoVar:         lowerFirst(entityName) + "Repo",
			RepoInterface:   repoInterfaceName(plan),
			RepoConstructor: "New" + repoInterfaceName(plan),
			HasRepo:         plan.Resource.Generate.EffectiveRepoCRUD(),
			ServiceVar:      lowerFirst(entityName) + "Service",
			ServiceName:     plan.Binding.ServiceName,
			Constructor:     "New" + plan.Binding.ServiceName,
		}

		for _, method := range plan.Resource.ServiceMethods {
			for _, extraRepo := range method.Repos {
				extraPlan, ok := findPlanByRepoInterface(resourceIndex, extraRepo.Interface)
				if !ok {
					continue
				}
				extraEntityName := strings.TrimSuffix(extraPlan.Binding.ServiceName, "Service")
				extraRepoVar := lowerFirst(extraEntityName) + "Repo"
				if !slices.Contains(data.ExtraRepoVars, extraRepoVar) {
					data.ExtraRepoVars = append(data.ExtraRepoVars, extraRepoVar)
				}
			}
		}

		resources = append(resources, data)
	}
	return resources
}

func (r *Runner) bootstrapRepoResources(plans []resourcePlan) []bootstrapResourceData {
	resources := make([]bootstrapResourceData, 0, len(plans))
	for _, plan := range plans {
		if !plan.Resource.Generate.EffectiveRepoCRUD() {
			continue
		}
		entityName := strings.TrimSuffix(plan.Binding.ServiceName, "Service")
		resources = append(resources, bootstrapResourceData{
			FieldName:       plan.ResourceField,
			RepoVar:         lowerFirst(entityName) + "Repo",
			RepoInterface:   repoInterfaceName(plan),
			RepoConstructor: "New" + repoInterfaceName(plan),
		})
	}
	return resources
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
	data := bootstrapTemplateData{
		templateBase:    r.templateBase(),
		Module:          r.project.Module,
		ServiceName:     r.config.Service,
		AppName:         r.project.Module,
		ServerImport:    r.internalImport("server"),
		DataBootImport:  r.internalImport("data", "bootstrap"),
		RepoImport:      r.internalImport("data", "repo"),
		ServiceImport:   r.internalImport("service"),
		RepoResources:   r.bootstrapRepoResources(plans),
		ServerResources: serverResources,
		HTTPResources: bootstrapRegisteredResources(serverResources, plans, func(flags config.GenerateFlags) bool {
			return flags.EffectiveRestRegister()
		}),
		GRPCResources: bootstrapRegisteredResources(serverResources, plans, func(flags config.GenerateFlags) bool {
			return flags.EffectiveGRPCRegister()
		}),
	}

	if err := r.removeObsoleteGeneratedFile(filepath.Join(r.project.Root, "internal", "data", "bootstrap", "ent_client.go")); err != nil {
		return Result{}, err
	}

	files := []struct {
		path     string
		template string
	}{
		{path: filepath.Join(r.project.Root, "internal", "bootstrap", "generated_servers.gen.go"), template: codegentemplate.BootstrapGeneratedServers},
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
	hasRepo := plan.Resource.Generate.EffectiveRepoCRUD()
	extraRepos := r.extraServiceRepos(plan)
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
		Embeds:          serviceEmbeds(plan.APIPackageAlias, plan.Binding.ServiceName),
		Methods:         methods,
		HasRepo:         hasRepo,
		RepoField:       repoField,
		RepoType:        repoType,
		ExtraRepos:      extraRepos,
		ResourceName:    plan.Resource.Name,
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
	for _, methodConfig := range plan.Resource.ServiceMethods {
		for _, repoConfig := range methodConfig.Repos {
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
	}
	return repos
}

func (r *Runner) serviceConfiguredImports(plan resourcePlan) []importSpec {
	var imports []importSpec
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

func (r *Runner) repoMethodBody(plan resourcePlan, methodName string, params []namedType, responseType string) string {
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
	dtoType := dtoAlias + "." + dtoName
	repoName := strings.TrimSpace(plan.Resource.RepoInterface)
	if repoName == "" {
		repoName = entityName + "Repo"
	}

	imports := []importSpec{
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
	filters := repoFilters(plan.Schema.Fields, plan.Resource.Filters.Allow)
	usesFilters := len(filters) > 0
	usesAuditFields := hasGeneratedAuditFields(plan.Schema.Fields)
	if usesFilters {
		imports = append(imports,
			importSpec{Path: "strconv"},
			importSpec{Alias: "paginationv1", Path: "github.com/chnxq/x-crud/api/gen/pagination/v1"},
		)
	}
	if usesAuditFields {
		imports = append(imports,
			importSpec{Alias: "crudviewer", Path: "github.com/chnxq/x-crud/viewer"},
			importSpec{Path: "time"},
		)
	}
	usedAliases := make(map[string]struct{})
	var methods []repoMethodData
	usesEnumSetter := false
	usesTimeSetter := false
	for _, method := range plan.Binding.Methods {
		if !isCRUDMethod(method.Name) && !isSupportedRepoSpecialMethod(plan.Resource, method.Name) {
			continue
		}
		kind := repoMethodKind(method.Name)
		if !resourceOperationEnabled(plan.Resource, kind) {
			continue
		}
		normalizedParams := normalizeTypeAliases(method.Params, plan.Binding.Imports, dtoImport, dtoAlias)
		normalizedResults := normalizeTypeAliases(method.Results, plan.Binding.Imports, dtoImport, dtoAlias)
		enumHelperName := lowerFirst(entityName) + "EnumPtrFromProto"
		timeHelperName := lowerFirst(entityName) + "TimePtrFromProto"
		setters := repoSetters(plan.Schema.Fields, r.dtoFieldNames(dtoImport, dtoName), method.Name, strings.ToLower(entityName), enumHelperName, timeHelperName)
		auditSetters := repoAuditSetters(plan.Schema.Fields, method.Name)
		usesEnumSetter = usesEnumSetter || settersUseKind(setters, "Enum")
		usesTimeSetter = usesTimeSetter || settersUseKind(setters, "Time")
		methodData := repoMethodData{
			Name:            method.Name,
			Params:          nameParams(normalizedParams),
			ResponseType:    firstResult(normalizedResults),
			Kind:            kind,
			Body:            r.repoMethodBody(plan, method.Name, nameParams(normalizedParams), firstResult(normalizedResults)),
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
			ZeroReturn:      zeroReturn(firstResult(normalizedResults)),
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

	data := repoTemplateData{
		templateBase:     r.templateBase(),
		Imports:          uniqueImports(imports),
		RepoName:         repoName,
		RepoStructName:   lowerFirst(repoName),
		ConstructorName:  "New" + repoName,
		EntityName:       entityName,
		EntOperationName: entOperationName(entityName),
		ResourceName:     plan.Resource.Name,
		EntPackage:       strings.ToLower(entityName),
		PredicateType:    entityName,
		DTOType:          dtoType,
		IDType:           idGoType(plan.Schema.Fields),
		Fields:           plan.Schema.Fields,
		UsesEnumSetter:   usesEnumSetter,
		UsesTimeSetter:   usesTimeSetter,
		EnumHelperName:   lowerFirst(entityName) + "EnumPtrFromProto",
		TimeHelperName:   lowerFirst(entityName) + "TimePtrFromProto",
		Filters:          filters,
		UsesFilters:      usesFilters,
		UsesAuditFields:  usesAuditFields,
	}
	data.Methods = methods

	return renderTemplate(codegentemplate.RepoFile, data)
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
	}
}

func (r *Runner) version() string {
	if strings.TrimSpace(r.options.Version) == "" {
		return "dev"
	}
	return r.options.Version
}

func (r *Runner) writeExtensionFile(path string, content []byte, result *Result) error {
	return r.writeFile(path, content, result, true)
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
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	result.Written = append(result.Written, path)
	return nil
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
	var out []importSpec
	for _, spec := range imports {
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

func serviceEmbeds(alias, serviceName string) []string {
	return []string{
		alias + "." + serviceName + "HTTPServer",
		alias + ".Unimplemented" + serviceName + "Server",
	}
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
					for _, name := range field.Names {
						out[name.Name] = struct{}{}
					}
				}
			}
		}
	}
	return out
}

func repoSetters(fields []entschema.Field, dtoFields map[string]struct{}, methodName, entPackage, enumHelperName, timeHelperName string) []setterData {
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
		if len(dtoFields) > 0 {
			if _, ok := dtoFields[dtoName]; !ok {
				continue
			}
		}
		method := "Set" + entName
		expr := "req.Data." + dtoName
		kind := field.Kind
		if field.Optional {
			method = "SetNillable" + entName
		}
		switch field.Kind {
		case "Enum":
			expr = fmt.Sprintf("%s[%s.%s](req.Data.%s)", enumHelperName, entPackage, entName, dtoName)
		case "Time":
			expr = timeHelperName + "(req.Data." + dtoName + ")"
		}
		if !field.Optional {
			expr = "req.Data.Get" + dtoName + "()"
			switch field.Kind {
			case "Enum":
				expr = fmt.Sprintf("%s.%s(req.Data.Get%s().String())", entPackage, entName, dtoName)
			case "Time":
				expr = "req.Data.Get" + dtoName + "().AsTime()"
			}
		}
		setters = append(setters, setterData{
			Method: method,
			Expr:   expr,
			Kind:   kind,
		})
	}
	return setters
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
	case "String", "Enum", "Time", "Bool", "Uint", "Uint8", "Uint16", "Uint32", "Uint64", "Int", "Int8", "Int16", "Int32", "Int64", "Float", "Float32":
		return true
	default:
		return false
	}
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
		})
		seen[name] = struct{}{}
	}
	return filters
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
