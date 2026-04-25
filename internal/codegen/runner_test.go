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
  rpc Count (CountRequest) returns (CountResponse) {}
  rpc Create (CreateUserRequest) returns (User) {}
  rpc Delete (DeleteUserRequest) returns (Empty) {}
  rpc UserExists (UserExistsRequest) returns (UserExistsResponse) {}
  rpc EditUserPassword (EditUserPasswordRequest) returns (Empty) {}
}`)
	writeFile(t, filepath.Join(root, "api", "gen", "admin", "v1", "i_user_grpc.pb.go"), `package admin

import (
	context "context"
	v1 "example.com/xadmin-web/api/gen/identity/v1"
)

type UserServiceServer interface {
	List(context.Context, *v1.ListRequest) (*v1.ListResponse, error)
	Get(context.Context, *v1.GetRequest) (*v1.GetResponse, error)
	Count(context.Context, *v1.CountRequest) (*v1.CountResponse, error)
	Create(context.Context, *v1.CreateUserRequest) (*v1.User, error)
	Delete(context.Context, *v1.DeleteUserRequest) (*v1.Empty, error)
	UserExists(context.Context, *v1.UserExistsRequest) (*v1.UserExistsResponse, error)
	EditUserPassword(context.Context, *v1.EditUserPasswordRequest) (*v1.Empty, error)
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
		field.Enum("status").Optional().Nillable(),
		field.Time("last_login_at").Optional().Nillable(),
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
				ExistsFields:  []string{"id", "username"},
				Filters: config.FilterConfig{
					Allow: []string{"id", "username", "status"},
				},
				Operations: config.OperationFlags{
					"list":   true,
					"get":    true,
					"count":  true,
					"create": true,
					"update": true,
					"delete": true,
					"exists": true,
				},
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
	}, cfg, Options{Version: "test-version"})
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
		filepath.Join(root, "internal", "service", "user_service.gen.go"),
		filepath.Join(root, "internal", "service", "user_service_ext.go"),
		filepath.Join(root, "internal", "data", "user_repo.gen.go"),
		filepath.Join(root, "internal", "server", "rest_register.gen.go"),
		filepath.Join(root, "internal", "server", "grpc_register.gen.go"),
		filepath.Join(root, "internal", "service", "providers", "wire_set.gen.go"),
		filepath.Join(root, "internal", "data", "providers", "wire_set.gen.go"),
	}
	for _, path := range expectedPaths {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected generated file %s: %v", path, err)
		}
	}

	serviceFile := readFile(t, filepath.Join(root, "internal", "service", "user_service.gen.go"))
	if !strings.Contains(serviceFile, "// xkit version: test-version") || !strings.Contains(serviceFile, "// generated at:") {
		t.Fatalf("service file is missing generated metadata header")
	}
	if !strings.Contains(serviceFile, "UnimplementedUserServiceServer") {
		t.Fatalf("service file is missing embedded unimplemented server")
	}
	if !strings.Contains(serviceFile, "UserServiceHTTPServer") {
		t.Fatalf("service file is missing embedded HTTP server")
	}
	if !strings.Contains(serviceFile, "func NewUserService(ctx *app.AppCtx, userRepo data.UserRepo) *UserService") {
		t.Fatalf("service file is missing repo-injected constructor")
	}
	if !strings.Contains(serviceFile, "log *log.Helper") || !strings.Contains(serviceFile, "userRepo data.UserRepo") {
		t.Fatalf("service file is missing log or repo fields")
	}
	if !strings.Contains(serviceFile, "return s.userRepo.List(ctx, req)") || !strings.Contains(serviceFile, "return s.userRepo.UserExists(ctx, req)") {
		t.Fatalf("service file is missing CRUD repo delegation")
	}
	if strings.Contains(serviceFile, "not implemented") || strings.Contains(serviceFile, "MethodNotImplemented") {
		t.Fatalf("service file should not contain not implemented stubs")
	}
	if !strings.Contains(serviceFile, "TODO: implement UserService.EditUserPassword business logic manually") {
		t.Fatalf("service file is missing TODO for manual business logic")
	}

	repoFile := readFile(t, filepath.Join(root, "internal", "data", "user_repo.gen.go"))
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
	if !strings.Contains(repoFile, "entities, err := builder.Limit(limit).Offset(int(req.GetOffset())).All(ctx)") {
		t.Fatalf("repo file is missing generated List body")
	}
	if !strings.Contains(repoFile, "entity, err := builder.Where(user.IDEQ(req.GetId())).Only(ctx)") {
		t.Fatalf("repo file is missing generated Get body")
	}
	if !strings.Contains(repoFile, "func (r *userRepo) Count") || !strings.Contains(repoFile, "total, err := builder.Count(ctx)") || !strings.Contains(repoFile, "return &identityv1.CountResponse{Count: uint64(total)}, nil") {
		t.Fatalf("repo file is missing generated Count body")
	}
	if !strings.Contains(repoFile, "func (r *userRepo) applyGeneratedFilters") || !strings.Contains(repoFile, "case \"username\":") || !strings.Contains(repoFile, "builder.Where(user.UsernameContains(condition.GetValue()))") {
		t.Fatalf("repo file is missing generated string filters")
	}
	if !strings.Contains(repoFile, "case \"id\":") || !strings.Contains(repoFile, "builder.Where(user.IDEQ(uint32(value)))") {
		t.Fatalf("repo file is missing generated id filter")
	}
	if !strings.Contains(repoFile, "case \"status\":") || !strings.Contains(repoFile, "builder.Where(user.StatusEQ(user.Status(condition.GetValue())))") {
		t.Fatalf("repo file is missing generated enum filter")
	}
	if !strings.Contains(repoFile, "func (r *userRepo) Create") {
		t.Fatalf("repo file is missing Create method")
	}
	if !strings.Contains(repoFile, "builder.SetNillableUsername(req.Data.Username)") {
		t.Fatalf("repo file is missing generated username setter")
	}
	if strings.Contains(repoFile, "SetNillableCreatedAt") || strings.Contains(repoFile, "SetNillableDeletedAt") || strings.Contains(repoFile, "SetNillableDeletedBy") {
		t.Fatalf("repo file should not generate unsafe audit/delete setters")
	}
	if !strings.Contains(repoFile, "func userEnumPtrFromProto") || !strings.Contains(repoFile, "builder.SetNillableStatus(userEnumPtrFromProto[user.Status](req.Data.Status))") {
		t.Fatalf("repo file is missing generated enum setter conversion")
	}
	if !strings.Contains(repoFile, "func userTimePtrFromProto") || !strings.Contains(repoFile, "builder.SetNillableLastLoginAt(userTimePtrFromProto(req.Data.LastLoginAt))") {
		t.Fatalf("repo file is missing generated time setter conversion")
	}
	if strings.Contains(repoFile, "*v1.") {
		t.Fatalf("repo file contains unnormalized generated alias: %s", repoFile)
	}
	if !strings.Contains(repoFile, "DeleteOneID(req.GetId()).Exec(ctx)") {
		t.Fatalf("repo file is missing generated delete body")
	}
	if !strings.Contains(repoFile, "case *identityv1.UserExistsRequest_Id:") || !strings.Contains(repoFile, "builder.Where(user.IDEQ(req.GetId()))") || !strings.Contains(repoFile, "builder.Where(user.UsernameEQ(req.GetUsername()))") {
		t.Fatalf("repo file is missing generated query_by exists body")
	}

	serviceWireFile := readFile(t, filepath.Join(root, "internal", "service", "providers", "wire_set.gen.go"))
	if !strings.Contains(serviceWireFile, "var ProviderSet = wire.NewSet") {
		t.Fatalf("service wire file is missing ProviderSet")
	}
	if !strings.Contains(serviceWireFile, "servicepkg.NewUserService") {
		t.Fatalf("service wire file is missing service constructor")
	}

	dataWireFile := readFile(t, filepath.Join(root, "internal", "data", "providers", "wire_set.gen.go"))
	if !strings.Contains(dataWireFile, "datapkg.NewUserRepo") {
		t.Fatalf("data wire file is missing repo constructor")
	}

	registerFile := readFile(t, filepath.Join(root, "internal", "server", "grpc_register.gen.go"))
	if !strings.Contains(registerFile, "RegisterUserServiceServer") {
		t.Fatalf("grpc register file is missing register call")
	}
}

func TestResourceOperationEnabled(t *testing.T) {
	t.Parallel()

	resource := config.Resource{
		Operations: config.OperationFlags{
			"list":   true,
			"create": false,
			"exists": true,
		},
	}

	if !resourceOperationEnabled(resource, "list") {
		t.Fatalf("list should be enabled")
	}
	if resourceOperationEnabled(resource, "create") {
		t.Fatalf("create should be disabled")
	}
	if resourceOperationEnabled(resource, "delete") {
		t.Fatalf("missing operation should be disabled when operations are configured")
	}
	if !resourceOperationEnabled(resource, "query_exists") {
		t.Fatalf("query_exists should map to exists")
	}
	if !resourceOperationEnabled(config.Resource{}, "delete") {
		t.Fatalf("empty operations should keep backward-compatible generation enabled")
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
