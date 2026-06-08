# XAdmin Context Handoff

Generated at: 2026-06-08 00:00 +08:00

This file is the latest resume entry for the current workspace state.
It should be treated as the primary handoff after:

- `xkit/NEXT_CONTEXT_HANDOFF_20260501.md`
- `xkit/NEXT_CONTEXT_HANDOFF_20260505.md`
- `xkit/NEXT_CONTEXT_HANDOFF_20260510.md`
- `xkit/NEXT_CONTEXT_HANDOFF_20260522.md`
- `xkit/NEXT_CONTEXT_HANDOFF_20260530.md`

This handoff explicitly incorporates the current thread's work, including
the task runtime refactor round, task schema/config regeneration alignment,
task frontend delivery progress, and the default data seed flow correction.

## Scope And Incorporation

- This handoff explicitly includes facts from the current thread.
- It covers:
  - `admin` task runtime layering and bootstrap wiring refactor
  - `xkit` example/config alignment for the current task model
  - task-related backend fixes after live runtime verification
  - task frontend delivery status in `admin-ui`
  - `default_data_ext.go` seed-flow correction and its new semantics
  - current repository state and latest validated commits

## Environment

- Workspace root: `D:\GoProjects\XAdmin`
- Shell: Windows PowerShell
- Timezone: `Asia/Shanghai`
- Main backend: `admin`
- Main frontend: `admin-ui`
- Generator/tooling repo: `xkit`

## Repository State As Of 2026-06-08

### `admin`

- Branch: `chnxq/dev`
- HEAD: `8bcb2ac fix default data seed flow`
- Working tree: clean

Recent backend landmarks:

- `5a1a0f3` `feat(task): add runtime scheduler and executor registry`
- `5ff38e2` `feat(task): add executor conventions and demo fixtures`
- `674a11e` `fix: fit tesk initial data`
- `865177b` `fix: support generated time filters in repos`
- `5c76297` `fix: add created_at filters for audit log repos`
- `aedb2ad` `fix: Delete duplicate logical code`
- `bbf2647` task decoupling design draft
- `cbecd49` `refactor task runtime wiring`
- `5edc326` `add task runtime regression tests`
- `5f70eee` `fix task runtime loading and failure logging`
- `9800ef5` fix runtime repo calls missing ViewerContext
- `8bcb2ac` `fix default data seed flow`

### `admin-ui`

- Branch: `xadmin-api-integration`
- HEAD: `4dd6c669 fix: correct paging query encoding for filters`
- Working tree: clean

Recent frontend landmarks in this round:

- task management page delivered and iterated repeatedly
- task log page aligned toward audit-log style
- multiple management pages migrated back to `VbenVxeGrid`
- toolbar/button placement unified across migrated pages
- cron editor introduced and then restyled toward the provided reference
- task log now resolves task name by `taskId`

Representative recent commits:

- `6cab61af` `refactor(ui): show task name in task log`
- `4969cc65` `fix(ui): resolve task log names by task id`
- `a222eee2` `fix(ui): repair permission group and task log display`
- `e2f99d6e` `feat(task): improve cron editor and task form`
- `78c1cb2d` `fix(task): normalize cron editor to 6 fields`
- `20406722` `style(task): format cron editor`
- `4dd6c669` `fix: correct paging query encoding for filters`

### `xkit`

- Branch: `main`
- HEAD: `3f537c7 align task example config with admin runtime`
- Working tree: clean

Recent generator landmarks:

- `be97b43` `fix(codegen): preserve handwritten bootstrap ext hooks`
- `f31b0a6` `feat:change defin of task data stucture and api`
- `6d6dd3d` fix generator should not rewrite files when only generated timestamp changes
- `73e5764` `fix: auto append created_at filters for log resources`
- `5eaa1da` `enhance xkit service generation config`
- `3f537c7` `align task example config with admin runtime`

## 1. Task Model And Generator Alignment

The task model was reworked around:

- `task_group`
- `task`
- `task_log`

Current backend-aligned structure:

- `task.group_id` is required
- `task.status` is retained
- `is_enabled` is removed
- `task.group_name` is not stored redundantly on `task`
- `task_log.job_id` was renamed to `task_id`
- scheduling is normalized to cron expression only

The `xkit/examples/admin` example and target config were updated to match the
current `admin` runtime/service structure instead of the earlier interim
variants.

Important consequence:

- future generation for task-related service fields must follow the current
  runtime injection model, not the earlier direct repo-field model

Relevant `xkit` commits:

- `5eaa1da`
- `3f537c7`

## 2. Current Task Runtime Architecture In `admin`

The task runtime layering has been intentionally restructured and documented.

Authoritative documents:

- `admin/docs/task-runtime-architecture.md`
- `admin/docs/task-decoupling-design.md`
- `admin/docs/task-executor-convention.md`

Current intended layering:

