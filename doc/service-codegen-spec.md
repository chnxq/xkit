# Service Code Generation Spec

## Purpose

This document defines the first workable specification for service-oriented code generation in `xkit`.

`xkit` is the implementation repository for the generator.
The primary target project is `D:\GoProjects\XAdmin\xadmin-web`.
The primary implementation reference for service/data/server code is `D:\GoProjects\chnxq\XAdmin`. Generator behavior, project discovery, scaffold conventions, and service output should follow the XAdmin reference first.

The goal is not to generate complete business systems.
The goal is to standardize and regenerate the highly repetitive parts of an XAdmin-style Go service safely:

- service interface adapters
- repository CRUD scaffolds
- transport registration code
- generated bootstrap glue for explicit startup assembly

The design must preserve manual business logic and allow repeated generation without overwriting hand-written code.

## Scope

This specification targets XAdmin-aligned source inputs and generated service outputs.

### Current Source Layout In XAdmin

- raw source directory, currently `source/`
- `source/api/protos/<domain>/v1/*.proto`
- `source/schema/*.go`
- active project copies under `api/protos/<domain>/v1/*.proto`
- `api/gen/<domain>/v1/*`
- `internal/data/ent/schema/*.go`

### Target Generated Layout

- `internal/service/*`
- `internal/data/*`
- `internal/server/*`

Representative examples in the current target workspace:

- `xadmin-web/internal/service/user_service.gen.go`
- `xadmin-web/internal/data/user_repo.gen.go`
- `xadmin-web/internal/server/rest_register.gen.go`
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

- `<target>/source/<project-name>-config/<service>.yaml` when config is derived from a raw source directory
- an explicit path passed with `--config` when a project wants a different layout
- `xkit/examples/<target>/<service>.yaml` only for generator development fixtures

## Raw Source Import

`xkit init source` prepares a target project from a raw source directory before `xkit gen ...` runs.

Supported source layouts:

```text
<source>/
  api/protos/
  schema/
```

Also accepted:

- `<source>/protos`
- `<source>/data/schema`
- `<source>/internal/data/ent/schema`

Command:

```text
xkit init source <source-path> --project <target> --service <service>
```

Example:

```text
xkit init source D:\GoProjects\XAdmin\xadmin-web\source \
  --project D:\GoProjects\XAdmin\xadmin-web \
  --service admin
```

The command copies:

- `<source>/api/protos` to `<target>/api/protos`
- `<source>/schema` to `<target>/internal/data/ent/schema`

It also derives a generation config from the imported proto services and Ent schemas. By default, the config is written to:

```text
<source>/<project-name>-config/<service>.yaml
```

For `xadmin-web` and `admin`, this becomes:

```text
xadmin-web/source/xadmin-web-config/admin.yaml
```

The config generation rules are intentionally conservative:

- Prefer `<service>.service.v1.<Entity>Service`, for example `admin.service.v1.UserService`.
- If there is no admin-facing service, keep a domain service only when it has supported repo operations, for example `authentication.service.v1.UserCredentialService` with `ResetCredential`.
- Skip schemas that have no matching service proto, such as pure relation tables or embedded/detail entities.
- Infer `dto_import` from method request/response packages.
- Infer `filters.allow` from Ent fields whose kinds are supported by generated filtering.
- Infer `exists_fields` from `*ExistsRequest.oneof query_by` when present.

By default, existing target files are not overwritten. Use `--force` to overwrite target proto/schema/config files, and `--dry-run` to inspect the write plan without changing files.

After source import, the normal generation command consumes the produced config:

```text
xkit gen all admin \
  --project D:\GoProjects\XAdmin\xadmin-web \
  --config D:\GoProjects\XAdmin\xadmin-web\source\xadmin-web-config\admin.yaml
```

## File Ownership Model

The generator must follow strict ownership boundaries.

### Generated Files

Generated files are fully owned by the generator and may be overwritten:

- `*.gen.go`

Examples:

- `user_service.gen.go`
- `user_repo.gen.go`
- `rest_register.gen.go`
- `generated_servers.gen.go`

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
internal/
  gen/
    <service>.yaml

  service/
    <resource>_service.gen.go
    <resource>_service_ext.go
    <resource>_service_manual.go

  data/
    <resource>_repo.gen.go
    <resource>_repo_ext.go
    <resource>_repo_manual.go

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

