package codegen

import (
	"path/filepath"
	"slices"
	"strings"

	codegentemplate "github.com/chnxq/xkit/internal/codegen/template"
	"github.com/chnxq/xkit/internal/config"
)

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
		Module:            r.layout.ModuleImport,
		ServiceName:       r.config.Service,
		AppName:           r.project.Module,
		ServerImport:      r.internalImport("server"),
		DataBootImport:    r.internalImport("data", "bootstrap"),
		RepoImport:        r.internalImport("data", "repo"),
		ServiceImport:     r.internalImport("service"),
		EntImport:         r.layout.EntImportRoot,
		EntMigrateImport:  filepath.ToSlash(filepath.Join(r.layout.EntImportRoot, "migrate")),
		EntRuntimeImport:  filepath.ToSlash(filepath.Join(r.layout.EntImportRoot, "runtime")),
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

	if err := r.removeObsoleteGeneratedFile(filepath.Join(r.layout.DataBootstrapRoot, "ent_client.go")); err != nil {
		return Result{}, err
	}
	files := []struct {
		path     string
		template string
	}{
		{path: filepath.Join(r.layout.BootstrapRoot, "generated_servers.gen.go"), template: codegentemplate.BootstrapGeneratedServers},
		{path: filepath.Join(r.layout.BootstrapRoot, "generated_data_providers.gen.go"), template: codegentemplate.BootstrapGeneratedDataProviders},
		{path: filepath.Join(r.layout.DataBootstrapRoot, "ent_client.gen.go"), template: codegentemplate.BootstrapEntClient},
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
	hooksPath := filepath.Join(r.layout.BootstrapRoot, "generated_hooks_ext.go")
	if err := r.writeExtensionFile(hooksPath, hooksContent, &result); err != nil {
		return result, err
	}

	if r.isModuleMode() {
		if err := r.generateModuleResourcesFile(&result); err != nil {
			return result, err
		}
		moduleRuntimeContent, err := renderTemplate(codegentemplate.ModuleRuntimeExt, r.templateBase())
		if err != nil {
			return result, err
		}
		moduleRuntimePath := filepath.Join(r.layout.BootstrapRoot, "module_runtime_ext.go")
		if err := r.writeFile(moduleRuntimePath, moduleRuntimeContent, &result, true); err != nil {
			return result, err
		}
	}

	entHooksContent, err := renderAnyTemplate(codegentemplate.BootstrapEntClientExt, data)
	if err != nil {
		return result, err
	}
	entHooksPath := filepath.Join(r.layout.DataBootstrapRoot, "ent_client_ext.go")
	if err := r.writeExtensionFile(entHooksPath, entHooksContent, &result); err != nil {
		return result, err
	}

	return result, nil
}
