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
	return templateBase{
		Generated: generatedMeta{
			Version:     r.version(),
			GeneratedAt: time.Now().Format("2006-01-02 15:04:05 MST"),
		},
		Module: r.layout.ModuleImport,
	}
}
