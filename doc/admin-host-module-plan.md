# Admin Host + Module Plan

Generated at: 2026-06-18 +08:00

## 1. Goal

This document evaluates and refines the proposed direction:

- `admin` is treated as a host application
- new business capabilities are added as modules
- the system still exposes one unified Admin service externally
- module code stays relatively isolated internally
- `xkit` evolves from flat resource generation to host-aware module generation

The target is not an abstract plugin system first. The target is a practical,
incremental structure that works with the current `admin` and `xkit`
repositories.

## 2. Conclusion

The proposed direction is feasible and is the right long-term direction.

However, it is only feasible if we split three concerns clearly:

1. host-owned infrastructure
2. module-owned business code
3. stable shared code moved out of `admin/internal/*` into a host-visible shared package

If this split is not made explicit, the module layout will only become a
different folder shape around the same flat coupling.

The core rule should be:

- `admin` owns runtime assembly and infrastructure
- each module owns its own schema/proto/repo/service/server glue
- shared cross-module helpers move to `admin/shared` or `admin/common`
- module code must not directly depend on unrelated host internals

## 3. Why The Direction Is Sound

The current `admin` repo already behaves like a host in runtime terms:

- `internal/bootstrap/generated_servers.gen.go` assembles generated repos and services
- `internal/bootstrap/generated_data_providers.gen.go` exposes repo providers through `GeneratedData`
- `internal/server/http.go` and `grpc.go` start the transport servers
- `internal/server/rest_register.gen.go` and `grpc_register.gen.go` aggregate generated service registration
- `internal/server/manual_http_data.go` is already a host-side extension point

So the runtime model is already close to:

- one host
- many business services attached to that host

The main problem is not runtime architecture. The main problem is code layout:

- generated resource code is still flat under:
  - `internal/service`
  - `internal/data/repo`
  - `internal/server`
- cross-resource handwritten helpers also accumulate there

That layout does not scale when many future modules are added.

## 4. Proposed Repository Shape

### 4.1 Host-owned directories

These should remain in the host project:

```text
admin/
  cmd/
  configs/
  internal/
    bootstrap/
    server/
    data/
      ent/
      bootstrap/
      repo/
        shared/
    service/
      shared/
  shared/
```

Responsibilities:

- `internal/bootstrap`
  - application assembly
  - generated host aggregation
  - lifecycle hooks
- `internal/server`
  - HTTP/gRPC server construction
  - registration orchestration
  - host-level auth/viewer/middleware/manual endpoints
- `internal/data/bootstrap`
  - DB/Redis/cache/queue/object-storage initialization
- `internal/data/repo/shared`
  - host-visible shared repo helpers
- `internal/service/shared`
  - host-visible shared service helpers
- `shared`
  - stable cross-module contracts and utility code that modules are allowed to import

### 4.2 Module-owned directories

New module code should prefer this structure:

```text
admin/modules/
  device/
    api/
      protos/
      gen/
    data/
      schema/
      repo/
      bootstrap/
    service/
    server/
    bootstrap/
    module.go
  order/
  asset/
```

Responsibilities:

- `api/protos`
  - module-owned proto source
- `api/gen`
  - generated module dto / pb / http / grpc code
- `data/schema`
  - module-owned Ent schemas
- `data/repo`
  - generated and handwritten repo code for this module
- `service`
  - generated and handwritten service code for this module
- `server`
  - module-specific transport adapters or manual route helpers
- `bootstrap`
  - module-local data or service wiring helpers
- `module.go`
  - module entrypoint used by host assembly

## 5. Import Boundary Rules

This is the most important part of the design.

### 5.1 Allowed dependency direction

The dependency direction should be:

```text
admin host
  -> shared
  -> modules/*

modules/*
  -> shared
  -> xkit-generated dto under their own module

modules/*
  -X-> unrelated modules/*
  -X-> arbitrary admin/internal host implementation packages
```

In practice:

- module repo/service code may import:
  - `admin/shared/...`
  - their own `admin/modules/<module>/api/gen/...`
  - their own module-local packages
- module code should not import:
  - `admin/internal/server/...`
  - unrelated module implementation packages
  - host-only bootstrap internals unless explicitly exposed

### 5.2 Why `admin/internal` is a problem for future module repos

You already identified the key issue:

- if `admin/modules/device` is expected to become an independent git repo later,
  then direct imports from `admin/internal/...` are structurally wrong

