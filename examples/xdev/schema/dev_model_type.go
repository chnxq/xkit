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

// DeviceModelType maps to xdev_dev_model_type.
type DeviceModelType struct {
	ent.Schema
}

func (DeviceModelType) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "xdev_dev_model_type",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("设备型号分类"),
	}
}

func (DeviceModelType) Fields() []ent.Field {
	return []ent.Field{
		field.String("model_type_name").
			Optional().
			Nillable().
			MaxLen(100).
			Comment("型号分类名"),
		field.String("use_case").
			Optional().
			Nillable().
			MaxLen(200).
			Comment("型号用途"),
		field.String("type_desc").
			Optional().
			Nillable().
			MaxLen(200).
			Comment("描述"),
	}
}

func (DeviceModelType) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.CreatedBy{},
		mixin.TimeAt{},
		mixin.TenantID[uint32]{},
	}
}

func (DeviceModelType) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("models", DeviceModel.Type),
	}
}

func (DeviceModelType) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_by").
			StorageKey("idx_xdev_dev_model_type_created_by"),
	}
}
