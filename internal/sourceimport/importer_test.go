package sourceimport

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/chnxq/xkit/internal/config"
)

func TestImportCopiesSourceAndGeneratesConfig(t *testing.T) {
	root := t.TempDir()
	projectRoot := filepath.Join(root, "xadmin-web")
	sourceRoot := filepath.Join(projectRoot, "source")

	writeTestFile(t, filepath.Join(projectRoot, "go.mod"), "module xadmin-web\n")
	writeTestFile(t, filepath.Join(sourceRoot, "api", "protos", "admin", "v1", "i_user.proto"), `syntax = "proto3";
package admin.service.v1;

import "pagination/v1/pagination.proto";
import "identity/v1/user.proto";
import "google/protobuf/empty.proto";

service UserService {
  rpc List (pagination.PagingRequest) returns (identity.service.v1.ListUserResponse) {}
  rpc Get (identity.service.v1.GetUserRequest) returns (identity.service.v1.User) {}
  rpc Create (identity.service.v1.CreateUserRequest) returns (google.protobuf.Empty) {}
  rpc Update (identity.service.v1.UpdateUserRequest) returns (google.protobuf.Empty) {}
  rpc Delete (identity.service.v1.DeleteUserRequest) returns (google.protobuf.Empty) {}
  rpc UserExists (identity.service.v1.UserExistsRequest) returns (identity.service.v1.UserExistsResponse) {}
  rpc EditUserPassword (identity.service.v1.EditUserPasswordRequest) returns (google.protobuf.Empty) {}
}
`)
	writeTestFile(t, filepath.Join(sourceRoot, "api", "protos", "identity", "v1", "user.proto"), `syntax = "proto3";
package identity.service.v1;

message User {}
message ListUserResponse {}
message GetUserRequest {
  oneof query_by {
    uint32 id = 1;
    string username = 2;
  }
}
message CreateUserRequest {}
message UpdateUserRequest {}
message DeleteUserRequest {}
message UserExistsRequest {
  oneof query_by {
    uint32 id = 1;
    string username = 2;
  }
}
message UserExistsResponse {}
message EditUserPasswordRequest {}
`)
	writeTestFile(t, filepath.Join(sourceRoot, "api", "protos", "authentication", "v1", "user_credential.proto"), `syntax = "proto3";
package authentication.service.v1;

import "pagination/v1/pagination.proto";
import "google/protobuf/empty.proto";

service UserCredentialService {
  rpc List (pagination.PagingRequest) returns (ListUserCredentialResponse) {}
  rpc Get (GetUserCredentialRequest) returns (UserCredential) {}
  rpc ResetCredential (ResetCredentialRequest) returns (google.protobuf.Empty) {}
}

message UserCredential {}
message ListUserCredentialResponse {}
message GetUserCredentialRequest {}
message ResetCredentialRequest {}
`)
	writeTestFile(t, filepath.Join(sourceRoot, "schema", "user.go"), `package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type User struct{ ent.Schema }

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username"),
		field.String("nickname").Optional(),
		field.Uint32("tenant_id").Optional(),
		field.Enum("status").Values("enabled", "disabled"),
	}
}
`)
	writeTestFile(t, filepath.Join(sourceRoot, "schema", "user_credential.go"), `package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type UserCredential struct{ ent.Schema }

func (UserCredential) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("user_id"),
		field.Enum("identity_type").Values("username"),
		field.String("identifier"),
		field.Enum("credential_type").Values("password_hash"),
	}
}
`)

	result, err := Import(Options{
		SourceRoot:  sourceRoot,
		ProjectRoot: projectRoot,
		Service:     "admin",
		Force:       true,
	})
	if err != nil {
		t.Fatalf("import source: %v", err)
	}

	assertFileExists(t, filepath.Join(projectRoot, "api", "protos", "admin", "v1", "i_user.proto"))
	assertFileExists(t, filepath.Join(projectRoot, "internal", "data", "ent", "schema", "user.go"))
	if result.ConfigPath != filepath.Join(sourceRoot, "xadmin-web-config", "admin.yaml") {
		t.Fatalf("unexpected config path: %s", result.ConfigPath)
	}

	cfg, err := config.Load(result.ConfigPath)
	if err != nil {
		t.Fatalf("load generated config: %v", err)
	}
	if cfg.Service != "admin" || cfg.Module != "xadmin-web" {
		t.Fatalf("unexpected config header: service=%q module=%q", cfg.Service, cfg.Module)
	}

	user := findResource(t, cfg, "user")
	if user.ProtoService != "admin.service.v1.UserService" {
		t.Fatalf("unexpected user proto service: %s", user.ProtoService)
	}
	if user.DTOImport != "xadmin-web/api/gen/identity/v1" {
		t.Fatalf("unexpected user dto import: %s", user.DTOImport)
	}
	if !user.Operations["list"] || !user.Operations["exists"] {
		t.Fatalf("user operations should include list and exists: %#v", user.Operations)
	}
	if !slices.Equal(user.ExistsFields, []string{"id", "username"}) {
		t.Fatalf("unexpected user exists fields: %#v", user.ExistsFields)
	}
	if _, ok := user.ServiceMethods["EditUserPassword"]; !ok {
		t.Fatalf("user should include EditUserPassword service method config")
	}
	if !slices.Contains(user.Filters.Allow, "tenant_id") || !slices.Contains(user.Filters.Allow, "status") {
		t.Fatalf("user filters should include schema fields: %#v", user.Filters.Allow)
	}

	credential := findResource(t, cfg, "user_credential")
	if credential.ProtoService != "authentication.service.v1.UserCredentialService" {
		t.Fatalf("unexpected credential proto service: %s", credential.ProtoService)
	}
	if credential.Generate.ServiceStub || credential.Generate.RestRegister || credential.Generate.GRPCRegister {
		t.Fatalf("domain-only resource should not generate public service/register: %#v", credential.Generate)
	}
	if !credential.Generate.RepoCRUD || !credential.Generate.Wire {
		t.Fatalf("domain-only resource should generate repo and wire: %#v", credential.Generate)
	}
	if !credential.Operations["resetcredential"] {
		t.Fatalf("credential operations should include resetcredential: %#v", credential.Operations)
	}
}