Go `internal/` visibility is only allowed inside the same parent tree. That works
inside the host repo today, but it creates the wrong dependency shape for future
module separation.

Therefore:

- code that modules must see must move out of `admin/internal/...`
- such code should live under:
  - `admin/shared/...`
  - or `admin/common/...`

Recommended naming:

- prefer `shared/` for stable host-visible reusable code
- avoid `common/` if it turns into a vague dumping ground

## 6. What Should Move To `admin/shared`

Move code that is:

1. needed by multiple modules
2. not tied to one resource
3. safe for module code to import

Likely candidates:

- viewer / operator / tenant context helpers
- shared repo filter helpers
- list sorting helpers
- shared auth-facing error conversion helpers
- transport metadata helpers
- stable DTO mapping helpers
- module-facing interfaces / provider contracts

Likely not good candidates:

- host HTTP server construction
- host grpc server construction
- host-only registration defaults
- manual auth flow handlers tightly coupled to host startup

## 7. Module Entry Contract

The proposed `Module` interface is directionally correct, but the exact method
shape should be adjusted to fit the current host.

Current proposal:

```go
type Module interface {
    Name() string
    RegisterData(*bootstrap.GeneratedData) error
    RegisterServices(*bootstrap.GeneratedServices) error
    RegisterHTTP(*httptransport.Server, bootstrap.GeneratedServices) error
    RegisterGRPC(grpc.ServiceRegistrar, bootstrap.GeneratedServices) error
}
```

This is feasible, but there is a practical issue:

- `GeneratedData` and `GeneratedServices` are currently generated host structs,
  not stable contracts

If modules depend on those concrete generated structs directly, the module
boundary is weak.

Recommended refinement:

### 7.1 Near-term practical form

Use a thin function-based module entry first:

```go
type HostData interface {
    GetAppCtx() *app.AppCtx
}

type HostServices interface{}

type Module interface {
    Name() string
    RegisterData(HostData) error
    RegisterServices(HostData, HostServices) error
    RegisterHTTP(*httptransport.Server, HostServices) error
    RegisterGRPC(grpc.ServiceRegistrar, HostServices) error
}
```

This keeps the first version simple while allowing `GeneratedData` and
`GeneratedServices` to satisfy host-side interfaces.

### 7.2 Better long-term form

Later, stabilize explicit host contracts:

- `shared/hostctx`
- `shared/providers`
- `shared/moduleapi`

Then modules depend on stable host contracts rather than generated bootstrap
structs.

## 8. How The Host Should Mount Modules

Your statement that later it may be enough to add one line in `http.go` and one
line in `grpc.go` is conceptually right, but the host should mount modules one
level earlier.

Preferred mounting flow:

1. host builds base infra
2. host constructs generated host data/services
3. host loads module list
4. host lets each module register data/services if needed
5. host builds HTTP/gRPC servers
6. host lets each module register transport handlers

This means:

- `internal/bootstrap` should own the module list
- `internal/server/http.go` and `grpc.go` should not become the main place that
  knows every module name

Better shape:

- `internal/bootstrap/modules.go`
  - returns enabled modules
- `internal/bootstrap/generated_hooks_ext.go`
  - can be used as a transitional hook

Then:

- `http.go` and `grpc.go` only iterate over mounted modules
- they do not hardcode `device`, `order`, `asset`, etc.

## 9. Implications For `xkit`

Your proposed new commands are justified.

The current command split:

- `xkit init template`
- `xkit init source`
- `xkit gen service/repo/register/bootstrap/frontend-meta/all`

is resource-project oriented.

For a host + module future, `xkit` needs a module-oriented workflow.

### 9.1 `xkit init module`

This command is feasible and necessary.

It should combine parts of:

- `init template`
- `init source`

But only for a module scope, not a whole project scope.

Suggested purpose:

- initialize a new module inside an existing host
- lay down module skeleton
- copy module source schema/proto input
- rewrite module-local buf and go package paths
- prepare module config

Suggested command shape:

```text
xkit init module <source-path> \
  --host-project <path> \
  --module-name <name> \
  --service <service-name> \
  --module-root <relative-path> \
  --config <path> \
  --typescript-project <path> \
  [--force] [--dry-run]
```

Suggested responsibilities:

- create:
  - `admin/modules/<module>/api/protos`
  - `admin/modules/<module>/data/schema`
  - `admin/modules/<module>/data/repo`
  - `admin/modules/<module>/service`
  - `admin/modules/<module>/server`
  - `admin/modules/<module>/bootstrap`
  - `admin/modules/<module>/module.go`
