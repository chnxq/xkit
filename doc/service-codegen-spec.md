# Service Code Generation Spec

## Purpose

This document defines the first workable specification for service-oriented code generation in `xkit`.

`xkit` is the implementation repository for the generator.
The primary target project is `D:\GoProjects\XAdmin\xadmin-web`.
The primary implementation reference for service/data/server code is `D:\GoProjects\chnxq\XAdmin`. Generator behavior, project discovery, scaffold conventions, and service output should follow the XAdmin reference first.

The goal is not to generate complete business systems.
The goal is to standardize and regenerate the highly repetitive parts of a GoWind-style Go service safely:

- service interface adapters
- repository CRUD scaffolds
- Wire provider sets
- transport registration code

The design must preserve manual business logic and allow repeated generation without overwriting hand-written code.

## Scope

This specification targets XAdmin-aligned source inputs and GoWind-style generated service outputs.

### Current Source Layout In XAdmin

- `api/protos/<domain>/v1/*.proto`
- `api/gen/<domain>/v1/*`
- `internal/data/ent/schema/*.go`

### Target Generated Layout

- `app/<service>/service/internal/service/*`
- `app/<service>/service/internal/data/*`
- `app/<service>/service/internal/server/*`

Representative examples in the current target workspace:

- `xadmin-web/api/protos/admin/v1/i_user.proto`
- `xadmin-web/api/gen/admin/v1/i_user.pb.go`
- `xadmin-web/internal/data/ent/schema/user.go`

Primary reference files from `D:\GoProjects\chnxq\XAdmin` for generated service layers:

- `app/admin/service/internal/service/user_service.go`
- `app/admin/service/internal/data/user_repo.go`
- `app/admin/service/internal/server/rest_server.go`
- `app/admin/service/internal/service/providers/provider_set.go`
- `app/admin/service/internal/data/providers/provider_set.go`
- `app/admin/service/internal/server/providers/provider_set.go`
- `app/admin/service/cmd/server/wire.go`

## Goals

- Use stable source definitions to generate repeatable scaffolding.
- Keep transport, service, repo, and wiring code aligned.
- Reduce manual edits for new resources and new services.
- Make regeneration safe by separating generated and manual files.

## Non-Goals

- Do not generate complex business orchestration automatically.
- Do not encode all business rules into proto files.
- Do not overwrite hand-written files during regeneration.
- Do not treat generated CRUD code as a substitute for domain design.

## Source Of Truth

Generation uses three inputs with distinct responsibilities.

### 1. Proto

Proto files are the source of truth for:

- service names
- RPC names
- request and response types
- HTTP bindings
- transport registration lists

Example:

- `xadmin-web/api/protos/admin/v1/i_user.proto`

### 2. Ent Schema

Ent schema files are the source of truth for:

- entity names
- fields
- relations
- primary keys
- base repository model shape

Example:

- `xadmin-web/internal/data/ent/schema/user.go`

### 3. Generator Config

A per-service YAML file controls generation behavior that does not belong in proto or ent:

- which resources are enabled
- which RPCs are standard CRUD
- which files should be emitted
- filter, sort, paging, and enrich policies
- manual opt-outs

Recommended path:

- `app/<service>/service/gen/<domain>.yaml`

## File Ownership Model

The generator must follow strict ownership boundaries.

### Generated Files

Generated files are fully owned by the generator and may be overwritten:

- `*.gen.go`

Examples:

- `user_service.gen.go`
- `user_repo.gen.go`
- `wire_set.gen.go`
- `rest_register.gen.go`

### Manual Extension Files

Extension files are created once and never overwritten:

- `*_ext.go`

Examples:

- `user_service_ext.go`
- `user_repo_ext.go`

These files contain hook implementations and resource-specific customization.

### Manual Files

Manual files are fully owned by developers:

- `*_manual.go`
- existing custom bootstrap files such as `rest_server.go`

These files contain business orchestration, special handlers, authorization rules, non-standard transport behavior, and performance-sensitive queries.

## Directory Layout

Recommended layout inside one generated microservice:

```text
app/<service>/service/
  gen/
    <domain>.yaml

  internal/
    service/
      <resource>_service.gen.go
      <resource>_service_ext.go
      <resource>_service_manual.go
      providers/
        wire_set.gen.go

    data/
      <resource>_repo.gen.go
      <resource>_repo_ext.go
      <resource>_repo_manual.go
      providers/
        wire_set.gen.go

    server/
      rest_register.gen.go
      grpc_register.gen.go
      rest_server.go
      grpc_server.go
```

## Generation Boundaries

### Service Layer

The generator may emit:

- service struct definitions
- constructor functions
- standard CRUD RPC methods
- base request validation
- calls into repo methods
- hook points for custom logic

The generator must not emit:

- cross-resource aggregation
- custom authorization rules
- bootstrap data initialization
- domain-specific workflows
- upload and download transport handling

Typical generated shape:

```go
type UserService struct {
	adminv1.UserServiceHTTPServer
	log  *log.Helper
	repo data.UserRepo
}

func NewUserService(ctx *bootstrap.Context, repo data.UserRepo) *UserService { ... }

func (s *UserService) List(ctx context.Context, req *paginationv1.PagingRequest) (*identityv1.ListUserResponse, error) {
	return s.repo.List(ctx, req)
}
```

Hooks should be routed through extension files:

- `beforeCreate`
- `afterCreate`
- `beforeUpdate`
- `afterList`
- `enrichRelations`

### Repository Layer

The generator may emit:

- repo interface definitions
- base repo structs
- CRUD method skeletons
- standard transaction wrappers
- mapper initialization
- field-based filtering and sorting
- basic relation attach and detach templates

The generator must not emit:

