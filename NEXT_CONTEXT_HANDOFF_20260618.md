# XAdmin Context Handoff

Generated at: 2026-06-18 +08:00

This file is the latest resume entry for the current workspace state.
It should be treated as the primary handoff after:

- `xkit/NEXT_CONTEXT_HANDOFF_20260501.md`
- `xkit/NEXT_CONTEXT_HANDOFF_20260505.md`
- `xkit/NEXT_CONTEXT_HANDOFF_20260510.md`
- `xkit/NEXT_CONTEXT_HANDOFF_20260522.md`
- `xkit/NEXT_CONTEXT_HANDOFF_20260530.md`
- `xkit/NEXT_CONTEXT_HANDOFF_20260608.md`
- `xkit/NEXT_CONTEXT_HANDOFF_20260613.md`

This handoff explicitly incorporates the current thread's work.

## Scope And Incorporation

- This handoff explicitly includes facts from the current thread.
- It covers:
  - `xkit` frontend meta generation work after the 2026-06-13 handoff
  - `examples/admin` and `examples/admin-v2` frontend config alignment
  - generated frontend meta rollout in `admin-ui`
  - placeholder/i18n generation behavior adjustments
  - current social-auth bind captcha finding
  - current repo state and recent commits needed for the next resume

## Environment

- Workspace root: `D:\GoProjects\XAdmin`
- Shell: Windows PowerShell
- Timezone: `Asia/Shanghai`
- Main repos in scope:
  - `xkit`
  - `admin-ui`
  - `admin`

## Repository State As Of 2026-06-18

### `xkit`

- Branch: `main`
- HEAD: `1f9808a` `fix: placeholder generate text`
- Working tree: clean

Relevant recent commits in this round:

- `1f9808a` `fix: placeholder generate text`
- `cbe29e7` `feat(config): sync frontend metadata configs`
- `ee32aa1` `feat(admin-config): add frontend form metadata for internal message`
- `3255a17` `feat(admin-config): add frontend form metadata for menu and task`
- `2dd01e8` `feat: frontend-view-meta-generation add content form-generate`
- `7837f52` `feat: frontend-view-meta-generation add enum values`
- `351fdfb` `feat: frontend-view-meta-generation`

### `admin-ui`

- Branch: `xadmin-api-integration`
- HEAD: `26259197` `fix: placeholder generate text`
- Working tree: dirty

Current local changes intentionally preserved:

- `apps/web-antd/src/views/app/log/permission-audit-log/index.vue`
- `apps/web-antd/src/views/task/log/index.vue`
- multiple generated files under:
  - `apps/web-antd/src/views/generated/admin/...`

Important note:

- this dirty state is expected from local regeneration / page wiring work and
  should be inspected before any reset or regeneration

Relevant recent commits in this round:

- `26259197` `fix: placeholder generate text`
- `51cd2d1b` `feat(web-antd): wire generated meta for internal message`
- `7e78d1e8` `feat(web-antd): wire generated meta for menu and task`
- `644f5ff6` `feat(web-antd): checkpoint generated admin meta pages`
- `e806b28e` `feat: add generated admin search meta resources`
- `ddc45590` `fix: refresh auth and profile completion flow`

### `admin`

- Branch: `chnxq/dev`
- HEAD: `33ee57e` `change merge task_service_ext`
- Working tree: dirty

Current local changes intentionally preserved:

- `configs/auth.yaml`

Recent backend commits relevant to the current frontend/meta round:

- `33ee57e` `change merge task_service_ext`
- `69a17fe` `refactor: consolidate task domain logic`
- `81bc8d8` `refactor: consolidate bootstrap and file modules`
- `1339312` `refactor: use shared server utils`

## 1. Main Work Completed Since The 2026-06-13 Handoff

This round was mainly about building a usable:

- `xkit config -> generated frontend meta -> admin-ui page wiring`

pipeline for real `admin` / `admin-ui` rather than only example code.

The work moved in three layers:

1. `xkit` frontend meta generation capability
2. `examples/admin*` config alignment
3. `admin-ui` page adoption and generated resource wiring

## 2. What Changed In `xkit`

### 2.1 Frontend Meta Generation Landed

`xkit` now has a real `frontend-meta` generation flow producing:

- generated view meta files
- generated `page_i18n.zh-CN.json`
- generated `page_i18n.en-US.json`
- generated enum resource copies under:
  - `views/generated/admin/langs/...`

At that stage the docs still described the output as a fixed `admin-ui` path,
but the intended rule is now:

- frontend meta output is derived from `TypeScriptRoot`
- it follows the same root-resolution logic as generated TypeScript API code
- it should not depend on a top-level `admin.yaml` output-root setting

