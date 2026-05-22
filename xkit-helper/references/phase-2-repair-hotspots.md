# Phase 2 Repair Hotspots

Use this checklist after Phase 1 if the new project still differs from the chosen reference implementation.

## 0. Resolve the reference project first

The repair phase must start from a local reference directory.

Allowed inputs:

- local path, for example `D:\GoProjects\XAdmin\admin`
- git repository URL

If a git URL is provided:

- clone it to a local working path first
- use the cloned directory as the reference repo root
- record the exact path and revision used during the repair session

For the rest of this document, call that directory `<ReferenceProjectRoot>`.

## 1. Preserved server files fell back to template stubs

### Highest-risk files

- `internal/server/manual_http_data.go`
- `internal/server/options.go`
- `internal/server/http_options.go`
- `internal/server/grpc_options.go`
- `internal/server/sse.go`

### Typical symptoms

- login, captcha, refresh token, profile, analytics, navigation, or SSE behavior is missing
- `internal/server` tests fail because manual types or helpers are gone
- message service cannot resolve shared SSE server

### Repair

Diff against `<ReferenceProjectRoot>` and bring over the current hand-written implementation, then rewrite imports from the reference module path to `<Module>/...`.

## 2. Extension helper names now collide with generated code

### Typical symptoms

- compile errors like `redeclared in this block`

### Confirmed examples from qadmin

- `permissionFieldMaskContains`
- `roleFieldMaskContains`

### Repair

Check whether generated code now defines the same helper. If it does, remove or rename the extension-side helper instead of patching unrelated generated files first.

## 3. Ent query types are missing `Modify(...)`

### Typical symptoms

- repo package build errors mentioning `missing method Modify`
- failures inside `BuildListSelectorWithPaging`

### Main repair file

- `internal/data/ent/query_modify_ext.go`

### Types that needed coverage in qadmin

- `ApiAuditLogQuery`
- `DataAccessAuditLogQuery`
- `DictEntryQuery`
- `DictTypeQuery`
- `FileQuery`
- `InternalMessageCategoryQuery`
- `LanguageQuery`
- `LoginAuditLogQuery`
- `LoginPolicyQuery`
- `OperationAuditLogQuery`
- `PermissionAuditLogQuery`
- `PolicyEvaluationLogQuery`
- `TaskQuery`
- `UserCredentialQuery`

Scan all generated repo files and cover every list-builder query type in one pass.

## 4. Hand-written business features commonly required from the reference project

Compare whether the target project also needs these groups from `<ReferenceProjectRoot>`:

- auth password verification and token handling
- captcha support
- viewer auth middleware
- manual HTTP services for login, admin portal, profile, analytics, and menu sync
- internal message repo/service extensions
- audit log repo wrappers and DB logging integration
- default data bootstrap extensions
- SQL default/demo data

## 5. Verification target

At minimum:

```powershell
go test ./...
```

The practical success condition is:

- target repo compiles
- bootstrap and core runtime paths are present
- target repo is behaviorally close enough to the resolved reference project for the intended iteration
- the real process is written into `<target>/readme.md`
