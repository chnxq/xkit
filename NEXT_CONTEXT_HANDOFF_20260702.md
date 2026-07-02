# XAdmin Next Context Handoff

Generated at: 2026-07-02 +08:00

This file is the resume handoff for the current XAdmin / xkit / admin / admin-ui / xdev-ui workspace after the recent module-generation, frontend integration, tenant-boundary, and handwritten-page rounds.

It should be read together with, but takes precedence over, the older handoff files under `xkit/NEXT_CONTEXT_HANDOFF_*`.

## 1. Current Scope

Recent work was concentrated in four connected areas:

1. `xkit` module-mode generation and config contract stabilization
2. `admin` host-side module mounting, seed/resource sync, and tenant-boundary hardening
3. `admin-ui` / `xdev-ui` frontend generation, handwritten `device-center`, and route/resource integration
4. `xdev` example module as the real verification target for the full chain

The practical baseline is now:

- `xdev` is the reference module for validating module-mode generation
- `xdev-ui` is the reference for module-contained frontend code
- `device-center` is a handwritten page and must not be treated as generated output
- `xkit/examples/xdev/xdev-target-config/xdev.yaml` is the effective config used for accepted generation

## 2. Stable Architecture Decisions

### 2.1 Module generation contract

- `generateModule.ps1` is the real module verification entry, not just an example helper.
- Its behavior must follow the same canonical-config principle as `generateAll.ps1`.
- `*-target-config` is the effective working config after first materialization.
- `*-config` is the canonical seed template.
- If `*-target-config` does not exist, copy from `*-config` once and replace project/module-specific placeholders.
- If `*-target-config` already exists, do not overwrite it implicitly.
- This rule matters because module-specific accepted edits such as `host_module` and handwritten service hooks live in `target-config`.

### 2.2 Frontend generation contract

- Frontend generated output root comes from `TypeScriptRoot`, not from an extra top-level frontend output config knob.
- Module generated frontend code now lives under module-owned paths, not mixed into legacy `src/views/generated/...` only.
- Generated frontend artifacts are disposable and should be safe to delete and regenerate.
- Handwritten frontend code must live in module-owned non-generated paths and must not be overwritten by xkit.

Current practical separation:

- generated frontend API/provider/meta/i18n artifacts: module-owned generated directories
- handwritten frontend page code: `admin-ui/apps/web-antd/src/modules/xdev-ui/views/...`

### 2.3 Host-module integration contract

- Host maintains the module registration table.
- Module mounts itself through its own bootstrap/module entry.
- Backend manual host integration for a new module is intentionally minimal:
  - import the module once in host module registration, for example `_ "admin/modules/xdev"`
- Frontend module code should also stay module-owned as much as possible, avoiding scattered host-side handwritten glue.

### 2.4 Generated vs handwritten boundary

This boundary is critical and must stay explicit:

- `*.gen.go`, generated TS API, generated meta, generated i18n sync outputs: disposable, overwriteable
- `*_ext.go`, handwritten Vue pages, host-specific handwritten bootstrap/resource logic: durable, must not be overwritten
- If a behavior is accepted as project-specific business behavior, do not hide it inside a generated-once `_ext.go` unless it is intentionally one-time bootstrap scaffolding
- If a file is intended for regeneration safety, the generator must detect existing durable files and skip overwriting them

### 2.5 Shared extension placement

- Shared cross-module extension helpers belong in host shared/module-level locations rather than resource-local duplicated files.
- The previous direction around `module_shared_ext.go` was correct:
  - shared basic capability only
  - no business-specific policy hidden there
- `module_shared_ext.go` should provide foundation helpers, not resource-specific mutation logic

## 3. What Was Proven In Real Usage

### 3.1 End-to-end module chain

The following chain has been exercised on real `xdev`:

1. schema/proto/config/langs update
2. `generateModule.ps1`
3. backend generation into `admin/modules/xdev`
4. frontend generation into module-owned frontend paths
5. host startup and module registration
6. menu/resource/API visibility in admin
7. CRUD pages and handwritten `device-center` validation

### 3.2 `xdev` is not only generated CRUD now

`xdev` contains both:

- generated CRUD resources
- handwritten integrated business page: `device-center`

This distinction matters for future work:

- generator issues should be fixed in generator/config/template level
- `device-center` issues should usually be fixed in handwritten page/provider logic, not forced back into generic generator behavior

### 3.3 Module i18n source of truth

- For module frontend language resources, the durable source of truth is under `xkit/examples/xdev/langs/...`
- Target module/frontend language files can be regenerated or resynced from there
- When target language files are corrupted, prefer restoring from the xkit example source rather than patching a damaged file in place

This was especially important for `page.json` and UTF-8 corruption recovery.

## 4. Tenant Boundary Strategy That Is Now Preferred

### 4.1 General rule

The working direction is:

- cross-tenant read may be allowed in selected platform-admin views
- cross-tenant mutate is not allowed unless explicitly classified otherwise
- repo/service enforcement is the hard safety boundary
- frontend only improves UX and reduces accidental invalid operations