### Bootstrap Assembly Layer

The default generator flow should use explicit generated startup assembly:

- `internal/bootstrap/generated_servers.gen.go`

This file constructs generated data, services, and HTTP/gRPC transport servers directly. Because the template startup path no longer calls a Wire injector, `xkit gen all` must not emit `internal/service/providers/wire_set.gen.go` or `internal/data/providers/wire_set.gen.go`. Existing generated Wire provider-set files are treated as manual migration leftovers and are not removed automatically.

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
- `generated_servers.gen.go`

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
    dto_import: xadmin-web/api/gen/identity/v1
    dto_type: User
    repo_interface: UserRepo
    exists_fields:
      - id
      - username

    operations:
      list: true
      get: true
      count: true
      create: true
      update: true
      delete: true
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
```

Field notes:

- `entity`: Ent schema/entity name used to discover fields and build query/create/update code.
- `dto_import`: Go import path for the DTO type used by generated repo methods.
- `dto_type`: DTO type name, for example `User`.
- `repo_interface`: generated data-layer repository interface name.
- `exists_fields`: allowed fields for generated existence queries; this avoids hardcoding business fields in the generator.

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
xkit init source <source-path>
xkit gen service <service>
xkit gen repo <service>
xkit gen wire <service>
xkit gen register <service>
xkit gen all <service>
```

Behavior:

- `xkit init source <source-path>`: copy raw schema/proto files into the target project and derive the per-service YAML config
- `xkit gen service <service>`: generate service `*.gen.go` and extension stubs
- `xkit gen repo <service>`: generate repo `*.gen.go` and extension stubs
- `xkit gen wire <service>`: legacy explicit command for service and data provider sets; not used by the default template startup path
- `xkit gen register <service>`: generate HTTP and gRPC registration helpers
- `xkit gen all <service>`: generate service, repo, register, and bootstrap glue, overwriting only generated files; it does not generate or clean Wire provider sets

## Current Rollout

The first implementation has moved beyond the original phase-1 service-only
slice and now covers the core generated service stack.

### Implemented

- proto-driven service method stubs
- proto-driven HTTP and gRPC registration files
- explicit generated bootstrap assembly in `internal/bootstrap/generated_servers.gen.go`, split into generated data, services, components, and transport construction
- YAML-driven resource selection
- Ent-schema-driven repository CRUD scaffolds
- config-driven existence-query field selection via `exists_fields`
- generated headers containing `xkit` version and generation timestamp

### Next Additions

- standard filter and sort generation
- richer hook contracts
- explicit safe patching for extension files
- validation generation
- batch operation scaffolds

## Practical Boundaries

The current combined reference context shows three realities:

1. `xadmin-web` already has regular proto and ent schema inputs that are stable enough for generation.
2. `xadmin-web` does not yet have a mature generated service/data/server layer, so the initial tool must create that structure instead of learning from local implementations.
3. `D:\GoProjects\chnxq\XAdmin` contains the primary service/data/server implementation patterns that `xkit` should reproduce safely.

That means the generator targets the stable structural layer first:

- service shells
- repo CRUD scaffolds
- registration lists
- bootstrap glue

It should not try to auto-solve the business-heavy parts until the generated ownership boundaries are stable.

## Open Questions

These questions should be answered before expanding repo generation aggressively:

- How much filter behavior should come from YAML versus ent metadata?
- Should relation enrichment be declarative or always manual?
- Should generated repos continue to target `ent` only, or later support `gorm` as well?
- Should `xkit gen all` fail when required YAML metadata is missing?
- Does `xkit` need a later compatibility wrapper for existing `gow` workflows?

## Recommended Next Step

Continue hardening `xkit gen repo` against real XAdmin resources:

- broaden Ent field-to-DTO conversion coverage
- add declarative filters and sorting
- keep business-specific fields in YAML instead of generator code
- validate generated output against `xadmin-web` service packages after each generator change

That keeps the generator useful while preserving explicit ownership boundaries between generated scaffolding and manual business logic.



