package codegen

import (
	"path/filepath"
	"strings"

	codegentemplate "github.com/chnxq/xkit/internal/codegen/template"
)

type frontendProviderTemplateData struct {
	templateBase
	GeneratedAPIImportPath string
	CreateClientFunc       string
	EntityType             string
	ListResponseType       string
	GetResponseType        string
	CreateRequestType      string
	UpdateRequestType      string
	FrontendEntityType     string
	FrontendSaveInputType  string
	FrontendListParamsType string
	FrontendListResultType string
	ListFuncName           string
	GetFuncName            string
	CreateFuncName         string
	UpdateFuncName         string
	DeleteFuncName         string
	UpdateMask             string
	HasGet                 bool
	HasCreate              bool
	HasUpdate              bool
	HasDelete              bool
	FilterFields           []frontendProviderFilterField
}

type frontendProviderFilterField struct {
	FieldName    string
	SchemaField  string
	Type         string
	Operator     string
	UseCleanText bool
}

func (r *Runner) generateFrontendProviderFiles() (Result, error) {
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

		data := r.frontendProviderData(plan)
		content, err := renderAnyTemplate(codegentemplate.FrontendProvider, data)
		if err != nil {
			return result, err
		}

		targetPath := filepath.Join(baseDir, filepath.FromSlash(plan.Resource.Frontend.ViewPath+".provider.ts"))
		if err := r.writeGeneratedFile(targetPath, content, &result); err != nil {
			return result, err
		}
	}

	return result, nil
}

func (r *Runner) supportsStandardFrontendProvider(plan resourcePlan) bool {
	if plan.Resource.Frontend == nil || strings.TrimSpace(plan.Resource.Frontend.ViewPath) == "" {
		return false
	}
	if !plan.Resource.Generate.EffectiveServiceStub() {
		return false
	}
	ops := plan.Resource.Operations
	return ops["list"] && ops["get"] && ops["create"] && ops["update"] && ops["delete"]
}

func (r *Runner) frontendProviderData(plan resourcePlan) frontendProviderTemplateData {
	entityType := plan.DTOTypeOrDefault()
	serviceName := strings.TrimSpace(plan.Binding.ServiceName)
	resourceField := plan.ResourceField

	return frontendProviderTemplateData{
		templateBase:           r.templateBase(),
		GeneratedAPIImportPath: "#/api/generated/" + r.templateBase().Frontend + "/service/v1",
		CreateClientFunc:       "create" + serviceName + "Client",
		EntityType:             entityType,
		ListResponseType:       serviceName + "ListResponse",
		GetResponseType:        serviceName + "GetResponse",
		CreateRequestType:      serviceName + "CreateRequest",
		UpdateRequestType:      serviceName + "UpdateRequest",
		FrontendEntityType:     "Admin" + resourceField,
		FrontendSaveInputType:  "Admin" + resourceField + "SaveInput",
		FrontendListParamsType: "Admin" + resourceField + "ListParams",
		FrontendListResultType: "Admin" + resourceField + "ListResult",
		ListFuncName:           "list" + resourceField + "Page",
		GetFuncName:            "get" + resourceField + "ById",
		CreateFuncName:         "create" + resourceField,
		UpdateFuncName:         "update" + resourceField,
		DeleteFuncName:         "delete" + resourceField,
		UpdateMask:             r.frontendProviderUpdateMask(plan),
		HasGet:                 true,
		HasCreate:              true,
		HasUpdate:              true,
		HasDelete:              true,
		FilterFields:           r.frontendProviderFilterFields(plan),
	}
}

func (plan resourcePlan) DTOTypeOrDefault() string {
	if strings.TrimSpace(plan.Resource.DTOType) != "" {
		return strings.TrimSpace(plan.Resource.DTOType)
	}
	return strings.TrimSpace(plan.Resource.Entity)
}

func (r *Runner) frontendProviderUpdateMask(plan resourcePlan) string {
	fields := make([]string, 0, len(plan.Schema.Fields))
	for _, field := range plan.Schema.Fields {
		name := strings.TrimSpace(field.Name)
		if name == "" || name == "created_at" || name == "updated_at" || name == "deleted_at" {
			continue
		}
		fields = append(fields, name)
	}
	return strings.Join(fields, ",")
}

func (r *Runner) frontendProviderFilterFields(plan resourcePlan) []frontendProviderFilterField {
	if plan.Resource.Frontend == nil || plan.Resource.Frontend.List == nil {
		return nil
	}
	filters := normalizedFrontendFilters(plan.Resource.Frontend.List.Filters)
	items := make([]frontendProviderFilterField, 0, len(filters))
	for _, filter := range filters {
		fieldName := simpleFieldName(filter.Field)
		items = append(items, frontendProviderFilterField{
			FieldName:    fieldName,
			SchemaField:  frontendSnakeCase(fieldName),
			Type:         frontendProviderFilterType(filter.Component),
			Operator:     frontendProviderFilterOperator(filter.Component),
			UseCleanText: frontendProviderNeedsCleanText(filter.Component),
		})
	}
	return items
}

func frontendProviderFilterType(component string) string {
	switch strings.TrimSpace(component) {
	case "InputNumber":
		return "number"
	default:
		return "string"
	}
}

func frontendProviderFilterOperator(component string) string {
	switch strings.TrimSpace(component) {
	case "InputNumber":
		return "EQ"
	default:
		return "CONTAINS"
	}
}

func frontendProviderNeedsCleanText(component string) bool {
	switch strings.TrimSpace(component) {
	case "Input":
		return true
	default:
		return false
	}
}
