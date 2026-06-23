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

// DeviceGroupDevice maps to xdev_dev_group_device.
type DeviceGroupDevice struct {
	ent.Schema
}

func (DeviceGroupDevice) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "xdev_dev_group_device",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("设备组与设备关联"),
	}
}

func (DeviceGroupDevice) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("group_id").
			Comment("设备组ID"),
		field.Uint32("device_id").
			Comment("设备ID"),
	}
}

func (DeviceGroupDevice) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.CreatedBy{},
		mixin.TimeAt{},
		mixin.TenantID[uint32]{},
	}
}

func (DeviceGroupDevice) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("group", DeviceGroup.Type).
			Ref("device_relations").
			Field("group_id").
			Unique().
			Required(),
		edge.From("device", Device.Type).
			Ref("group_relations").
			Field("device_id").
			Unique().
			Required(),
	}
}

func (DeviceGroupDevice) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "group_id", "device_id").
			Unique().
			StorageKey("uidx_xdev_dev_group_device_tenant_group_device"),
		index.Fields("tenant_id", "group_id").
			StorageKey("idx_xdev_dev_group_device_tenant_group"),
		index.Fields("tenant_id", "device_id").
			StorageKey("idx_xdev_dev_group_device_tenant_device"),
	}
}