- non-trivial relation intersection logic
- multi-tenant business branches
- performance-tuned custom SQL
- cross-entity consistency rules

### Wire Layer

The generator should fully own:

- `internal/service/providers/wire_set.gen.go`
- `internal/data/providers/wire_set.gen.go`

These files are list-oriented and have no business logic.

### Transport Registration Layer

The generator should emit registration lists only:

- `rest_register.gen.go`
- `grpc_register.gen.go`

Manual server bootstrap files remain hand-written and call generated registration helpers.

Recommended pattern:

```go
func NewRestServer(...) (*http.Server, error) {
	srv, err := ...
	if err != nil {
		return nil, err
	}

	registerGeneratedHTTPServices(srv, services)
	registerManualHTTPServices(srv, specialHandlers)

	return srv, nil
}
```

## Resource Naming Rules

- Resource names use singular nouns: `user`, `role`, `tenant`
- Service types use PascalCase: `UserService`
- Repo types use PascalCase: `UserRepo`
- Generated file names use snake_case with `.gen.go`

Examples:

- `user_service.gen.go`
- `user_repo.gen.go`
- `wire_set.gen.go`

## CRUD Classification

RPCs are classified into three buckets.

### 1. Standard CRUD

These are safe to generate by default:

- `List`
- `Get`
- `Create`
- `Update`
- `Delete`
- `Count`
- `Exists`

### 2. Semi-Standard Operations

These should generate a function shell and hook points, but not full business logic:

- `BatchCreate`
- `BatchDelete`
- `Import`
- `Export`
- `AssignRole`
- `EditUserPassword`

### 3. Special Operations

These should only participate in registration generation and must remain manual:

- login and auth flows
- file upload and download
- portal aggregation APIs
- workflow triggers
- long-running task orchestration

## YAML Configuration Schema

Recommended minimum schema:

```yaml
service: admin
module: xadmin-web

resources:
  - name: user
    proto_service: admin.service.v1.UserService
    entity: User
    repo_interface: UserRepo

    operations:
      list: true
      get: true
      create: true
      update: true
      delete: true
      count: true
      exists: true

    paging:
      enabled: true
      request: pagination.v1.PagingRequest

    filters:
      allow:
        - id
        - username
        - tenant_id
        - status

    relation_filters:
      - role_ids
      - org_unit_ids
      - position_ids

    enrich:
      enabled: true
      handlers:
        - roles
        - tenant
        - org_units
        - positions

    generate:
      service_stub: true
      repo_crud: true
      rest_register: true
      grpc_register: true
      wire: true
```

## Regeneration Rules

These rules are mandatory.

1. Only overwrite `*.gen.go`.
2. Never overwrite `*_ext.go`.
3. Never overwrite `*_manual.go`.
4. If an extension file is missing, generate it once with empty hooks.
5. If a manual file exists, do not patch it automatically unless an explicit code-mod mode is introduced later.

Without these rules, repeated generation will not be safe enough for real projects.

## Command Design

Recommended initial CLI surface inside `xkit`:

```text
xkit gen service <service>
xkit gen repo <service>
xkit gen wire <service>
xkit gen register <service>
xkit gen all <service>
```

Behavior:

- `xkit gen service <service>`: generate service `*.gen.go` and extension stubs
- `xkit gen repo <service>`: generate repo `*.gen.go` and extension stubs
- `xkit gen wire <service>`: generate service and data provider sets
- `xkit gen register <service>`: generate HTTP and gRPC registration helpers
- `xkit gen all <service>`: run the full pipeline, overwriting only `*.gen.go`

## Minimum Viable Rollout

The first implementation should stay narrow.

### Phase 1

Implement:

- proto-driven service method stubs
- proto-driven transport registration files
- constructor list generation for Wire
- YAML-driven resource selection

Do not implement full repo CRUD generation in phase 1.

### Phase 2

Add:

- ent-driven repo interface generation
- base CRUD repo skeletons
- standard filter and sort generation

### Phase 3

Add:

- richer hook contracts
- explicit safe patching for extension files
- validation generation
- batch operation scaffolds

## Practical Boundaries

The current combined reference context shows three realities:

1. `xadmin-web` already has regular proto and ent schema inputs that are stable enough for generation.
2. `xadmin-web` does not yet have a mature generated service/data/server layer, so the initial tool must create that structure instead of learning from local implementations.
3. `D:\GoProjects\chnxq\XAdmin` contains the primary service/data/server implementation patterns that `xkit` should reproduce safely.

That means the generator should first target the stable structural layer:

- service shells
- repo shells
- wire provider lists
- registration lists

It should not try to auto-solve the business-heavy parts until the generated ownership boundaries are stable.

## Open Questions

These questions should be answered before implementing repo generation aggressively:

- How much filter behavior should come from YAML versus ent metadata?
- Should relation enrichment be declarative or always manual?
- Should generated repos target `ent`, `gorm`, or both in the first implementation?
- Should `xkit gen all` fail when required YAML metadata is missing?
- Does `xkit` need a later compatibility wrapper for existing `gow` workflows?

## Recommended Next Step

Implement phase 1 in `xkit` first:

- bootstrap the `xkit` CLI entry and `gen` command group
- model output layout after `D:\GoProjects\chnxq\XAdmin\app\admin\service`
- add generation for `service/*.gen.go`
- add generation for `server/*_register.gen.go`
- add generation for `providers/wire_set.gen.go`
- validate the first slice against `xadmin-web/api/protos/admin/v1/i_user.proto`
- prepare ent-driven repo generation against `xadmin-web/internal/data/ent/schema/user.go` after phase 1 is stable

That gives immediate value with low risk and establishes the file ownership model before repo generation is introduced.



