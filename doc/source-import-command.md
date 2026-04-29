# Source Import Command

## Purpose

`xkit init source` imports raw schema and proto sources into a target project and derives the YAML config used by `xkit gen all`. It also normalizes Buf output paths for generated artifacts that are owned outside the Go package tree, such as OpenAPI and TypeScript clients.

This command exists because raw generation inputs may now live outside the active generated-code locations. For `xadmin-web`, the raw input directory is:

```text
D:\GoProjects\XAdmin\xadmin-web\source
```

## Command

```text
xkit init source <source-path> [--project <path>] [--service <name>] [--config <path>] [--typescript-project <path>] [--force] [--dry-run]
```

Example:

```text
xkit init source D:\GoProjects\XAdmin\xadmin-web\source \
  --project D:\GoProjects\XAdmin\xadmin-web \
  --service admin \
  --typescript-project D:\GoProjects\XAdmin\xadmin-web-ui
```

## Source Layout

Preferred layout:

```text
source/
  api/
    buf.yaml
    buf.gen.yaml
    README.md
    protos/
  schema/
```

Also supported:

- `source/protos`
- `source/data/schema`
- `source/internal/data/ent/schema`

## Target Layout

The command copies source files into the target project:

```text
api/*                         <- files directly under source api/
api/protos/                   <- source api/protos or protos
internal/data/ent/schema/     <- source schema, data/schema, or internal/data/ent/schema
```

When copying `buf*.gen.yaml` or `buf*.gen.yml`, `xkit init source` validates and forcibly rewrites Go package options under `managed.override`:

- active `go_package_prefix` becomes `<target-module>/api/gen`
- each `go_package` becomes `<target-module>/api/gen/<proto-path>;<go-package-name>`
- local proto package names that collide with generated Ent packages are version-suffixed, such as `permission/v1;permissionv1` and `task/v1;taskv1`
- OpenAPI plugin output paths are normalized to `../cmd/server/assets`, so `buf generate --template buf.admin.openapi.gen.yaml` writes the document where the template embeds it.
- TypeScript plugin output paths are normalized under the configured TypeScript project root. By default the root is a sibling of the Go project named `<project>-ui`, such as `D:\GoProjects\XAdmin\admin-02-ui`. Relative `--typescript-project` values are resolved beside the Go project.

Current TypeScript output convention:

```text
buf.vue.<service>.typescript.gen.yaml -> <typescript-project>/apps/<service>/src/generated/api
```

For example, `path: authentication/v1` in project `admin-01` is written as:

```yaml
value: admin-01/api/gen/authentication/v1;authentication
```

This correction is applied even when the target Buf generation YAML already exists, so stale module names copied from another project do not block `buf generate`.

When copying schema `.go` files, local `*/api/gen/<domain>/...` imports are also normalized to the target module if `<domain>` exists under the imported proto root. External imports, such as `github.com/chnxq/x-crud/api/gen/pagination/v1`, are preserved.

Existing files are skipped by default. Use `--force` to overwrite them. Use `--dry-run` to print the write plan without changing the project.

## Config Generation

By default, the generated config is written to:

```text
<source>/<project-name>-config/<service>.yaml
```

For `xadmin-web`:

```text
D:\GoProjects\XAdmin\xadmin-web\source\xadmin-web-config\admin.yaml
```

Use `--config <path>` to write it elsewhere.

## Resource Derivation

The generated config is derived from Ent schemas and proto services:

- resource name: Ent schema name converted to snake_case
- `entity`: Ent schema type name
- `proto_service`: matching `<Entity>Service`
- `dto_import`: inferred from proto request/response packages
- `dto_type`: Ent schema type name
- `repo_interface`: `<Entity>Repo`
- `filters.allow`: filterable Ent fields
- `exists_fields`: `*ExistsRequest.oneof query_by` fields when present
- `operations`: supported CRUD and special repo operations found in service methods

The command prefers admin-facing services such as:

```text
admin.service.v1.UserService
```

Domain services are kept only when they have currently supported repo operations. The primary example is:

```text
authentication.service.v1.UserCredentialService
```

Schemas without a matching service proto are skipped and listed in command output. This is expected for relation/detail schemas such as `UserRole`, `PermissionApi`, or `DictEntryI18n`.

## Follow-Up Generation

After import:

```text
cd D:\GoProjects\XAdmin\xadmin-web\api
buf generate --template buf.gen.yaml
buf generate --template buf.admin.openapi.gen.yaml
buf generate --template buf.vue.admin.typescript.gen.yaml

cd D:\GoProjects\XAdmin\xkit
xkit gen all admin \
  --project D:\GoProjects\XAdmin\xadmin-web \
  --config D:\GoProjects\XAdmin\xadmin-web\source\xadmin-web-config\admin.yaml
```

`init source` does not generate service/repo/register/bootstrap Go files. It only materializes raw source inputs and prepares the config consumed by the normal generator.
