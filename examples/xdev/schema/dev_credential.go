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

// DeviceCredential maps to xdev_dev_credential.
type DeviceCredential struct {
	ent.Schema
}

func (DeviceCredential) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "xdev_dev_credential",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("设备凭据"),
	}
}

func (DeviceCredential) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("device_id").
			Comment("设备ID"),
		field.String("credential_name").
			Optional().
			Nillable().
			MaxLen(64).
			Comment("凭据名称"),
		field.Enum("credential_type").
			NamedValues(
				"String", "STRING",
				"JSON", "JSON",
				"Other", "OTHER",
			).
			Default("STRING").
			Comment("凭据类型"),
		field.String("credential_value").
			Optional().
			Nillable().
			MaxLen(2048).
			Comment("凭据内容"),
		field.String("remark").
			Optional().
			Nillable().
			MaxLen(200).
			Comment("备注"),
	}
}

func (DeviceCredential) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TenantID[uint32]{},
	}
}

func (DeviceCredential) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("device", Device.Type).
			Ref("credentials").
			Field("device_id").
			Unique().
			Required(),
	}
}

func (DeviceCredential) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "device_id").
			StorageKey("idx_xdev_dev_credential_tenant_device"),
		index.Fields("tenant_id", "credential_type").
			StorageKey("idx_xdev_dev_credential_tenant_type"),
		index.Fields("tenant_id", "device_id", "credential_name").
			StorageKey("idx_xdev_dev_credential_tenant_device_name"),
	}
}
