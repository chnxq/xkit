---
name: xkit-helper
description: Bootstrap a new Go backend project from `xkit/examples/admin` and then align the generated project with the current `admin` repository's hand-written business details. Use when creating repos like `qadmin`, when `generateAll.ps1` needs parameter substitution for `ProjectName`, `Module`, `AppName`, or `TypeScriptRoot`, when canonical config must be verified after generation, or when a fresh generated project still differs from `admin` in preserved extension files and real runtime behavior.
---

# Xkit Helper

## Overview

Use this skill for the full backend bootstrap-and-repair workflow around `xkit`.

Keep the work in two explicit phases:

1. Phase 1: copy the static template, align config/api/schema inputs, and run the same generation path that `xkit/examples/generateAll.ps1` is supposed to cover.
2. Phase 2: align the newly generated project with the current `admin` repository's hand-written and extension-heavy business details.

Do not split these phases into separate skills unless they become independently reusable with weak coupling. Right now the second phase depends directly on artifacts and verification results produced by the first phase.

## Required inputs

Before running Phase 1, collect and substitute these values explicitly:

- `ProjectName`
- `Module`
- `AppName`
- `TypeScriptRoot`

Use these input prompts:

- `ProjectName`: target repo directory name, for example `qadmin`
- `Module`: Go module name, usually the same as `ProjectName` unless the user wants a different module path
- `AppName`: human-facing app name used by template/bootstrap metadata, for example `QAdmin`
- `TypeScriptRoot`: where generated TypeScript API output should land; prefer a target-local path such as `<target>\.generated-ui`

If the user does not specify `TypeScriptRoot`, default it to:

- `D:\GoProjects\XAdmin\<ProjectName>\.generated-ui`

## Phase 1: Generate from xkit

### 1. Confirm workspace paths

Identify these fixed roots first:

- Workspace root, usually `D:\GoProjects\XAdmin`
- Generator repo: `xkit`
- Template repo: `xkit-template`
- Example source: `xkit/examples/admin`
- Reference project: `admin`
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

## Phase 2: Align with admin hand-written details

### 1. Compare against `admin` at the preserved-file boundary

Do not assume fresh generated output matches the real runtime behavior of `admin`.

Compare the target project with `admin`, especially:

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
- auth, captcha, viewer auth, manual HTTP services, analytics, messages, or audit behaviors exist in `admin` but were not carried into the target repo

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
- whether canonical config replacement worked automatically or needed repair
- which hand-written files or extension files were synchronized from `admin`
- the final verification result

Do not leave the process only in chat.

## Working rules

- Treat `admin` as the behavior reference, not just the example source.
- Prefer fixing the generation chain when a defect is systemic, but still repair the target repo immediately so it becomes usable.
- Never assume template-owned files and preserved hand-written files follow the same update path.
- Before deleting or renaming helper functions in extension files, check whether generated code now defines the same helper.
- When copying files from `admin` to the target repo, rewrite module imports consistently from `admin/...` to `<Module>/...`.
- Keep the target repo buildable after every repair round.

## References

- Read `references/phase-1-generate.md` for the command template and the validated generation sequence.
- Read `references/phase-2-repair-hotspots.md` before diffing `admin` and the target repo.
