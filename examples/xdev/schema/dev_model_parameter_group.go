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

// DeviceModelParameterGroup maps to xdev_dev_model_parameter_group.
type DeviceModelParameterGroup struct {
	ent.Schema
}

func (DeviceModelParameterGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "xdev_dev_model_parameter_group",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("设备型号参数组绑定"),
	}
}

func (DeviceModelParameterGroup) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("model_id").
			Comment("设备型号ID"),
		field.Uint32("parameter_group_id").
			Comment("参数组ID"),
	}
}

func (DeviceModelParameterGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TenantID[uint32]{},
	}
}

func (DeviceModelParameterGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("model", DeviceModel.Type).
			Ref("parameter_group_relations").
			Field("model_id").
			Unique().
			Required(),
		edge.From("parameter_group", DeviceParameterGroup.Type).
			Ref("model_relations").
			Field("parameter_group_id").
			Unique().
			Required(),
	}
}

func (DeviceModelParameterGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "model_id").
			StorageKey("idx_xdev_dev_model_parameter_group_tenant_model"),
		index.Fields("tenant_id", "parameter_group_id").
			StorageKey("idx_xdev_dev_model_parameter_group_tenant_group"),
		index.Fields("tenant_id", "model_id", "parameter_group_id").
			Unique().
			StorageKey("uidx_xdev_dev_model_parameter_group_tenant_model_group"),
	}
}
