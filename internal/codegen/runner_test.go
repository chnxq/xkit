package codegen

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
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
	writeFile(t, filepath.Join(root, "api", "protos", "admin", "v1", "i_dict_label.proto"), `syntax = "proto3";

package admin.service.v1;

service DictLabelService {
  rpc Create (CreateDictLabelRequest) returns (DictLabel) {}
  rpc Update (UpdateDictLabelRequest) returns (DictLabel) {}
}`)
	writeFile(t, filepath.Join(root, "api", "gen", "admin", "v1", "i_dict_label_grpc.pb.go"), `package admin

import (
	context "context"
	v1 "example.com/xadmin-web/api/gen/dict/v1"
)

type DictLabelServiceServer interface {
	Create(context.Context, *v1.CreateDictLabelRequest) (*v1.DictLabel, error)
	Update(context.Context, *v1.UpdateDictLabelRequest) (*v1.DictLabel, error)
	mustEmbedUnimplementedDictLabelServiceServer()
}

var DictLabelService_ServiceDesc = struct{
	ServiceName string
}{
	ServiceName: "admin.service.v1.DictLabelService",
}
`)
	writeFile(t, filepath.Join(root, "api", "gen", "admin", "v1", "i_dict_label_http.pb.go"), `package admin

type DictLabelServiceHTTPServer interface{}
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
	writeFile(t, filepath.Join(root, "api", "gen", "dict", "v1", "dict_label.pb.go"), `package dictv1

type DictLabel struct {
	LabelCode   *string
	PayloadJson *string
	IsBuiltin   *bool
	Status      *DictLabel_Status
}

type DictLabel_Status string

func (x *DictLabel) GetLabelCode() string {
	if x != nil && x.LabelCode != nil {
		return *x.LabelCode
	}
	return ""
}

func (x *DictLabel) GetPayloadJson() string {
	if x != nil && x.PayloadJson != nil {
		return *x.PayloadJson
	}
	return ""
}

func (x *DictLabel) GetIsBuiltin() bool {
	if x != nil && x.IsBuiltin != nil {
		return *x.IsBuiltin
	}
	return false
}

func (x *DictLabel) GetStatus() DictLabel_Status {
	if x != nil && x.Status != nil {
		return *x.Status
	}
	return ""
}

func (x *DictLabel_Status) String() string {
	if x == nil {
		return ""
	}
	return string(*x)
}

type CreateDictLabelRequest struct {
	Data *DictLabel
}

type UpdateDictLabelRequest struct {
	Id         uint32
	Data       *DictLabel
	UpdateMask *FieldMask
}

func (x *CreateDictLabelRequest) GetData() *DictLabel {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *UpdateDictLabelRequest) GetId() uint32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *UpdateDictLabelRequest) GetData() *DictLabel {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *UpdateDictLabelRequest) GetUpdateMask() *FieldMask {
	if x != nil {
		return x.UpdateMask
	}
	return nil
}

type FieldMask struct {
	Paths []string
}

