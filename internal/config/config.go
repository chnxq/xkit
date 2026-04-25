package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Service   string     `yaml:"service"`
	Module    string     `yaml:"module"`
	Resources []Resource `yaml:"resources"`
}

type Resource struct {
	Name          string         `yaml:"name"`
	ProtoService  string         `yaml:"proto_service"`
	Entity        string         `yaml:"entity"`
	DTOImport     string         `yaml:"dto_import"`
	DTOType       string         `yaml:"dto_type"`
	RepoInterface string         `yaml:"repo_interface"`
	ExistsFields  []string       `yaml:"exists_fields"`
	Filters       FilterConfig   `yaml:"filters"`
	Operations    OperationFlags `yaml:"operations"`
	Generate      GenerateFlags  `yaml:"generate"`
}

type FilterConfig struct {
	Allow []string `yaml:"allow"`
}

type OperationFlags map[string]bool

type GenerateFlags struct {
	ServiceStub  bool `yaml:"service_stub"`
	RepoCRUD     bool `yaml:"repo_crud"`
	RestRegister bool `yaml:"rest_register"`
	GRPCRegister bool `yaml:"grpc_register"`
	Wire         bool `yaml:"wire"`
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
