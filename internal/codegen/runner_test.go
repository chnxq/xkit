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
  rpc Update (UpdateUserRequest) returns (User) {}
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
	Update(context.Context, *v1.UpdateUserRequest) (*v1.User, error)
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
	writeFile(t, filepath.Join(root, "api", "protos", "authentication", "v1", "user_credential.proto"), `syntax = "proto3";

package authentication.service.v1;

service UserCredentialService {
  rpc ResetCredential (ResetCredentialRequest) returns (Empty) {}
}`)
	writeFile(t, filepath.Join(root, "api", "gen", "authentication", "v1", "user_credential_grpc.pb.go"), `package authenticationv1

import context "context"

type UserCredentialServiceServer interface {
	ResetCredential(context.Context, *ResetCredentialRequest) (*Empty, error)
}

var UserCredentialService_ServiceDesc = struct{
	ServiceName string
}{
	ServiceName: "authentication.service.v1.UserCredentialService",
}

type ResetCredentialRequest struct{}
type UserCredential struct{}
type Empty struct{}
`)
	writeFile(t, filepath.Join(root, "internal", "data", "ent", "schema", "user_credential.go"), `package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type UserCredential struct { ent.Schema }

func (UserCredential) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("user_id").Optional().Nillable(),
		field.Enum("identity_type").Optional().Nillable(),
		field.String("identifier").Optional().Nillable(),
		field.Enum("credential_type").Optional().Nillable(),
		field.String("credential").Optional().Nillable(),
		field.Enum("status").Optional().Nillable(),
	}
}
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
		field.Time("created_at").Optional().Nillable().Immutable(),
		field.Time("updated_at").Optional().Nillable(),
		field.Uint32("created_by").Optional().Nillable(),
		field.Uint32("updated_by").Optional().Nillable(),
		field.Time("deleted_at").Optional().Nillable(),
		field.Uint32("deleted_by").Optional().Nillable(),
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
				ServiceMethods: map[string]config.ServiceMethodConfig{
					"EditUserPassword": {
						Imports: []config.ImportConfig{
							{Alias: "authenticationv1", Path: "{{module}}/api/gen/authentication/v1"},
						},
						Repos: []config.RepoConfig{
							{Field: "userCredentialRepo", Interface: "UserCredentialRepo"},
						},
						Body: `user, err := s.{{repoField}}.Get({{ctx}}, &v1.GetUserRequest{
	QueryBy: &v1.GetUserRequest_Id{
		Id: {{param.req}}.GetUserId(),
	},
})
if err != nil {
	return nil, err
}

if _, err = s.userCredentialRepo.ResetCredential({{ctx}}, &authenticationv1.ResetCredentialRequest{
	IdentityType:  authenticationv1.UserCredential_USERNAME,
	Identifier:    user.GetUsername(),
	NewCredential: {{param.req}}.GetNewPassword(),
	NeedDecrypt:   false,
}); err != nil {
	s.log.Errorf("reset user password err: %v", err)
	return nil, err
}

return {{successReturn}}, nil`,
					},
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
			{
				Name:          "user_credential",
				ProtoService:  "authentication.service.v1.UserCredentialService",
				Entity:        "UserCredential",
				DTOImport:     "example.com/xadmin-web/api/gen/authentication/v1",
				DTOType:       "UserCredential",
				RepoInterface: "UserCredentialRepo",
				Filters: config.FilterConfig{
					Allow: []string{"id", "identity_type", "identifier"},
				},
				Operations: config.OperationFlags{
					"resetcredential": true,
				},
				Generate: config.GenerateFlags{
					RepoCRUD: true,
					Wire:     true,
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

	if len(result.Written) != 10 {
		t.Fatalf("written file count mismatch: got %d want %d", len(result.Written), 10)
	}

	expectedPaths := []string{
		filepath.Join(root, "internal", "service", "user_service.gen.go"),
		filepath.Join(root, "internal", "service", "user_service_ext.go"),
		filepath.Join(root, "internal", "data", "repo", "user_repo.gen.go"),
		filepath.Join(root, "internal", "data", "repo", "user_repo_ext.go"),
		filepath.Join(root, "internal", "data", "repo", "user_credential_repo_ext.go"),
		filepath.Join(root, "internal", "server", "rest_register.gen.go"),
		filepath.Join(root, "internal", "server", "grpc_register.gen.go"),
		filepath.Join(root, "internal", "bootstrap", "generated_servers.gen.go"),
		filepath.Join(root, "internal", "data", "bootstrap", "ent_client.gen.go"),
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
	if !strings.Contains(serviceFile, "func NewUserService(ctx *app.AppCtx, userRepo repo.UserRepo, userCredentialRepo repo.UserCredentialRepo) *UserService") {
		t.Fatalf("service file is missing repo-injected constructor")
	}
	if !strings.Contains(serviceFile, "log *log.Helper") || !strings.Contains(serviceFile, "userRepo repo.UserRepo") || !strings.Contains(serviceFile, "userCredentialRepo repo.UserCredentialRepo") {
		t.Fatalf("service file is missing log or repo fields")
	}
	if !strings.Contains(serviceFile, "return s.userRepo.List(ctx, req)") || !strings.Contains(serviceFile, "return s.userRepo.UserExists(ctx, req)") {
		t.Fatalf("service file is missing CRUD repo delegation")
	}
	if strings.Contains(serviceFile, "not implemented") || strings.Contains(serviceFile, "MethodNotImplemented") {
		t.Fatalf("service file should not contain not implemented stubs")
	}
	if !strings.Contains(serviceFile, "s.userRepo.Get(ctx,") || !strings.Contains(serviceFile, "GetUserRequest{") {
		t.Fatalf("service file is missing generated EditUserPassword user lookup")
	}
	if !strings.Contains(serviceFile, "userCredentialRepo repo.UserCredentialRepo") || !strings.Contains(serviceFile, "s.userCredentialRepo.ResetCredential") {
		t.Fatalf("service file is missing generated EditUserPassword credential reset")
	}

	repoFile := readFile(t, filepath.Join(root, "internal", "data", "repo", "user_repo.gen.go"))
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
	if strings.Contains(repoFile, "SetNillableCreatedAt") || strings.Contains(repoFile, "SetNillableUpdatedAt") || strings.Contains(repoFile, "SetNillableCreatedBy") || strings.Contains(repoFile, "SetNillableUpdatedBy") || strings.Contains(repoFile, "SetNillableDeletedAt") || strings.Contains(repoFile, "SetNillableDeletedBy") {
		t.Fatalf("repo file should not generate unsafe audit/delete setters")
	}
	if !strings.Contains(repoFile, `"github.com/chnxq/x-crud/viewer"`) || !strings.Contains(repoFile, "func (r *userRepo) generatedAuditContext(ctx context.Context) (time.Time, crudviewer.Context)") {
		t.Fatalf("repo file is missing generated audit context helper")
	}
	if !strings.Contains(repoFile, "now, viewer := r.generatedAuditContext(ctx)") || !strings.Contains(repoFile, "builder.SetCreatedAt(now)") || !strings.Contains(repoFile, "builder.SetUpdatedAt(now)") {
		t.Fatalf("repo file is missing generated audit time setters")
	}
	if !strings.Contains(repoFile, "builder.SetCreatedBy(uint32(viewer.UserID()))") || !strings.Contains(repoFile, "builder.SetUpdatedBy(uint32(viewer.UserID()))") {
		t.Fatalf("repo file is missing generated audit user setters")
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
	if !strings.Contains(repoFile, "case interface{ GetId() uint32 }:") || !strings.Contains(repoFile, "DeleteOneID(typedReq.GetId()).Exec(ctx)") {
		t.Fatalf("repo file is missing generated single-id delete body")
	}
	if !strings.Contains(repoFile, "case interface{ GetIds() []uint32 }:") || !strings.Contains(repoFile, "IDIn(typedReq.GetIds()...)") {
		t.Fatalf("repo file is missing generated multi-id delete body")
	}
	if !strings.Contains(repoFile, "case *identityv1.UserExistsRequest_Id:") || !strings.Contains(repoFile, "builder.Where(user.IDEQ(req.GetId()))") || !strings.Contains(repoFile, "builder.Where(user.UsernameEQ(req.GetUsername()))") {
		t.Fatalf("repo file is missing generated query_by exists body")
	}

	credentialRepoFile := readFile(t, filepath.Join(root, "internal", "data", "repo", "user_credential_repo.gen.go"))
	if !strings.Contains(credentialRepoFile, "func (r *userCredentialRepo) ResetCredential") || !strings.Contains(credentialRepoFile, "SetCredential(credential)") {
		t.Fatalf("credential repo file is missing generated ResetCredential body")
	}
	if strings.Contains(credentialRepoFile, "IDentity") || strings.Contains(credentialRepoFile, "IDentifier") {
		t.Fatalf("credential repo file has broken initialism conversion")
	}

	if _, err := os.Stat(filepath.Join(root, "internal", "service", "providers", "wire_set.gen.go")); !os.IsNotExist(err) {
		t.Fatalf("gen all should not generate service wire provider sets, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "internal", "data", "providers", "wire_set.gen.go")); !os.IsNotExist(err) {
		t.Fatalf("gen all should not generate data wire provider sets, stat err=%v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "configs", "server.yaml")); !os.IsNotExist(err) {
		t.Fatalf("gen all should not create static config files, stat err=%v", err)
	}

	bootstrapFile := readFile(t, filepath.Join(root, "internal", "bootstrap", "generated_servers.gen.go"))
	if !strings.Contains(bootstrapFile, "func NewGeneratedServers") || strings.Contains(bootstrapFile, "func Initialize") {
		t.Fatalf("bootstrap generation should only write generated server glue")
	}
	if !strings.Contains(bootstrapFile, "type GeneratedData struct") || !strings.Contains(bootstrapFile, "type GeneratedServices struct") || !strings.Contains(bootstrapFile, "func NewGeneratedComponents") {
		t.Fatalf("bootstrap generation should split data, services, and component assembly")
	}
	if !strings.Contains(bootstrapFile, "httpServer, err := server.NewHTTPServer(appCtx, components.Services.HTTP(), components.Data)") || !strings.Contains(bootstrapFile, "new generated grpc server") {
		t.Fatalf("bootstrap generation should handle transport constructor errors")
	}
	if strings.Contains(bootstrapFile, "UserCredential:") {
		t.Fatalf("bootstrap generation should not register resources without generated service stubs")
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

func TestGeneratedEntNamesUseInitialismsWhereEntDoes(t *testing.T) {
	t.Parallel()

	if got := entOperationName("Api"); got != "API" {
		t.Fatalf("entOperationName(Api) = %q, want API", got)
	}
	if got := entOperationName("ApiAuditLog"); got != "ApiAuditLog" {
		t.Fatalf("entOperationName(ApiAuditLog) = %q, want ApiAuditLog", got)
	}
	if got := toGoName("http_method"); got != "HTTPMethod" {
		t.Fatalf("toGoName(http_method) = %q, want HTTPMethod", got)
	}
	if got := toGoName("api_module"); got != "APIModule" {
		t.Fatalf("toGoName(api_module) = %q, want APIModule", got)
	}
	if got := toGoName("sql_digest"); got != "SQLDigest" {
		t.Fatalf("toGoName(sql_digest) = %q, want SQLDigest", got)
	}
	if got := toGoName("file_guid"); got != "FileGUID" {
		t.Fatalf("toGoName(file_guid) = %q, want FileGUID", got)
	}
	if got := toPascal("api_module"); got != "ApiModule" {
		t.Fatalf("toPascal(api_module) = %q, want ApiModule", got)
	}
	if got := filterCastType("Int32"); got != "int32" {
		t.Fatalf("filterCastType(Int32) = %q, want int32", got)
	}
	if got := filterParseBitSize("Int32"); got != "32" {
		t.Fatalf("filterParseBitSize(Int32) = %q, want 32", got)
	}
	if supportsGeneratedSetterKind("Strings") {
		t.Fatalf("Strings fields should be left for manual conversion")
	}
}

func TestServiceMethodBodyWithoutConfigFallsBackToTODO(t *testing.T) {
	t.Parallel()

	runner := &Runner{}
	body := runner.serviceMethodBody(resourcePlan{
		Resource: config.Resource{Name: "user"},
	}, "EditUserPassword", []namedType{
		{Name: "ctx", Type: "context.Context"},
		{Name: "req", Type: "*identityv1.EditUserPasswordRequest"},
	}, "*emptypb.Empty", "userRepo", true)

	if body != "" {
		t.Fatalf("service method body should be empty without service_methods config, got %q", body)
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
