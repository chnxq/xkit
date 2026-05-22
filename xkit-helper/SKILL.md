---
name: xkit-helper
description: Bootstrap a new Go backend project from `xkit/examples/admin`, align the generated project with a reference backend repository's hand-written business details, and then review the remaining differences from a feature-first perspective. Use when creating repos like `qadmin`, when `generateAll.ps1` needs parameter substitution for `ProjectName`, `Module`, `AppName`, or `TypeScriptRoot`, when canonical config must be verified after generation, or when a fresh generated project still differs from the chosen reference project in preserved extension files and real runtime behavior. The reference project may be provided as a local path or as a git repository URL.
---

# Xkit Helper

## Overview

Use this skill for the full backend bootstrap, repair, and gap-review workflow around `xkit`.

Keep the work in three explicit phases:

1. Phase 1: copy the static template, align config/api/schema inputs, and run the same generation path that `xkit/examples/generateAll.ps1` is supposed to cover.
2. Phase 2: align the newly generated project with the chosen reference repository's hand-written and extension-heavy business details.
3. Phase 3: compare the generated target and the reference project again, but this time organize the result around functional differences, missing behavior, and operational gaps instead of raw file diffs.

Do not split these phases into separate skills unless they become independently reusable with weak coupling. Right now the later phases depend directly on artifacts and verification results produced by the earlier phases.

## Required inputs

Before running Phase 1, collect and substitute these values explicitly:

- `ProjectName`
- `Module`
- `AppName`
- `TypeScriptRoot`
- `ReferenceSource`

Use these input prompts:

- `ProjectName`: target repo directory name, for example `qadmin`
- `Module`: Go module name, usually the same as `ProjectName` unless the user wants a different module path
- `AppName`: human-facing app name used by template/bootstrap metadata, for example `QAdmin`
- `TypeScriptRoot`: where generated TypeScript API output should land; prefer a target-local path such as `<target>\.generated-ui`
- `ReferenceSource`: the backend project used for Phase 2 comparison; allow either:
  - a local path, for example `D:\GoProjects\XAdmin\admin`
  - a git repository URL, for example `https://...` or `git@...`

If the user does not specify `TypeScriptRoot`, default it to:

- `D:\GoProjects\XAdmin\<ProjectName>\.generated-ui`

Handle `ReferenceSource` like this:

- if the user gives a local path, use it directly
- if the user gives a git repository URL, clone it to a temporary or user-approved local working path before Phase 2
- if the user gives neither, ask explicitly; do not assume `admin` unless the user agrees that it is the intended default reference

## Phase 1: Generate from xkit

### 1. Confirm workspace paths

Identify these fixed roots first:

- Workspace root, usually `D:\GoProjects\XAdmin`
- Generator repo: `xkit`
- Template repo: `xkit-template`
- Example source: `xkit/examples/admin`
- Target project root: `<WorkspaceRoot>\<ProjectName>`

### 2. Substitute the four user inputs into the generation command

Run the command pattern in `references/phase-1-generate.md`.

This phase is responsible for:

- copying the static template from `xkit-template`
- importing example source assets from `xkit/examples/admin`
- generating target-specific config
- generating Go API, OpenAPI, TypeScript, Ent, and xkit dynamic code
- running `go test ./...`

### 3. Verify canonical config replacement happened

The critical canonical config is:

- `xkit/examples/admin/admin-config/admin.yaml`

The target-specific generated config is:

- `xkit/examples/admin/<ProjectName>-config/admin.yaml`

After `init source`, verify that the target config was overwritten from the canonical config and that these replacements occurred:

- `module: admin` -> `module: <Module>`
- `admin/api/gen/` -> `<Module>/api/gen/`

If this did not happen, Phase 1 is incomplete even if files were generated.

### 4. Re-run dynamic generation if config was repaired

If the config had to be corrected after the initial bootstrap, rerun:

```powershell
go run ./cmd/xkit gen all admin --project <target-project> --config <target-config>
```

Then re-run:

```powershell
go test ./...
```

## Phase 2: Align with reference hand-written details

### 0. Resolve the reference project

Before diffing or copying any hand-written logic, resolve `ReferenceSource`.

If `ReferenceSource` is a local path:

- verify the path exists
- use it directly as the reference repo root

If `ReferenceSource` is a git repository URL:

- clone it to a local working directory
- use the cloned directory as the reference repo root
- record the clone path and revision used during the repair session

