package codegen

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

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
	pathParts := append([]string{r.layout.InternalRoot}, parts...)
	return filepath.Join(pathParts...)
}

func (r *Runner) internalImport(parts ...string) string {
	pathParts := append([]string{r.layout.InternalImportRoot}, parts...)
	return filepath.ToSlash(filepath.Join(pathParts...))
}

func (r *Runner) templateBase() templateBase {
	frontend := r.config.Service
	if strings.TrimSpace(r.options.ModuleName) != "" {
		frontend = strings.TrimSpace(r.options.ModuleName)
	}
	if strings.TrimSpace(frontend) == "" {
		frontend = "admin"
	}
	return templateBase{
		Generated: generatedMeta{
			Version:     r.version(),
			GeneratedAt: time.Now().Format("2006-01-02 15:04:05 MST"),
		},
		Module:   r.layout.ModuleImport,
		Project:  r.project.Module,
		Frontend: frontend,
		Shared:   r.sharedModuleImport(),
	}
}

func (r *Runner) frontendModuleRoot() string {
	typeScriptRoot, err := resolveTypeScriptRoot(r.project.Root, r.options.TypeScriptRoot)
	if err != nil || !r.isModuleMode() {
		return ""
	}
	return filepath.Join(
		typeScriptRoot,
		"apps",
		"web-antd",
		"src",
		"modules",
		r.frontendModuleDirName(),
	)
}

func (r *Runner) frontendGeneratedMetaRoot() string {
	typeScriptRoot, err := resolveTypeScriptRoot(r.project.Root, r.options.TypeScriptRoot)
	if err != nil {
		return ""
	}
	if r.isModuleMode() {
		return filepath.Join(r.frontendModuleRoot(), "meta")
	}
	return filepath.Join(typeScriptRoot, "apps", "web-antd", "src", "views", "generated", r.templateBase().Frontend)
}

func (r *Runner) frontendGeneratedProviderRoot() string {
	typeScriptRoot, err := resolveTypeScriptRoot(r.project.Root, r.options.TypeScriptRoot)
	if err != nil {
		return ""
	}
	if r.isModuleMode() {
		return filepath.Join(r.frontendModuleRoot(), "provider")
	}
	return filepath.Join(typeScriptRoot, "apps", "web-antd", "src", "views", "generated", r.templateBase().Frontend)
}

func (r *Runner) frontendGeneratedPageRoot() string {
	typeScriptRoot, err := resolveTypeScriptRoot(r.project.Root, r.options.TypeScriptRoot)
	if err != nil {
		return ""
	}
	if r.isModuleMode() {
		return filepath.Join(r.frontendModuleRoot(), "views")
	}
	return filepath.Join(typeScriptRoot, "apps", "web-antd", "src", "views", "generated", r.templateBase().Frontend)
}

func (r *Runner) frontendGeneratedLangRoot() string {
	typeScriptRoot, err := resolveTypeScriptRoot(r.project.Root, r.options.TypeScriptRoot)
	if err != nil {
		return ""
	}
	if r.isModuleMode() {
		return filepath.Join(r.frontendModuleRoot(), "langs")
	}
	return filepath.Join(typeScriptRoot, "apps", "web-antd", "src", "views", "generated", r.templateBase().Frontend, "langs")
}

func (r *Runner) frontendGeneratedAPIImportPath() string {
	if r.isModuleMode() {
		moduleName := strings.TrimSpace(r.options.ModuleName)
		return "#/modules/" + r.frontendModuleDirName() + "/api/generated/" + moduleName + "/service/v1"
	}
	return "#/api/generated/" + r.templateBase().Frontend + "/service/v1"
}

func (r *Runner) frontendLegacyGeneratedRoot() string {
	typeScriptRoot, err := resolveTypeScriptRoot(r.project.Root, r.options.TypeScriptRoot)
	if err != nil || !r.isModuleMode() {
		return ""
	}
	return filepath.Join(typeScriptRoot, "apps", "web-antd", "src", "views", "generated", strings.TrimSpace(r.options.ModuleName))
}

func (r *Runner) frontendLegacyHostViewRoot() string {
	typeScriptRoot, err := resolveTypeScriptRoot(r.project.Root, r.options.TypeScriptRoot)
	if err != nil || !r.isModuleMode() {
		return ""
	}
	return filepath.Join(typeScriptRoot, "apps", "web-antd", "src", "views", strings.TrimSpace(r.options.ModuleName))
}

func (r *Runner) frontendLegacyGeneratedAPIRoot() string {
	typeScriptRoot, err := resolveTypeScriptRoot(r.project.Root, r.options.TypeScriptRoot)
	if err != nil || !r.isModuleMode() {
		return ""
	}
	return filepath.Join(typeScriptRoot, "apps", "web-antd", "src", "api", "generated", strings.TrimSpace(r.options.ModuleName))
}

func (r *Runner) frontendModuleDirName() string {
	moduleName := strings.TrimSpace(r.options.ModuleName)
	if moduleName == "" {
		return ""
	}
	if strings.HasSuffix(moduleName, "-ui") {
		return moduleName
	}
	return moduleName + "-ui"
}