### 4.2 Responsibility layering

- frontend:
  - hide or downgrade edit actions to detail/read-only when mutation is not allowed
  - when editing a tenant-owned object, relation candidates should be scoped to that object's tenant where appropriate
- service/repo:
  - reject cross-tenant mutation even if frontend leaked an invalid candidate
- shared helper layer:
  - provide tenant capability primitives only
  - do not bury business policies there

### 4.3 Current accepted practical rule

For many ordinary tenant-scoped resources:

- platform users may have broader `Get` / `List`
- mutation still follows target data tenant rules
- editing relation choices should usually be filtered to the tenant of the object being edited, not the viewer's broad visibility

This rule was already validated in:

- user management
- org-unit / position / role rounds
- xdev device/device-model related selection behavior

## 5. Frontend Runtime Lessons That Must Be Preserved

### 5.1 Meta generation was mostly correct; problems were usually in final Vue code

One important lesson from the xdev frontend rounds:

- generated `*.meta.ts` was usually structurally correct
- many runtime UX/data issues were introduced later in handwritten/generated Vue pages or providers

So for similar future issues:

1. first verify meta output
2. then verify provider mapping
3. only then modify final page rendering/interaction

Do not randomly patch generic shared form runtime first unless evidence points there.

### 5.2 Search/list/form consistency

The current preferred direction is:

- search fields, list columns, and form fields may differ intentionally
- but relation/enum semantics should stay consistent across them
- runtime capability determination should be centralized as much as possible
- tree/relation/enum behavior should be carried by config/meta/provider contracts, not page-local guesses where possible

### 5.3 Tree behavior

Tree resources such as `deviceGroup` need explicit treatment.

What is already known:

- plain list behavior is not enough for the final runtime experience
- tree display may require service-side `ListTree` or equivalent shaped data
- frontend tree pages should not assume generic flat CRUD UX is sufficient

### 5.4 Sorting

There was a long round of confusion around sorting.

Current stable conclusion:

- backend sorting path works
- many old pages had only superficial or partial frontend sorting behavior
- relation fields often should not expose sortable controls unless backend mapping is real

Rule:

- if a field cannot be mapped safely to backend sorting, disable or omit sorting for that field
- do not pretend relation sort works when it resolves to invalid DB columns

## 6. `device-center` Specific Context

`device-center` is a handwritten integration page under `xdev-ui`.

Its intended shape:

- left: device-group tree
- right: selected group scoped device list
- top summary area:
  - group summary
  - compact org-unit relation display
  - compact user relation display

Important constraints already established:

- relation summary rows must stay inside the top summary banner, not in a separate extra card
- the original summary content must remain visible
- relation edit buttons show only for leaf groups
- platform-admin cross-tenant display should resolve names using the selected group's tenant, not fallback to IDs because of dialog-option reuse
- handwritten page behavior should not be over-generalized into xkit unless a reusable pattern is clearly proven

Deletion semantics clarified:

- deleting a device group should clear device-group/device, device-group/org-unit, and device-group/user relations before deleting the group
- deleting a single device must not accidentally wipe unrelated group-level org-unit/user bindings

## 7. Config And Generation Rules That Must Not Be Regressed

### 7.1 Effective config rule

- `xdev-target-config/xdev.yaml` is the real accepted working config after generation/bootstrap
- do not silently drop accepted fields such as `host_module`
- do not regenerate target config from canonical config once the target config already exists

### 7.2 `host_module` matters

`host_module` is not optional decoration. It carries host integration intent such as:

- menus
- route/resource related host metadata
- module-facing host wiring

Losing it during config sync broke downstream generated module resource behavior before. This is a known regression pattern.

### 7.3 Generated-once durable business behavior should move out of accidental overwrite zones

If a capability is expected to survive regeneration and evolve manually:

- do not place it in a file/path that xkit later fully regenerates
- prefer explicit handwritten durable files or generated-if-missing strategies

## 8. Encoding And File Safety Rules

This section is intentionally explicit. UTF-8 damage was a repeated real problem.

### 8.1 Required encoding rules

- All newly written or rewritten text/code/config/json/yaml/proto files must be UTF-8.
- Prefer UTF-8 without accidental console-encoding conversions.
- Do not trust PowerShell `Get-Content` console display as proof that file content is correct.
- When Chinese text matters, verify file content by reading it as UTF-8 with a real parser/runtime if there is any doubt.

### 8.2 Safe editing rules for Chinese content

- Avoid rewriting a whole Chinese JSON/YAML file through shell pipelines unless encoding behavior is fully controlled.
- Prefer `apply_patch` for small edits.
- If a target Chinese i18n file is already corrupted, restore from the canonical source file instead of patching corrupted fragments.
- When console output looks garbled, determine whether:
  - only terminal rendering is wrong
  - or file bytes are actually damaged

### 8.3 Practical recovery rule

For module i18n:

