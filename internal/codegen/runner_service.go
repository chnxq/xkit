package codegen

import (
	"path/filepath"
	"slices"
	"strings"

	codegentemplate "github.com/chnxq/xkit/internal/codegen/template"
	xproto "github.com/chnxq/xkit/internal/proto"
)

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
	NeedsIdentity   bool
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

func (r *Runner) generateServiceFiles() (Result, error) {
	plans, err := r.plans()
	if err != nil {
		return Result{}, err
	}

	var result Result
	sharedContent, err := renderTemplate(codegentemplate.ServiceSharedExt, struct {
		templateBase
		NeedsIdentity bool
	}{
		templateBase:  r.templateBase(),
		NeedsIdentity: serviceNeedsIdentity(plans),
	})
	if err != nil {
		return result, err
	}
	sharedPath := filepath.Join(r.internalDir("service"), "service_shared_ext.go")
	if err := r.writeGeneratedFile(sharedPath, sharedContent, &result); err != nil {
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

func serviceNeedsIdentity(plans []resourcePlan) bool {
	for _, plan := range plans {
		if strings.TrimSpace(plan.Resource.TenantScope) == "tenant_scoped" {
			return true
		}
	}
	return false
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
		normalizedParams := normalizeTypeAliases(method.Params, plan.Binding.Imports, plan.Binding.ImportPath, plan.APIPackageAlias)
		normalizedResults := normalizeTypeAliases(method.Results, plan.Binding.Imports, plan.Binding.ImportPath, plan.APIPackageAlias)
		for _, typeText := range append(slices.Clone(normalizedParams), normalizedResults...) {
			for _, alias := range aliasesInType(typeText) {
				usedAliases[alias] = struct{}{}
			}
		}

		kind := repoMethodKind(method.Name)
		delegate := hasRepo && isCRUDMethod(method.Name) && resourceOperationEnabled(plan.Resource, kind)
		classification := lookupClassification(plan.Proto.Methods, method.Name)
		responseType := firstResult(normalizedResults)
		params := nameParams(normalizedParams)
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
		NeedsIdentity:   strings.TrimSpace(plan.Resource.TenantScope) == "tenant_scoped",
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

func lookupClassification(methods []xproto.Method, name string) string {
	for _, method := range methods {
		if method.Name == name {
			return method.Classification
		}
	}
	return "special"
}
