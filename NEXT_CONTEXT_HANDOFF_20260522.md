# XAdmin Context Handoff

Generated at: 2026-05-22 16:20 +08:00

This file is the new resume entry for the current workspace state.  
It inherits `xkit/NEXT_CONTEXT_HANDOFF_20260501.md`,
`xkit/NEXT_CONTEXT_HANDOFF_20260505.md`, and
`xkit/NEXT_CONTEXT_HANDOFF_20260510.md`, and includes all major verified
changes through the current thread.  
The filename target is `20260522` per corrected user request, and the verified runtime
and repository snapshot below was collected on 2026-05-22.

## Scope And Incorporation

- This handoff **explicitly includes** facts from the current thread.
- It covers:
  - `admin/docs/generated-code-files.md` ownership-boundary refresh
  - `xkit` bootstrap provider/service injection convergence
  - `xkit-template` and `admin` `manual_http.go` / `manual_http_data.go`
    boundary refactor
  - verified generator re-run of `xkit gen bootstrap admin`
  - final three-repo commits that closed this convergence round

## Environment

- Workspace root: `D:\GoProjects\XAdmin`
- Shell: Windows PowerShell
- Timezone: `Asia/Shanghai`
- Main backend: `admin`
- Main frontend: `admin-ui` (Vben, branch `xadmin-api-integration`)
- Generator/tooling repos: `xkit`, `xkit-template`, `x-crud`, `x-utils`,
  `xkitmod`, `xkitpkg`

## Repository State As Of 2026-05-22

### `admin`

- Branch: `main`
- HEAD: `cba6fe1 refactor(bootstrap): 对齐 manual http 与 generated data 边界`
- Working tree: clean
- Verified:
  - `go test ./...` passed on 2026-05-22
  - `go run ./cmd/xkit gen bootstrap admin --project D:\GoProjects\XAdmin\admin --config D:\GoProjects\XAdmin\xkit\examples\xadmin\admin-config\admin.yaml` completed successfully from `xkit`

### `admin-ui`

- Branch: `xadmin-api-integration`
- HEAD: `c7a41c6c2 ...`
- Working tree: clean
- No new frontend code was changed in this thread

### `xkit`

- Branch: `main`
- HEAD: `885016f test(scaffold): 锁定 manual http 模板同步边界`
- Working tree: clean
- Verified: `go test ./...` passed on 2026-05-22

### `xkit-template`

- Branch: `main`
- HEAD: `eb96cc4 refactor(template): 收紧 manual http preserve 边界`
- Working tree: clean
- Verified: `go test ./...` passed on 2026-05-22

### Related Repos

- `x-utils`: `417cf92 fix mapper proto enum conversion`
- `x-crud`: `a61448f update go.mod`
- `xkitmod`: `698fedc ...`
- `xkitpkg`: `80e4b48 add enggo readme.md; upgrade go mod`

## Runtime Status (Snapshot)

At handoff time, services are running:

- `5666`: Vite dev server (`admin-ui`)
- `7788`: `admin.exe` REST
- `7789`: `admin.exe` SSE
- `7790`: `admin.exe` gRPC

Process note:

- Ports `7788/7789/7790` are owned by the same backend process.

## Major Changes Since 2026-05-10

## 1. Generated-Code Ownership Inventory Was Rebuilt

Key result:

- `admin/docs/generated-code-files.md` was rewritten around the current real
  boundary instead of older 2026-05-01 assumptions.

Current four-way split recorded there:

- repeatable `xkit` generated files
- one-time `xkit` extension files
- `xkit-template` copied startup skeleton
- `codex` handwritten project extensions

Important boundary facts now recorded:

- `internal/bootstrap/generated_hooks_ext.go` is one-time extension scaffolding,
  not repeatable generation
- `internal/data/bootstrap/ent_client_ext.go` is the handwritten post-schema
  hook behind `afterEntSchemaCreate(...)`
- `internal/server/manual_http.go` is now template-owned baseline
- `internal/server/manual_http_data.go` is the project-owned handwritten HTTP
  hook that needs `GeneratedData`

Relevant commit:

- `759fc67 docs(inventory): 更新生成代码文件清单`

## 2. Bootstrap Generator/Template Boundary Was Tightened

This thread continued the earlier bootstrap convergence work that had already
produced:

- `7092e2a feat(codegen): 修正 bootstrap 服务依赖注入`
- `a0b7498 refactor(bootstrap): 对齐新生成器 bootstrap 输出`

Those earlier changes established:

- `resource.service_repos` support in `xkit`
- multiple repo injection for generated services
- shared `repo_interface` handling
- `internal/data/bootstrap/ent_client_ext.go` as the handwritten migration/data
  hook target

This thread then pushed the boundary further in the `manual_http` area.

## 3. `manual_http.go` Was Moved Back To Template Baseline

Before this thread:

- `admin/internal/server/manual_http.go` had accumulated a large amount of
  project-specific business logic
- this caused drift against `xkit-template`
- template sync could not safely realign the baseline file

After this thread:

- `admin/internal/server/manual_http.go` is back to the empty template-owned
  hook shape
- project-specific logic was moved into:
  - `admin/internal/server/manual_http_data.go`
- `admin/internal/server/http.go` now calls:
  - `RegisterManualHTTPServices(srv, appCtx)`
  - `RegisterManualHTTPServicesWithData(srv, appCtx, data)`

The moved business logic still includes:

- login/logout/refresh token manual HTTP binding
- portal navigation and dashboard analytics handlers
- profile update and password change handlers
- manual menu sync route

This was the largest structural change in the current thread.

Relevant files:

- `admin/internal/server/http.go`
- `admin/internal/server/manual_http.go`
- `admin/internal/server/manual_http_data.go`
- `xkit-template/internal/server/manual_http.go`
- `xkit-template/internal/server/manual_http_data.go`

## 4. `generated_data_ext.go` Was Removed From `admin`

Earlier `admin` state still had handwritten bridge helpers in:

- `admin/internal/bootstrap/generated_data_ext.go`

That file had been carrying methods such as:

- `GetAppCtx()`
- `UserRepoProvider()`
- `UserCredentialRepoProvider()`
- `ApiAuditLogRepoProvider()`
- audit-log writer accessors

During this thread:

- `db_logging_ext.go` was changed to depend directly on formal generated repo
  provider methods instead of handwritten writer helpers
- `generated_data_ext.go` became unnecessary and was deleted
- `xkit gen bootstrap admin` was rerun successfully
- `generated_data_providers.gen.go` now formally generates:
  - `GetAppCtx()`
  - `ApiAuditLogRepoProvider()`
  - `UserRepoProvider()`
  - `UserCredentialRepoProvider()`
  - all other repo provider methods

This is important because the project has moved from “handwritten compatibility
bridge” back to “real generated provider layer”.

Relevant files:

- `admin/internal/bootstrap/db_logging_ext.go`
- `admin/internal/bootstrap/generated_data_providers.gen.go`
- deleted `admin/internal/bootstrap/generated_data_ext.go`

## 5. Template Sync Rule Was Updated And Locked By Test

`xkit-template` changes:

- `template.yaml` no longer preserves `internal/server/manual_http.go`
- `manual_http.go` is now explicitly treated as template-owned baseline
- `manual_http_data.go` remains preserved as project-owned logic
- Chinese and English README files were updated to explain this boundary

`xkit` changes:

- added scaffold test covering force sync behavior:
  - `manual_http.go` must be replaced by template baseline
  - `manual_http_data.go` must remain project-owned

This is the key durable protection against future drift.

Relevant files:

- `xkit-template/template.yaml`
- `xkit-template/README.md`
- `xkit-template/README.en.md`
- `xkit/internal/scaffold/template_test.go`
- `xkit/README.md`

## 6. Verification Results

Verified commands in this thread:

### `admin`

```powershell
cd D:\GoProjects\XAdmin\admin
go test ./...
```

Passed on 2026-05-22.

### `xkit`

```powershell
cd D:\GoProjects\XAdmin\xkit
go test ./...
```

Passed on 2026-05-22.

### `xkit-template`

```powershell
cd D:\GoProjects\XAdmin\xkit-template
go test ./...
```

Passed on 2026-05-22.

### Bootstrap regeneration

```powershell
cd D:\GoProjects\XAdmin\xkit
go run ./cmd/xkit gen bootstrap admin --project D:\GoProjects\XAdmin\admin --config D:\GoProjects\XAdmin\xkit\examples\xadmin\admin-config\admin.yaml
```

Observed result:

- wrote `generated_servers.gen.go`
- wrote `generated_data_providers.gen.go`
- wrote `ent_client.gen.go`
- preserved `generated_hooks_ext.go`
- preserved `ent_client_ext.go`

This verified that the provider/accessor layer is again owned by generator
output rather than local handwritten bridge code.

## 7. Final Commit Landmarks For This Thread

### `xkit-template`

- `eb96cc4 refactor(template): 收紧 manual http preserve 边界`

### `xkit`

- `885016f test(scaffold): 锁定 manual http 模板同步边界`

### `admin`

- `cba6fe1 refactor(bootstrap): 对齐 manual http 与 generated data 边界`

## 8. Current Behavior Notes

- `manual_http.go` should now be treated as template-owned baseline in future
  sync/convergence work.
- `manual_http_data.go` is the correct place for project manual HTTP handlers
  that need generated repositories/services.
- `generated_data_providers.gen.go` is now authoritative for `GeneratedData`
  accessor/provider methods.
- `db_logging_ext.go` is still project-owned handwritten business logging logic,
  but it no longer needs the old handwritten provider bridge file.
- `generated_servers.gen.go` and `ent_client.gen.go` changed in this thread only
  by regenerated timestamps/content alignment, not by new business logic.

## Risks / Follow-Up Items

- `admin-ui` was not modified in this thread; any frontend alignment for
  `manual_http_data.go` changes is indirect only, via backend behavior staying
  compatible.
- The backend/provider convergence is now much cleaner, but the next meaningful
  step is to push this from “project aligned to generator/template” toward
  “generator/template fully own the intended boundary by default”.
- If another project still has old handwritten `manual_http.go` drift, template
  force sync behavior should now be usable to realign it, but that should still
  be applied carefully repo by repo.
- `generated-code-files.md` should be updated again if later bootstrap/template
  refactors introduce new extension points or remove existing ones.

## Recommended Resume Sequence

1. Read this file first.
2. Then open:
   - `xkit/NEXT_CONTEXT_HANDOFF_20260510.md`
   - `xkit/doc/bootstrap-template-generated-boundary.md`
   - `admin/docs/generated-code-files.md`
3. Confirm current repo state:
   - `git status` in `admin`
   - `git status` in `xkit`
   - `git status` in `xkit-template`
4. If continuing generator/template convergence, inspect in this order:
   - `xkit-template/template.yaml`
   - `xkit/internal/scaffold/template_test.go`
   - `admin/internal/server/manual_http.go`
   - `admin/internal/server/manual_http_data.go`
   - `admin/internal/bootstrap/generated_data_providers.gen.go`
5. If a future refactor needs bootstrap regeneration, rerun:
   - `go run ./cmd/xkit gen bootstrap admin --project D:\GoProjects\XAdmin\admin --config D:\GoProjects\XAdmin\xkit\examples\xadmin\admin-config\admin.yaml`
6. Re-verify with:
   - `go test ./...` in `admin`
   - `go test ./...` in `xkit`
   - `go test ./...` in `xkit-template`

## Commit Landmarks Since 2026-05-10

### `admin`

- `feb256b` station-message revoke policy and visibility constraints
- `8e094d4` captcha endpoint and login captcha enforcement
- `33917df` internal message chain and menu
- `1c8c27a` API docs first-click / Swagger UI access fix
- `dc09252` bootstrap accessor convergence
- `a0b7498` align admin bootstrap output to new generator
- `759fc67` generated-code inventory refresh
- `cba6fe1` manual HTTP and generated-data boundary alignment

### `admin-ui`

- `539623c27` login captcha and localized auth errors
- `44bd31d23` login hero image / auth branding unification
- `104fe25ba` internal-message notification + send page
- `3a2dd8ee1` message-page button permission and “view all” fix
- `cc46dd818` profile basic settings and password change wired to backend

### `xkit`

- `31cce6d` admin portal analytics proto sync
- `8af6bd4` xadmin example enum alignment
- `a95b52e` prior 20260510 handoff
- `c34fb85` bootstrap provider and hook generation boundary convergence
- `7092e2a` bootstrap service dependency injection fix
- `885016f` manual HTTP template sync boundary test

### `xkit-template`

- `65adc16` reusable server extension points
- `42ded86` bootstrap data-access and manual HTTP hook bridge
- `eb96cc4` tighten manual HTTP preserve boundary