func TestImportDryRunDoesNotWriteFiles(t *testing.T) {
	root := t.TempDir()
	projectRoot := filepath.Join(root, "app")
	sourceRoot := filepath.Join(root, "source")

	writeTestFile(t, filepath.Join(projectRoot, "go.mod"), "module example.com/app\n")
	writeTestFile(t, filepath.Join(sourceRoot, "api", "protos", "admin", "v1", "i_role.proto"), `syntax = "proto3";
package admin.service.v1;
import "permission/v1/role.proto";
service RoleService {
  rpc Get (permission.service.v1.GetRoleRequest) returns (permission.service.v1.Role) {}
}
`)
	writeTestFile(t, filepath.Join(sourceRoot, "api", "protos", "permission", "v1", "role.proto"), `syntax = "proto3";
package permission.service.v1;
message Role {}
message GetRoleRequest {}
`)
	writeTestFile(t, filepath.Join(sourceRoot, "schema", "role.go"), `package schema

import "entgo.io/ent"

type Role struct{ ent.Schema }
`)

	result, err := Import(Options{
		SourceRoot:  sourceRoot,
		ProjectRoot: projectRoot,
		Service:     "admin",
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("dry-run import source: %v", err)
	}
	if len(result.Written) == 0 {
		t.Fatalf("dry-run should plan writes")
	}
	if _, err := os.Stat(filepath.Join(projectRoot, "api", "protos", "admin", "v1", "i_role.proto")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not write target proto, stat err=%v", err)
	}
	if _, err := os.Stat(result.ConfigPath); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not write config, stat err=%v", err)
	}
}

func findResource(t *testing.T, cfg config.Config, name string) config.Resource {
	t.Helper()
	for _, resource := range cfg.Resources {
		if resource.Name == name {
			return resource
		}
	}
	t.Fatalf("missing resource %q", name)
	return config.Resource{}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}
