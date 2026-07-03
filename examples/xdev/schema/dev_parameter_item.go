package schema

import (
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
		field.Enum("parameter_type").
			NamedValues(
				"Equal", "EQ",
				"Min", "MIN",
				"Max", "MAX",
				"Between", "BETWEEN",
				"Bool", "BOOL",
				"Regex", "REGEX",
				"Contains", "CONTAINS",
				"String", "STRING",
				"JSON", "JSON",
			).
			Default("EQ").
			Comment("参数运算"),
		field.String("parameter_value").
			Optional().
			Nillable().
			MaxLen(1024).
			Comment("参数值"),
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
