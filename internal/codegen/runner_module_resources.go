package codegen

import (
	"encoding/json"
	"os"
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
	titleAux := strings.TrimSpace(cfg.Meta.TitleAux)
	if titleAux == "" {
		titleAux = strings.TrimSpace(cfg.Meta.Title)
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
			TitleAux:        titleAux,
			HasTitleAux:     titleAux != "",
		},
		Children: children,
	}
}

func (r *Runner) enrichModuleMenuTitleAux() {
	if r == nil || r.config.HostModule == nil || r.config.HostModule.Resources == nil {
		return
	}
	langPath := filepath.Join(r.project.Root, "langs", "zh-CN", "menu.json")
	content, err := os.ReadFile(langPath)
	if err != nil {
		return
	}
	var payload map[string]any
	if err := json.Unmarshal(content, &payload); err != nil {
		return
	}
	for index := range r.config.HostModule.Resources.Menus {
		enrichHostModuleMenuTitleAux(&r.config.HostModule.Resources.Menus[index], payload)
	}
}

func enrichHostModuleMenuTitleAux(menu *config.HostModuleMenuConfig, payload map[string]any) {
	if menu == nil {
		return
	}
	if strings.TrimSpace(menu.Meta.TitleAux) == "" {
		menu.Meta.TitleAux = lookupMenuTitleAux(payload, strings.TrimSpace(menu.Meta.TitleKey))
	}
	for index := range menu.Children {
		enrichHostModuleMenuTitleAux(&menu.Children[index], payload)
	}
}

func lookupMenuTitleAux(payload map[string]any, key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	parts := strings.Split(key, ".")
	var current any = payload
	for _, part := range parts {
		node, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current, ok = node[part]
		if !ok {
			return ""
		}
	}
	value, _ := current.(string)
	return strings.TrimSpace(value)
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
	r.enrichModuleMenuTitleAux()
	content, err := renderAnyTemplate(codegentemplate.ModuleGeneratedResources, r.moduleResourcesData())
	if err != nil {
		return err
	}
	path := filepath.Join(r.layout.BootstrapRoot, "generated_module_resources.gen.go")
	return r.writeGeneratedFile(path, content, result)
}
