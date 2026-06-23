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

// DeviceModel maps to xdev_dev_model.
type DeviceModel struct {
	ent.Schema
}

func (DeviceModel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "xdev_dev_model",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("设备型号"),
	}
}

func (DeviceModel) Fields() []ent.Field {
	return []ent.Field{
		field.String("model_name").
			Optional().
			Nillable().
			MaxLen(16).
			Comment("型号"),
		field.Uint32("model_type_id").
			Optional().
			Nillable().
			Comment("型号分类ID"),
		field.String("description").
			Optional().
			Nillable().
			MaxLen(200).
			Comment("型号描述"),
		field.String("remark").
			Optional().
			Nillable().
			MaxLen(100).
			Comment("备注"),
	}
}

func (DeviceModel) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.CreatedBy{},
		mixin.TimeAt{},
		mixin.TenantID[uint32]{},
	}
}

func (DeviceModel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("model_type", DeviceModelType.Type).
			Ref("models").
			Field("model_type_id").
			Unique(),
		edge.To("devices", Device.Type),
	}
}

func (DeviceModel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("model_type_id").
			StorageKey("inx_dev_model_model_type_id"),
		index.Fields("created_by").
			StorageKey("idx_dev_model_created_by"),
		index.Fields("tenant_id", "model_name").
			StorageKey("idx_dev_model_tenant_model_name"),
	}
}
