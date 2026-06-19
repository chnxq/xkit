# Admin Host Module Implementation Plan

Generated at: 2026-06-20 +08:00

## 1. Purpose

This document turns the host + module idea into an implementation-oriented plan.

It follows the requested exploration order:

1. use current `xkit` behavior to explore copying `examples/xdev` schema and
   proto into `admin/modules/xdev`
2. identify which code from `xkit-template` would need module-local copies
3. analyze which current `admin/internal/*` dependencies are actually shared and
   should move into `xkit-template/shared` so future generated module code can
   import them safely

The output of this document is intended to become the basis for:

- `xkit init module`
- `xkit gen module`
- a concrete pilot implementation using `device` / `xdev`

This document now also records the post-exploration pilot status after the
first real `xdev` module generation pass against the host project.

## 2. Executive Summary

The exploration shows that the direction is feasible, but current `xkit` is
still structurally project-flat.

Current blockers are not mainly in schema/proto syntax anymore. They are in
hardcoded output and import assumptions:

- `init source` always copies proto to `<project>/api/protos`
- `init source` always copies schema to `<project>/internal/data/ent/schema`
- `gen all` assumes:
  - `api/protos`
  - `api/gen`
  - `internal/data/ent/schema`
  - `internal/data/repo`
  - `internal/service`
  - `internal/server`
  - `internal/bootstrap`
- generated imports assume `{{module}}/internal/...`

So the first useful exploration result is:

- current `xkit` cannot generate a module into `admin/modules/xdev/...` only by
  changing config and CLI parameters

That is not a failure of the idea. It is a confirmation that `xkit` needs
module-aware output roots and host-aware shared import boundaries.

## 3. Exploration Result 1: Reusing Current xkit To Copy xdev Into `admin/modules/xdev`

### 3.1 Goal of the exploration

Try to reuse current `xkit` flow as much as possible:

- source example remains `xkit/examples/xdev`
- target host remains `D:\GoProjects\XAdmin\admin`
- intended module target would be `admin/modules/xdev`

### 3.2 What works today

The current `xdev` source input was already refined so that:

- `buf lint` passes
- `buf generate --template buf.gen.yaml` passes
- `xkit init source ... --dry-run` passes for the source input itself

This means the source sample is now valid as a controlled CRUD input set.

### 3.3 What does not work with current xkit

The current code hardcodes project-flat roots.

Observed hardcoded target roots in `init source`:

- proto copy target:
  - `<project>/api/protos`
- API root files:
  - `<project>/api`
- schema copy target:
  - `<project>/internal/data/ent/schema`

Observed hardcoded assumptions in `gen all`:

- proto service lookup under:
  - `api/protos`
- binding service lookup under:
  - `api/gen`
- Ent schema lookup under:
  - `internal/data/ent/schema`
- generated output under:
  - `internal/service`
  - `internal/data/repo`
  - `internal/server`
  - `internal/bootstrap`

Observed hardcoded import routing:

- `internalImport("service")` -> `<module>/internal/service`
- `internalImport("data", "repo")` -> `<module>/internal/data/repo`
- `internalImport("server")` -> `<module>/internal/server`
- generated bootstrap imports host internals directly

### 3.4 Conclusion of exploration 1

Trying to make current `xkit` generate into `admin/modules/xdev` only by:

- editing Buf config
- changing module/service name
- changing CLI args

is not enough.

Current `xkit` needs explicit support for:

1. module-local source destinations
2. module-local generated output destinations
3. host-owned shared import roots
4. host aggregation mounting points

This validates the need for `xkit init module` and `xkit gen module`.

## 4. Exploration Result 2: What From `xkit-template` Needs Module-local Copies

### 4.1 Current `xkit-template` shape

Current template content is still host-oriented. It contains:

- `cmd/server/*`
- `configs/*`
- `internal/bootstrap/*`
- `internal/server/*`
- `internal/data/bootstrap/*`

It also currently contains placeholder generated files:

- `internal/bootstrap/generated_servers.gen.go`
- `internal/bootstrap/generated_data_providers.gen.go`
- `internal/server/rest_register.gen.go`
- `internal/server/grpc_register.gen.go`

### 4.2 What should remain host-only