This was introduced across these commits:

- `351fdfb`
- `7837f52`
- `2dd01e8`

The generated module shape in `admin-ui` includes:

- `buildSearchFormOptions`
- `buildListGridColumns`
- `buildFormOptions` for resources with dialog form metadata

### 2.2 `admin` / `admin-v2` Example Configs Were Expanded

`xkit/examples/admin/admin-target-config/admin.yaml` is the effective source of
truth for current frontend generation work.

Later in the round, its existing `frontend` sections were synced back into:

- `xkit/examples/admin/admin-config/admin.yaml`
- `xkit/examples/admin-v2/admin-config/admin.yaml`

Important result:

- the three files now expose the same frontend coverage set for the resources
  already adopted in this round

The sync commit is:

- `cbe29e7` `feat(config): sync frontend metadata configs`

### 2.2.1 What `admin-v2` Work Clarified

The `admin-v2` round was important not because it became the final production
config baseline, but because it forced several generation boundaries to become
explicit.

Confirmed outcomes from that round:

- schema resource name and proto service name cannot be assumed to align
  one-to-one
- some resources are valid backend data resources even when they do not have a
  directly bindable proto/admin wrapper service yet
- for those cases, generation must not silently pretend the mapping exists
- `admin-target-config` remains the place to carry the validated explicit
  frontend/resource decisions after real project verification

The skipped-resource logs from the earlier `admin-v2` generation round were a
real signal rather than noise, for example:

- `DictCategoryI18n`
- `Membership`
- `MembershipOrgUnit`
- `MembershipPosition`
- `MembershipRole`
- `PermissionApi`
- `PermissionMenu`
- `PermissionPolicy`
- `RoleMetadata`
- `RolePermission`
- `UserOrgUnit`
- `UserPosition`
- `UserRole`

Those names should be treated as:

- resources that may exist at schema level
- but not automatically eligible for repo/service generation unless the proto /
  admin wrapper side is explicitly aligned

The practical rule refined through `admin-v2` was:

1. do not force xkit to invent false schema-to-proto matches
2. allow the config to carry explicit frontend/repo/service intent only after
   the real project proves the mapping
3. if a resource only supports partial generation, generate the safe framework
   and leave handwritten extension code in `_ext.go` or UI page code

### 2.3 Placeholder Generation Was Tightened

The old generated placeholder behavior produced poor zh-CN UI strings such as:

- `创建时间 Range`
- `Search 任务ID`
- `Select 状态`

`xkit/internal/codegen/runner.go` was then adjusted so that in zh-CN locale:

- `Input` / `InputNumber` placeholders fall back to the field label
- `Select` placeholders fall back to the field label
- `RangePicker` placeholders fall back to the field label

This removed the mixed English verb artifacts in zh-CN generated page i18n.

Files changed:

- `xkit/internal/codegen/runner.go`
- `xkit/internal/codegen/runner_test.go`

Commit:

- `1f9808a` `fix: placeholder generate text`

Validation done:

- `go test ./internal/codegen -count=1`
- reran `xkit gen frontend-meta ...`

Important nuance:

- the current committed behavior specifically fixed zh-CN noise
- en-US placeholders may still intentionally keep `Search ...` / `Select ...`
  phrasing depending on the current runner logic at resume time

## 3. Current `examples/admin*` Frontend Coverage

After the sync work, the frontend config coverage in the `admin-target-config`
set includes these `view_path` resources:

- `system/api`
- `app/log/api-audit-log`
- `system/file`
- `app/internal-message/message`
- `app/log/login-audit-log`
- `system/menu`
- `system/org-unit`
- `app/log/permission-audit-log`
- `system/position`
- `system/role`
- `task/task`
- `task/log`
- `system/tenant`
- `system/user`

These are the key resources currently participating in generated meta output.

## 4. What Changed In `admin-ui`

### 4.1 Generated Resource Tree Exists

Generated resources now live under the frontend project resolved from
`TypeScriptRoot`, for example:

- `admin-ui/apps/web-antd/src/views/generated/admin/`

Examples:

- `system/api.meta.ts`
- `system/tenant.meta.ts`
- `system/user.meta.ts`
- `system/menu.meta.ts`
- `task/task.meta.ts`
- `task/log.meta.ts`
- `app/log/api-audit-log.meta.ts`
- `app/log/login-audit-log.meta.ts`
- `app/log/permission-audit-log.meta.ts`
- `app/internal-message/message.meta.ts`
- `page_i18n.zh-CN.json`
- `page_i18n.en-US.json`
- `langs/zh-CN/enum.json`
- `langs/en-US/enum.json`

