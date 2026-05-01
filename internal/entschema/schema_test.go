package entschema

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ParsesEntSchemaFields(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "internal", "data", "ent", "schema", "user.go")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir schema dir: %v", err)
	}

	content := `package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type User struct { ent.Schema }

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").NotEmpty().Immutable().Optional().Nillable(),
		field.Enum("status").Default("NORMAL").Optional().Nillable(),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.OperatorID{},
		mixin.TimeAt{},
		mixin.Remark{},
		mixin.TenantID[uint32]{},
	}
}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	items, err := Load(root)
	if err != nil {
		t.Fatalf("load schema: %v", err)
	}

	user, ok := items["User"]
	if !ok {
		t.Fatalf("missing User schema")
	}
	if got, want := len(user.Fields), 11; got != want {
		t.Fatalf("field count mismatch: got %d want %d", got, want)
	}
	if !user.Fields[0].Optional || !user.Fields[0].Nillable || !user.Fields[0].Immutable {
		t.Fatalf("username flags were not parsed: %+v", user.Fields[0])
	}
	if user.Fields[1].Kind != "Enum" || user.Fields[1].Name != "status" {
		t.Fatalf("enum field mismatch: %+v", user.Fields[1])
	}
	assertField(t, user.Fields, Field{Name: "tenant_id", Kind: "Uint32", Optional: true, Nillable: true, Immutable: true})
	assertField(t, user.Fields, Field{Name: "created_at", Kind: "Time", Optional: true, Nillable: true, Immutable: true})
	assertField(t, user.Fields, Field{Name: "remark", Kind: "String", Optional: true, Nillable: true})
}

func TestLoad_ParsesJSONFieldWithTypeArgument(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "internal", "data", "ent", "schema", "menu.go")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir schema dir: %v", err)
	}

	content := `package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	resourceV1 "example.com/xadmin/api/gen/resource/v1"
)

type Menu struct { ent.Schema }

func (Menu) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("meta", &resourceV1.MenuMeta{}).Optional(),
	}
}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	items, err := Load(root)
	if err != nil {
		t.Fatalf("load schema: %v", err)
	}

	menu, ok := items["Menu"]
	if !ok {
		t.Fatalf("missing Menu schema")
	}
	assertField(t, menu.Fields, Field{Name: "meta", Kind: "JSON", Optional: true})
}

func TestKnownMixinFields_PositionMixins(t *testing.T) {
	t.Parallel()

	assertField(t, knownMixinFields("SortOrder"), Field{Name: "sort_order", Kind: "Uint32", Optional: true, Nillable: true})
	assertField(t, knownMixinFields("SwitchStatus"), Field{Name: "status", Kind: "Enum", Nillable: true})
	assertField(t, knownMixinFields("Tree"), Field{Name: "parent_id", Kind: "Uint32", Optional: true, Nillable: true})
	assertField(t, knownMixinFields("TreePath"), Field{Name: "path", Kind: "String", Optional: true, Nillable: true})
	assertField(t, knownMixinFields("TreePathIDs"), Field{Name: "ancestor_ids", Kind: "JSON", Optional: true})
}

func TestKnownMixinFields_AuditAliasMixins(t *testing.T) {
	t.Parallel()

	assertField(t, knownMixinFields("CreatedAt"), Field{Name: "created_at", Kind: "Time", Optional: true, Nillable: true, Immutable: true})
	assertField(t, knownMixinFields("UpdatedAt"), Field{Name: "updated_at", Kind: "Time", Optional: true, Nillable: true})
	assertField(t, knownMixinFields("CreateTime"), Field{Name: "create_time", Kind: "Time", Optional: true, Nillable: true, Immutable: true})
	assertField(t, knownMixinFields("UpdateTime"), Field{Name: "update_time", Kind: "Time", Optional: true, Nillable: true})
	assertField(t, knownMixinFields("CreatedAtTimestamp"), Field{Name: "created_at", Kind: "Int64", Optional: true, Nillable: true, Immutable: true})
	assertField(t, knownMixinFields("Timestamp"), Field{Name: "update_time", Kind: "Int64", Optional: true, Nillable: true})
	assertField(t, knownMixinFields("CreatorId"), Field{Name: "creator_id", Kind: "Uint32", Optional: true, Nillable: true, Immutable: true})
	assertField(t, knownMixinFields("CreateBy64"), Field{Name: "create_by", Kind: "Uint64", Optional: true, Nillable: true})
	assertField(t, knownMixinFields("UpdatedBy"), Field{Name: "updated_by", Kind: "Uint32", Optional: true, Nillable: true})
	assertField(t, knownMixinFields("AuditorID64"), Field{Name: "deleted_by", Kind: "Uint64", Optional: true, Nillable: true})
}

func assertField(t *testing.T, fields []Field, want Field) {
	t.Helper()

	for _, got := range fields {
		if got.Name != want.Name {
			continue
		}
		if got.Kind != want.Kind || got.Optional != want.Optional || got.Nillable != want.Nillable || got.Immutable != want.Immutable {
			t.Fatalf("field %s mismatch: got %+v want %+v", want.Name, got, want)
		}
		return
	}
	t.Fatalf("missing field %s", want.Name)
}
