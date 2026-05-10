# XAdmin Context Handoff

Generated at: 2026-05-10 16:48 +08:00

This file is the new resume entry for the current workspace state.  
It inherits `xkit/NEXT_CONTEXT_HANDOFF_20260501.md` and
`xkit/NEXT_CONTEXT_HANDOFF_20260505.md`, and includes all major verified
changes through 2026-05-10.

## Scope And Incorporation

- This handoff **explicitly includes** facts from the current thread and recent
  implementation/verification work (permissions, logs, auth hardening,
  analytics backend integration, analytics zero-count fix, and enum alignment).

## Environment

- Workspace root: `D:\GoProjects\XAdmin`
- Shell: Windows PowerShell
- Timezone: `Asia/Shanghai`
- Main backend: `admin`
- Main frontend: `admin-ui` (Vben, branch `xadmin-api-integration`)
- Generator/tooling repos: `xkit`, `xkit-template`, `x-crud`, `x-utils`,
  `xkitmod`, `xkitpkg`

## Repository State As Of 2026-05-10

### `admin`

- Branch: `main`
- HEAD: `d25ac64 fix(portal): 修复分析页访问量统计与审计仓库包装透传`
- Working tree: clean
- Verified: `go test ./...` passed on 2026-05-10

### `admin-ui`

- Branch: `xadmin-api-integration`
- HEAD: `9d79a364c chore(@vben/locales): 调整登录页中文文案`
- Working tree: clean
- Verified: `pnpm -F @vben/web-antd run typecheck` passed on 2026-05-10

### `xkit`

- Branch: `main`
- HEAD: `8af6bd4 fix(xadmin-example): 对齐组织与岗位相关枚举定义`
- Working tree: clean

### `xkit-template`

- Branch: `main`
- HEAD: `8c136d6 update go mod`
- Working tree: clean

### Related Repos

- `x-utils`: `417cf92` (clean)
- `x-crud`: `a61448f` (clean)
- `xkitmod`: `698fedc` (clean)
- `xkitpkg`: `80e4b48` (clean)

## Runtime Status (Snapshot)

At handoff time, services are running:

- `5666`: `node.exe` (Vite dev server, `admin-ui`)
- `7788`: `admin.exe` REST
- `7789`: `admin.exe` SSE
- `7790`: `admin.exe` gRPC

Process snapshot:

- `admin.exe server -c ./configs`
- `node ... vite.js --mode development`

## Major Changes Since 2026-05-05

## 1. Frontend Admin UX, Permission Controls, And i18n

Key area:

- Permission page/table-page scroll behavior and access-code alignment
- Log pages (login/api/permission audit) added and iterated
- i18n expansion for table/common actions (refresh/export/column settings/etc.)
- Menu and page language keys aligned in CN/EN

Representative commits (`admin-ui`):

- `b7e57d5e3` fix: 修正分页滚动和权限控制
- `f058f8b76` feat: add audit log pages and expand admin i18n
- `3350fa4c8` feat: add permission audit log page and i18n updates
- `92d74022f` fix: correct permission sync button i18n label
- `5c6f4288f` fix: 补齐表格操作列文案键
- `5ed1fe550` feat: 日志页支持时间范围筛选
- `2c06a0d97` feat: 分析页改为后端聚合数据驱动

Current permission-page button access codes in UI are aligned to:

- `permission:groups:create|edit|delete`
- `permissions:sync:perms:create`
- `permissions:export`

## 2. Backend Auth, Logging, Audit, And Enum Consistency

Key outcomes (`admin`):

- Password verification hardened with bcrypt-aware verify/upgrade path
- JWT signing/validation strengthened and used consistently
- `/admin/v1/logout` in public path list to avoid noisy unauthorized spam in
  wrong-password flows
- Permission-audit required-field persistence regression covered by tests
- Log list `created_at` filter normalization and audit-log sort support
- Proto/schema enum alignment fixes integrated

Representative commits:

- `8b3c8bf` auth/log hardening and logout whitelist
- `38ac7dd` test(repo): permission audit required fields regression tests
- `d71d7c5` fix(ent): proto/schema enum alignment
- `a1aeb98` fix(audit): normalize created_at filters

## 3. Analytics Dashboard Real-Data Integration + Zero-Count Root Fix

What was implemented:

- Backend portal API expanded to return aggregated analytics data.
- Frontend analytics page switched from static/demo data to backend-driven data.
- Stats source is API audit log aggregation path.

Critical fix (root cause of "访问量=0"):

- `WrapAuditLogRepos()` wrapper previously hid extension capabilities
  (`Write*AuditLog`, `AnalyticsSummary`) after wrapping.
