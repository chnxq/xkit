# Frontend CRUD Page Generation Plan

Generated at: 2026-06-20 +08:00

## 1. Purpose

This document defines a practical next step for `xkit` frontend generation:

- keep the existing generated frontend meta flow
- add generated frontend API provider wrappers
- add generated standard CRUD Vue pages

The goal is not to generate every complex business page automatically.

The goal is to generate a working baseline CRUD page for simple resources, so
manual frontend work only needs to focus on resource-specific enhancements.

## 2. Current Baseline

The current frontend generation already provides two important building blocks.

### 2.1 Generated page meta

For each resource, `xkit gen frontend-meta` already generates:

- `buildSearchFormOptions(...)`
- `buildListGridColumns(...)`
- `buildFormOptions(...)`
- `i18nPrefix`
- `defaultSortField`
- `defaultSortDirection`

Examples:

- `admin-ui/apps/web-antd/src/views/generated/admin/system/position.meta.ts`
- `admin-ui/apps/web-antd/src/views/generated/xdev/asset/device.meta.ts`

This is already enough to describe:

- search form layout
- table columns
- dialog form fields
- page i18n key prefix
- default sorting

### 2.2 Generated TypeScript API

Buf + `protoc-gen-typescript-http` already generate stable client-like TS APIs.

Examples:

- `admin-ui/apps/web-antd/src/api/generated/admin/service/v1/index.ts`
- `admin-ui/apps/web-antd/src/api/generated/xdev/service/v1/index.ts`

For standard CRUD resources, the generated TS API already has stable methods:

- `List`
- `Get`
- `Create`
- `Update`
- `Delete`

That means `xkit` does not need to invent API knowledge from scratch.

## 3. Reference Simple Page Pattern

The current best reference is a simple manually assembled page such as:

- `admin-ui/apps/web-antd/src/api/admin/positions.ts`
- `admin-ui/apps/web-antd/src/views/system/position/index.vue`

This pattern is important because it shows the actual layering already used by
the host frontend:

1. generated TS API
2. handwritten provider wrapper in `src/api/admin/*.ts`
3. generated page meta in `src/views/generated/**/*.meta.ts`
4. handwritten Vue page that combines provider + meta

The `position` page is simple enough to be a template target because:

- list query is standard
- CRUD operations are standard
- page mainly adds enum options, placeholders, and table slots

This should be the first-class reference model for automatic standard CRUD page
generation.

## 4. Main Decision

Do not try to generate a full business page directly from `meta.ts` alone.

Instead, split the problem into three generated layers:

1. `meta.ts`
   - describes the page structure
2. `provider.ts`
   - adapts generated TS API to frontend CRUD usage
3. `crud.vue`
   - builds a standard CRUD page from `meta.ts` + `provider.ts`

This keeps the interfaces stable and limits the template complexity.

## 5. Proposed Generated Outputs

For each resource, generate these files under the existing generated frontend
tree:

```text
admin-ui/apps/web-antd/src/views/generated/<module>/<view-path>/
  <resource>.meta.ts
  <resource>.provider.ts
  <resource>.crud.vue
```

For example:

```text
admin-ui/apps/web-antd/src/views/generated/xdev/asset/
  device.meta.ts
  device.provider.ts
  device.crud.vue
  device-model.meta.ts
  device-model.provider.ts
  device-model.crud.vue
```

Optionally, later add thin host-facing wrapper pages:

```text
admin-ui/apps/web-antd/src/views/<module>/<view-path>/<resource>/index.vue
```

That wrapper would only import the generated CRUD page, so later manual takeover
does not require editing generated files.

## 6. Generated Provider Layer

### 6.1 Why a provider layer is needed

The generated TS API is usable, but it is still proto-oriented:

- paging input is proto-shaped
- response shapes follow proto conventions
- create/update payloads need normalization
- update often needs `updateMask`
- list results need frontend-friendly total extraction

The current handwritten files under `src/api/admin/*.ts` already solve exactly
this problem.

So `provider.ts` should be generated as a standard CRUD adapter layer.

### 6.2 Suggested provider exports

For each resource:

