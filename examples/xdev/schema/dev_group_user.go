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

// DeviceGroupUser maps to xdev_dev_group_user.
type DeviceGroupUser struct {
	ent.Schema
}

func (DeviceGroupUser) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "xdev_dev_group_user",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("用户与设备组关联"),
	}
}

func (DeviceGroupUser) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("group_id").
			Comment("设备组ID"),
		field.Uint32("user_id").
			Comment("用户ID"),
	}
}

func (DeviceGroupUser) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.CreatedBy{},
		mixin.TimeAt{},
		mixin.TenantID[uint32]{},
	}
}

func (DeviceGroupUser) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("group", DeviceGroup.Type).
			Ref("user_relations").
			Field("group_id").
			Unique().
			Required(),
	}
}

func (DeviceGroupUser) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "group_id", "user_id").
			Unique().
			StorageKey("uidx_xdev_dev_group_user_tenant_group_user"),
		index.Fields("tenant_id", "group_id").
			StorageKey("idx_xdev_dev_group_user_tenant_group"),
		index.Fields("tenant_id", "user_id").
			StorageKey("idx_xdev_dev_group_user_tenant_user"),
	}
}
