# Admin Host Module Implementation Plan

Generated at: 2026-06-18 +08:00

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

Recommended new root:

```text
admin/shared/
  authctx/
  tenant/
  repohelpers/
  servicehelpers/
  moduleapi/
```

If needed, keep a small compatibility bridge in host internals for old code
during migration.

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

Before module generation becomes real, a first shared extraction pass is needed.

Without it:

- generated module code will still rely on host internals
- future module isolation will be fake

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

- `admin/shared/...`
- a host module mounting hook
- a first module list file in host bootstrap

Do not yet migrate all old resources.

### 8.3 Pilot generator scope

For the pilot, focus only on:

- one module
- one resource
- one host
- dry-run import
- dry-run codegen

That is enough to expose the real generator delta.

## 9. Implementation Task List

### Phase A: shared boundary exploration and extraction

1. Create `admin/shared/`.
2. Audit module-visible candidates in:
   - `internal/data/repo`
   - `internal/service`
   - `internal/server`
3. Move only the first minimal helper set required by the `xdev` pilot.
4. Leave compatibility wrappers in old locations if necessary.

### Phase B: module scaffold definition

1. Define final `admin/modules/<module>` directory contract.
2. Add a first manual `modules/xdev/` skeleton in `admin`.
3. Decide module-local proto and schema roots.
4. Define `module.go` contract.

### Phase C: host mounting points

1. Add host-side module list / mounting hook in `internal/bootstrap`.
2. Keep `http.go` and `grpc.go` generic.
3. Add transition-friendly host hooks for module registration.

### Phase D: xkit command design

1. Design `xkit init module` CLI arguments.
2. Design `xkit gen module` CLI arguments.
3. Define module-aware config extensions.
4. Define module template files.

### Phase E: xkit implementation

1. Refactor `sourceimport` to support configurable destination roots.
2. Refactor `codegen.Runner` to support configurable source and output roots.
3. Refactor import generation to support shared import roots.
4. Add host mounting integration generation or update hooks.

### Phase F: pilot execution

1. Initialize `admin/modules/xdev`.
2. Import `examples/xdev`.
3. Generate module-local code.
4. Mount into host.
5. Verify build and runtime registration.

## 10. Decision

The requested exploration is complete enough to proceed.

Main decisions confirmed:

- current `xkit` cannot reach `admin/modules/xdev` by config-only tuning
- a module scaffold distinct from the host template is required
- some host internals must move to `shared` before module generation is real
- `xkit init module` and `xkit gen module` are justified by real structural gaps,
  not by stylistic preference

The next best step is to turn this plan into a concrete implementation backlog
and start with:

1. `admin/shared` first extraction candidates
2. `admin/modules/xdev` skeleton
3. module-aware root configuration design in `xkit`
