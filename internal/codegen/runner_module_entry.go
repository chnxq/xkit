package codegen

import (
	"path/filepath"
	"strings"

	codegentemplate "github.com/chnxq/xkit/internal/codegen/template"
)

type moduleEntryTemplateData struct {
	templateBase
	PackageName string
	ModuleName  string
	Imports     []importSpec
}

func (r *Runner) generateModuleEntryFile() (Result, error) {
	if !r.isModuleMode() {
		return Result{}, nil
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
