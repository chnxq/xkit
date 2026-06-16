package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Service   string          `yaml:"service"`
	Module    string          `yaml:"module"`
	Frontend  *FrontendConfig `yaml:"frontend,omitempty"`
	Resources []Resource      `yaml:"resources"`
}

type FrontendConfig struct {
	OutputRoot string `yaml:"output_root,omitempty"`
}

type Resource struct {
	Name           string                         `yaml:"name"`
	ProtoService   string                         `yaml:"proto_service"`
	Entity         string                         `yaml:"entity,omitempty"`
	DTOImport      string                         `yaml:"dto_import,omitempty"`
	DTOType        string                         `yaml:"dto_type,omitempty"`
	TenantScope    string                         `yaml:"tenant_scope,omitempty"`
	RepoInterface  string                         `yaml:"repo_interface,omitempty"`
	Frontend       *FrontendResourceConfig        `yaml:"frontend,omitempty"`
	Tree           *TreeConfig                    `yaml:"tree,omitempty"`
	Aggregates     []AggregateConfig              `yaml:"aggregates,omitempty"`
	ServiceImports []ImportConfig                 `yaml:"service_imports,omitempty"`
	ServiceRepos   []RepoConfig                   `yaml:"service_repos,omitempty"`
	ServiceFields  []ServiceFieldConfig           `yaml:"service_fields,omitempty"`
	ExistsFields   []string                       `yaml:"exists_fields,omitempty"`
	Filters        FilterConfig                   `yaml:"filters,omitempty"`
	ServiceMethods map[string]ServiceMethodConfig `yaml:"service_methods,omitempty"`
	RepoMethods    map[string]RepoMethodConfig    `yaml:"repo_methods,omitempty"`
	Operations     OperationFlags                 `yaml:"operations,omitempty"`
	Generate       GenerateFlags                  `yaml:"generate,omitempty"`
}

type ServiceMethodConfig struct {
	Imports []ImportConfig `yaml:"imports,omitempty"`
	Repos   []RepoConfig   `yaml:"repos,omitempty"`
	Body    string         `yaml:"body,omitempty"`
}

type ServiceFieldConfig struct {
	Field string `yaml:"field"`
	Type  string `yaml:"type"`
}

type RepoMethodConfig struct {
	Imports []ImportConfig `yaml:"imports,omitempty"`
	Body    string         `yaml:"body,omitempty"`
}

type FrontendResourceConfig struct {
	ViewPath   string                `yaml:"view_path,omitempty"`
	I18nPrefix string                `yaml:"i18n_prefix,omitempty"`
	List       *FrontendListConfig   `yaml:"list,omitempty"`
	Form       *FrontendDialogConfig `yaml:"form,omitempty"`
}

type FrontendListConfig struct {
	Columns []FrontendColumn  `yaml:"columns,omitempty"`
	Filters map[string]string `yaml:"filters,omitempty"`
}

type FrontendDialogConfig struct {
	Enabled *bool            `yaml:"enabled,omitempty"`
	Fields  []FrontendColumn `yaml:"fields,omitempty"`
}

type FrontendColumn struct {
	Field string
	EN    string
	CN    string
}

func (c *FrontendColumn) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		c.Field = strings.TrimSpace(value.Value)
		return nil
	case yaml.SequenceNode:
		if len(value.Content) == 0 {
			return nil
		}
		if len(value.Content) > 0 {
			c.Field = strings.TrimSpace(value.Content[0].Value)
		}
		if len(value.Content) > 1 {
			c.EN = strings.TrimSpace(value.Content[1].Value)
		}
		if len(value.Content) > 2 {
			c.CN = strings.TrimSpace(value.Content[2].Value)
		}
		return nil
	default:
		return fmt.Errorf("frontend column must be a string or sequence")
	}
}

type TreeConfig struct {
	ParentField   string `yaml:"parent_field,omitempty"`
	PathField     string `yaml:"path_field,omitempty"`
	ChildrenField string `yaml:"children_field,omitempty"`
	ListMethod    string `yaml:"list_method,omitempty"`
}

type AggregateConfig struct {
	Name            string `yaml:"name"`
	Resource        string `yaml:"resource,omitempty"`
	RepoInterface   string `yaml:"repo_interface,omitempty"`
	ForeignKey      string `yaml:"foreign_key,omitempty"`
	ParentField     string `yaml:"parent_field,omitempty"`
	CollectionField string `yaml:"collection_field,omitempty"`
	CurrentField    string `yaml:"current_field,omitempty"`
	Primary         bool   `yaml:"primary,omitempty"`
}

type ImportConfig struct {
	Alias string `yaml:"alias,omitempty"`
	Path  string `yaml:"path"`
}

type RepoConfig struct {
	Field     string `yaml:"field"`
	Interface string `yaml:"interface"`
}

type FilterConfig struct {
	Allow []string `yaml:"allow,omitempty"`
}

type OperationFlags map[string]bool

type GenerateFlags struct {
	ServiceStub  bool `yaml:"service_stub,omitempty"`
	RepoCRUD     bool `yaml:"repo_crud,omitempty"`
	RestRegister bool `yaml:"rest_register,omitempty"`
	GRPCRegister bool `yaml:"grpc_register,omitempty"`
	Wire         bool `yaml:"wire,omitempty"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) Validate() error {
	if strings.TrimSpace(cfg.Service) == "" {
		return fmt.Errorf("config service is required")
	}
	if len(cfg.Resources) == 0 {
		return fmt.Errorf("config must include at least one resource")
	}

	for _, resource := range cfg.Resources {
		if strings.TrimSpace(resource.Name) == "" {
			return fmt.Errorf("resource name is required")
		}
		if strings.TrimSpace(resource.ProtoService) == "" {
			return fmt.Errorf("resource %q is missing proto_service", resource.Name)
		}
	}

	return nil
}

func (g GenerateFlags) EffectiveServiceStub() bool {
	return g.ServiceStub || (!g.ServiceStub && !g.RepoCRUD && !g.RestRegister && !g.GRPCRegister && !g.Wire)
}

func (g GenerateFlags) EffectiveRepoCRUD() bool {
	return g.RepoCRUD || (!g.ServiceStub && !g.RepoCRUD && !g.RestRegister && !g.GRPCRegister && !g.Wire)
}

func (g GenerateFlags) EffectiveRestRegister() bool {
	return g.RestRegister || (!g.ServiceStub && !g.RepoCRUD && !g.RestRegister && !g.GRPCRegister && !g.Wire)
}

func (g GenerateFlags) EffectiveGRPCRegister() bool {
	return g.GRPCRegister || (!g.ServiceStub && !g.RepoCRUD && !g.RestRegister && !g.GRPCRegister && !g.Wire)
}

func (g GenerateFlags) EffectiveWire() bool {
	return g.Wire || (!g.ServiceStub && !g.RepoCRUD && !g.RestRegister && !g.GRPCRegister && !g.Wire)
}
