package codegen

import (
	"path/filepath"
	"strings"

	codegentemplate "github.com/chnxq/xkit/internal/codegen/template"
	"github.com/chnxq/xkit/internal/config"
)

type moduleEntryTemplateData struct {
	templateBase
	PackageName   string
	ModuleName    string
	Imports       []importSpec
	HTTPResources []bootstrapResourceData
	GRPCResources []bootstrapResourceData
}

func (r *Runner) generateModuleEntryFile() (Result, error) {
	if !r.isModuleMode() {
		return Result{}, nil
	}
	plans, err := r.plans()
	if err != nil {
		return Result{}, err
	}

	data := moduleEntryTemplateData{
		templateBase: r.templateBase(),
		PackageName:  filepath.Base(r.layout.ModuleRoot),
		ModuleName:   strings.TrimSpace(r.options.ModuleName),
		Imports: []importSpec{
			{Path: "github.com/chnxq/xkitpkg/app"},
			{Alias: "httptransport", Path: "github.com/chnxq/xkitpkg/transport/http"},
			{Path: "google.golang.org/grpc"},
			{Alias: "bootstrap", Path: r.layout.BootstrapImportRoot},
			{Alias: "server", Path: r.internalImport("server")},
		},
		HTTPResources: bootstrapRegisteredResources(r.bootstrapResources(plans), plans, func(flags config.GenerateFlags) bool {
			return flags.EffectiveRestRegister()
		}),
		GRPCResources: bootstrapRegisteredResources(r.bootstrapResources(plans), plans, func(flags config.GenerateFlags) bool {
			return flags.EffectiveGRPCRegister()
		}),
	}

	content, err := renderAnyTemplate(codegentemplate.ModuleEntry, data)
	if err != nil {
		return Result{}, err
	}

	var result Result
	path := filepath.Join(r.layout.ModuleRoot, "module.go")
	if err := r.writeGeneratedFile(path, content, &result); err != nil {
		return result, err
	}
	return result, nil
}
