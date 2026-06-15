# Phase 1 Generate

This reference covers the part that `xkit/examples/generateAll.ps1` is supposed to handle.

## Collect inputs first

Ask for or confirm these values before generation:

- `ProjectName`
- `Module`
- `AppName`
- `TypeScriptRoot`
- `ConfigPath`
- `CanonicalConfigPath`

Recommended defaults:

- `Module = ProjectName`
- `TypeScriptRoot = D:\GoProjects\XAdmin\<ProjectName>\.generated-ui`
- `ConfigPath = D:\GoProjects\XAdmin\xkit\examples\admin\<ProjectName>-config\admin.yaml`
- `CanonicalConfigPath = D:\GoProjects\XAdmin\xkit\examples\admin\admin-config\admin.yaml`

When `ProjectName = admin`, use this collision-free default for `ConfigPath`:

- `D:\GoProjects\XAdmin\xkit\examples\admin\admin-target-config\admin.yaml`

`generateAll.ps1` now supports two modes:

- pass the values explicitly as parameters
- omit one or more of `ProjectName`, `Module`, `AppName`, `TypeScriptRoot`, `ConfigPath`, `CanonicalConfigPath` and let the script prompt for them interactively

## Canonical command template

```powershell
& 'D:\GoProjects\XAdmin\xkit\examples\generateAll.ps1' `
  -WorkspaceRoot 'D:\GoProjects\XAdmin' `
  -ProjectName '<ProjectName>' `
  -Module '<Module>' `
  -AppName '<AppName>' `
  -ServiceName 'admin' `
  -TypeScriptRoot '<TypeScriptRoot>' `
  -ConfigPath '<ConfigPath>' `
  -CanonicalConfigPath '<CanonicalConfigPath>' `
  -SkipGoGetUpdateAll
```

## Interactive usage

If you want the script to prompt for the four key values, you can run:

```powershell
& 'D:\GoProjects\XAdmin\xkit\examples\generateAll.ps1' `
  -WorkspaceRoot 'D:\GoProjects\XAdmin' `
  -ServiceName 'admin' `
  -SkipGoGetUpdateAll
```

The script will then prompt for:

- `ProjectName`
- `Module`
- `AppName`
- `TypeScriptRoot`
- `ConfigPath`
- `CanonicalConfigPath`

## Validated example from qadmin

```powershell
& 'D:\GoProjects\XAdmin\xkit\examples\generateAll.ps1' `
  -WorkspaceRoot 'D:\GoProjects\XAdmin' `
  -ProjectName 'qadmin' `
  -Module 'qadmin' `
  -AppName 'QAdmin' `
  -ServiceName 'admin' `
  -TypeScriptRoot 'D:\GoProjects\XAdmin\qadmin\.generated-ui' `
  -ConfigPath 'D:\GoProjects\XAdmin\xkit\examples\admin\qadmin-config\admin.yaml' `
  -CanonicalConfigPath 'D:\GoProjects\XAdmin\xkit\examples\admin\admin-config\admin.yaml' `
  -SkipGoGetUpdateAll
```

## Expected internal stages

The script is expected to do this sequence:

1. `go run ./cmd/xkit init template ...`
2. `go run ./cmd/xkit init source ...`
3. overwrite `ConfigPath` from `CanonicalConfigPath`
4. `buf generate --template buf.gen.yaml`
5. `buf generate --template buf.admin.openapi.gen.yaml`
6. `buf generate --template buf.vue.admin.typescript.gen.yaml`
7. `go mod tidy`
8. `go run -mod=mod entgo.io/ent/cmd/ent generate ./internal/data/ent/schema --feature privacy,sql/upsert,sql/versioned-migration,sql/modifier`
9. `go run ./cmd/xkit gen all admin --project <target> --config <target-config>`
10. `go get -u all`
11. `go mod tidy`
12. `go test ./...`

## Mandatory config verification

After `init source`, verify the target config:

- source: `CanonicalConfigPath`
- target: `ConfigPath`

Default paths are:

- source: `xkit/examples/admin/admin-config/admin.yaml`
- target: `xkit/examples/admin/<ProjectName>-config/admin.yaml`

When `ProjectName` is `admin`, the default collision-free target path is:

- target: `xkit/examples/admin/admin-target-config/admin.yaml`

Confirm:

- module line was rewritten to `<Module>`
- API import root was rewritten to `<Module>/api/gen/`

If not, repair the target config and rerun:

```powershell
Set-Location D:\GoProjects\XAdmin\xkit
go run ./cmd/xkit gen all admin `
  --project D:\GoProjects\XAdmin\<ProjectName> `
  --config <ConfigPath>
```
