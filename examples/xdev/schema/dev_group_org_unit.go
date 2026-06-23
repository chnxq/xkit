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

// DeviceGroupOrgUnit maps to xdev_dev_group_org_unit.
type DeviceGroupOrgUnit struct {
	ent.Schema
}

func (DeviceGroupOrgUnit) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "xdev_dev_group_org_unit",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("组织单元与设备组关联"),
	}
}

func (DeviceGroupOrgUnit) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("group_id").
			Comment("设备组ID"),
		field.Uint32("org_unit_id").
			Comment("组织单元ID"),
	}
}

func (DeviceGroupOrgUnit) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.CreatedBy{},
		mixin.TimeAt{},
		mixin.TenantID[uint32]{},
	}
}

func (DeviceGroupOrgUnit) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("group", DeviceGroup.Type).
			Ref("org_unit_relations").
			Field("group_id").
			Unique().
			Required(),
	}
}

func (DeviceGroupOrgUnit) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "group_id", "org_unit_id").
			Unique().
			StorageKey("uidx_xdev_dev_group_org_unit_tenant_group_org_unit"),
		index.Fields("tenant_id", "group_id").
			StorageKey("idx_xdev_dev_group_org_unit_tenant_group"),
		index.Fields("tenant_id", "org_unit_id").
			StorageKey("idx_xdev_dev_group_org_unit_tenant_org_unit"),
	}
}
