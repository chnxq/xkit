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
	if got, want := len(user.Fields), 2; got != want {
		t.Fatalf("field count mismatch: got %d want %d", got, want)
	}
	if !user.Fields[0].Optional || !user.Fields[0].Nillable || !user.Fields[0].Immutable {
		t.Fatalf("username flags were not parsed: %+v", user.Fields[0])
	}
	if user.Fields[1].Kind != "Enum" || user.Fields[1].Name != "status" {
		t.Fatalf("enum field mismatch: %+v", user.Fields[1])
	}
}