These should remain in the host and should not be copied per module:

- `cmd/server/*`
- `configs/*`
- `internal/bootstrap/app.go`
- `internal/bootstrap/infra.go`
- `internal/bootstrap/cleanup.go`
- `internal/server/http.go`
- `internal/server/grpc.go`
- `internal/server/http_options.go`
- `internal/server/grpc_options.go`
- `internal/server/server.go`
- `internal/data/bootstrap/*`

Reason:

- they represent one runtime host, not one business module

### 4.3 What a module needs as its own skeleton

A module should get its own local scaffold, but not a copy of the full host
template.

Suggested module-local skeleton:

```text
modules/<module>/
  api/
    protos/
  data/
    schema/
    repo/
    bootstrap/
  service/
  server/
  bootstrap/
  module.go
```

From current template logic, the following categories need module-local
equivalents:

- module bootstrap placeholder
- module register placeholder
- module-local manual server extension point if needed
- module entry file

### 4.4 Minimal files that should come from a future module template

The first version of `xkit init module` should create at least:

- `modules/<module>/module.go`
- `modules/<module>/bootstrap/hooks.go`
- `modules/<module>/server/manual_http.go`
- `modules/<module>/server/manual_grpc.go` or skip until needed
- `modules/<module>/README.md` optional

These are not full host template copies. They are module scaffold files.

### 4.5 Conclusion of exploration 2

The current `xkit-template` should not be copied into each module.

Instead:

- host template remains host-only
- a new module template layer is needed
- `xkit init module` should combine:
  - module scaffold creation
  - source import
  - config derivation

## 5. Exploration Result 3: Which `admin/internal/*` Dependencies Are Actually Shared

### 5.1 Problem statement

Future module code under `admin/modules/xdev` should not directly depend on
random host internals, especially if a module may later become an independent
repository.

Therefore, we need to identify which current host internals are actually shared
capabilities and should be moved into a host-visible shared location.

### 5.2 Observed shared-pattern candidates in current `admin`

Likely shared candidates:

- tenant context guard logic
- viewer context helpers
- list sorting helpers
- runtime viewer bridge
- repo filter helper patterns
- cross-tenant access guard helpers
- generic server auth/viewer metadata helpers

Files or concepts that likely belong in future `shared`:

- `internal/data/repo/tenant_scope_ext.go`
- `internal/data/repo/list_sorting_ext.go`
- `internal/data/repo/runtime_viewer_ext.go`
- `internal/service/tenant_context_guard_ext.go`
- parts of `internal/server/viewer_auth.go`
- stable host-visible contracts currently implicit in `GeneratedData`

### 5.3 What should not move to shared

These are host-owned, not generally shared:

- `internal/server/http.go`
- `internal/server/grpc.go`
- `internal/server/manual_http_data.go`
- auth flow handlers tightly tied to host startup
- host bootstrap construction
- generated host registration files

### 5.4 Recommended shared location

Do not keep module-visible shared code under `admin/internal/...`.

The pilot settled on a concrete first shared root:

```text
admin/shared/
  modulex/
    module_shared_ext.go
```

Rationale:

- it is visible to generated module code without importing `admin/internal/...`
- it keeps the first extraction small and focused
- it gives a stable hand-written extension point for future modules

The current `modulex` helper file already centralizes:

- viewer context helpers
- tenant visibility and mutation guards
- default paging and sorting helpers
- tenant display-name helper functions
- tenant-name lookup by host Ent client

This is intentionally not yet the final shared package taxonomy. It is the
first stable landing zone that works with real generated code.

### 5.5 Role of `xkit-template/shared`

To support future generation, `xkit-template` should eventually contain a
`shared/` directory that is copied to target projects by `xkit init template`.

That directory should hold only:

- stable module-visible shared helpers
- interfaces and contracts intended for generated module code

It should not hold:

- business resource logic
- host-only startup logic

### 5.6 Conclusion of exploration 3

The first shared extraction pass is now partially complete.

Implemented in the pilot:

- `admin/shared/modulex/module_shared_ext.go`
- module-mode repo generation imports `admin/shared/modulex`
- module-mode no longer depends on per-module generated shared helper files

Still not complete:

- existing host resources have not been migrated from `admin/internal/...` to
  the shared package
