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

// DeviceParameterGroup maps to xdev_dev_parameter_group.
type DeviceParameterGroup struct {
	ent.Schema
}

func (DeviceParameterGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "xdev_dev_parameter_group",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("设备参数组"),
	}
}

func (DeviceParameterGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("group_code").
			NotEmpty().
			MaxLen(64).
			Comment("参数组编码"),
		field.String("group_name").
			Optional().
			Nillable().
			MaxLen(100).
			Comment("参数组名称"),
		field.Enum("group_type").
			NamedValues(
				"Communication", "COMMUNICATION",
				"Control", "CONTROL",
				"Acquisition", "ACQUISITION",
				"UserDefinition", "USER_DEFINITION",
				"Integration", "INTEGRATION",
				"NoClassify", "NO_CLASSIFY",
			).
			Default("NO_CLASSIFY").
			Comment("参数组类型"),
		field.Bool("editable").
			Default(true).
			Comment("允许编辑"),
		field.String("description").
			Optional().
			Nillable().
			MaxLen(200).
			Comment("参数组描述"),
		field.Uint32("version").
			Default(1).
			Comment("版本"),
	}
}

func (DeviceParameterGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TenantID[uint32]{},
	}
}

func (DeviceParameterGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("items", DeviceParameterItem.Type),
		edge.To("model_relations", DeviceModelParameterGroup.Type),
	}
}

func (DeviceParameterGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "group_code").
			Unique().
			StorageKey("uidx_xdev_dev_parameter_group_tenant_code"),
		index.Fields("tenant_id", "group_name").
			StorageKey("idx_xdev_dev_parameter_group_tenant_name"),
		index.Fields("tenant_id", "group_type").
			StorageKey("idx_xdev_dev_parameter_group_tenant_type"),
	}
}