- Result: wrapped repos could lose audit-write and analytics-read behavior,
  causing dashboard numbers to stay zero.
- Fixed by explicitly delegating:
  - `WriteApiAuditLog`
  - `WriteLoginAuditLog`
  - `WritePermissionAuditLog`
  - `AnalyticsSummary`
- Added wrapper regression tests to prevent recurrence.

Additional analytics correction:

- `totalUsages` now uses **all-time distinct user_id** instead of mirroring
  total access count.

Relevant files:

- `admin/internal/server/manual_http.go`
- `admin/internal/data/repo/api_audit_log_repo_ext.go`
- `admin/internal/data/repo/audit_log_repo_wrappers_ext.go`
- `admin/internal/data/repo/audit_log_repo_wrappers_ext_test.go`
- `admin-ui/apps/web-antd/src/views/dashboard/analytics/index.vue`
- `admin-ui/apps/web-antd/src/api/admin/portal.ts`

Important clarification:

- Analytics page is **not SSE-driven**.
- Current analytics page fetches once on mounted via
  `GET /admin/v1/dashboard/analytics`.
- SSE capability exists in system but is unrelated to this page.

## 4. xkit Example Sync Updates

`xkit` updates include:

- `31cce6d` sync admin portal analytics proto
- `8af6bd4` align xadmin example enums:
  - `OrgUnit.type`: add `SUBSIDIARY`, `BRANCH`
  - `Position.type`: `LEAD` -> `LEADER`
  - position status: add `RESIGNED` in related schemas
  - `examples/xadmin/go.mod` and `go.sum` synced

## Verified Commands (Latest)

Backend:

```powershell
cd D:\GoProjects\XAdmin\admin
go test ./...
```

Frontend:

```powershell
cd D:\GoProjects\XAdmin\admin-ui
pnpm -F @vben/web-antd run typecheck
```

Both passed on 2026-05-10.

## Current Behavior Notes

- `server.rest.enable_db_logging` is enabled in `admin/configs/server.yaml`.
- Analytics metrics depend on records in `sys_api_audit_logs`.
- If no logs are written (or wrong runtime config), access/usage metrics can be
  low or zero; this is data-state related, not SSE related.
- Current auth seed still uses default password `123456` in local `admin`
  bootstrap source:
  `admin/internal/data/bootstrap/default_data_ext.go`.

## Risks / Follow-Up Items

- Dashboard correctness still depends on ongoing audit-log write health and
  deployment config parity.
- API sync behavior is still delete-and-recreate oriented; if downstream logic
  starts depending on stable API IDs, an upsert strategy should be introduced.
- Default seed password policy may need hardening for non-dev deployments.
- Continue promoting stable manual backend enhancements into
  `xkit`/`xkit-template` to reduce future regeneration drift.

## Recommended Resume Sequence

1. Read this file, then `git status` in `admin`, `admin-ui`, `xkit`.
2. Confirm services needed for your task:
   - backend `7788/7789/7790`
   - frontend `5666`
3. Validate key flows:
   - login (correct/incorrect password paths)
   - permission-based button visibility
   - permission page layout/scroll behavior
   - login/api/permission log pages (time-range filters)
   - analytics page numbers (ensure they change with real access traffic)
4. If analytics looks wrong, check in order:
   - `server.rest.enable_db_logging`
   - `sys_api_audit_logs` row growth
   - wrapper delegation path in
     `audit_log_repo_wrappers_ext.go`
5. Continue generator/template convergence for durable fixes.

## Commit Landmarks Since 2026-05-05

### `admin`

- `cfb2213` fix: audit sort + localized menu titles
- `8b3c8bf` auth/password/token/logout whitelist hardening
- `38ac7dd` permission-audit required-field regression tests
- `d71d7c5` proto/schema enum alignment
- `a1aeb98` audit created_at filter normalization
- `97000cf` analytics backend real-data aggregation
- `d25ac64` analytics zero-count root fix + wrapper delegation tests

### `admin-ui`

- `b7e57d5e3` pagination/permission-control fixes
- `f058f8b76` log pages + broad i18n expansion
- `3350fa4c8` permission-audit log page + i18n updates
- `92d74022f` permission sync i18n fix
- `5c6f4288f` table operation i18n keys fix
- `5ed1fe550` log time-range filtering
- `2c06a0d97` analytics page switched to backend data
- `9d79a364c` auth page zh-CN text adjustment

### `xkit`

- `e10a858` previous handoff整理
- `31cce6d` analytics portal proto sync
- `8af6bd4` enum alignment in xadmin example schemas
