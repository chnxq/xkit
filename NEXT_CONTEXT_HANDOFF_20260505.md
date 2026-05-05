# XAdmin Context Handoff

Generated at: 2026-05-05 +08:00

This file should be the first document read in a new context for the current
XAdmin workspace. It carries forward the durable background from
`xkit/NEXT_CONTEXT_HANDOFF_20260501.md` and adds all material changes and
verified state through 2026-05-05.

## Environment

- Workspace root: `D:\GoProjects\XAdmin`
- Shell: Windows PowerShell
- Timezone used in this work: `Asia/Shanghai`
- Main backend project: `admin`
- Main frontend project: `admin-ui`
- Frontend baseline: Vben, branch `xadmin-api-integration`
- Backend was regenerated from `xkit` and `xkit-template`
- Previous generated/copy baseline backup: `D:\GoProjects\XA-backup`
- Generated frontend API client location:
  `admin-ui\apps\web-antd\src\api\generated\admin\service\v1`

## High-Level Profile

- `admin` is the active backend. It contains generated code plus hand-maintained
  `_ext.go` files and manual server/service additions.
- `admin-ui` is the active frontend. It is a manual Vben implementation and
  should not be regenerated from `xkit`.
- `xkit` owns repeatable generation rules.
- `xkit-template` owns reusable backend runtime/bootstrap skeleton code.
- Prefer root-cause fixes in generator/template/shared-lib layers instead of
  frontend-only or one-off backend workarounds.
- Preserve user changes. Do not revert unrelated worktree state.
- For a resumed context, read this file first, then inspect current git status,
  then start services explicitly if UI validation is required.

## Carry-Forward Direction From 2026-05-01

- Backend generation ownership:
  - `xkit` for generator logic, templates, schema interpretation, repo output.
  - `xkit-template` for stable startup/runtime/bootstrap/server wiring.
- Frontend ownership:
  - `admin-ui` is the real UI target.
  - Do not try to make `xkit` generate the frontend.
- Provenance policy remains:
  - copied from `xkit-template`: mark as template-generated
  - directly handwritten by Codex: mark as Codex-generated
  - existing generator-owned files keep their own generator headers
- Existing architectural decisions still stand:
  - menu sync is non-destructive and should preserve existing menu IDs and
    permission associations
  - API sync currently uses delete-and-recreate behavior
  - manual HTTP routes are injected into API sync through manual route support
  - shared bugs should be fixed in `xkit`, `xkit-template`, `x-utils`,
    `x-crud`, or `xkitmod` according to ownership

## Repository State As Of 2026-05-05

### `admin`

- Branch: `main`
- Latest commit: `7f38aca`
- Latest commit summary:
  token-auth requests without an established viewer now receive a temporary
  default viewer during user/permission lookup, then switch to the real viewer
- Working tree: clean
- Verified on 2026-05-05:
  - `go test ./...` passed

### `admin-ui`

- Branch: `xadmin-api-integration`
- Latest commit: `9ae4cee24`
- Latest commit summary:
  merge from remote `origin/xadmin-api-integration`
- Working tree: not clean
  - one staged file:
    `apps/web-antd/src/views/system/role/index.vue`
- Staged diff summary in that file:
  - only a small parenthesization change around
    `handleAuthorizePermissionChange(...)`
- Verified on 2026-05-05:
  - `pnpm -F @vben/web-antd run typecheck` failed
  - current errors are all in
    `apps/web-antd/src/views/system/role/index.vue`
  - error shape: strict nullability around permission/menu/api preview items

### `xkit`

- Branch: `main`
- Latest commit: `9d72308`
- Latest commit summary:
  generated repo list methods now apply paging and sorting
- Working tree: clean

### `xkit-template`

- Branch: `main`
- Latest commit: `4b53493`
- Latest commit summary:
  changed default password in template layer
- Working tree: clean
- Important note:
  current local `admin` seed code still uses `defaultPassword = "123456"`.
  Do not assume the template password change has already flowed into `admin`.

### Related Repositories

- `x-utils`
  - Branch: `main`
  - Latest commit: `417cf92 fix mapper proto enum conversion`
  - Working tree: clean
- `x-crud`
  - Branch: `main`
  - Latest commit: `a61448f update go.mod`
  - Working tree: clean
- `xkitmod`
  - Branch: `main`
  - Latest commit: `698fedc`
  - Latest commit summary:
    role authorization support around permission/menu/api selection flow
  - Working tree: clean
- `xkitpkg`
  - Branch: `main`
  - Latest commit: `80e4b48 add enggo readme.md; upgrade go mod`
  - Working tree: clean

## Main Work Completed After 2026-05-01

### 1. Generator And Generated Repo Fixes

Relevant commits:

