package codegen

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chnxq/xkit/internal/binding"
	"github.com/chnxq/xkit/internal/config"
	"github.com/chnxq/xkit/internal/entschema"
	"github.com/chnxq/xkit/internal/project"
	xproto "github.com/chnxq/xkit/internal/proto"
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
	writeFile(t, filepath.Join(root, "api", "gen", "identity", "v1", "user.pb.go"), `package identityv1

type User struct {
	Username    *string
	Status      *User_Status
	LastLoginAt *Timestamp
	ParentId    *uint32
	Path        *string
	Meta        *UserMeta
}

type UserMeta struct {
	Title *string
}

type User_Status string

func (x *User) GetUsername() string {
	if x != nil && x.Username != nil {
		return *x.Username
	}
	return ""
}

func (x *User) GetStatus() User_Status {
	if x != nil && x.Status != nil {
		return *x.Status
	}
	return ""
}

func (x *User_Status) String() string {
	if x == nil {
		return ""
	}
	return string(*x)
}

func (x *User) GetLastLoginAt() *Timestamp {
	if x != nil && x.LastLoginAt != nil {
		return x.LastLoginAt
	}
	return &Timestamp{}
}

func (x *User) GetParentId() uint32 {
	if x != nil && x.ParentId != nil {
		return *x.ParentId
	}
	return 0
}

func (x *User) GetPath() string {
	if x != nil && x.Path != nil {
		return *x.Path
	}
	return ""
}

func (x *User) GetMeta() *UserMeta {
	if x != nil {
		return x.Meta
	}
	return nil
}

type Timestamp struct{}

func (*Timestamp) AsTime() Time { return Time{} }

type Time struct{}
`)
	writeFile(t, filepath.Join(root, "internal", "data", "ent", "schema", "user_credential.go"), `package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/chnxq/x-crud/entgo/mixin"
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
			field.JSON("meta", &struct{}{}).Optional(),
			field.Time("created_at").Optional().Nillable().Immutable(),
			field.Time("updated_at").Optional().Nillable(),
			field.Uint32("created_by").Optional().Nillable(),
		field.Uint32("updated_by").Optional().Nillable(),
		field.Time("deleted_at").Optional().Nillable(),
		field.Uint32("deleted_by").Optional().Nillable(),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Tree[User]{},
		mixin.TreePath{},
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
				TenantScope:   "tenant_scoped",
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

	if len(result.Written) != 13 {
		t.Fatalf("written file count mismatch: got %d want %d", len(result.Written), 13)
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
		filepath.Join(root, "internal", "bootstrap", "generated_data_providers.gen.go"),
		filepath.Join(root, "internal", "bootstrap", "generated_hooks_ext.go"),
		filepath.Join(root, "internal", "data", "bootstrap", "ent_client.gen.go"),
		filepath.Join(root, "internal", "data", "bootstrap", "ent_client_ext.go"),
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
	if !strings.Contains(repoFile, "if tenantID := viewerTenantID(ctx); tenantID != nil {") || !strings.Contains(repoFile, "builder.Where(user.TenantIDEQ(*tenantID))") {
		t.Fatalf("repo file is missing generated tenant-scoped list filter")
	}
	if !strings.Contains(repoFile, `listReq.Sorting = withDefaultSorting(`) || !strings.Contains(repoFile, `"id"`) || !strings.Contains(repoFile, `paginationv1.Sorting_ASC`) {
		t.Fatalf("repo file is missing shared default id ascending sorting")
	}
	if !strings.Contains(repoFile, "if _, _, err := r.repository.BuildListSelectorWithPaging(builder, listReq); err != nil") || !strings.Contains(repoFile, "entities, err := builder.All(ctx)") {
		t.Fatalf("repo file is missing generated List body")
	}
	if !strings.Contains(repoFile, "userEnrichListDTOs(context.Context, []*ent.User) ([]*identityv1.User, error)") || !strings.Contains(repoFile, "custom.userEnrichListDTOs(ctx, entities)") {
		t.Fatalf("repo file is missing optional list enrichment hook")
	}
	if !strings.Contains(repoFile, "entity, err := builder.Where(user.IDEQ(req.GetId())).Only(ctx)") {
		t.Fatalf("repo file is missing generated Get body")
	}
	if !strings.Contains(repoFile, "if err := ensureTenantAccessible(ctx, entity.TenantID); err != nil {") {
		t.Fatalf("repo file is missing generated tenant-scoped get guard")
	}
	if !strings.Contains(repoFile, "userEnrichGetDTO(context.Context, []*ent.User) ([]*identityv1.User, error)") || !strings.Contains(repoFile, "custom.userEnrichGetDTO(ctx, []*ent.User{entity})") {
		t.Fatalf("repo file is missing optional get enrichment hook")
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
	if !strings.Contains(repoFile, "userCustomCreate(ctx context.Context, req *identityv1.CreateUserRequest)") || !strings.Contains(repoFile, "return custom.userCustomCreate(ctx, req)") {
		t.Fatalf("repo file is missing optional create hook")
	}
	if !strings.Contains(repoFile, "userCustomUpdate(ctx context.Context, req *identityv1.UpdateUserRequest)") || !strings.Contains(repoFile, "return custom.userCustomUpdate(ctx, req)") {
		t.Fatalf("repo file is missing optional update hook")
	}
	if !strings.Contains(repoFile, "userCustomDelete(ctx context.Context, req *identityv1.DeleteUserRequest)") || !strings.Contains(repoFile, "return custom.userCustomDelete(ctx, req)") {
		t.Fatalf("repo file is missing optional delete hook")
	}
	if !strings.Contains(repoFile, "builder.SetNillableUsername(req.Data.Username)") {
		t.Fatalf("repo file is missing generated username setter")
	}
	if !strings.Contains(repoFile, "builder.SetNillableParentID(req.Data.ParentId)") || !strings.Contains(repoFile, "builder.SetNillablePath(req.Data.Path)") {
		t.Fatalf("repo file is missing generated tree mixin setters")
	}
	if !strings.Contains(repoFile, "if req.Data.Meta != nil {") || !strings.Contains(repoFile, "builder.SetMeta(req.Data.GetMeta())") || !strings.Contains(repoFile, "builder.ClearMeta()") {
		t.Fatalf("repo file is missing generated JSON field setter and clear logic")
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
	if !strings.Contains(repoFile, "get user before delete failed") || !strings.Contains(repoFile, "ensureTenantAccessible(ctx, entity.TenantID)") {
		t.Fatalf("repo file is missing generated tenant-scoped delete guard")
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
	if !strings.Contains(bootstrapFile, "AppContext") || !strings.Contains(bootstrapFile, "*app.AppCtx") || !strings.Contains(bootstrapFile, "data := &GeneratedData{AppContext: appCtx}") {
		t.Fatalf("bootstrap generation should retain app context on generated data")
	}
	if !strings.Contains(bootstrapFile, "data.afterInit()") || !strings.Contains(bootstrapFile, "services.afterInit(data)") {
		t.Fatalf("bootstrap generation should call generated bootstrap hooks")
	}
	if !strings.Contains(bootstrapFile, "httpServer, err := server.NewHTTPServer(appCtx, components.Services.HTTP(), components.Data)") || !strings.Contains(bootstrapFile, "grpcServer, err := server.NewGRPCServer(appCtx, components.Services.GRPC(), components.Data)") || !strings.Contains(bootstrapFile, "new generated grpc server") {
		t.Fatalf("bootstrap generation should handle transport constructor errors")
	}
	if strings.Contains(bootstrapFile, "UserCredential:") {
		t.Fatalf("bootstrap generation should not register resources without generated service stubs")
	}
	hooksFile := readFile(t, filepath.Join(root, "internal", "bootstrap", "generated_hooks_ext.go"))
	if !strings.Contains(hooksFile, "func (data *GeneratedData) afterInit() {}") || !strings.Contains(hooksFile, "func (services *GeneratedServices) afterInit(data *GeneratedData)") {
		t.Fatalf("bootstrap hooks extension file is missing generated hook stubs")
	}

	entHooksFile := readFile(t, filepath.Join(root, "internal", "data", "bootstrap", "ent_client_ext.go"))
	if !strings.Contains(entHooksFile, "func afterEntSchemaCreate(ctx *app.AppCtx, entClient *entCrud.EntClient[*ent.Client]) error") {
		t.Fatalf("ent client extension file is missing schema-create hook stub")
	}

	providersFile := readFile(t, filepath.Join(root, "internal", "bootstrap", "generated_data_providers.gen.go"))
	if !strings.Contains(providersFile, `"github.com/chnxq/xkitpkg/app"`) || !strings.Contains(providersFile, "func (data *GeneratedData) GetAppCtx() *app.AppCtx") {
		t.Fatalf("bootstrap provider file is missing GetAppCtx")
	}
	if !strings.Contains(providersFile, `"example.com/xadmin-web/internal/data/repo"`) || !strings.Contains(providersFile, "func (data *GeneratedData) UserRepoProvider() repo.UserRepo") || !strings.Contains(providersFile, "return data.UserRepo") {
		t.Fatalf("bootstrap provider file is missing generated repo provider")
	}
	if !strings.Contains(providersFile, "func (data *GeneratedData) UserCredentialRepoProvider() repo.UserCredentialRepo") {
		t.Fatalf("bootstrap provider file should generate providers for extra repos")
	}

	registerFile := readFile(t, filepath.Join(root, "internal", "server", "grpc_register.gen.go"))
	if !strings.Contains(registerFile, "RegisterUserServiceServer") {
		t.Fatalf("grpc register file is missing register call")
	}
}

func TestRunnerGenerateBootstrap_SkipsExistingGeneratedDataMethods(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/xadmin-web\n\ngo 1.26.0\n")
	writeFile(t, filepath.Join(root, "api", "protos", "admin", "v1", "i_user.proto"), `syntax = "proto3";

package admin.service.v1;

service UserService {
  rpc List (ListRequest) returns (ListResponse) {}
}`)
	writeFile(t, filepath.Join(root, "api", "gen", "admin", "v1", "i_user_grpc.pb.go"), `package admin

import (
	context "context"
	v1 "example.com/xadmin-web/api/gen/identity/v1"
)

type UserServiceServer interface {
	List(context.Context, *v1.ListRequest) (*v1.ListResponse, error)
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
	writeFile(t, filepath.Join(root, "api", "gen", "identity", "v1", "user.pb.go"), `package identityv1

type User struct{}
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
	writeFile(t, filepath.Join(root, "internal", "bootstrap", "generated_data_ext.go"), `package bootstrap

import (
	"example.com/xadmin-web/internal/data/repo"
	"github.com/chnxq/xkitpkg/app"
)

func (data *GeneratedData) GetAppCtx() *app.AppCtx { return nil }
func (data *GeneratedData) UserRepoProvider() repo.UserRepo { return nil }
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
				Operations: config.OperationFlags{
					"list": true,
				},
				Generate: config.GenerateFlags{
					ServiceStub:  true,
					RepoCRUD:     true,
					RestRegister: true,
					GRPCRegister: true,
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

	if _, err := runner.Generate("bootstrap"); err != nil {
		t.Fatalf("generate bootstrap: %v", err)
	}

	providersFile := readFile(t, filepath.Join(root, "internal", "bootstrap", "generated_data_providers.gen.go"))
	if strings.Contains(providersFile, "func (data *GeneratedData) GetAppCtx() *app.AppCtx") {
		t.Fatalf("bootstrap provider file should skip existing GetAppCtx")
	}
	if strings.Contains(providersFile, "func (data *GeneratedData) UserRepoProvider() repo.UserRepo") {
		t.Fatalf("bootstrap provider file should skip existing repo provider")
	}
}

func TestRunnerGenerateBootstrap_PreservesHandwrittenBootstrapExtBodies(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/xadmin-web\n\ngo 1.26.0\n")
	writeFile(t, filepath.Join(root, "api", "protos", "admin", "v1", "i_user.proto"), `syntax = "proto3";

package admin.service.v1;

service UserService {
  rpc List (ListRequest) returns (ListResponse) {}
}`)
	writeFile(t, filepath.Join(root, "api", "gen", "admin", "v1", "i_user_grpc.pb.go"), `package admin

import (
	context "context"
	v1 "example.com/xadmin-web/api/gen/identity/v1"
)

type UserServiceServer interface {
	List(context.Context, *v1.ListRequest) (*v1.ListResponse, error)
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
	writeFile(t, filepath.Join(root, "api", "gen", "identity", "v1", "user.pb.go"), `package identityv1

type User struct{}
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
	writeFile(t, filepath.Join(root, "internal", "bootstrap", "generated_hooks_ext.go"), `// Code generated by xkit. DO NOT EDIT.
// legacy header

package bootstrap

func (data *GeneratedData) afterInit() {
	data.WrapAuditLogRepos()
}

func (services *GeneratedServices) afterInit(data *GeneratedData) {
	_, _ = services, data
}
`)
	writeFile(t, filepath.Join(root, "internal", "data", "bootstrap", "ent_client_ext.go"), `// Code generated by xkit. DO NOT EDIT.
// legacy header

package bootstrap

import (
	entCrud "github.com/chnxq/x-crud/entgo"
	"github.com/chnxq/xkitpkg/app"

	"example.com/xadmin-web/internal/data/ent"
)

func afterEntSchemaCreate(ctx *app.AppCtx, entClient *entCrud.EntClient[*ent.Client]) error {
	return ensureDefaultData(ctx, entClient)
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
				Operations: config.OperationFlags{
					"list": true,
				},
				Generate: config.GenerateFlags{
					ServiceStub:  true,
					RepoCRUD:     true,
					RestRegister: true,
					GRPCRegister: true,
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

	if _, err := runner.Generate("bootstrap"); err != nil {
		t.Fatalf("generate bootstrap: %v", err)
	}

	hooksFile := readFile(t, filepath.Join(root, "internal", "bootstrap", "generated_hooks_ext.go"))
	if !strings.Contains(hooksFile, "func (data *GeneratedData) afterInit() {\n\tdata.WrapAuditLogRepos()\n}") {
		t.Fatalf("bootstrap hooks ext body should be preserved, got:\n%s", hooksFile)
	}
	if strings.Contains(hooksFile, "DO NOT EDIT") {
		t.Fatalf("bootstrap hooks ext header should be refreshed, got:\n%s", hooksFile)
	}

	entHooksFile := readFile(t, filepath.Join(root, "internal", "data", "bootstrap", "ent_client_ext.go"))
	if !strings.Contains(entHooksFile, "return ensureDefaultData(ctx, entClient)") {
		t.Fatalf("ent client ext body should be preserved, got:\n%s", entHooksFile)
	}
	if strings.Contains(entHooksFile, "DO NOT EDIT") {
		t.Fatalf("ent client ext header should be refreshed, got:\n%s", entHooksFile)
	}
}

func TestDefaultListSortFieldPrefersSortOrder(t *testing.T) {
	t.Parallel()

	plan := resourcePlan{
		Resource: config.Resource{Name: "role", Entity: "Role"},
		Schema: entschema.Schema{Fields: []entschema.Field{
			{Name: "id"},
			{Name: "name"},
			{Name: "sort_order"},
		}},
	}

	if got := defaultListSortField(plan); got != "sort_order" {
		t.Fatalf("defaultListSortField() = %q, want sort_order", got)
	}
	if got := defaultListSortDirection(plan); got != "ASC" {
		t.Fatalf("defaultListSortDirection() = %q, want ASC", got)
	}
}

func TestDefaultListSortFieldFallsBackToID(t *testing.T) {
	t.Parallel()

	plan := resourcePlan{
		Resource: config.Resource{Name: "user", Entity: "User"},
		Schema: entschema.Schema{Fields: []entschema.Field{
			{Name: "id"},
			{Name: "name"},
			{Name: "created_at"},
		}},
	}

	if got := defaultListSortField(plan); got != "id" {
		t.Fatalf("defaultListSortField() = %q, want id", got)
	}
	if got := defaultListSortDirection(plan); got != "ASC" {
		t.Fatalf("defaultListSortDirection() = %q, want ASC", got)
	}
}

func TestDefaultListSortForLogUsesCreatedAtDesc(t *testing.T) {
	t.Parallel()

	plan := resourcePlan{
		Resource: config.Resource{Name: "api_audit_log", Entity: "ApiAuditLog"},
		Schema: entschema.Schema{Fields: []entschema.Field{
			{Name: "id"},
			{Name: "created_at"},
		}},
	}

	if got := defaultListSortField(plan); got != "created_at" {
		t.Fatalf("defaultListSortField() = %q, want created_at", got)
	}
	if got := defaultListSortDirection(plan); got != "DESC" {
		t.Fatalf("defaultListSortDirection() = %q, want DESC", got)
	}
}

func TestDefaultListSortForTaskUsesIDDesc(t *testing.T) {
	t.Parallel()

	plan := resourcePlan{
		Resource: config.Resource{Name: "task", Entity: "Task"},
		Schema: entschema.Schema{Fields: []entschema.Field{
			{Name: "id"},
			{Name: "created_at"},
			{Name: "name"},
		}},
	}

	if got := defaultListSortField(plan); got != "id" {
		t.Fatalf("defaultListSortField() = %q, want id", got)
	}
	if got := defaultListSortDirection(plan); got != "DESC" {
		t.Fatalf("defaultListSortDirection() = %q, want DESC", got)
	}
}

func TestDefaultListSortForMessageCategoryStillUsesSortOrderAsc(t *testing.T) {
	t.Parallel()

	plan := resourcePlan{
		Resource: config.Resource{Name: "internal_message_category", Entity: "InternalMessageCategory"},
		Schema: entschema.Schema{Fields: []entschema.Field{
			{Name: "id"},
			{Name: "sort_order"},
			{Name: "created_at"},
		}},
	}

	if got := defaultListSortField(plan); got != "sort_order" {
		t.Fatalf("defaultListSortField() = %q, want sort_order", got)
	}
	if got := defaultListSortDirection(plan); got != "ASC" {
		t.Fatalf("defaultListSortDirection() = %q, want ASC", got)
	}
}

func TestGeneratedBootstrapProvidersFileCompiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), `module example.com/xadmin-web

go 1.26.0

require github.com/chnxq/xkitpkg/app v0.0.0-20260421141638-80e4b484ff8f
`)
	writeFile(t, filepath.Join(root, "api", "protos", "admin", "v1", "i_user.proto"), `syntax = "proto3";

package admin.service.v1;

service UserService {
  rpc List (ListRequest) returns (ListResponse) {}
}`)
	writeFile(t, filepath.Join(root, "api", "gen", "admin", "v1", "i_user_grpc.pb.go"), `package admin

import (
	context "context"
	v1 "example.com/xadmin-web/api/gen/identity/v1"
)

type UserServiceServer interface {
	List(context.Context, *v1.ListRequest) (*v1.ListResponse, error)
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
	writeFile(t, filepath.Join(root, "api", "gen", "identity", "v1", "user.pb.go"), `package identityv1

type User struct{}
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
				Operations: config.OperationFlags{
					"list": true,
				},
				Generate: config.GenerateFlags{
					ServiceStub:  true,
					RepoCRUD:     true,
					RestRegister: true,
					GRPCRegister: true,
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

	if _, err := runner.Generate("bootstrap"); err != nil {
		t.Fatalf("generate bootstrap: %v", err)
	}

	providersFile := readFile(t, filepath.Join(root, "internal", "bootstrap", "generated_data_providers.gen.go"))
	if _, err := parser.ParseFile(token.NewFileSet(), "generated_data_providers.gen.go", providersFile, parser.AllErrors); err != nil {
		t.Fatalf("generated providers file should parse as valid Go: %v", err)
	}
}

func TestBootstrapResourcesCollectServiceReposWithoutRepoCRUD(t *testing.T) {
	t.Parallel()

	runner := &Runner{}
	plans := []resourcePlan{
		{
			Resource: config.Resource{
				Name:          "user",
				ProtoService:  "admin.service.v1.UserService",
				RepoInterface: "UserRepo",
				Generate: config.GenerateFlags{
					ServiceStub: true,
					RepoCRUD:    true,
				},
			},
			Binding:       binding.ServiceBinding{ServiceName: "UserService"},
			ResourceField: "User",
		},
		{
			Resource: config.Resource{
				Name:          "user_portal",
				ProtoService:  "admin.service.v1.UserPortalService",
				RepoInterface: "UserRepo",
				ServiceRepos: []config.RepoConfig{
					{Field: "userCredentialRepo", Interface: "UserCredentialRepo"},
				},
				Generate: config.GenerateFlags{
					ServiceStub: true,
				},
			},
			Binding:       binding.ServiceBinding{ServiceName: "UserPortalService"},
			ResourceField: "UserPortal",
		},
		{
			Resource: config.Resource{
				Name:          "user_credential",
				ProtoService:  "authentication.service.v1.UserCredentialService",
				RepoInterface: "UserCredentialRepo",
				Generate: config.GenerateFlags{
					RepoCRUD: true,
				},
			},
			Binding:       binding.ServiceBinding{ServiceName: "UserCredentialService"},
			ResourceField: "UserCredential",
		},
	}

	serverResources := runner.bootstrapResources(plans)
	if len(serverResources) != 2 {
		t.Fatalf("bootstrapResources count mismatch: got %d want 2", len(serverResources))
	}

	var userPortal bootstrapResourceData
	for _, item := range serverResources {
		if item.FieldName == "UserPortal" {
			userPortal = item
			break
		}
	}
	if !userPortal.HasRepo {
		t.Fatalf("service-only resource with repo_interface should still inject its main repo")
	}
	if userPortal.RepoVar != "userRepo" {
		t.Fatalf("user portal repo var mismatch: got %q want userRepo", userPortal.RepoVar)
	}
	if len(userPortal.ServiceRepoVars) != 1 || userPortal.ServiceRepoVars[0] != "userCredentialRepo" {
		t.Fatalf("user portal extra repos mismatch: got %#v", userPortal.ServiceRepoVars)
	}

	repoResources := runner.bootstrapRepoResources(plans)
	if len(repoResources) != 2 {
		t.Fatalf("bootstrapRepoResources count mismatch: got %d want 2", len(repoResources))
	}
	if repoResources[0].RepoVar != "userRepo" && repoResources[1].RepoVar != "userRepo" {
		t.Fatalf("bootstrapRepoResources should include shared userRepo once")
	}
	if repoResources[0].RepoVar != "userCredentialRepo" && repoResources[1].RepoVar != "userCredentialRepo" {
		t.Fatalf("bootstrapRepoResources should include userCredentialRepo")
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

func TestServiceMethodBodyUsesConfiguredSpecialMethod(t *testing.T) {
	t.Parallel()

	runner := &Runner{}
	body := runner.serviceMethodBody(resourcePlan{
		Resource: config.Resource{
			Name: "internal_message",
			ServiceMethods: map[string]config.ServiceMethodConfig{
				"ListMessage": {
					Body: "return s.{{repoField}}.ListByPaging({{ctx}}, repo.PagingRequest{})",
				},
			},
		},
	}, "ListMessage", []namedType{
		{Name: "ctx", Type: "context.Context"},
		{Name: "req", Type: "*v1.PagingRequest"},
	}, "*v11.ListInternalMessageResponse", "internalMessageRepo", true)

	if body == "" {
		t.Fatalf("expected configured special method body to be generated")
	}
	if !strings.Contains(body, "s.internalMessageRepo.ListByPaging(ctx, repo.PagingRequest{})") {
		t.Fatalf("configured special method body was not rendered correctly: %q", body)
	}
}

func TestRenderServiceFileIncludesConfiguredSpecialMethods(t *testing.T) {
	t.Parallel()

	runner := &Runner{
		project: project.Info{Module: "admin"},
	}
	plan := resourcePlan{
		Resource: config.Resource{
			Name:          "internal_message",
			Entity:        "InternalMessage",
			RepoInterface: "InternalMessageRepo",
			ServiceRepos: []config.RepoConfig{
				{Field: "internalMessageRecipientRepo", Interface: "InternalMessageRecipientRepo"},
			},
			ServiceMethods: map[string]config.ServiceMethodConfig{
				"ListMessage": {Body: "return s.{{repoField}}.ListByPaging({{ctx}}, repo.PagingRequest{})"},
				"GetMessage":  {Body: "return s.{{repoField}}.GetByID({{ctx}}, {{param.req}}.GetId())"},
			},
		},
		Proto: xproto.Service{
			Methods: []xproto.Method{
				{Name: "ListMessage", Classification: "query"},
				{Name: "GetMessage", Classification: "query"},
			},
		},
		Binding: binding.ServiceBinding{
			ServiceName: "InternalMessageService",
			ImportPath:  "admin/api/gen/admin/v1",
			Imports: map[string]string{
				"repo": "admin/internal/data/repo",
				"v11":  "admin/api/gen/internal_message/v1",
			},
			Methods: []binding.Method{
				{
					Name:    "ListMessage",
					Params:  []string{"context.Context", "*paginationv1.PagingRequest"},
					Results: []string{"*v11.ListInternalMessageResponse"},
				},
				{
					Name:    "GetMessage",
					Params:  []string{"context.Context", "*v11.GetInternalMessageRequest"},
					Results: []string{"*v11.InternalMessage"},
				},
			},
		},
		APIPackageAlias: "adminv1",
	}

	content, err := runner.renderServiceFile(plan)
	if err != nil {
		t.Fatalf("render service file: %v", err)
	}
	got := string(content)
	if !strings.Contains(got, "func (s *InternalMessageService) ListMessage") {
		t.Fatalf("rendered service file missing ListMessage: %s", got)
	}
	if !strings.Contains(got, "return s.internalMessageRepo.ListByPaging(ctx, repo.PagingRequest{})") {
		t.Fatalf("rendered service file missing configured ListMessage body: %s", got)
	}
	if !strings.Contains(got, "return s.internalMessageRepo.GetByID(ctx, req.GetId())") {
		t.Fatalf("rendered service file missing configured GetMessage body: %s", got)
	}
	if !strings.Contains(got, "internalMessageRecipientRepo repo.InternalMessageRecipientRepo") {
		t.Fatalf("rendered service file missing extra repo injection: %s", got)
	}
}

func TestRenderRepoFileIncludesConfiguredRepoMethods(t *testing.T) {
	t.Parallel()

	runner := &Runner{
		project: project.Info{Module: "admin"},
	}
	plan := resourcePlan{
		Resource: config.Resource{
			Name:          "role",
			Entity:        "Role",
			RepoInterface: "RoleRepo",
			RepoMethods: map[string]config.RepoMethodConfig{
				"List": {Body: "return &permissionv1.ListRoleResponse{}, nil"},
				"Get":  {Body: "return &permissionv1.Role{}, nil"},
			},
			Operations: map[string]bool{
				"list": true,
				"get":  true,
			},
		},
		Binding: binding.ServiceBinding{
			ServiceName: "RoleService",
			ImportPath:  "admin/api/gen/admin/v1",
			Imports: map[string]string{
				"permissionv1": "admin/api/gen/permission/v1",
			},
			Methods: []binding.Method{
				{
					Name:    "List",
					Params:  []string{"context.Context", "*paginationv1.PagingRequest"},
					Results: []string{"*permissionv1.ListRoleResponse"},
				},
				{
					Name:    "Get",
					Params:  []string{"context.Context", "*permissionv1.GetRoleRequest"},
					Results: []string{"*permissionv1.Role"},
				},
			},
		},
		APIPackageAlias: "adminv1",
		Schema:          entschema.Schema{},
	}

	content, err := runner.renderRepoFile(plan)
	if err != nil {
		t.Fatalf("render repo file: %v", err)
	}
	got := string(content)
	if !strings.Contains(got, "func (r *roleRepo) List") {
		t.Fatalf("rendered repo file missing List: %s", got)
	}
	if !strings.Contains(got, "return &permissionv1.ListRoleResponse{}, nil") {
		t.Fatalf("rendered repo file missing configured List body: %s", got)
	}
	if !strings.Contains(got, "func (r *roleRepo) Get") {
		t.Fatalf("rendered repo file missing Get: %s", got)
	}
	if !strings.Contains(got, "return &permissionv1.Role{}, nil") {
		t.Fatalf("rendered repo file missing configured Get body: %s", got)
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
