package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/chnxq/x-crud/entgo/mixin"
)

// DeviceParameterItem maps to xdev_dev_parameter_item.
type DeviceParameterItem struct {
	ent.Schema
}

func (DeviceParameterItem) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "xdev_dev_parameter_item",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("设备参数项"),
	}
}

func (DeviceParameterItem) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("parameter_group_id").
			Comment("参数组ID"),
		field.String("parameter_name").
			Optional().
			Nillable().
			MaxLen(100).
			Comment("参数名称"),
		field.String("parameter_key").
			NotEmpty().
			MaxLen(64).
			Comment("参数键"),
		field.Enum("value_type").
			NamedValues(
				"Number", "NUMBER",
				"Bool", "BOOL",
				"String", "STRING",
				"JSON", "JSON",
			).
			Default("STRING").
			Comment("参数值类型"),
		field.String("default_value").
			Optional().
			Nillable().
			MaxLen(4096).
			Comment("默认值"),
		field.Enum("constraint_type").
			NamedValues(
				"None", "NONE",
				"Range", "RANGE",
				"Length", "LENGTH",
			).
			Default("NONE").
			Comment("约束类型"),
		field.String("constraint_config").
			Optional().
			Nillable().
			MaxLen(4096).
			Validate(func(value string) error {
				if strings.TrimSpace(value) == "" {
					return nil
				}
				var config map[string]any
				if err := json.Unmarshal([]byte(value), &config); err != nil {
					return fmt.Errorf("constraint config must be a JSON object: %w", err)
				}
				return nil
			}).
			Comment("JSON格式约束配置"),
		field.String("unit").
			Optional().
			Nillable().
			MaxLen(32).
			Comment("单位"),
		field.Bool("required").
			Default(false).
			Comment("是否必填"),
		field.String("remark").
			Optional().
			Nillable().
			MaxLen(200).
			Comment("备注"),
	}
}

func (DeviceParameterItem) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TenantID[uint32]{},
	}
}

func (DeviceParameterItem) Hooks() []ent.Hook {
	return []ent.Hook{
		func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, mutation ent.Mutation) (ent.Value, error) {
				if err := validateParameterItemMutation(mutation); err != nil {
					return nil, err
				}
				return next.Mutate(ctx, mutation)
			})
		},
	}
}

func validateParameterItemMutation(mutation ent.Mutation) error {
	valueType := mutationStringField(mutation, "value_type")
	defaultValue := mutationStringField(mutation, "default_value")
	constraintType := mutationStringField(mutation, "constraint_type")
	constraintConfig := mutationStringField(mutation, "constraint_config")

	if defaultValue != "" {
		switch valueType {
		case "NUMBER":
			if _, err := strconv.ParseFloat(defaultValue, 64); err != nil {
				return fmt.Errorf("default value must be a number: %w", err)
			}
		case "BOOL":
			if defaultValue != "true" && defaultValue != "false" {
				return fmt.Errorf("boolean default value must be true or false")
			}
		case "JSON":
			if !json.Valid([]byte(defaultValue)) {
				return fmt.Errorf("JSON default value is invalid")
			}
		}
	}

	if constraintType == "RANGE" {
		if valueType != "" && valueType != "NUMBER" {
			return fmt.Errorf("RANGE constraint requires NUMBER value type")
		}
		var config struct {
			Max *float64 `json:"max"`
			Min *float64 `json:"min"`
		}
		if err := json.Unmarshal([]byte(constraintConfig), &config); err != nil {
			return fmt.Errorf("invalid RANGE constraint config: %w", err)
		}
		if config.Min == nil || config.Max == nil || *config.Min > *config.Max {
			return fmt.Errorf("RANGE constraint requires min <= max")
		}
	}
	if constraintType == "LENGTH" {
		if valueType != "" && valueType != "STRING" {
			return fmt.Errorf("LENGTH constraint requires STRING value type")
		}
		var config struct {
			MaxLength int `json:"maxLength"`
		}
		if err := json.Unmarshal([]byte(constraintConfig), &config); err != nil {
			return fmt.Errorf("invalid LENGTH constraint config: %w", err)
		}
		if config.MaxLength <= 0 {
			return fmt.Errorf("LENGTH constraint requires a positive maxLength")
		}
		if defaultValue != "" && len([]rune(defaultValue)) > config.MaxLength {
			return fmt.Errorf("default value exceeds maxLength")
		}
	}
	return nil
}

func mutationStringField(mutation ent.Mutation, name string) string {
	value, ok := mutation.Field(name)
	if !ok || value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func (DeviceParameterItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("parameter_group", DeviceParameterGroup.Type).
			Ref("items").
			Field("parameter_group_id").
			Unique().
			Required(),
	}
}

func (DeviceParameterItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "parameter_group_id").
			StorageKey("idx_xdev_dev_parameter_item_tenant_group"),
		index.Fields("tenant_id", "parameter_key").
			StorageKey("idx_xdev_dev_parameter_item_tenant_key"),
		index.Fields("tenant_id", "parameter_group_id", "parameter_key").
			Unique().
			StorageKey("uidx_xdev_dev_parameter_item_tenant_group_key"),
	}
}