- `admin/internal/task/runtime`
  - stable runtime contracts
  - `Registry`
  - `Runner`
  - `Scheduler`
- `admin/internal/task`
  - task-domain assembly glue
  - loader
  - runtime store adapter
- `admin/internal/task/tasks/<name>`
  - one concrete task per directory
  - task-specific executor/factory/tests
- `admin/internal/bootstrap`
  - runtime bootstrap entry
  - service binding and scheduler startup wiring
- `admin/internal/service`
  - consume already-wired `Runner` / `Scheduler`
  - no direct ownership of task-specific business repo dependencies

Current sample tasks:

- `auditlogcleanup`
- `taskruntimesummary`
- `echo`

### What Was Explicitly Removed

The following old directions are now considered deprecated and should not be
restored casually:

- `RegisterRuntimeDeps(...)`-style service-level dependency plumbing
- building task executors/registry directly inside service constructors
- letting task platform code know concrete task business repos

## 3. Current Task Runtime Wiring Chain

Current startup chain:

1. `admin/internal/bootstrap/generated_hooks_ext.go`
2. `configureTaskRuntime(...)`
3. `admin/internal/bootstrap/task_runtime_ext.go`
4. `task.NewRuntimeBundleFromRepos(...)`
5. `service.BindTaskServices(...)`
6. app startup path calls `registerTaskRuntime(...)`
7. `service.RegisterTaskScheduler(...)`
8. `scheduler.Start()` and `scheduler.RestoreTasks(ctx)`

This chain is the current stable resume point.

If future refactor continues, use this chain as the source of truth rather
than the older service-local task runtime setup attempts.

## 4. Runtime Bugs Found During Live Verification And Their Fixes

Several bugs only surfaced after real startup / scheduled execution checks.

### 4.1 Missing ViewerContext When Runtime Loaded Task

Observed symptom:

- scheduled runtime path failed with:
  - `security: missing ViewerContext in context`

Root cause:

- scheduler-triggered runtime task reload used repo access that still expected
  business-side viewer context

Fix:

- added runtime-specific repo loading entry
- runtime store prefers the runtime-safe repo entry

Main commit:

- `5f70eee` `fix task runtime loading and failure logging`

### 4.2 Scheduled Pre-run Failure Did Not Write `task_log`

Observed symptom:

- task execution failed before `Runner.RunTask(...)`
- no execution log was written

Fix:

- `Runner` gained failure-log recording capability
- scheduler now records failure log even when pre-run reload fails

Main commit:

- `5f70eee`

### 4.3 Runtime Repo Calls Missing ViewerContext In Cleanup / TaskLog Paths

Observed symptom:

- cleanup task execution failed with:
  - `cleanup api audit logs failed: missing ViewerContext in context`
  - `write task log failed: missing ViewerContext in context`

Root cause:

- some runtime-only repo methods still executed on ordinary repo context

Fix:

- runtime viewer context was added to:
  - audit log cleanup repo paths
  - task log write path

Main commit:

- `9800ef5`

### 4.4 Invalid Cron Should Not Block Service Startup

Observed live behavior:

- a malformed cron expression in a task row produced restore errors
- service startup should continue while skipping the single bad task

Current state:

- runtime restore path now logs and skips the bad task instead of blocking the
  process startup

Practical note:

- frontend cron editor and backend runtime now both expect 6-field cron
  expressions for the current implementation

## 5. Tests Added For Task Runtime Refactor

This round finally restored meaningful backend regression coverage for task
runtime behavior.

Added tests include:

- `admin/internal/service/task_bootstrap_ext_test.go`
- `admin/internal/task/loader_test.go`
- `admin/internal/task/runtime/runner_test.go`
- `admin/internal/task/runtime/scheduler_test.go`
- `admin/internal/task/runtime_store_test.go`

Validated command during this round:

```bash
go test ./internal/task/... ./internal/service/...
```

This passed after the runtime fixes.

Additional bootstrap validation also passed during the later seed-flow work:

```bash
go test ./internal/data/bootstrap/...
```

Not yet established in this round:

- full `go test ./...`
- database-backed integration tests for task scheduling with real infra

## 6. Default Data Seed Flow Was Incorrect And Was Reworked

This was an important late discovery in the thread.

### Previous Wrong Behavior

`admin/internal/data/bootstrap/default_data_ext.go` previously mixed together:

- resource sync
- default seed creation
- existing-data reconciliation

This caused several problems:

1. `ensureTaskSeeds(ctx)` ran before any existing-data decision.
   Result: task seeds were always refreshed.

2. `hasExternalSeedData()` only checked a small subset such as tenant/user and
   did not cover the whole seed domain.

3. `reconcileExistingSeedData()` did not merely detect and skip.
   It actively called role/user relation repair paths and still rewrote data.

4. `ensurePlatformSuperRole()` was called redundantly.

