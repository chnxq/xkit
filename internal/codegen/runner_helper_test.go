package codegen

import (
	"testing"

	"github.com/chnxq/xkit/internal/entschema"
)

func TestRepoSetters_DefaultEnumUsesSchemaDefaultWhenAbsent(t *testing.T) {
	fields := []entschema.Field{{
		Name:    "share_scope",
		Kind:    "Enum",
		Default: true,
	}}
	dtoFields := map[string]string{"ShareScope": "*v1.DeviceParameterGroup_ShareScope"}

	create := repoSetters(fields, dtoFields, "Create", "deviceparametergroup", "enumPtr", "timePtr", "maskContains", "nil")
	if len(create) != 1 || create[0].Condition != "req.Data.ShareScope != nil" || create[0].Method != "SetShareScope" {
		t.Fatalf("unexpected create setter: %+v", create)
	}

	update := repoSetters(fields, dtoFields, "Update", "deviceparametergroup", "enumPtr", "timePtr", "maskContains", "nil")
	if len(update) != 1 || update[0].Condition != "req.Data.ShareScope != nil" || update[0].ClearMethod != "" {
		t.Fatalf("unexpected update setter: %+v", update)
	}
}