- copy:
  - schema
  - proto
  - module-local buf config
- generate or derive:
  - module-local YAML config
  - go_package normalization
  - imports pointing at host module path

### 9.2 `xkit gen module`

This command is also justified.

It should not merely alias `gen all`. It should scope output to one module.

Suggested command shape:

```text
xkit gen module <module-name> \
  --host-project <path> \
  --config <path> \
  --service <service-name> \
  --typescript-project <path> \
  [--domain <name>] [--dry-run]
```

Suggested responsibilities:

- generate module-local:
  - repo
  - service
  - transport register glue
  - bootstrap glue
  - optional frontend meta
- update host aggregation points:
  - module mounting list
  - generated host registration adapters if needed

### 9.3 Additional xkit requirements

To make module generation work cleanly, `xkit` will also need:

1. configurable generation root
   - today output assumptions are mostly flat under `internal/service`,
     `internal/data/repo`, `internal/server`
   - module mode must let these roots become:
     - `modules/<name>/service`
     - `modules/<name>/data/repo`
     - `modules/<name>/server`

2. host-aware shared import rules
   - module-generated code should import `admin/shared/...` for shared helpers
   - not hardcode host internals

3. host aggregation templates
   - generation must optionally touch:
     - host module list
     - host register adapter
     - host bootstrap mounting hooks

4. stable module template support
   - `module.go`
   - module-local bootstrap placeholders
   - module-local README if desired

## 10. Recommended Migration Strategy

Do not refactor all existing `admin` resources into modules at once.

Recommended sequence:

### Phase 1: define boundaries

- create `admin/shared`
- identify and move a minimal set of module-visible shared helpers there
- keep current flat generated resources working

### Phase 2: pilot one module

- use `device` as the pilot module
- place its code under `admin/modules/device/...`
- manually adapt host mounting for this one module
- verify:
  - no import cycles
  - host startup still clean
  - generated code can be re-run safely

### Phase 3: add xkit module mode

- implement `xkit init module`
- implement `xkit gen module`
- add templates for module layout

### Phase 4: migrate selected future additions

- all future new business capabilities use module mode by default
- old flat resources stay as-is until there is clear migration value

### Phase 5: optionally back-migrate old domains

- only migrate old flat resources to module layout when:
  - they are actively changing
  - the migration buys real clarity
  - tests are sufficient

## 11. Risks

### 11.1 Fake isolation risk

If modules still import `admin/internal/...` directly, the new structure only
creates folder nesting, not real isolation.

### 11.2 Over-abstracted plugin risk

If we overdesign a generic plugin system before one module works end-to-end, the
implementation cost will outrun the actual problem.

### 11.3 Generator drift risk

If host aggregation and module-local output paths are both generated without a
clear ownership boundary, repeated regeneration will create unstable diffs.

### 11.4 Future split risk

If a module is expected to become an independent repository later, every direct
dependency on host internals becomes migration debt.

## 12. Recommended Immediate Next Steps

1. Write and adopt a concrete host/module boundary rule for `admin`.
2. Create `admin/shared` and move only the minimum module-visible helpers there.
3. Use `device` as the first pilot module under `admin/modules/device`.
4. Do not generate into flat `internal/service` and `internal/data/repo` for
   this pilot.
5. Add a host-side module mounting hook in `internal/bootstrap` rather than
   hardcoding module names directly in `http.go` and `grpc.go`.
6. Design `xkit init module` and `xkit gen module` around this pilot rather than
   inventing command semantics first.

## 13. Proposed xkit Work Items

### 13.1 CLI

- add `xkit init module`
- add `xkit gen module`
- keep existing project/resource commands unchanged for backward compatibility

### 13.2 Config

- add module-aware config fields, for example:

```yaml
host_project: admin
module_name: device
module_root: modules/device
shared_import_root: admin/shared
```

### 13.3 Templates

- add module-local templates for:
  - `module.go`
  - module bootstrap hooks
  - module server register wrappers

### 13.4 Codegen routing

- make output roots configurable per layer:
  - schema root
  - repo root
  - service root
  - server root
  - bootstrap root

### 13.5 Host integration

- define generated or semi-generated module mounting points in the host

## 14. Decision

Proceed with the host + module direction.

But proceed with a narrow pilot:

- one real module
- one host
- one shared boundary cleanup
- then evolve `xkit`

That is the lowest-risk path that still moves toward the long-term modular goal.
