package codegen

import (
	"path/filepath"
	"strings"

	codegentemplate "github.com/chnxq/xkit/internal/codegen/template"
)

type frontendPageTemplateData struct {
	templateBase
	EntityType        string
	ProviderBaseName  string
	MetaBaseName      string
	ListFuncName      string
	GetFuncName       string
	CreateFuncName    string
	UpdateFuncName    string
	DeleteFuncName    string
	GridClassName     string
	PageTitleKey      string
	PageModuleNameKey string
}

func (r *Runner) generateFrontendPageFiles() (Result, error) {
	plans, err := r.plans()
	if err != nil {
		return Result{}, err
	}

	typeScriptRoot, err := resolveTypeScriptRoot(r.project.Root, r.options.TypeScriptRoot)
	if err != nil {
		return Result{}, err
	}
	baseDir := filepath.Join(typeScriptRoot, "apps", "web-antd", "src", "views", "generated", r.templateBase().Frontend)

	var result Result
	for _, plan := range plans {
		if !r.supportsStandardFrontendProvider(plan) {
			continue
		}
		data := r.frontendPageData(plan)
		content, err := renderAnyTemplate(codegentemplate.FrontendCRUDPage, data)
		if err != nil {
			return result, err
		}
		targetPath := filepath.Join(baseDir, filepath.FromSlash(plan.Resource.Frontend.ViewPath+".crud.vue"))
		if err := r.writeGeneratedFile(targetPath, content, &result); err != nil {
			return result, err
		}
	}

	return result, nil
}

func (r *Runner) frontendPageData(plan resourcePlan) frontendPageTemplateData {
	resourceField := plan.ResourceField
	viewPath := strings.TrimSpace(plan.Resource.Frontend.ViewPath)
	pageTitleKey := "menu." + strings.ReplaceAll(viewPath, "/", ".")
	pageModuleNameKey := strings.TrimSpace(plan.Resource.Frontend.I18nPrefix) + ".moduleName"
	if strings.TrimSpace(plan.Resource.Frontend.I18nPrefix) == "" {
		pageModuleNameKey = "page." + lowerFirst(resourceField) + ".moduleName"
	}

	return frontendPageTemplateData{
		templateBase:      r.templateBase(),
		EntityType:        "Admin" + resourceField,
		ProviderBaseName:  filepath.Base(viewPath),
		MetaBaseName:      filepath.Base(viewPath),
		ListFuncName:      "list" + resourceField + "Page",
		GetFuncName:       "get" + resourceField + "ById",
		CreateFuncName:    "create" + resourceField,
		UpdateFuncName:    "update" + resourceField,
		DeleteFuncName:    "delete" + resourceField,
		GridClassName:     "generated-" + frontendSnakeCase(resourceField) + "-grid",
		PageTitleKey:      pageTitleKey,
		PageModuleNameKey: pageModuleNameKey,
	}
}