### 4.2 Shared Generated Form Renderer Was Added / Extended

The generated form renderer component in:

- `admin-ui/apps/web-antd/src/components/admin-generated-form/index.vue`

was extended earlier in the round to support more generated form cases such as:

- `Password`
- `AutoComplete`
- `IconPicker`
- custom component instances
- extra props such as `formItemClass`

This component is now part of the expected generated form path.

### 4.3 Real Pages Already Adopted The Pattern

Pages already moved onto the generated-meta pattern include:

- `system/api`
- `system/tenant`
- `system/org-unit`
- `system/position`
- `system/role`
- `system/user`
- `system/menu`
- `task/task`
- `app/internal-message/message`

Also specifically handled in this round:

- `app/log/permission-audit-log`
- `task/log`

For the last two:

- search form and list columns are generated
- page layer keeps only business-specific overrides and slots

### 4.4 `task/log` Specific Outcome

`task/log` was not converted from a plain table to VXE in this round; it was
already on `VxeTableGrid`.

What changed was:

- the handwritten top search form was removed
- search was moved into `useVbenVxeGrid(... formOptions ...)`
- then the page was connected to generated meta

Business logic intentionally kept handwritten:

- detail modal
- task name backfill
- cron rendering
- right-side `action` column

### 4.5 `permission-audit-log` Specific Outcome

`permission-audit-log` was identified as a good generated-meta candidate
because it needs:

- generated search form
- generated list columns

but does not require a dialog CRUD form.

It was connected to generated meta while keeping:

- action tag rendering
- old/new value slots
- request parameter mapping

in the page layer.

## 5. Important Runtime / Process Notes

### 5.1 `admin-target-config` Is The Operational Source

During the round it became explicit that:

- `xkit/examples/admin/admin-config/admin.yaml` can lag
- `xkit/examples/admin/admin-target-config/admin.yaml` was the file actually
  being iterated on and used for generation

For future resume work:

- start from `admin-target-config`
- only backport/sync to the other config files once the target config is known good

### 5.2 Large YAML Edits Were Error-Prone

There were multiple user-reported issues earlier in the round around:

- YAML structure accidentally getting mangled
- UTF-8 / Chinese text corruption
- partial config loss when editing large blocks manually

The safest practice going forward is:

- avoid broad manual rewrites of the big config files
- prefer narrow, verified insertions
- immediately regenerate and inspect deltas after config edits

### 5.2.1 Additional Rules Refined From `admin-v2`

The `admin-v2` work also clarified a few operating rules that should be carried
forward:

- treat `examples/admin/admin-target-config/admin.yaml` as the live iteration
  baseline
- only sync back into:
  - `examples/admin/admin-config/admin.yaml`
  - `examples/admin-v2/admin-config/admin.yaml`
  after the target config has already been validated by generation and real
  project usage
- when a generated page needs list/search/form metadata, prefer to encode that
  intent in `frontend` config first rather than hand-spreading the same rule
  across page code
- if the generator can safely produce an empty framework and the rest is still
  handwritten, that is acceptable; xkit does not need to solve the whole page
  on day one
- placeholder/i18n quality problems should be fixed either:
  - in generator fallback rules
  - or in explicit `frontend` config
  rather than patched ad hoc in each generated page
- enum i18n is a special case: copied/generated language resources under
  `views/generated/admin/langs/` are a better current strategy than trying to
  infer every enum label directly inside xkit

### 5.3 Generated Files Can Cause Wide Diff Drift

Running `xkit gen frontend-meta` can rewrite:

- timestamps
- formatting in generated meta files
- page i18n files

even when only a small subset of resources materially changed.

Before committing `admin-ui`, always inspect whether a diff is:

- semantic
- or only generated timestamp / formatting churn

## 6. Current Social Auth Finding

Near the end of this round, a separate issue was investigated:

- GitHub bind-existing-account flow failed with:
  - `401 UNAUTHORIZED`
  - `invalid captcha code`

The backend path is:

- `admin/internal/server/social_auth_ext.go`
  - `ConfirmBindOrRegister(...)`
  - `BIND_EXISTING` branch explicitly calls `verifyCaptcha(...)`

The captcha verifier is:

- `admin/internal/server/captcha_ext.go`

Important confirmed behavior:

- bind-existing-account requires its own captcha
- captcha is one-time consumed via `captchaStore.Verify(..., true)`
- this is not a GitHub OAuth callback issue
- it is not a password-check failure
- it is a true captcha mismatch / stale captcha / reused captcha issue

Frontend path involved:

- `admin-ui/apps/web-antd/src/views/_core/authentication/social-auth-bind.vue`
- `admin-ui/apps/web-antd/src/api/admin/social-auth.ts`

Additional note:

- `admin-ui/apps/web-antd/src/api/request.ts` currently treats all `401`
  responses as unauthorized session problems and may trigger logout handling
  even for business errors like invalid captcha

That front-end 401 handling remains a reasonable follow-up item.

## 7. Validation Performed In This Round

### `xkit`

- `go test ./internal/codegen -count=1`
- frontend-meta generation reruns against:
  - `examples/admin/admin-target-config/admin.yaml`

### `admin-ui`

Used repeatedly during the round:

- `pnpm exec eslint <changed-files> --fix`
- `pnpm exec oxfmt <changed-files>`
- `pnpm -F @vben/web-antd run typecheck`

At key checkpoints, typecheck passed after page wiring changes.

## 8. Current Risks / Open Ends

### 8.1 `admin-ui` Is Still Dirty

`admin-ui` is not at a clean checkpoint in the current local state.

This matters because the next session should not assume:

- generated files are already committed
- local page wiring changes were fully reviewed
- all diffs are safe to discard

Check `git status --short` first in `admin-ui`.

### 8.2 English Placeholder Strategy Is Not Fully Settled

Current round outcome focused on:

- fixing ugly zh-CN placeholder strings

English placeholders may still legitimately appear as:

- `Search ...`
- `Select ...`

If the team later prefers a shorter English style, that would require another
small generator pass.

### 8.3 Remaining Page Rollout Is Still Selective

Although many pages were adopted, the generated pattern is not yet universal.

Some pages are still intentionally manual or only partially migrated.

Examples of caution areas discussed in this round:

- `system/file` is not a clean full-dialog CRUD candidate because upload /
  preview / download behavior is custom
- pages with richer business forms may only want generated list/search sections

## 9. Recommended Resume Sequence

For the next continuation on this frontend-meta track:

1. read this file first
2. inspect current dirty state in:
   - `xkit`
   - `admin-ui`
   - `admin`
3. treat `xkit/examples/admin/admin-target-config/admin.yaml` as the current
   operational config baseline
4. if continuing generation work, start from:
   - `xkit/internal/codegen/runner.go`
   - `xkit/internal/codegen/template/frontend_view_meta.tmpl`
   - `xkit/examples/admin/admin-target-config/admin.yaml`
5. if continuing UI rollout, inspect:
   - `admin-ui/apps/web-antd/src/views/generated/admin/`
   - currently wired page `index.vue` files
6. if continuing social-auth debugging, inspect:
   - `admin/internal/server/social_auth_ext.go`
   - `admin/internal/server/captcha_ext.go`
   - `admin-ui/apps/web-antd/src/views/_core/authentication/social-auth-bind.vue`
   - `admin-ui/apps/web-antd/src/api/request.ts`

## 10. Good Next Steps

The most sensible next-step options are:

1. continue `admin-ui` rollout to additional pages that are good generated-meta
   candidates
2. review and reduce `admin-ui` dirty generated diffs into a commit-safe set
3. tighten `admin-ui` 401 handling so business captcha failures do not behave
   like session-expired errors
4. optionally unify the English placeholder generation style if desired
5. continue evaluating which pages should use:
   - generated search/list only
   - generated search/list/form
   - or remain mostly manual
6. if generation rules need another proving ground, use `admin-v2` as the
   bounded experimental surface first, then promote only the validated subset
   back into `admin`-family baseline configs

## 11. Commits Produced In This Round

### `xkit`

- `351fdfb` `feat: frontend-view-meta-generation`
- `7837f52` `feat: frontend-view-meta-generation add enum values`
- `2dd01e8` `feat: frontend-view-meta-generation add content form-generate`
- `3255a17` `feat(admin-config): add frontend form metadata for menu and task`
- `ee32aa1` `feat(admin-config): add frontend form metadata for internal message`
- `cbe29e7` `feat(config): sync frontend metadata configs`
- `1f9808a` `fix: placeholder generate text`

### `admin-ui`

- `e806b28e` `feat: add generated admin search meta resources`
- `644f5ff6` `feat(web-antd): checkpoint generated admin meta pages`
- `7e78d1e8` `feat(web-antd): wire generated meta for menu and task`
- `51cd2d1b` `feat(web-antd): wire generated meta for internal message`
- `26259197` `fix: placeholder generate text`

### `admin`

- `1339312` `refactor: use shared server utils`
- `81bc8d8` `refactor: consolidate bootstrap and file modules`
- `69a17fe` `refactor: consolidate task domain logic`
- `33ee57e` `change merge task_service_ext`
