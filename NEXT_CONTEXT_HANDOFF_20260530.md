# XAdmin Context Handoff

Generated at: 2026-05-30 00:00 +08:00

This file is the latest resume entry for the current workspace state.  
It should be treated as the primary handoff after:

- `xkit/NEXT_CONTEXT_HANDOFF_20260501.md`
- `xkit/NEXT_CONTEXT_HANDOFF_20260505.md`
- `xkit/NEXT_CONTEXT_HANDOFF_20260510.md`
- `xkit/NEXT_CONTEXT_HANDOFF_20260522.md`

This handoff explicitly incorporates the current multi-tenant stabilization
round across `admin`, `admin-ui`, and `xkit`.

## Scope And Incorporation

- This handoff explicitly includes facts from the current thread.
- It covers:
  - single-database multi-tenant refactor progress
  - platform tenant semantics (`tenantId=0`)
  - backend tenant-aware data resolution and UI-facing `tenantName` behavior
  - frontend tenant/resource ownership presentation unification
  - encoding regression guard added to `admin-ui`
  - current repository state and latest validated commits

## Environment

- Workspace root: `D:\GoProjects\XAdmin`
- Shell: Windows PowerShell
- Timezone: `Asia/Shanghai`
- Main backend: `admin`
- Main frontend: `admin-ui`
- Generator/tooling repo: `xkit`

## Repository State As Of 2026-05-30

### `admin`

- Branch: `main`
- HEAD: `d6a9a37 fix(admin): resolve tenant names for tenant-scoped resources`
- Working tree: clean

Recent tenant-related backend landmarks:

- `4d53d4c` `feat: enforce hybrid tenant access for roles and permissions`
- `82012a8` `refactor(admin): stabilize tenant-aware backend behavior`
- `6d5a157` `refactor(admin): unify platform tenant display name`
- `d6a9a37` `fix(admin): resolve tenant names for tenant-scoped resources`

### `admin-ui`

- Branch: `xadmin-api-integration`
- HEAD: `925218b1 fix(ui): restore permission layout and align table toolbar helpers`
- Working tree: clean

Recent frontend landmarks:

- `1dc9a09d` `refactor(ui): refine tenant-aware admin ui`
- `df656835` `refactor(ui): unify tenant name presentation`
- `65dd3875` `fix(ui): align internal message tenant label`
- `b9f4d368` `fix(ui): normalize tenant labels and guard encoding regressions`
- `1ce30916` `refactor(ui): refresh generated admin client formatting`
- `925218b1` `fix(ui): restore permission layout and align table toolbar helpers`

### `xkit`

- Branch: `main`
- HEAD: `d6cb0dc feat: add tenant scoped repo generation`
- Working tree: clean

Recent generator landmarks:

- `da0e3f2` `feat: add default generated list sorting`
- `6b92e34` generated list sorting policy consolidation
- `d6cb0dc` `feat: add tenant scoped repo generation`

## What Changed In This Multi-Tenant Round

## 1. Tenant Semantics Were Made Explicit

The current working rule is:

- `tenantId = 0` means platform semantics
- non-zero `tenantId` means ordinary tenant data

Important note:

- The project does **not** force an actual `sys_tenants.id = 0` record in Ent
  schema migration because the Ent-side auto-increment/validation path had
  already shown range/compatibility issues.
- Platform display is therefore a semantic rule first, not a DB-row guarantee.

This distinction matters for future work:

- backend can still treat `tenantId=0` as platform scope
- frontend should prefer backend-resolved `tenantName`
- future split-database design should preserve this semantic boundary

## 2. Resource Classification Became The Core Tenant Design Lens

The current tenant refactor no longer treats all resources the same.
The effective design direction is:

- global resources
- tenant resources
- hybrid resources

This classification is recorded in:

- `admin/docs/tenant-resource-classification-v1.md`

It is now the correct basis for:

- repo filtering rules
- write validation rules
- frontend ownership labels
- platform-state vs tenant-state menu/button visibility

## 3. Backend Tenant-Aware Behavior Was Stabilized

The backend has already moved beyond “just having `tenant_id` fields”.

Key outcomes already landed:

- tenant-aware backend behavior stabilized for core resources
- platform tenant display name handling unified
- tenant-scoped resources now resolve actual `tenantName` instead of only
  special-casing platform

The last backend fix in this area (`d6a9a37`) addressed a concrete bug:

- platform rows (`tenantId=0`) displayed correctly
- ordinary tenant rows showed `-`

Root cause:

- code only special-cased platform rows
- it did not batch-load `sys_tenants.name` for `tenantId > 0`