- `xkit 1ba9b76 feat: generate repo setters for optional tree and json fields`
- `xkit 9d72308 fix: apply paging and sorting in generated repo lists`
- `admin 134dec6 fix: persist generated repo parent and metadata fields`
- `admin 8936644 feat: enrich admin user repo relations and sorting`

What changed:

- Generator output now handles optional tree parent fields and JSON metadata
  fields more correctly in generated repos.
- Generated list methods now apply paging and sorting consistently.
- Generated backend repo code in `admin` was refreshed to pick up those fixes.
- User-related list/detail responses were enriched so frontend pages can display
  organization, position, and role relationships without custom ad hoc shaping.

Problem threads these changes were tied to:

- org-unit parent selection not persisting correctly
- duplicate or broken handling around parent-related save fields
- menu title/metadata persistence path being ineffective because menu metadata
  is stored in `sys_menu.metadata`
- list pages needing sortable columns backed by real backend sorting

Important files:

- `xkit/internal/codegen/template/repo_file.tmpl`
- `admin/internal/data/repo/menu_repo.gen.go`
- `admin/internal/data/repo/org_unit_repo.gen.go`
- `admin/internal/data/repo/user_repo.gen.go`
- `admin/internal/data/ent/query_modify_ext.go`

### 2. Unified Admin List Toolbar And User Management Expansion

Relevant commits:

- `admin-ui 44bdae98a feat: unify admin list table toolbars`
- `admin-ui 4c7c65304 feat: enhance admin user management and list sorting`
- `admin-ui 542d04c25 fix: stabilize admin user edit form behavior`
- `admin-ui ddc9a3b64 fix: stabilize admin user edit form behavior`

What changed in the frontend:

- All admin list pages were moved toward a common toolbar with:
  - export
  - refresh
  - fullscreen
  - column settings
- User list now supports sortable columns.
- User filters were expanded to include:
  - username
  - realname
  - mobile
  - telephone
  - org unit
  - position
  - role
  - status
- User create/edit forms were expanded to handle:
  - org-unit assignment
  - position assignment
  - role assignment
  - more profile fields
- User list display now shows relationship tags for org units, positions, and
  roles.
- Form layout and toolbar alignment were reworked to be more consistent.
- Browser autofill interference was mitigated in the user form by adding hidden
  autofill guards and explicit field names/autocomplete behavior.

Important files:

- `admin-ui/apps/web-antd/src/components/admin-table-toolbar/index.vue`
- `admin-ui/apps/web-antd/src/components/admin-table-toolbar/shared.ts`
- `admin-ui/apps/web-antd/src/views/system/user/index.vue`
- `admin-ui/apps/web-antd/src/api/admin/users.ts`

### 3. Role Authorization Flow

Relevant commits:

- `admin 228a1dd`
- `xkitmod 698fedc`

Commit summary:

- role list gained an `Authorize` action
- a dedicated role-authorization dialog was added
- dialog loads role detail, permissions, menus, and APIs
- selected permissions drive a live preview of linked menus and APIs
- saving still uses the existing update-role contract instead of changing the
  backend API surface

Important backend/frontend files:

- `admin/internal/service/permission_service_ext.go`
- `admin/internal/data/repo/role_repo_ext.go`
- `admin/internal/data/repo/permission_repo_ext.go`
- `admin/internal/server/manual_http.go`
- `admin/internal/server/http_options.go`
- `admin-ui/apps/web-antd/src/views/system/role/index.vue`

Current caution:

- `admin-ui/apps/web-antd/src/views/system/role/index.vue` is also the current
  blocking file for frontend typecheck.
- Before continuing broader frontend work, clear the nullability errors in this
  file and decide whether to commit the staged change.

### 4. Fresh-Database Seed Data And Viewer-Context Repair

Relevant commit:

- `admin 7f38aca`

What changed:

- Backend migrate/bootstrap now seeds default business data after schema create.
- Auth viewer middleware was fixed so permission lookup works even when a token
  is present but a real viewer has not yet been attached to context.

Why this mattered:

- With an empty database plus migrate, `admin` login succeeded but CRUD buttons
  disappeared in the UI.
- Root cause was not missing permissions in seed data.
- Root cause was `security: missing ViewerContext in context` during user and
  permission lookup inside viewer middleware.
- That failure caused `/admin/v1/perm-codes` to return an empty permission set,
  which made frontend `v-access:code` hide buttons.

Current fix path:

- `authViewerMiddleware` now:
  - short-circuits only when a real viewer already exists
  - injects a temporary default viewer for lookup queries
  - loads user, role permissions, and permission codes through that context
  - then replaces the temporary viewer with a real user viewer

Important files:

- `admin/internal/server/viewer_auth.go`
- `admin/internal/data/bootstrap/default_data_ext.go`
- `admin/internal/data/bootstrap/ent_client.gen.go`

## Seed Data Behavior In Current `admin` Checkout

