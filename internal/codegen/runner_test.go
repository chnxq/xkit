package codegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chnxq/xkit/internal/config"
	"github.com/chnxq/xkit/internal/project"
)

func TestRunnerGenerateAll_WritesPhaseOneFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/xadmin-web\n\ngo 1.26.0\n")
	writeFile(t, filepath.Join(root, "api", "protos", "admin", "v1", "i_user.proto"), `syntax = "proto3";

package admin.service.v1;

service UserService {
  rpc List (ListRequest) returns (ListResponse) {}
  rpc Get (GetRequest) returns (GetResponse) {}
}`)
	writeFile(t, filepath.Join(root, "api", "gen", "admin", "v1", "i_user_grpc.pb.go"), `package admin

import (
	context "context"
	v1 "example.com/xadmin-web/api/gen/identity/v1"
)

type UserServiceServer interface {
	List(context.Context, *v1.ListRequest) (*v1.ListResponse, error)
	Get(context.Context, *v1.GetRequest) (*v1.GetResponse, error)
	mustEmbedUnimplementedUserServiceServer()
}

var UserService_ServiceDesc = struct{
	ServiceName string
}{
	ServiceName: "admin.service.v1.UserService",
}
`)
	writeFile(t, filepath.Join(root, "api", "gen", "admin", "v1", "i_user_http.pb.go"), `package admin

type UserServiceHTTPServer interface{}
`)
	writeFile(t, filepath.Join(root, "internal", "data", "ent", "schema", "user.go"), `package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type User struct { ent.Schema }

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").Optional().Nillable(),
	}
}
`)

	cfg := config.Config{
		Service: "admin",
		Module:  "example.com/xadmin-web",
		Resources: []config.Resource{
			{
				Name:          "user",
				ProtoService:  "admin.service.v1.UserService",
				Entity:        "User",
				DTOImport:     "example.com/xadmin-web/api/gen/identity/v1",
				DTOType:       "User",
				RepoInterface: "UserRepo",
				Generate: config.GenerateFlags{
					ServiceStub:  true,
					RepoCRUD:     true,
					RestRegister: true,
					GRPCRegister: true,
					Wire:         true,
				},
			},
		},
	}

	runner, err := New(project.Info{
		Root:   root,
		Module: "example.com/xadmin-web",
	}, cfg, Options{})
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}

	result, err := runner.Generate("all")
	if err != nil {
		t.Fatalf("generate all: %v", err)
	}

	if len(result.Written) != 7 {
		t.Fatalf("written file count mismatch: got %d want %d", len(result.Written), 7)
	}

	expectedPaths := []string{
		filepath.Join(root, "app", "admin", "service", "internal", "service", "user_service.gen.go"),
		filepath.Join(root, "app", "admin", "service", "internal", "service", "user_service_ext.go"),
		filepath.Join(root, "app", "admin", "service", "internal", "data", "user_repo.gen.go"),
		filepath.Join(root, "app", "admin", "service", "internal", "server", "rest_register.gen.go"),
		filepath.Join(root, "app", "admin", "service", "internal", "server", "grpc_register.gen.go"),
		filepath.Join(root, "app", "admin", "service", "internal", "service", "providers", "wire_set.gen.go"),
		filepath.Join(root, "app", "admin", "service", "internal", "data", "providers", "wire_set.gen.go"),
	}
	for _, path := range expectedPaths {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected generated file %s: %v", path, err)
		}
	}

	serviceFile := readFile(t, filepath.Join(root, "app", "admin", "service", "internal", "service", "user_service.gen.go"))
	if !strings.Contains(serviceFile, "UnimplementedUserServiceServer") {
		t.Fatalf("service file is missing embedded unimplemented server")
	}
	if !strings.Contains(serviceFile, "UserServiceHTTPServer") {
		t.Fatalf("service file is missing embedded HTTP server")
	}
	if !strings.Contains(serviceFile, "func NewUserService() *UserService") {
		t.Fatalf("service file is missing constructor")
	}

	repoFile := readFile(t, filepath.Join(root, "app", "admin", "service", "internal", "data", "user_repo.gen.go"))
	if !strings.Contains(repoFile, "func NewUserRepo") {
		t.Fatalf("repo file is missing constructor")
	}
	if !strings.Contains(repoFile, "entCrud.Repository") {
		t.Fatalf("repo file is missing ent CRUD repository")
	}
	if !strings.Contains(repoFile, "type UserRepo interface") {
		t.Fatalf("repo file is missing repo interface")
	}
	if !strings.Contains(repoFile, "func (r *userRepo) List") {
		t.Fatalf("repo file is missing List method skeleton")
	}

	serviceWireFile := readFile(t, filepath.Join(root, "app", "admin", "service", "internal", "service", "providers", "wire_set.gen.go"))
	if !strings.Contains(serviceWireFile, "var ProviderSet = wire.NewSet") {
		t.Fatalf("service wire file is missing ProviderSet")
	}

	registerFile := readFile(t, filepath.Join(root, "app", "admin", "service", "internal", "server", "grpc_register.gen.go"))
	if !strings.Contains(registerFile, "RegisterUserServiceServer") {
		t.Fatalf("grpc register file is missing register call")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
