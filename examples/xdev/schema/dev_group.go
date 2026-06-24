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

// DeviceGroup maps to xdev_dev_group.
type DeviceGroup struct {
	ent.Schema
}

func (DeviceGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "xdev_dev_group",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("设备组目录"),
	}
}

func (DeviceGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("group_name").
			Optional().
			Nillable().
			MaxLen(16).
			Comment("设备组名称"),
		field.Enum("type").
			NamedValues(
				"Function", "FUNCTION",
				"Network", "NETWORK",
				"User", "USER",
				"Department", "DEPARTMENT",
			).
			Default("FUNCTION").
			Comment("设备组类型"),
		field.Bool("is_leaf_node").
			Default(false).
			Comment("叶子节点"),
		field.String("descript").
			Optional().
			Nillable().
			MaxLen(64).
			Comment("设备组描述信息"),
		field.Bool("visible").
			Optional().
			Nillable().
			Default(false).
			Comment("是否显示"),
	}
}

func (DeviceGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.CreatedBy{},
		mixin.UpdatedBy{},
		mixin.TimeAt{},
		mixin.TenantID[uint32]{},
		mixin.SwitchStatus{},
		mixin.SortOrder{},
		mixin.Tree[DeviceGroup]{},
		mixin.TreePath{},
	}
}

func (DeviceGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("device_relations", DeviceGroupDevice.Type),
		edge.To("user_relations", DeviceGroupUser.Type),
		edge.To("org_unit_relations", DeviceGroupOrgUnit.Type),
	}
}

func (DeviceGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_by").
			StorageKey("idx_xdev_dev_group_created_by"),
		index.Fields("updated_by").
			StorageKey("idx_xdev_dev_group_updated_by"),
		index.Fields("deleted_at").
			StorageKey("idx_xdev_dev_group_deleted_at"),
		index.Fields("tenant_id", "parent_id", "group_name").
			StorageKey("idx_xdev_dev_group_tenant_parent_name"),
		index.Fields("tenant_id", "path").
			StorageKey("idx_xdev_dev_group_tenant_path"),
		index.Fields("tenant_id", "type").
			StorageKey("idx_xdev_dev_group_tenant_type"),
		index.Fields("tenant_id", "status").
			StorageKey("idx_xdev_dev_group_tenant_status"),
	}
}