For the rest of Phase 2, treat the resolved local directory as `<ReferenceProjectRoot>`.

### 1. Compare against `admin` at the preserved-file boundary

Do not assume fresh generated output matches the real runtime behavior of the reference project.

Compare the target project with `<ReferenceProjectRoot>`, especially:

- `internal/server`
- `internal/service/*_ext.go`
- `internal/data/repo/*_ext.go`
- `internal/data/ent/query_modify_ext.go`
- `internal/bootstrap/*_ext.go`
- `internal/data/bootstrap/default_data_ext.go`
- `sql/*.sql`

Use `references/phase-2-repair-hotspots.md`.

### 2. Repair the common mismatch patterns

The main patterns seen so far are:

- canonical config was not applied, so generated wiring is incomplete
- preserved hand-written files regressed to template stubs
- generated code now defines helpers that collide with older extension helpers
- generated repos require additional Ent `Modify(...)` methods
- auth, captcha, viewer auth, manual HTTP services, analytics, messages, or audit behaviors exist in the reference project but were not carried into the target repo

### 3. Re-verify after each repair round

At minimum, rerun:

```powershell
go test ./...
```

If startup behavior matters to the current task, optionally run:

```powershell
go run ./cmd/server server -config_path ./configs
```

### 4. Record the actual process in the target repo

Write `<target>/readme.md` in Chinese with:

- the exact command sequence actually used
- how `ProjectName`, `Module`, `AppName`, and `TypeScriptRoot` were filled
- how `ReferenceSource` was filled and resolved
- whether canonical config replacement worked automatically or needed repair
- which hand-written files or extension files were synchronized from the reference project
- the final verification result

Do not leave the process only in chat.

## Phase 3: Review remaining differences by feature

### 1. Re-scan the target against the resolved reference project

After Phase 2 repairs are complete, compare the target repo with `<ReferenceProjectRoot>` again.

Do not lead with "which files still differ". Lead with:

- which user-visible features are still absent
- which runtime behaviors still differ
- which data, config, auth, logging, messaging, analytics, or admin workflows are still incomplete
- which differences are intentional template/generator boundaries and should stay different

Use `references/phase-3-feature-gap-review.md`.

### 2. Group the result by functional area

Prefer grouping by areas like:

- authentication and session
- captcha and security hardening
- viewer auth and permission checks
- menus, portal, and profile
- audit logs and analytics
- internal messaging and SSE
- default data and bootstrap behavior
- SQL assets and operational scripts

For each area, state:

- current target status
- reference behavior
- whether the gap blocks delivery, is optional, or is an intentional divergence
- which files likely own the gap

### 3. Produce an actionable gap summary

The output of Phase 3 should help decide the next iteration, not just describe diffs.

At minimum, produce:

- a list of missing or divergent features
- a list of risky runtime inconsistencies
- a list of acceptable intentional differences
- a prioritized next-step recommendation

### 4. Persist the result into the target repo

Phase 3 is not complete until the difference summary is written into a real file inside the target repo.

Use this default path:

- `<target>/docs/xkit-helper-feature-gap-review.md`

Write that file in Chinese and include:

- review scope and date
- resolved `ReferenceSource` path or clone path, plus revision when available
- what was checked in Phase 3
- which remaining gaps were found
- which gaps were intentionally left unresolved
- which differences belong to template/generator boundaries
- what should be implemented next

If the target repo already has a process log in `<target>/readme.md`, add a short pointer there to the Phase 3 review file, but keep the full difference description in the dedicated document.

## Working rules

- Treat the resolved reference project as the behavior reference, not just the example source.
- Prefer fixing the generation chain when a defect is systemic, but still repair the target repo immediately so it becomes usable.
- Never assume template-owned files and preserved hand-written files follow the same update path.
- Before deleting or renaming helper functions in extension files, check whether generated code now defines the same helper.
- When copying files from the reference project to the target repo, rewrite module imports consistently from the reference module path to `<Module>/...`.
- In Phase 3, do not stop at file-level differences; convert them into feature-level conclusions.
- In Phase 3, do not leave the gap summary only in chat; write it into the target repo file.
- Keep the target repo buildable after every repair round.

## References

- Read `references/phase-1-generate.md` for the command template and the validated generation sequence.
- Read `references/phase-2-repair-hotspots.md` before diffing `admin` and the target repo.
- Read `references/phase-3-feature-gap-review.md` when summarizing the remaining differences after repair.