- `xkit-template/shared` has not yet been introduced as a reusable source
- the host module mounting chain is still manual / pending

Without the remaining work:

- future modules will still depend on host-specific shared layout assumptions
- shared helper evolution will not yet be template-driven

## 6. What `xkit init module` Should Do

Based on the three exploration results, `xkit init module` should not just be a
wrapper around existing commands. It needs new behavior.

### 6.1 Intended responsibility

`xkit init module` should:

1. create a module scaffold under an existing host project
2. copy module schema/proto source into module-local directories
3. normalize module-local Buf config
4. derive a module-local generation config
5. prepare host aggregation hooks or mounting stubs

### 6.2 Suggested command shape

```text
xkit init module <source-path> \
  --host-project <path> \
  --module-name <name> \
  --module-root <relative-path> \
  --service <service-name> \
  --config <path> \
  --typescript-project <path> \
  [--force] [--dry-run]
```

Suggested first pilot values:

- `source-path`: `xkit/examples/xdev`
- `host-project`: `D:\GoProjects\XAdmin\admin`
- `module-name`: `xdev`
- `module-root`: `modules/xdev`
- `service-name`: `admin` or a host-facing service group name depending on final API packaging

### 6.3 Files it should create

At minimum:

- `admin/modules/xdev/api/protos/...`
- `admin/modules/xdev/data/schema/...`
- `admin/modules/xdev/data/repo/...`
- `admin/modules/xdev/service/...`
- `admin/modules/xdev/server/...`
- `admin/modules/xdev/bootstrap/...`
- `admin/modules/xdev/module.go`
- module-local buf config files
- module-local YAML config

Current pilot status:

- this is already working for `xdev`
- `init module` uses `moduleRoot`, defaulting to
  `hostProject/modules/<moduleName>`
- source import is module-aware and no longer routes through project-flat
  `api/protos` or `internal/data/ent/schema`

### 6.4 What it must not do

It should not copy:

- host server main entry
- host configs
- host `internal/bootstrap`
- host `internal/server/http.go`
- host `internal/server/grpc.go`

## 7. What `xkit gen module` Should Do

### 7.1 Intended responsibility

`xkit gen module` should:

- generate module-local repo/service/register/bootstrap glue
- generate or reuse host shared helper glue required by module-mode code
- optionally generate frontend meta
- update host mounting integration if that part is generated or semi-generated

### 7.2 Suggested command shape

```text
xkit gen module <module-name> \
  --host-project <path> \
  --module-root <relative-path> \
  --config <path> \
  --typescript-project <path> \
  [--domain <name>] [--dry-run]
```

### 7.3 New capabilities required in generator

The current generator will need:

1. configurable source roots
   - proto root
   - schema root
   - api gen root

2. configurable output roots
   - repo root
   - service root
   - server root
   - bootstrap root

3. configurable import roots
   - module local imports
   - shared imports
   - host mount imports

4. host integration output
   - module mounting registry or hook update

### 7.4 Current pilot status

The following module-generation behavior is now real and verified:

1. module-local generation roots
   - proto lookup under `modules/<module>/api/protos`
   - binding lookup under `modules/<module>/api/gen`
   - schema lookup under `modules/<module>/data/schema`

2. module-local output roots
   - repo output under `modules/<module>/data/repo`
   - service output under `modules/<module>/service`
   - server output under `modules/<module>/server`
   - bootstrap output under `modules/<module>/bootstrap`

3. host shared helper output
   - module mode writes or reuses:
     - `admin/shared/modulex/module_shared_ext.go`
   - if the file already exists, generation skips it

4. module-mode helper import behavior
   - generated repo code imports `admin/shared/modulex`
   - module mode no longer requires generated:
     - `modules/<module>/service/service_shared_ext.go`
     - `modules/<module>/data/repo/repo_shared_ext.go`

5. frontend-meta behavior
   - module mode only generates resource-local `*.meta.ts`
   - module mode does not generate:
     - shared frontend `config.ts`
     - shared frontend `page_i18n.*`
     - shared frontend `langs/*`

6. verified pilot commands
   - `powershell -ExecutionPolicy Bypass -File xkit/examples/generateModule.ps1 -SkipDryRun`
   - `go test ./internal/codegen`
   - `go test ./modules/xdev/...`

