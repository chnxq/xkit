package sourceimport

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/chnxq/xkit/internal/config"
)

func TestImportCopiesSourceAndGeneratesConfig(t *testing.T) {
	root := t.TempDir()
	projectRoot := filepath.Join(root, "xadmin-web")
	sourceRoot := filepath.Join(projectRoot, "source")

	writeTestFile(t, filepath.Join(projectRoot, "go.mod"), "module xadmin-web\n")
	writeTestFile(t, filepath.Join(sourceRoot, "api", "buf.yaml"), "version: v2\n")
	writeTestFile(t, filepath.Join(sourceRoot, "api", "buf.admin.openapi.gen.yaml"), `version: v2
managed:
  enabled: true
  override:
    - file_option: go_package_prefix
      value: admin-01/api/gen
plugins: []
`)
	writeTestFile(t, filepath.Join(sourceRoot, "api", "buf.gen.yaml"), `version: v2
managed:
  enabled: true
  override:
    - file_option: go_package_prefix
      value: wrong-module/api/gen
    - file_option: go_package
      path: admin/v1
      value: wrong-module/api/gen/admin/v1;wrongadmin
    - file_option: go_package
      path: authentication/v1
      value: admin-02/api/gen/authentication/v1;wrongauth
    - file_option: go_package
      path: internal_message/v1
      value: wrong-module/api/gen/internal_message/v1;wrongmessage
    - file_option: go_package
      path: permission/v1
      value: wrong-module/api/gen/permission/v1;permission
    - file_option: go_package
      path: task/v1
      value: wrong-module/api/gen/task/v1;task
plugins: []
`)
	writeTestFile(t, filepath.Join(sourceRoot, "api", "buf.lock"), "# lock\n")
	writeTestFile(t, filepath.Join(sourceRoot, "api", "README.md"), "source api notes\n")
	writeTestFile(t, filepath.Join(sourceRoot, "api", "notes.custom"), "any root file\n")
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

	paginationV1 "github.com/chnxq/x-crud/api/gen/pagination/v1"

	identityV1 "admin-01/api/gen/identity/v1"
)

var _ *paginationV1.PagingRequest
var _ *identityV1.User

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
	assertFileExists(t, filepath.Join(projectRoot, "api", "buf.yaml"))
	assertFileExists(t, filepath.Join(projectRoot, "api", "buf.admin.openapi.gen.yaml"))
	assertFileExists(t, filepath.Join(projectRoot, "api", "buf.gen.yaml"))
	assertFileExists(t, filepath.Join(projectRoot, "api", "buf.lock"))
	assertFileExists(t, filepath.Join(projectRoot, "api", "README.md"))
	assertFileExists(t, filepath.Join(projectRoot, "api", "notes.custom"))
	bufGen := readTestFile(t, filepath.Join(projectRoot, "api", "buf.gen.yaml"))
	for _, expected := range []string{
		"value: xadmin-web/api/gen\n",
		"value: xadmin-web/api/gen/admin/v1;admin\n",
		"value: xadmin-web/api/gen/authentication/v1;authentication\n",
		"value: xadmin-web/api/gen/internal_message/v1;internalmessage\n",
		"value: xadmin-web/api/gen/permission/v1;permissionv1\n",
		"value: xadmin-web/api/gen/task/v1;taskv1\n",
	} {
		if !strings.Contains(bufGen, expected) {
			t.Fatalf("buf.gen.yaml missing corrected value %q:\n%s", expected, bufGen)
		}
	}
	if strings.Contains(bufGen, "admin-02") || strings.Contains(bufGen, "wrong-module") || strings.Contains(bufGen, "wrongauth") {
		t.Fatalf("buf.gen.yaml should not keep stale go package values:\n%s", bufGen)
	}
	openapiBufGen := readTestFile(t, filepath.Join(projectRoot, "api", "buf.admin.openapi.gen.yaml"))
	if !strings.Contains(openapiBufGen, "value: xadmin-web/api/gen\n") || strings.Contains(openapiBufGen, "admin-01") {
		t.Fatalf("buf.*.gen.yaml should normalize go package values:\n%s", openapiBufGen)
	}
	assertFileExists(t, filepath.Join(projectRoot, "internal", "data", "ent", "schema", "user.go"))
	userSchema := readTestFile(t, filepath.Join(projectRoot, "internal", "data", "ent", "schema", "user.go"))
	if !strings.Contains(userSchema, `identityV1 "xadmin-web/api/gen/identity/v1"`) {
		t.Fatalf("schema import should be corrected to target module:\n%s", userSchema)
	}
	if !strings.Contains(userSchema, `paginationV1 "github.com/chnxq/x-crud/api/gen/pagination/v1"`) {
		t.Fatalf("external api/gen import should be preserved:\n%s", userSchema)
	}
	if strings.Contains(userSchema, `admin-01/api/gen/identity/v1`) {
		t.Fatalf("schema import should not keep stale local module:\n%s", userSchema)
	}
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
	if !credential.Generate.RepoCRUD || credential.Generate.Wire {
		t.Fatalf("domain-only resource should generate repo but not default wire: %#v", credential.Generate)
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
	writeTestFile(t, filepath.Join(sourceRoot, "api", "buf.yaml"), "version: v2\n")
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
	if _, err := os.Stat(filepath.Join(projectRoot, "api", "buf.yaml")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not write api root file, stat err=%v", err)
	}
	if _, err := os.Stat(result.ConfigPath); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not write config, stat err=%v", err)
	}
}

func TestImportCorrectsExistingBufGenYAMLWithoutForce(t *testing.T) {
	root := t.TempDir()
	projectRoot := filepath.Join(root, "app")
	sourceRoot := filepath.Join(root, "source")

	writeTestFile(t, filepath.Join(projectRoot, "go.mod"), "module example.com/app\n")
	targetBufGen := filepath.Join(projectRoot, "api", "buf.gen.yaml")
	writeTestFile(t, targetBufGen, `version: v2
managed:
  enabled: true
  override:
    - file_option: go_package
      path: admin/v1
      value: stale-module/api/gen/admin/v1;stale
plugins: []
`)
	writeTestFile(t, filepath.Join(sourceRoot, "api", "buf.gen.yaml"), `version: v2
managed:
  enabled: true
  override:
    - file_option: go_package
      path: admin/v1
      value: admin-02/api/gen/admin/v1;wrong
plugins: []
`)
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
	})
	if err != nil {
		t.Fatalf("import source: %v", err)
	}
	if !slices.Contains(result.Written, targetBufGen) {
		t.Fatalf("expected existing buf.gen.yaml to be corrected, written=%#v", result.Written)
	}
	bufGen := readTestFile(t, targetBufGen)
	if !strings.Contains(bufGen, "value: example.com/app/api/gen/admin/v1;admin\n") {
		t.Fatalf("buf.gen.yaml missing corrected go package:\n%s", bufGen)
	}
	if strings.Contains(bufGen, "admin-02") || strings.Contains(bufGen, "stale-module") || strings.Contains(bufGen, "wrong") {
		t.Fatalf("buf.gen.yaml should not keep stale values:\n%s", bufGen)
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

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}

func assertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s not to exist, stat err=%v", path, err)
	}
}
