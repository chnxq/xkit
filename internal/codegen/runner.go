package codegen

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/chnxq/xkit/internal/binding"
	codegentemplate "github.com/chnxq/xkit/internal/codegen/template"
	"github.com/chnxq/xkit/internal/config"
	"github.com/chnxq/xkit/internal/entschema"
	"github.com/chnxq/xkit/internal/project"
	xproto "github.com/chnxq/xkit/internal/proto"
)

type Options struct {
	DryRun         bool
	Version        string
	TypeScriptRoot string
	ModuleName     string
	ModuleRoot     string
}

type generatedMeta struct {
	Version     string
	GeneratedAt string
}

type templateBase struct {
	Generated generatedMeta
	Module    string
	Frontend  string
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
	layout       layout
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
	EntImport         string
	EntMigrateImport  string
	EntRuntimeImport  string
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

type i18nEntry struct {
	Key   string
	Value string
}

var aliasPattern = regexp.MustCompile(`\b([A-Za-z_][A-Za-z0-9_]*)\.`)

func NewProjectRunner(info project.Info, cfg config.Config, options Options) (*Runner, error) {
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
		layout:       newProjectLayout(info),
	}, nil
}

func New(info project.Info, cfg config.Config, options Options) (*Runner, error) {
	return NewProjectRunner(info, cfg, options)
}

func NewModuleRunner(info project.Info, cfg config.Config, options Options) (*Runner, error) {
	moduleName := strings.TrimSpace(options.ModuleName)
	if moduleName == "" {
		return nil, fmt.Errorf("module name is required")
	}

	layout, err := newModuleLayout(info, moduleName, options.ModuleRoot)
	if err != nil {
		return nil, err
	}
	if cfg.Module != "" && cfg.Module != layout.ModuleImport {
		return nil, fmt.Errorf("config module %q does not match target module %q", cfg.Module, layout.ModuleImport)
	}

	protoIndex, err := xproto.LoadServicesDir(filepath.Join(layout.ModuleRoot, "api", "protos"))
	if err != nil {
		return nil, err
	}

	bindingIndex, err := binding.Load(layout.ModuleRoot, layout.ModuleImport)
	if err != nil {
		return nil, err
	}

	schemaIndex, err := entschema.LoadDir(filepath.Join(layout.ModuleRoot, "data", "schema"))
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
		layout:       layout,
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
	case "module-entry":
		return r.generateModuleEntryFile()
	case "module":
		return r.generateModule()
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

func (r *Runner) generateModule() (Result, error) {
	var result Result
	parts := []string{"service", "repo", "register", "bootstrap", "module-entry", "frontend-meta"}
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