Actual observed outcome:

- regardless of whether relevant seed data already existed, startup kept
  refreshing default seed data again.

### Current Corrected Model

The file was simplified toward:

- `syncResources(ctx)` always runs
- business default seed runs only for empty seed-domain state

New gating logic:

- `shouldSeedDefaultData(ctx)` checks whether any core seed-domain table
  already has data

Current checks include:

- `tenant`
- `org_unit`
- `position`
- `role`
- `user`
- `membership`
- `task_group`
- `task`

If any of them already contain records:

- default seed is skipped
- resource synchronization still runs

This means current semantics are:

- sync menus/apis/permissions every startup
- do not keep reapplying default tenant/user/role/task business seeds after the
  system has already been initialized

Main backend commit:

- `8bcb2ac` `fix default data seed flow`

## 7. Current Frontend Task Status

The task frontend is materially advanced compared with the earlier empty-shell
stage.

Delivered or substantially iterated:

- task scheduling top-level menu
- merged task-group + task-management page structure
- task log page
- toolbar normalization
- cron editor integration
- cron natural-language display refinement
- task log task-name resolution by `taskId`

Current practical state:

- the task frontend is no longer at scaffolding stage
- it already passed multiple rounds of manual testing
- remaining work should focus on refinement and behavior verification, not
  redoing the page skeleton

Important note:

- several UI refinements were done interactively and incrementally; when
  revisiting task UI, continue from current code rather than from the original
  plan doc alone

Authoritative feature plan:

- `admin/docs/task-management-development-plan.md`

## 8. Time-Range Filter Work

Another thread in this round addressed generated list filter behavior for time
range queries.

Relevant facts:

- generated repo filters needed support for created-at style range conditions
- this was partly pushed into `xkit`
- but there was also live-project reconciliation in `admin`

Relevant commits:

- `73e5764` in `xkit`
- `865177b`
- `5c76297`

Important caution:

- this area had at least one round of refactor that was judged structurally
  unsatisfactory and partly rolled back at the `admin` side
- future work here should prefer generator-level consistency, but not at the
  cost of forcing unnatural service coupling

## 9. Important Files To Reopen First In A New Context

If resuming task/backend work, reopen these first:

- `admin/docs/task-runtime-architecture.md`
- `admin/docs/task-decoupling-design.md`
- `admin/docs/task-management-development-plan.md`
- `admin/internal/bootstrap/task_runtime_ext.go`
- `admin/internal/task/loader.go`
- `admin/internal/task/runtime_store.go`
- `admin/internal/task/runtime/runner.go`
- `admin/internal/task/runtime/scheduler.go`
- `admin/internal/task/tasks/auditlogcleanup/*`
- `admin/internal/task/tasks/taskruntimesummary/*`
- `admin/internal/data/bootstrap/default_data_ext.go`

If resuming generator alignment work, reopen:

- `xkit/examples/admin/admin-config/admin.yaml`
- `xkit/examples/admin/admin-target-config/admin.yaml`

If resuming frontend task work, reopen:

- task management page under `admin-ui/apps/web-antd/src/views`
- task log page under the same area
- relevant cron editor component files introduced in this round

## 10. Resume Guidance

When opening the next context, the safest resume order is:

1. Read this file.
2. Read `admin/docs/task-runtime-architecture.md`.
3. Inspect current `admin/internal/task` and `admin/internal/bootstrap/task_runtime_ext.go`.
4. Confirm current `admin` HEAD is still based on `8bcb2ac`.
5. Only then continue with:
   - new task executor work
   - further runtime decoupling
   - seed-data semantics changes
   - task frontend refinements

## 11. Suggested Next Backend Work

The most reasonable backend-next items after this handoff are:

- continue executor decoupling patterns using the current runtime layering
- add real integration verification around:
  - restore runnable tasks on startup
  - scheduled execution
  - failure log persistence
  - invalid cron skip behavior
- decide whether default task seeds should stay "only on empty seed domain" or
  later evolve into a more explicit versioned seed migration mechanism

Avoid immediately doing:

- another large restructure of `internal/task` package boundaries
- restoring service-level task dependency injection
- reintroducing reconciliation paths that silently mutate existing business
  seed data on every startup

## 12. Suggested Next Frontend Work

The next frontend work should be incremental rather than architectural:

- keep task pages on the current `VbenVxeGrid` direction
- continue i18n cleanup
- continue cron editor UX polish only from the current implementation
- verify export/modal/list behavior against the now-preferred unified patterns

Do not restart from the old "empty UI shell first" assumption.

## 13. Verified Current Workspace Summary

As of this handoff:

- `admin` is clean and currently at `8bcb2ac`
- `xkit` is clean and currently at `3f537c7`
- `admin-ui` is clean and currently at `4dd6c669`

This handoff should now be treated as the primary resume file for the current
task/runtime/default-seed stage.