- `type <Resource>ListParams`
- `type <Resource>ListResult`
- `type <Resource>FormModel`
- `type <Resource>SaveInput`
- `function createEmpty<Resource>FormModel()`
- `function to<Resource>FormModel(data)`
- `function to<Resource>CreateRequest(input)`
- `function to<Resource>UpdateRequest(id, input)`
- `async function list<Resource>Page(params)`
- `async function get<Resource>ById(id)`
- `async function create<Resource>(input)`
- `async function update<Resource>(id, input)`
- `async function delete<Resource>(id)`

Example target:

```ts
export async function listDevicePage(params: DeviceListParams)
export async function getDeviceById(id: number)
export async function createDevice(input: DeviceSaveInput)
export async function updateDevice(id: number, input: DeviceSaveInput)
export async function deleteDevice(id: number)
```

### 6.3 Provider input sources

The provider generator should combine:

1. generated TS API:
   - request/response types
   - client object / service methods
2. xkit config resource definition:
   - resource name
   - DTO type
   - frontend view path
   - operations enabled
3. frontend meta data:
   - default sort field
   - default sort direction
   - form fields

### 6.4 Provider responsibilities

The provider should be responsible for:

- converting page pagination to generated API paging request
- mapping frontend filter form values to query conditions or request fields
- converting `GetResponse.data` to form model
- converting form model to create/update payload
- generating update masks for standard updates
- returning a frontend-friendly list result

### 6.5 Provider should not do

It should not include:

- resource-specific remote option loading
- tree loading
- upload logic
- custom business actions
- tenant/session-specific UI branching

Those remain manual enhancements.

## 7. Generated CRUD Page Layer

### 7.1 Goal

Generate a standard Vue page that is directly runnable for simple CRUD
resources.

It should be intentionally thin and generic.

### 7.2 Suggested page structure

Each generated CRUD page should include:

1. search form
   - uses `buildSearchFormOptions($t)`
2. data grid
   - uses `buildListGridColumns($t)`
3. action column
   - standard edit/delete buttons
4. create/edit modal
   - uses `buildFormOptions($t)`
5. standard page-level state
   - `loading`
   - `modalOpen`
   - `editingId`
   - `submitting`
   - `formModel`

### 7.3 Required imports

The page template should standardize imports like:

- `Page`
- `AdminGeneratedForm`
- `useVbenVxeGrid`
- `message`, `Modal`, `Popconfirm`, `Button`, `Space`
- `$t`
- generated meta
- generated provider

### 7.4 Standard interactions

The generated page should implement:

- load list
- open create modal
- open edit modal
- submit create
- submit update
- confirm delete
- refresh after mutation
- reset form model

### 7.5 Deliberate limitations

The generated page should not try to handle:

- custom slots beyond a minimal default set
- complex enum display rules
- related-resource display lookups
- tree or nested forms
- uploads
- tabs / subresources
- bulk actions

Those should remain manual overlays.

## 8. I18n Plan

### 8.1 Existing baseline

Current frontend meta generation already outputs:

- `page_i18n.zh-CN.json`
- `page_i18n.en-US.json`

for the generated frontend root.

Project mode currently also generates shared `langs/*` and `config.ts`.

Module mode currently only generates resource-local `*.meta.ts`.

### 8.2 Recommended direction for CRUD page generation

For standard CRUD pages, keep using the existing `meta.ts` i18n pattern:

- `meta.ts` owns field labels and search labels
- page template uses `$t(...)`
- provider should avoid direct user-facing strings whenever possible

So the generated CRUD page should consume:

- `i18nPrefix` from `meta.ts`
- shared UI i18n keys already in host frontend
- generated page i18n entries when resource-specific text is needed

### 8.3 Additional i18n keys needed for generated CRUD page

The current meta i18n is not enough for a full page template.

The page template also needs generated keys like:

- `<prefix>.createTitle`
- `<prefix>.editTitle`
- `<prefix>.deleteConfirm`
- `<prefix>.createSuccess`
- `<prefix>.updateSuccess`
- `<prefix>.deleteSuccess`
- `<prefix>.emptyText`

These should be emitted together with existing `page_i18n.*.json`.

### 8.4 Host shared i18n reuse

Prefer reusing existing shared keys when available:

- `ui.table.action`
- `ui.action.create`
- `ui.action.edit`
- `ui.action.delete`
- `ui.action.confirm`
- `ui.action.cancel`
- `ui.formRules.required`
- `ui.formRules.selectRequired`