Fix direction:

- common tenant-scope helper now collects tenant IDs
- loads tenant name map in batch
- resolves both platform and ordinary tenant names consistently

Important backend files involved in this round:

- `admin/internal/data/repo/tenant_scope_ext.go`
- `admin/internal/data/repo/user_repo_ext.go`
- `admin/internal/data/repo/role_repo_ext.go`
- `admin/internal/data/repo/org_unit_repo_ext.go`
- `admin/internal/data/repo/position_repo_ext.go`

## 4. Frontend Ownership Expression Was Unified Further

The frontend moved from inconsistent “租户” wording toward explicit ownership
semantics.

Current intended wording:

- user list: `租户`
- internal message list: `消息归属`
- resource-oriented pages: `资源归属`

The relevant pages already adjusted in this round include:

- `system/user`
- `system/org-unit`
- `system/position`
- `system/role`
- `system/menu`
- `system/dict`
- `app/permission/permission`
- `app/internal-message/message`

In the same round:

- several toolbar containers were corrected from top-aligned visual drift to
  centered button alignment
- one mistaken layout regression in the permission page was later corrected:
  `.admin-permission-layout` must remain top-aligned, not centered

## 5. Frontend Reduced Hardcoded Platform Fallbacks

The preferred direction is now:

- use backend-returned `tenantName` when available
- avoid permanent frontend hardcoded `XAdmin平台` fallback logic

Practical consequence:

- tenant/resource ownership display should come from DTO-level `tenantName`
- pure global resources can still use explicit platform semantics where the DTO
  does not expose tenant ownership

## 6. Encoding Pollution Was Partially Root-Caused And Now Has A Guard

The workspace had recurring Chinese mojibake/history pollution issues.

What was confirmed:

- repository settings already lean toward UTF-8 + LF
- but historical polluted files still existed
- some UI strings had degraded into literal `????`

What was added:

- `admin-ui/scripts/check-encoding-mojibake.mjs`
- `admin-ui/package.json` script:
  - `check:encoding`

This is now the minimal regression guard for future frontend changes.

Validated in the current thread:

- `pnpm -C D:\GoProjects\XAdmin\admin-ui check:encoding` passed
- `pnpm -C D:\GoProjects\XAdmin\admin-ui -F @vben/web-antd run typecheck` passed

## 7. Generator Direction Shifted From Per-Project Repair To Rule Regression

The project already recognized that repeated fixes in `*.gen.go` or repeated UI
sorting/tenant patches should migrate back to generator/template layers.

This has already produced generator-facing work in `xkit`, including:

- generated default list sorting
- tenant-scoped repo generation

That direction should continue:

- avoid repeating project-local post-generation patches
- push stable tenant/sorting behavior into generator defaults

## Verified Current State

As of this handoff:

- `admin`: clean
- `admin-ui`: clean
- `xkit`: clean

## Important Documents To Read Next

If resuming multi-tenant work, read in this order:

1. `xkit/NEXT_CONTEXT_HANDOFF_20260530.md`
2. `admin/docs/tenant-improvements-summary-20260530.md`
3. `admin/docs/tenant-awareness-refactor-plan.md`
4. `admin/docs/tenant-resource-classification-v1.md`
5. `admin/docs/tenant-split-database-analysis.md`

## Recommended Resume Sequence

1. Confirm repo state:
   - `git status` in `admin`
   - `git status` in `admin-ui`
   - `git status` in `xkit`
2. Re-read current tenant design constraints:
   - platform semantics via `tenantId=0`
   - global / tenant / hybrid resource split
3. If continuing backend tenant refactor:
   - inspect repo/service/viewer boundaries first
   - do not start from random page-level UI patches
4. If continuing frontend tenant refactor:
   - preserve ownership wording rules
   - prefer backend `tenantName`
   - run both:
     - `pnpm -C D:\GoProjects\XAdmin\admin-ui check:encoding`
     - `pnpm -C D:\GoProjects\XAdmin\admin-ui -F @vben/web-antd run typecheck`
5. If continuing generator regression:
   - prioritize rules already validated in `admin`
   - move them into `xkit` rather than hand-maintaining future generated output

## Open Follow-Up Themes

- viewer / repo / service tenant isolation still needs to become more explicit
  and less scattered
- hybrid resources (`role`, `permission`, `menu`, `dict_*`) still need tighter
  long-term modeling
- platform-resource vs tenant-resource UX still needs more consistent visual
  language
- split-database remains a later-phase architecture target, not the current
  implementation baseline