- canonical source: `xkit/examples/xdev/langs/...`
- target copies: `admin-ui/.../modules/xdev-ui/langs/...`

If target i18n is damaged:

1. restore from canonical source
2. reapply minimal accepted target-specific differences only if they are truly needed

## 9. Explicit Coding Conventions For The Next Phase

This is the most important section for future work.

### 9.1 General code-change discipline

- Change the smallest correct layer first.
- Do not guess across layers.
- Verify whether the bug belongs to:
  - config
  - generator/template
  - generated provider/meta
  - handwritten runtime page/service/repo
- If the real problem is in a handwritten page, do not “fix” shared generator/runtime code first.

### 9.2 Generator discipline

- Never break existing host/admin generation while fixing module generation.
- Treat `admin` baseline behavior as regression-sensitive.
- Generated code should stay ugly-but-safe rather than clever-but-fragile.
- If a durable handwritten extension point is needed, add it intentionally; do not smuggle business behavior into generated files.

### 9.3 Host/module separation discipline

- Keep `xdev` dependence on `admin/shared` as the upper acceptable bound unless a broader dependency is explicitly discussed.
- Do not let module business code reach deep into host internals casually.
- Prefer host-registered module loading over host-scattered per-module handwritten glue.

### 9.4 Frontend discipline

- Generated files must be disposable.
- Handwritten module pages belong in module-owned paths and should stay isolated from unrelated host pages.
- Reuse existing admin i18n keys when a stable semantic key already exists.
- Do not create new i18n keys casually for words that already exist in the host vocabulary.
- For relation display, distinguish:
  - display-source options
  - dialog candidate options
  - mutation payload shape

Do not reuse one list for all three unless it is actually valid.

### 9.5 Tenant-safety discipline

- Frontend restrictions are advisory UX; backend restrictions are authoritative.
- For tenant-owned object edits, relation candidate dropdowns should prefer the edited object's tenant scope.
- A platform viewer's wide visibility must not silently widen mutation candidates.
- Cross-tenant read-only detail is acceptable only where explicitly intended.

### 9.6 Tree and relation discipline

- Tree resources are not ordinary flat CRUD.
- Relation resources are not ordinary scalar fields.
- Do not assume generic CRUD templates fully solve tree/relation UX.
- Add explicit service/provider/runtime treatment where needed.

### 9.7 Sorting/filtering discipline

- Only expose sorting/filtering that the backend/provider mapping can actually honor.
- If relation sorting is not safely supported, disable it rather than leaving a fake interactive affordance.
- Prefer honest capability over misleading UI.

### 9.8 File header and overwrite discipline

- Generated files should carry clear generated headers.
- Durable handwritten files should carry headers/comments that make overwrite expectations obvious.
- If a file must only be generated once, generator logic should enforce that instead of relying on memory.

## 10. Recommended Next-Step Checklist

When the next phase starts, use this order:

1. identify whether the change is generator-wide, xdev-specific generated, or handwritten `device-center` business logic
2. confirm the effective config in `xdev-target-config/xdev.yaml`
3. confirm whether the target file is disposable or durable
4. if Chinese/i18n is involved, verify UTF-8 before and after edit
5. if tenant behavior is involved, define:
   - read scope
   - candidate selection scope
   - mutation allowed scope
6. regenerate only when the change truly belongs to generated artifacts
7. verify on real xdev flow rather than only by code inspection

## 11. High-Risk Regression Patterns

Avoid repeating these:

- overwriting `target-config` from canonical config after the target has accepted local edits
- losing `host_module` during sync
- patching shared frontend runtime when the issue is actually in final page code
- mixing display options and edit candidate options across tenants
- editing Chinese JSON through unsafe shell encoding paths
- storing durable business behavior in files later regenerated by xkit
- broadening module->host dependency without explicit review
- enabling sorting on relation fields that do not map to real backend sort expressions

## 12. Files And Areas To Recheck First In Similar Future Work

### `xkit`

- `examples/generateModule.ps1`
- `examples/generateModule.README.md`
- `examples/xdev/xdev-config/xdev.yaml`
- `examples/xdev/xdev-target-config/xdev.yaml`
- `examples/xdev/langs/**/*`
- `internal/codegen/...`
- `internal/config/config.go`

### `admin`

- `modules/registeredHostModules.go`
- `modules/xdev/**`
- `shared/modulex/module_shared_ext.go`
- `internal/data/repo/tenant_scope_ext.go`

### `admin-ui`

- `apps/web-antd/src/modules/xdev-ui/**`
- especially handwritten:
  - `views/device-center/index.vue`

## 13. Final Summary

The most important outcome of the recent rounds is not just that `xdev` works.

It is that the project now has a clearer operating model:

- canonical config vs effective target config
- generated vs handwritten boundaries
- host-managed module registration with module self-mounting
- `TypeScriptRoot`-driven frontend generation
- repo/service as the hard tenant boundary
- UTF-8 safety as a mandatory editing discipline

If the next phase preserves these rules, progress should be much faster and with fewer regressions.
