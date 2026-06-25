package codegen

import (
	"path/filepath"
	"strings"

	codegentemplate "github.com/chnxq/xkit/internal/codegen/template"
	"github.com/chnxq/xkit/internal/config"
)

func (r *Runner) moduleResourcesData() moduleResourcesTemplateData {
	menus := make([]moduleMenuResourceData, 0)
	if r.config.HostModule != nil && r.config.HostModule.Resources != nil {
		for _, item := range r.config.HostModule.Resources.Menus {
			menus = append(menus, buildModuleMenuResourceData(item))
		}
	}
	return moduleResourcesTemplateData{
		templateBase:     r.templateBase(),
		ModuleHostImport: filepath.ToSlash(filepath.Join(r.project.Module, "shared", "modulehost")),
		HasMenus:         len(menus) > 0,
		Menus:            menus,
	}
}

func buildModuleMenuResourceData(cfg config.HostModuleMenuConfig) moduleMenuResourceData {
	children := make([]moduleMenuResourceData, 0, len(cfg.Children))
	for _, child := range cfg.Children {
		children = append(children, buildModuleMenuResourceData(child))
	}
	title := strings.TrimSpace(cfg.Meta.TitleKey)
	if title == "" {
		title = strings.TrimSpace(cfg.Meta.Title)
	}
	return moduleMenuResourceData{
		Name:      strings.TrimSpace(cfg.Name),
		Path:      strings.TrimSpace(cfg.Path),
		Component: strings.TrimSpace(cfg.Component),
		Redirect:  strings.TrimSpace(cfg.Redirect),
		Type:      normalizeModuleMenuType(cfg.Type),
		Meta: moduleMenuMetaData{
			Authority:       trimStringSlice(cfg.Meta.Authority),
			HasAuthority:    len(trimStringSlice(cfg.Meta.Authority)) > 0,
			Icon:            strings.TrimSpace(cfg.Meta.Icon),
			HasIcon:         strings.TrimSpace(cfg.Meta.Icon) != "",
			Link:            strings.TrimSpace(cfg.Meta.Link),
			HasLink:         strings.TrimSpace(cfg.Meta.Link) != "",
			OpenInNewWindow: cfg.Meta.OpenInNewWindow,
			HasOpenInNew:    cfg.Meta.OpenInNewWindow != nil,
			Title:           title,
			HasTitle:        title != "",
		},
		Children: children,
	}
}

func normalizeModuleMenuType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "catalog":
		return "Catalog"
	case "menu":
		return "Menu"
	case "embedded":
		return "Embedded"
	case "link":
		return "Link"
	case "button":
		return "Button"
	default:
		return ""
	}
}

func trimStringSlice(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		out = append(out, value)
	}
	return out
}

func (r *Runner) generateModuleResourcesFile(result *Result) error {
	if !r.isModuleMode() {
		return nil
	}
	content, err := renderAnyTemplate(codegentemplate.ModuleGeneratedResources, r.moduleResourcesData())
	if err != nil {
		return err
	}
	path := filepath.Join(r.layout.BootstrapRoot, "generated_module_resources.gen.go")
	return r.writeGeneratedFile(path, content, result)
}
