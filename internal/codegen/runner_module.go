package codegen

import (
	"fmt"
	"path/filepath"

	"github.com/chnxq/xkit/internal/project"
)

type layout struct {
	InternalRoot        string
	InternalImportRoot  string
	BootstrapRoot       string
	BootstrapImportRoot string
	DataBootstrapRoot   string
	DataBootstrapImport string
	EntRoot             string
	EntImportRoot       string
	ModuleImport        string
	ModuleRoot          string
}

func newProjectLayout(info project.Info) layout {
	return layout{
		InternalRoot:        filepath.Join(info.Root, "internal"),
		InternalImportRoot:  filepath.ToSlash(filepath.Join(info.Module, "internal")),
		BootstrapRoot:       filepath.Join(info.Root, "internal", "bootstrap"),
		BootstrapImportRoot: filepath.ToSlash(filepath.Join(info.Module, "internal", "bootstrap")),
		DataBootstrapRoot:   filepath.Join(info.Root, "internal", "data", "bootstrap"),
		DataBootstrapImport: filepath.ToSlash(filepath.Join(info.Module, "internal", "data", "bootstrap")),
		EntRoot:             filepath.Join(info.Root, "internal", "data", "ent"),
		EntImportRoot:       filepath.ToSlash(filepath.Join(info.Module, "internal", "data", "ent")),
		ModuleImport:        info.Module,
	}
}

func newModuleLayout(info project.Info, moduleName, configuredRoot string) (layout, error) {
	moduleRoot, moduleImport, err := resolveCodegenModuleRoot(info, moduleName, configuredRoot)
	if err != nil {
		return layout{}, err
	}
	return layout{
		InternalRoot:        moduleRoot,
		InternalImportRoot:  moduleImport,
		BootstrapRoot:       filepath.Join(moduleRoot, "bootstrap"),
		BootstrapImportRoot: filepath.ToSlash(filepath.Join(moduleImport, "bootstrap")),
		DataBootstrapRoot:   filepath.Join(moduleRoot, "data", "bootstrap"),
		DataBootstrapImport: filepath.ToSlash(filepath.Join(moduleImport, "data", "bootstrap")),
		EntRoot:             filepath.Join(moduleRoot, "data"),
		EntImportRoot:       filepath.ToSlash(filepath.Join(moduleImport, "data")),
		ModuleImport:        moduleImport,
		ModuleRoot:          moduleRoot,
	}, nil
}

func resolveCodegenModuleRoot(info project.Info, moduleName, configuredRoot string) (string, string, error) {
	if moduleName == "" {
		return "", "", fmt.Errorf("module name is required")
	}

	moduleRoot := configuredRoot
	if moduleRoot == "" {
		moduleRoot = filepath.Join(info.Root, "modules", moduleName)
	} else if !filepath.IsAbs(moduleRoot) {
		moduleRoot = filepath.Join(info.Root, moduleRoot)
	}
	moduleRoot, err := filepath.Abs(moduleRoot)
	if err != nil {
		return "", "", fmt.Errorf("resolve module root: %w", err)
	}
	moduleRoot = filepath.Clean(moduleRoot)

	rel, err := filepath.Rel(info.Root, moduleRoot)
	if err != nil {
		return "", "", fmt.Errorf("resolve module import root: %w", err)
	}
	rel = filepath.ToSlash(filepath.Clean(rel))
	if rel == "." {
		return "", "", fmt.Errorf("module root must not equal project root")
	}

	return moduleRoot, filepath.ToSlash(filepath.Join(info.Module, rel)), nil
}