## 8. Recommended Pilot Implementation

Use `xdev` / `device` as the first pilot.

### 8.1 Pilot target shape

```text
admin/
  modules/
    xdev/
      api/
        protos/
      data/
        schema/
        repo/
      service/
      server/
      bootstrap/
      module.go
```

### 8.2 Pilot host changes

Add:

- `admin/shared/modulex`
- a host module mounting hook
- a first module list file in host bootstrap

Do not yet migrate all old resources.

### 8.3 Pilot generator scope

The original pilot scope was:

- one module
- one or more tightly-related resources
- one host
- dry-run import
- dry-run codegen

Current pilot result:

- `xdev` now covers three related resources:
  - `Device`
  - `DeviceModel`
  - `DeviceModelType`
- module generation is no longer dry-run only
- the backend and frontend-meta generation chain has been exercised against the
  real `admin` and `admin-ui` targets

## 9. Implementation Task List

### Phase A: completed pilot baseline

1. `examples/xdev` source set built with:
   - split proto files
   - restored comments / `json_name`
   - three related resources
   - tenant mixin usage
2. `xkit init module` creates and fills `admin/modules/xdev`.
3. `xkit gen module` generates repo / service / register / bootstrap /
   module-entry code for module mode.
4. module-mode frontend-meta writes only under:
   - `admin-ui/apps/web-antd/src/views/generated/xdev`
5. host-visible shared helper extraction landed at:
   - `admin/shared/modulex/module_shared_ext.go`

### Phase B: remaining shared-boundary work

1. Audit additional module-visible candidates in:
   - `internal/data/repo`
   - `internal/service`
   - `internal/server`
2. Decide which existing host helpers should migrate into `modulex` now, and
   which should wait for package split beyond `modulex`.
3. Introduce `xkit-template/shared` so future host projects can receive the
   same shared helper baseline.
4. Decide whether compatibility wrappers are needed for host-internal callers.

### Phase C: module scaffold hardening

1. Freeze the `admin/modules/<module>` directory contract.
2. Keep `module.go` as the stable entry contract with:
   - `Name()`
   - `RegisterData(*app.AppCtx)`
   - `RegisterServices(*app.AppCtx, *bootstrap.GeneratedData)`
   - `RegisterHTTP(...)`
   - `RegisterGRPC(...)`
3. Decide whether additional per-module manual extension files are required in
   the scaffold beyond current generated outputs.

### Phase D: host mounting points

1. Add host-side module list / mounting hook in `internal/bootstrap`.
2. Keep `http.go` and `grpc.go` generic.
3. Add transition-friendly host hooks for module registration.

### Phase E: xkit command design refinement

1. Preserve the current `moduleName` + `moduleRoot` parameter structure.
2. Decide which `gen module` sub-targets should remain aligned with `gen all`
   and which should stay module-specific.
3. Define how `xkit-template/shared` participates in `init module`.
4. Decide whether host mounting update is generated, semi-generated, or manual.

### Phase F: xkit implementation follow-up

1. Keep project-mode generation behavior stable.
2. Continue isolating module-mode code paths where project-mode assumptions
   differ.
3. Add host mounting integration generation or update hooks.
4. Add tests that pin module shared-helper generation and skip-if-exists
   behavior.

### Phase G: host integration execution

1. Mount `xdev.Module` into the host bootstrap/server chain.
2. Verify host HTTP registration.
3. Verify host gRPC registration.
4. Verify runtime startup with the mounted module enabled.
5. Only after that, evaluate generating into additional business modules.

## 10. Decision

The requested exploration is complete enough to proceed.

Main decisions confirmed:

- current `xkit` cannot reach `admin/modules/xdev` by config-only tuning
- a module scaffold distinct from the host template is required
- the first shared extraction is now concretely implemented as
  `admin/shared/modulex`
- `xkit init module` and `xkit gen module` are justified by real structural gaps,
  not by stylistic preference

The next best step is to turn this plan into a concrete implementation backlog
and start with:

1. host mounting of `xdev.Module`
2. `xkit-template/shared` introduction
3. follow-up shared helper extraction beyond the first `modulex` baseline