On migrate, backend bootstrap now does all of the following:

- sync default menus
- sync APIs from embedded OpenAPI plus manual routes
- sync permissions
- create a default tenant
- create a root org unit plus one child department
- create positions
- create two roles:
  - `SUPER_ADMIN`
  - `USER`
- create two users:
  - `admin`
  - `user`
- create and refresh user/org/position/role/membership relations
- grant all permissions to `SUPER_ADMIN`
- grant a limited subset to `USER`

Current password fact in local `admin` source:

- `admin/internal/data/bootstrap/default_data_ext.go` still sets:
  - `admin / 123456`
  - `user / 123456`

## Current Verified Runtime Facts

Verified on 2026-05-05:

- no service is currently listening on these ports:
  - `5666`
  - `7788`
  - `7789`
  - `7790`
- backend verification passed:

```powershell
cd D:\GoProjects\XAdmin\admin
go test ./...
```

- frontend verification currently fails:

```powershell
cd D:\GoProjects\XAdmin\admin-ui
pnpm -F @vben/web-antd run typecheck
```

Current failing areas from that typecheck:

- `src/views/system/role/index.vue(269, 275)`
- `src/views/system/role/index.vue(940-997)`
- all reported issues are nullability / possible `undefined` handling

## Important Current Files To Inspect First

- `xkit/NEXT_CONTEXT_HANDOFF_20260505.md`
- `xkit/NEXT_CONTEXT_HANDOFF_20260501.md`
- `admin/internal/server/viewer_auth.go`
- `admin/internal/data/bootstrap/default_data_ext.go`
- `admin/internal/data/bootstrap/ent_client.gen.go`
- `admin/internal/data/repo/menu_repo.gen.go`
- `admin/internal/data/repo/org_unit_repo.gen.go`
- `admin/internal/data/repo/user_repo.gen.go`
- `admin-ui/apps/web-antd/src/views/system/user/index.vue`
- `admin-ui/apps/web-antd/src/views/system/role/index.vue`
- `admin-ui/apps/web-antd/src/components/admin-table-toolbar/index.vue`
- `xkit/internal/codegen/template/repo_file.tmpl`

## Known Risks And Open Follow-Up Items

- `admin-ui` has one staged but uncommitted change in
  `apps/web-antd/src/views/system/role/index.vue`.
- Frontend typecheck is not green until `role/index.vue` nullability handling is
  fixed.
- Services are currently stopped; do not assume the backend or frontend is
  running.
- API sync still uses delete-and-recreate behavior. This is workable for now,
  but it can become a problem if downstream logic starts relying on stable API
  IDs.
- Menu sync should remain non-destructive unless the user explicitly asks for a
  reset.
- Carry-forward risk from the previous investigation, not re-verified in this
  turn:
  permission-management page access codes may still mismatch backend-generated
  codes, especially around plural/singular naming like
  `permission:groups:*` versus `permission:group:*`.
- PowerShell output can still show garbled Chinese text if the console decoding
  is wrong. Treat that as a console issue first, not immediate file corruption.

## Recommended Resume Sequence

1. Read this file first.
2. Check `git status` in `admin`, `admin-ui`, and `xkit`.
3. Resolve the staged/frontend typecheck issue in
   `admin-ui/apps/web-antd/src/views/system/role/index.vue`.
4. If UI validation is needed, start services explicitly:

```powershell
cd D:\GoProjects\XAdmin\admin
go run ./cmd/server server -config_path ./configs
```

```powershell
cd D:\GoProjects\XAdmin\admin-ui
pnpm -F @vben/web-antd run dev
```

5. Re-test these flows in order:
   - login as `admin`
   - permission-driven CRUD button visibility
   - user create/edit
   - org-unit parent save
   - menu title/metadata save
   - role authorization dialog
6. If only the permission-management page still hides buttons, inspect frontend
   permission-code constants before changing backend permission generation.

## Commit Landmarks Since 2026-05-01

Backend:

```text
134dec6 fix: persist generated repo parent and metadata fields
8936644 feat: enrich admin user repo relations and sorting
228a1dd role authorization flow enhancements
7f38aca viewer context repair plus default-data/bootstrap updates
```

Frontend:

```text
44bdae98a feat: unify admin list table toolbars
4c7c65304 feat: enhance admin user management and list sorting
542d04c25 fix: stabilize admin user edit form behavior
ddc9a3b64 fix: stabilize admin user edit form behavior
9ae4cee24 merge remote xadmin-api-integration
```

Generator / template / shared:

```text
xkit       1ba9b76 feat: generate repo setters for optional tree and json fields
xkit       9d72308 fix: apply paging and sorting in generated repo lists
xkitmod    698fedc role authorization support work
x-template 4b53493 change default pwd
x-utils    417cf92 fix mapper proto enum conversion
```