func (x *FieldMask) GetPaths() []string {
	if x != nil {
		return x.Paths
	}
	return nil
}
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
	writeFile(t, filepath.Join(root, "internal", "data", "ent", "schema", "dict_label.go"), `package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type DictLabel struct { ent.Schema }

func (DictLabel) Fields() []ent.Field {
	return []ent.Field{
		field.String("label_code").Optional().Nillable(),
		field.JSON("payload_json", map[string]any{}).Optional(),
		field.Bool("is_builtin").Default(false).Optional(),
		field.Enum("status").Values("ON", "OFF").Default("ON").Optional(),
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
					Allow: []string{"id", "username", "status", "last_login_at"},
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
			{
				Name:          "dict_label",
				ProtoService:  "admin.service.v1.DictLabelService",
				Entity:        "DictLabel",
				DTOImport:     "example.com/xadmin-web/api/gen/dict/v1",
				DTOType:       "DictLabel",
				RepoInterface: "DictLabelRepo",
				Operations: config.OperationFlags{
					"create": true,
					"update": true,
				},
				Generate: config.GenerateFlags{
					RepoCRUD: true,
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

	if len(result.Written) != 17 {
		t.Fatalf("written file count mismatch: got %d want %d", len(result.Written), 17)
	}

	expectedPaths := []string{
		filepath.Join(root, "internal", "service", "user_service.gen.go"),
		filepath.Join(root, "internal", "service", "user_service_ext.go"),
		filepath.Join(root, "internal", "service", "service_shared_ext.go"),
		filepath.Join(root, "internal", "data", "repo", "user_repo.gen.go"),
		filepath.Join(root, "internal", "data", "repo", "repo_shared_ext.go"),
		filepath.Join(root, "internal", "data", "repo", "user_repo_ext.go"),
		filepath.Join(root, "internal", "data", "repo", "user_credential_repo_ext.go"),
		filepath.Join(root, "internal", "data", "repo", "dict_label_repo.gen.go"),
		filepath.Join(root, "internal", "data", "repo", "dict_label_repo_ext.go"),
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
	serviceSharedFile := readFile(t, filepath.Join(root, "internal", "service", "service_shared_ext.go"))
	if !strings.Contains(serviceSharedFile, "func requireViewerContext(") || !strings.Contains(serviceSharedFile, "func requirePlatformContext(") {
		t.Fatalf("service shared helper file is missing viewer/platform context helpers")
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
	repoSharedFile := readFile(t, filepath.Join(root, "internal", "data", "repo", "repo_shared_ext.go"))
	if !strings.Contains(repoSharedFile, "func withDefaultSorting(") || !strings.Contains(repoSharedFile, "paginationv1.Sorting_Direction") {
		t.Fatalf("repo shared helper file is missing withDefaultSorting helper")
	}
	if !strings.Contains(repoSharedFile, "func ensureTenantAccessible(") || !strings.Contains(repoSharedFile, "func viewerTenantID(") {
		t.Fatalf("repo shared helper file is missing tenant scope helpers")
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
	if !strings.Contains(repoFile, "case \"last_login_at\":") || !strings.Contains(repoFile, "userParseFilterTime(condition.GetValue(), \"last_login_at\")") || !strings.Contains(repoFile, "builder.Where(user.LastLoginAtGTE(value))") {
		t.Fatalf("repo file is missing generated time filter")
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

	dictLabelRepoFile := readFile(t, filepath.Join(root, "internal", "data", "repo", "dict_label_repo.gen.go"))
	if !strings.Contains(dictLabelRepoFile, "payloadJSON := strings.TrimSpace(req.Data.GetPayloadJson())") {
		t.Fatalf("dict label repo file is missing JSON payload trim: %s", dictLabelRepoFile)
	}
	if !strings.Contains(dictLabelRepoFile, "payloadValue := map[string]any{}") {
		t.Fatalf("dict label repo file is missing JSON payload map init: %s", dictLabelRepoFile)
	}
	if !strings.Contains(dictLabelRepoFile, "if err := json.Unmarshal([]byte(payloadJSON), &payloadValue); err != nil {") {
		t.Fatalf("dict label repo file is missing JSON payload decode: %s", dictLabelRepoFile)
	}
	if !strings.Contains(dictLabelRepoFile, "builder.SetPayloadJSON(payloadValue)") || !strings.Contains(dictLabelRepoFile, "builder.ClearPayloadJSON()") {
		t.Fatalf("dict label repo file is missing JSON payload setter or clear: %s", dictLabelRepoFile)
	}
	if !strings.Contains(dictLabelRepoFile, "builder.SetNillableIsBuiltin(req.Data.IsBuiltin)") {
		t.Fatalf("dict label repo file is missing bool nillable setter: %s", dictLabelRepoFile)
	}
	if !strings.Contains(dictLabelRepoFile, "builder.SetNillableStatus(dictLabelEnumPtrFromProto[dictlabel.Status](req.Data.Status))") {
		t.Fatalf("dict label repo file is missing enum nillable setter: %s", dictLabelRepoFile)
	}
	if !strings.Contains(dictLabelRepoFile, `"encoding/json"`) {
		t.Fatalf("dict label repo file is missing JSON import: %s", dictLabelRepoFile)
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
	if userPortal.HasRepo {
		t.Fatalf("service-only resource without repo CRUD should not inject its main repo")
	}
	if len(userPortal.ServiceRepoVars) != 1 || userPortal.ServiceRepoVars[0] != "userCredentialRepo" {
		t.Fatalf("user portal extra repos mismatch: got %#v", userPortal.ServiceRepoVars)
	}

	repoResources := runner.bootstrapRepoResources(plans)
	if len(repoResources) != 2 {
		t.Fatalf("bootstrapRepoResources count mismatch: got %d want 2", len(repoResources))
	}
	seenRepoVars := make(map[string]struct{}, len(repoResources))
	for _, item := range repoResources {
		seenRepoVars[item.RepoVar] = struct{}{}
	}
	if _, ok := seenRepoVars["userRepo"]; !ok {
		t.Fatalf("bootstrapRepoResources should include userRepo, got %#v", repoResources)
	}
	if _, ok := seenRepoVars["userCredentialRepo"]; !ok {
		t.Fatalf("bootstrapRepoResources should include userCredentialRepo, got %#v", repoResources)
	}
}

func TestBootstrapResourcesIgnoreRepoTypedServiceFields(t *testing.T) {
	t.Parallel()

	runner := &Runner{}
	plans := []resourcePlan{
		{
			Resource: config.Resource{
				Name:          "task_group",
				ProtoService:  "admin.service.v1.TaskGroupService",
				RepoInterface: "TaskGroupRepo",
				ServiceFields: []config.ServiceFieldConfig{
					{Field: "taskRepo", Type: "repo.TaskRepo"},
					{Field: "scheduler", Type: "*taskruntime.Scheduler"},
				},
				Generate: config.GenerateFlags{
					ServiceStub: true,
					RepoCRUD:    true,
				},
			},
			Binding:       binding.ServiceBinding{ServiceName: "TaskGroupService"},
			ResourceField: "TaskGroup",
		},
		{
			Resource: config.Resource{
				Name:          "task",
				ProtoService:  "admin.service.v1.TaskService",
				RepoInterface: "TaskRepo",
				ServiceFields: []config.ServiceFieldConfig{
					{Field: "taskGroupRepo", Type: "repo.TaskGroupRepo"},
					{Field: "scheduler", Type: "*taskruntime.Scheduler"},
				},
				Generate: config.GenerateFlags{
					ServiceStub: true,
					RepoCRUD:    true,
				},
			},
			Binding:       binding.ServiceBinding{ServiceName: "TaskService"},
			ResourceField: "Task",
		},
	}

	serverResources := runner.bootstrapResources(plans)
	if len(serverResources) != 2 {
		t.Fatalf("bootstrapResources count mismatch: got %d want 2", len(serverResources))
	}

	for _, item := range serverResources {
		if len(item.ServiceRepoVars) != 0 {
			t.Fatalf("repo-typed service_fields should not be treated as constructor repos: %#v", item)
		}
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
	if got := filterKind("Time"); got != "Time" {
		t.Fatalf("filterKind(Time) = %q, want Time", got)
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

func TestRenderServiceFileKeepsServiceFieldsOutOfConstructorInjection(t *testing.T) {
	t.Parallel()

	runner := &Runner{
		project: project.Info{Module: "admin"},
	}
	plan := resourcePlan{
		Resource: config.Resource{
			Name:          "task",
			Entity:        "Task",
			RepoInterface: "TaskRepo",
			ServiceImports: []config.ImportConfig{
				{Alias: "taskruntime", Path: "{{module}}/internal/task/runtime"},
			},
			ServiceFields: []config.ServiceFieldConfig{
				{Field: "taskGroupRepo", Type: "repo.TaskGroupRepo"},
				{Field: "scheduler", Type: "*taskruntime.Scheduler"},
			},
			ServiceMethods: map[string]config.ServiceMethodConfig{
				"Start": {Body: "return s.start({{ctx}}, {{param.req}})"},
			},
		},
		Proto: xproto.Service{
			Methods: []xproto.Method{
				{Name: "Start", Classification: "special"},
			},
		},
		Binding: binding.ServiceBinding{
			ServiceName: "TaskService",
			ImportPath:  "admin/api/gen/admin/v1",
			Imports: map[string]string{
				"taskv1": "admin/api/gen/task/v1",
			},
			Methods: []binding.Method{
				{
					Name:    "Start",
					Params:  []string{"context.Context", "*taskv1.StartTaskRequest"},
					Results: []string{"*emptypb.Empty"},
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
	if !strings.Contains(got, "taskGroupRepo repo.TaskGroupRepo") {
		t.Fatalf("rendered service file missing runtime-bound service field: %s", got)
	}
	if !strings.Contains(got, "scheduler     *taskruntime.Scheduler") {
		t.Fatalf("rendered service file missing scheduler service field: %s", got)
	}
	if !strings.Contains(got, "func NewTaskService(ctx *app.AppCtx, taskRepo repo.TaskRepo) *TaskService") {
		t.Fatalf("rendered service constructor should only inject the main repo: %s", got)
	}
	if strings.Contains(got, "func NewTaskService(ctx *app.AppCtx, taskRepo repo.TaskRepo, taskGroupRepo repo.TaskGroupRepo") {
		t.Fatalf("rendered service constructor should not inject repo-typed service fields: %s", got)
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

func TestRenderRepoFileNormalizesDTOTypeInitialisms(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/xadmin-web\n\ngo 1.26.0\n")
	writeFile(t, filepath.Join(root, "api", "gen", "dict", "v1", "dict_label_i18n.pb.go"), `package dictv1

type DictLabelI18N struct {
	Id *uint32
}

type GetDictLabelI18NRequest struct {
	Id *uint32
}
`)

	runner := &Runner{
		project: project.Info{
			Root:   root,
			Module: "example.com/xadmin-web",
		},
		options: Options{Version: "test"},
	}

	plan := resourcePlan{
		Resource: config.Resource{
			Name:          "dict_label_i18n",
			ProtoService:  "admin.service.v1.DictLabelI18nService",
			Entity:        "DictLabelI18n",
			DTOImport:     "example.com/xadmin-web/api/gen/dict/v1",
			DTOType:       "DictLabelI18n",
			RepoInterface: "DictLabelI18nRepo",
			Generate:      config.GenerateFlags{RepoCRUD: true},
		},
		Binding: binding.ServiceBinding{
			ServiceName: "DictLabelI18nService",
			ImportPath:  "example.com/xadmin-web/api/gen/admin/v1",
			Imports: map[string]string{
				"v1": "example.com/xadmin-web/api/gen/dict/v1",
			},
			Methods: []binding.Method{
				{
					Name:    "Get",
					Params:  []string{"context.Context", "*v1.GetDictLabelI18NRequest"},
					Results: []string{"*v1.DictLabelI18N"},
				},
			},
		},
		APIPackageAlias: "adminv1",
		Schema: entschema.Schema{
			Name: "DictLabelI18n",
			Fields: []entschema.Field{
				{Name: "id", Kind: "Uint32"},
			},
		},
	}

	content, err := runner.renderRepoFile(plan)
	if err != nil {
		t.Fatalf("render repo file: %v", err)
	}
	got := string(content)
	if !strings.Contains(got, "mapper *mapper.CopierMapper[dictv1.DictLabelI18N, ent.DictLabelI18n]") {
		t.Fatalf("rendered repo file missing normalized DTO type: %s", got)
	}
	if strings.Contains(got, "dictv1.DictLabelI18n") {
		t.Fatalf("rendered repo file still contains non-normalized DTO type: %s", got)
	}
}

func TestRenderServiceFileSkipsRepoInjectionWhenRepoCRUDDisabled(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/xadmin-web\n\ngo 1.26.0\n")
	writeFile(t, filepath.Join(root, "api", "gen", "admin", "v1", "i_internal_message_grpc.pb.go"), `package admin

import (
	context "context"
	v1 "example.com/xadmin-web/api/gen/internal_message/v1"
)

type InternalMessageServiceServer interface {
	ListMessage(context.Context, *v1.ListInternalMessageRequest) (*v1.ListInternalMessageResponse, error)
	mustEmbedUnimplementedInternalMessageServiceServer()
}
`)
	writeFile(t, filepath.Join(root, "api", "gen", "admin", "v1", "i_internal_message_http.pb.go"), `package admin

type InternalMessageServiceHTTPServer interface{}
`)
	writeFile(t, filepath.Join(root, "api", "gen", "internal_message", "v1", "internal_message.pb.go"), `package internalmessagev1

type ListInternalMessageRequest struct{}
type ListInternalMessageResponse struct{}
`)

	runner := &Runner{
		project: project.Info{
			Root:   root,
			Module: "example.com/xadmin-web",
		},
		options: Options{Version: "test"},
	}

	plan := resourcePlan{
		Resource: config.Resource{
			Name:          "internal_message",
			ProtoService:  "admin.service.v1.InternalMessageService",
			Entity:        "InternalMessage",
			DTOImport:     "example.com/xadmin-web/api/gen/internal_message/v1",
			DTOType:       "InternalMessage",
			RepoInterface: "InternalMessageRepo",
			Generate: config.GenerateFlags{
				ServiceStub:  true,
				RestRegister: true,
				GRPCRegister: true,
			},
		},
		Binding: binding.ServiceBinding{
			ServiceName: "InternalMessageService",
			ImportPath:  "example.com/xadmin-web/api/gen/admin/v1",
			Imports: map[string]string{
				"v1": "example.com/xadmin-web/api/gen/internal_message/v1",
			},
			Methods: []binding.Method{
				{
					Name:    "ListMessage",
					Params:  []string{"context.Context", "*v1.ListInternalMessageRequest"},
					Results: []string{"*v1.ListInternalMessageResponse"},
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
	if strings.Contains(got, "repo.InternalMessageRepo") {
		t.Fatalf("service file should not inject repo when repo CRUD is disabled: %s", got)
	}
	if strings.Contains(got, "ctx *app.AppCtx") {
		t.Fatalf("service constructor should not require app context when repo CRUD is disabled: %s", got)
	}
	if !strings.Contains(got, "func (s *InternalMessageService) ListMessage") {
		t.Fatalf("service file missing method: %s", got)
	}
}

func TestBootstrapSkipsRepoWhenRepoCRUDDisabled(t *testing.T) {
	t.Parallel()

	runner := &Runner{
		config: config.Config{
			Resources: []config.Resource{
				{
					Name:          "internal_message",
					RepoInterface: "InternalMessageRepo",
					Generate: config.GenerateFlags{
						ServiceStub:  true,
						RestRegister: true,
						GRPCRegister: true,
					},
				},
			},
		},
	}

	plans := []resourcePlan{
		{
			Resource: config.Resource{
				Name:          "internal_message",
				RepoInterface: "InternalMessageRepo",
				Generate: config.GenerateFlags{
					ServiceStub:  true,
					RestRegister: true,
					GRPCRegister: true,
				},
			},
			ResourceField: "InternalMessage",
			Binding: binding.ServiceBinding{
				ServiceName: "InternalMessageService",
			},
		},
	}

	serverResources := runner.bootstrapResources(plans)
	if len(serverResources) != 1 {
		t.Fatalf("expected one bootstrap resource, got %d", len(serverResources))
	}
	if serverResources[0].HasRepo {
		t.Fatalf("bootstrap resource should not mark repo present when repo CRUD is disabled: %+v", serverResources[0])
	}

	repoResources := runner.bootstrapRepoResources(plans)
	if len(repoResources) != 0 {
		t.Fatalf("bootstrap repo resources should be empty when repo CRUD is disabled: %+v", repoResources)
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

func TestWriteGeneratedFileSkipsTimestampOnlyChanges(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "internal", "service", "user_service.gen.go")

	runner := &Runner{}

	first := []byte(`// Code generated by xkit. DO NOT EDIT.
// xkit version: test-version
// generated at: 2026-06-03 10:00:00 CST

package service

func noop() {}
`)
	second := []byte(`// Code generated by xkit. DO NOT EDIT.
// xkit version: test-version
// generated at: 2026-06-03 10:05:00 CST

package service

func noop() {}
`)

	var firstResult Result
	if err := runner.writeGeneratedFile(path, first, &firstResult); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if len(firstResult.Written) != 1 {
		t.Fatalf("expected first write recorded, got written=%v skipped=%v", firstResult.Written, firstResult.Skipped)
	}

	var secondResult Result
	if err := runner.writeGeneratedFile(path, second, &secondResult); err != nil {
		t.Fatalf("second write: %v", err)
	}
	if len(secondResult.Skipped) != 1 {
		t.Fatalf("expected timestamp-only rewrite to be skipped, got written=%v skipped=%v", secondResult.Written, secondResult.Skipped)
	}

	got := readFile(t, path)
	if got != string(first) {
		t.Fatalf("file should keep original content when only timestamp changes, got:\n%s", got)
	}
}

func TestWriteGeneratedFileSkipsTimestampAndLineEndingOnlyChanges(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "internal", "service", "user_service.gen.go")

	runner := &Runner{}

	existing := []byte("// Code generated by xkit. DO NOT EDIT.\r\n// xkit version: test-version\r\n// generated at: 2026-06-03 10:00:00 CST\r\n\r\npackage service\r\n\r\nfunc noop() {}\r\n")
	generated := []byte("// Code generated by xkit. DO NOT EDIT.\n// xkit version: test-version\n// generated at: 2026-06-03 10:05:00 CST\n\npackage service\n\nfunc noop() {}\n")

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, existing, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}

	var result Result
	if err := runner.writeGeneratedFile(path, generated, &result); err != nil {
		t.Fatalf("write generated file: %v", err)
	}
	if len(result.Skipped) != 1 {
		t.Fatalf("expected timestamp and line-ending-only rewrite to be skipped, got written=%v skipped=%v", result.Written, result.Skipped)
	}

	got := readFile(t, path)
	if got != string(existing) {
		t.Fatalf("file should keep original content when only timestamp and line endings change, got:\n%s", got)
	}
}

func TestWriteExtensionFileSkipsTimestampAndLineEndingOnlyChanges(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "internal", "service", "user_service_ext.go")

	runner := &Runner{}

	existing := []byte("// Code generated by xkit.\r\n// xkit version: test-version\r\n// generated at: 2026-06-03 10:00:00 CST\r\n\r\npackage service\r\n\r\n// custom body stays untouched\r\nfunc noop() {}\r\n")
	generated := []byte("// Code generated by xkit.\n// xkit version: test-version\n// generated at: 2026-06-03 10:05:00 CST\n\npackage service\n\n// custom body from template would be ignored\n")

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, existing, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}

	var result Result
	if err := runner.writeExtensionFile(path, generated, &result); err != nil {
		t.Fatalf("write extension file: %v", err)
	}
	if len(result.Skipped) != 1 {
		t.Fatalf("expected extension timestamp and line-ending-only rewrite to be skipped, got written=%v skipped=%v", result.Written, result.Skipped)
	}

	got := readFile(t, path)
	if got != string(existing) {
		t.Fatalf("extension file should keep original content when only header timestamp and line endings change, got:\n%s", got)
	}
}

func TestEffectiveFilterAllowAppendsCreatedAtForRecordLikeResource(t *testing.T) {
	plan := resourcePlan{
		Resource: config.Resource{
			Name: "api_audit_log",
			Filters: config.FilterConfig{
				Allow: []string{"id", "user_id"},
			},
		},
		Schema: entschema.Schema{
			Fields: []entschema.Field{
				{Name: "id", Kind: "Uint32"},
				{Name: "created_at", Kind: "Time"},
				{Name: "user_id", Kind: "Uint32"},
			},
		},
	}

	got := effectiveFilterAllow(plan)
	if !slices.Contains(got, "created_at") {
		t.Fatalf("effectiveFilterAllow() = %v, want created_at to be appended", got)
	}
}

func TestEffectiveFilterAllowKeepsCreatedAtSingle(t *testing.T) {
	plan := resourcePlan{
		Resource: config.Resource{
			Name: "task_log",
			Filters: config.FilterConfig{
				Allow: []string{"id", "created_at", "task_id"},
			},
		},
		Schema: entschema.Schema{
			Fields: []entschema.Field{
				{Name: "id", Kind: "Uint32"},
				{Name: "created_at", Kind: "Time"},
				{Name: "task_id", Kind: "Uint64"},
			},
		},
	}

	got := effectiveFilterAllow(plan)
	count := 0
	for _, item := range got {
		if item == "created_at" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("effectiveFilterAllow() created_at count = %d, want 1; full=%v", count, got)
	}
}

func TestEffectiveFilterAllowDoesNotAppendCreatedAtForNonLogResource(t *testing.T) {
	plan := resourcePlan{
		Resource: config.Resource{
			Name: "login_policy",
			Filters: config.FilterConfig{
				Allow: []string{"id", "target_id"},
			},
		},
		Schema: entschema.Schema{
			Fields: []entschema.Field{
				{Name: "id", Kind: "Uint32"},
				{Name: "created_at", Kind: "Time"},
				{Name: "target_id", Kind: "Uint32"},
			},
		},
	}

	got := effectiveFilterAllow(plan)
	if slices.Contains(got, "created_at") {
		t.Fatalf("effectiveFilterAllow() = %v, want created_at not to be appended for non-log resource", got)
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