Only resource-specific texts should be newly generated.

### 8.5 Module-mode i18n output decision

If CRUD page generation is added for modules, module mode should no longer stop
at `*.meta.ts` only.

It should also generate module-local:

- `page_i18n.zh-CN.json`
- `page_i18n.en-US.json`

under:

```text
src/views/generated/<module>/
```

This is required if generated CRUD pages are expected to be directly runnable.

## 9. Provider + Meta Joint Generation

The user explicitly raised whether API transfer providers can be generated
together with frontend meta.

The answer is yes, and that is the recommended design.

But they should be generated as separate files, not mixed into `meta.ts`.

Recommended command-level behavior:

- `xkit gen frontend-meta`
  - remains responsible for `*.meta.ts` and page i18n
- `xkit gen frontend-provider`
  - generates CRUD provider wrappers
- `xkit gen frontend-page`
  - generates standard CRUD Vue pages
- `xkit gen module` / `xkit gen all`
  - may later include all three as composed stages

This separation avoids turning `meta.ts` into a mixed UI + API abstraction
file.

## 10. Template Design

### 10.1 New frontend templates

Add new templates under `xkit/internal/codegen/template/`:

```text
frontend_provider.tmpl
frontend_crud_page.tmpl
frontend_page_i18n_crud_zh.tmpl
frontend_page_i18n_crud_en.tmpl
```

### 10.2 Template data needed

The generator will need frontend template data such as:

- module name
- resource name
- frontend view path
- meta import path
- provider import path
- generated TS API import path
- service client name
- request/response type names
- enabled CRUD operations
- default sort field/direction
- search fields
- form fields

### 10.3 Naming stability requirement

Before provider generation is implemented, confirm that generated TS API naming
is stable enough for codegen.

For `xdev`, current TS API already exposes stable method groups such as:

- `DeviceService.List`
- `DeviceService.Get`
- `DeviceService.Create`
- `DeviceService.Update`
- `DeviceService.Delete`

This is good enough for a first implementation.

## 11. Applicability Boundary

Automatic CRUD page generation should be explicitly scoped to simple resources.

### Good candidates

- `position`
- `tenant`
- `role`
- `xdev/device`
- `xdev/device-model`
- `xdev/device-model-type`

### Poor candidates

- tree resources
- resources with rich relation selectors
- resources with custom action workflows
- resources with upload/file preview
- resources with many manual table slots
- aggregate pages like message center or task scheduler

For poor candidates, generated provider and meta may still be useful, but full
page generation should remain opt-in or skipped.

## 12. Suggested Rollout Plan

### Phase A: design freeze

1. Freeze generated provider interface contract.
2. Freeze generated CRUD page minimal UI contract.
3. Define required additional i18n keys.

### Phase B: provider pilot

1. Implement `xkit gen frontend-provider`.
2. Use `position` and `xdev/device` as pilot references.
3. Verify generated provider compiles against current `admin-ui`.

### Phase C: CRUD page pilot

1. Implement `xkit gen frontend-page`.
2. Generate a thin standard page for:
   - `xdev/device`
   - `xdev/device-model`
   - `xdev/device-model-type`
3. Verify page can load, create, update, delete.

### Phase D: i18n completion

1. Extend page i18n generation for CRUD page strings.
2. In module mode, generate module-local `page_i18n.*.json`.

### Phase E: host wrapper policy

1. Decide whether wrapper pages under `src/views/<module>/...` are generated.
2. Decide whether route registration stays manual or gets a generated helper.

## 13. Immediate Recommendation

The next concrete implementation step should be:

1. write `frontend-provider` generation first
2. keep `crud.vue` template very thin
3. pilot only on `xdev`
4. do not try to solve complex enum/remote option/tree cases in version 1

This order is important because:

- provider generation is the real normalization layer
- once provider shape is stable, page template becomes simple
- if page template is attempted first, it will become full of API-shape
  exceptions

## 14. Decision

This work is feasible.

Not as “generate every admin page automatically”, but as:

- generate frontend meta
- generate CRUD provider wrappers
- generate a standard baseline CRUD Vue page

That is both technically realistic and aligned with the current `admin-ui`
architecture already visible in pages like `position`.
