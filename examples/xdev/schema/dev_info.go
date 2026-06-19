package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/chnxq/x-crud/entgo/mixin"
)

// Device holds the schema definition for the Device entity.
type Device struct {
	ent.Schema
}

func (Device) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "dev_info",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("设备信息表"),
	}
}

// Fields of the DevInfo.
func (Device) Fields() []ent.Field {
	return []ent.Field{
		field.String("device_code").
			Optional().
			Nillable().
			MaxLen(48).
			Comment("设备编码"),

		field.String("name").
			Optional().
			Nillable().
			MaxLen(64).
			Comment("设备名称"),

		field.Uint64("model_id").
			Default(0).
			Comment("设备型号ID"),

		field.String("serial_number").
			Optional().
			Nillable().
			MaxLen(32).
			Comment("设备序列号"),

		field.Bytes("finger_print").
			Optional().
			Nillable().
			Comment("设备指纹"),

		field.String("use_status").
			NotEmpty().
			MaxLen(2).
			Comment("使用状态"),

		field.Bytes("meta_data").
			Optional().
			Nillable().
			Comment("其他数据"),
	}
}

// Mixin of the Device.
func (Device) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.CreatedBy{},
		mixin.UpdatedBy{},
		mixin.TimeAt{},
		mixin.TenantID[uint32]{},
	}
}

// Indexes of the Device.
func (Device) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_by").
			StorageKey("idx_dev_info_created_by"),
		index.Fields("updated_by").
			StorageKey("idx_dev_info_updated_by"),
		index.Fields("deleted_at").
			StorageKey("idx_dev_info_deleted_at"),
		index.Fields("id", "device_code", "model_id").
			StorageKey("inx_dev_info"),
		index.Fields("tenant_id", "device_code").
			StorageKey("idx_dev_info_tenant_device_code"),
		index.Fields("tenant_id", "model_id").
			StorageKey("idx_dev_info_tenant_model_id"),
		index.Fields("tenant_id", "use_status").
			StorageKey("idx_dev_info_tenant_use_status"),
	}
}
