package codegen

import (
	"fmt"
	"path/filepath"
	"strings"

	codegentemplate "github.com/chnxq/xkit/internal/codegen/template"
	"github.com/chnxq/xkit/internal/config"
)

type frontendProviderTemplateData struct {
	templateBase
	GeneratedAPIImportPath string
	CreateClientFunc       string
	ListPath              string
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
	RelationOptions        []frontendRelationOptionData
	ExtraTypeImports       []string
	ExtraClientFuncs       []string
}

type frontendProviderFilterField struct {
	FieldName    string
	SchemaField  string
	Type         string
	Operator     string
	UseCleanText bool
}

type frontendRelationOptionData struct {
	FuncName           string
	ResultType         string
	ItemType           string
	RelatedClientFunc  string
	RelatedListResult  string
	LabelField         string
	ValueField         string
}

func (r *Runner) generateFrontendProviderFiles() (Result, error) {
	plans, err := r.plans()
	if err != nil {
		return Result{}, err
	}

	baseDir := r.frontendGeneratedProviderRoot()
	if baseDir == "" {
		return Result{}, fmt.Errorf("resolve frontend provider root failed")
	}

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

		providerRelPath := filepath.FromSlash(plan.Resource.Frontend.ViewPath + ".provider.ts")
		if r.isModuleMode() {
			providerRelPath = filepath.Base(filepath.FromSlash(plan.Resource.Frontend.ViewPath)) + ".provider.ts"
		}
		targetPath := filepath.Join(baseDir, providerRelPath)
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

	relationOptions := r.frontendProviderRelationOptions(plan)
	extraTypeImports, extraClientFuncs := frontendProviderExtraImports(relationOptions)

	return frontendProviderTemplateData{
		templateBase:           r.templateBase(),
		GeneratedAPIImportPath: r.frontendGeneratedAPIImportPath(),
		CreateClientFunc:       "create" + serviceName + "Client",
		ListPath:               r.frontendProviderListPath(plan),
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
		RelationOptions:        relationOptions,
		ExtraTypeImports:       extraTypeImports,
		ExtraClientFuncs:       extraClientFuncs,
	}
}

func (r *Runner) frontendProviderListPath(plan resourcePlan) string {
	serviceName := strings.TrimSpace(plan.Binding.ServiceName)
	switch serviceName {
	case "DeviceService":
		return "/xdev/v1/devices"
	case "DeviceModelService":
		return "/xdev/v1/device-models"
	case "DeviceModelTypeService":
		return "/xdev/v1/device-model-types"
	default:
		resourceName := strings.TrimSpace(plan.Resource.Name)
		resourceName = strings.ReplaceAll(resourceName, "_", "-")
		if resourceName == "" {
			resourceName = strings.ToLower(strings.TrimSuffix(serviceName, "Service"))
		}
		return "/" + strings.Trim(r.templateBase().Frontend, "/") + "/v1/" + resourceName + "s"
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
		if name == "" {
			continue
		}
		if field.Immutable || skipGeneratedSetter(name) {
			continue
		}
		switch name {
		case "id", "tenant_id":
			continue
		}
		fields = append(fields, lowerFirst(toPascal(name)))
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

func (r *Runner) frontendProviderRelationOptions(plan resourcePlan) []frontendRelationOptionData {
	resourceNames := map[string]struct{}{}
	var items []frontendRelationOptionData

	appendRelation := func(spec *config.FrontendRelationSpec) {
		if spec == nil {
			return
		}
		resourceName := strings.TrimSpace(spec.Resource)
		if resourceName == "" {
			return
		}
		if _, ok := resourceNames[resourceName]; ok {
			return
		}
		resourceNames[resourceName] = struct{}{}

		relatedPlan, ok := r.findPlanByResourceName(resourceName)
		if !ok {
			return
		}
		relatedServiceName := strings.TrimSpace(relatedPlan.Binding.ServiceName)
		relatedEntityType := relatedPlan.DTOTypeOrDefault()
		resourceField := relatedPlan.ResourceField
		items = append(items, frontendRelationOptionData{
			FuncName:          "list" + resourceField + "Options",
			ResultType:        "Admin" + resourceField + "Option",
			ItemType:          relatedEntityType,
			RelatedClientFunc: "create" + relatedServiceName + "Client",
			RelatedListResult: relatedServiceName + "ListResponse",
			LabelField:        strings.TrimSpace(spec.LabelField),
			ValueField:        strings.TrimSpace(spec.ValueField),
		})
	}

	if plan.Resource.Frontend != nil {
		if plan.Resource.Frontend.List != nil {
			for _, column := range plan.Resource.Frontend.List.Columns {
				appendRelation(column.Relation)
			}
		}
		if plan.Resource.Frontend.Form != nil {
			for _, field := range plan.Resource.Frontend.Form.Fields {
				appendRelation(field.Relation)
			}
		}
	}

	return items
}

func (r *Runner) findPlanByResourceName(name string) (resourcePlan, bool) {
	plans, err := r.plans()
	if err != nil {
		return resourcePlan{}, false
	}
	for _, plan := range plans {
		if strings.TrimSpace(plan.Resource.Name) == name {
			return plan, true
		}
	}
	return resourcePlan{}, false
}

func frontendProviderExtraImports(items []frontendRelationOptionData) ([]string, []string) {
	typeSeen := map[string]struct{}{}
	funcSeen := map[string]struct{}{}
	var typeImports []string
	var clientFuncs []string
	for _, item := range items {
		if _, ok := typeSeen[item.ItemType]; !ok {
			typeSeen[item.ItemType] = struct{}{}
			typeImports = append(typeImports, item.ItemType)
		}
		if _, ok := typeSeen[item.RelatedListResult]; !ok {
			typeSeen[item.RelatedListResult] = struct{}{}
			typeImports = append(typeImports, item.RelatedListResult)
		}
		if _, ok := funcSeen[item.RelatedClientFunc]; !ok {
			funcSeen[item.RelatedClientFunc] = struct{}{}
			clientFuncs = append(clientFuncs, item.RelatedClientFunc)
		}
	}
	return typeImports, clientFuncs
}
