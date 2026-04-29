# Source Import Command

## Purpose

`xkit init source` imports raw schema and proto sources into a target project and derives the YAML config used by `xkit gen all`.

This command exists because raw generation inputs may now live outside the active generated-code locations. For `xadmin-web`, the raw input directory is:

```text
D:\GoProjects\XAdmin\xadmin-web\source
```

## Command

```text
xkit init source <source-path> [--project <path>] [--service <name>] [--config <path>] [--force] [--dry-run]
```

Example:

```text
xkit init source D:\GoProjects\XAdmin\xadmin-web\source \
  --project D:\GoProjects\XAdmin\xadmin-web \
  --service admin
```

## Source Layout

Preferred layout:

```text
source/
  api/protos/
  schema/
```

Also supported:

- `source/protos`
- `source/data/schema`
- `source/internal/data/ent/schema`

## Target Layout

The command copies source files into the target project:

```text
api/protos/                  <- source api/protos or protos
internal/data/ent/schema/    <- source schema, data/schema, or internal/data/ent/schema
```

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
xkit gen all admin \
  --project D:\GoProjects\XAdmin\xadmin-web \
  --config D:\GoProjects\XAdmin\xadmin-web\source\xadmin-web-config\admin.yaml
```

`init source` does not generate service/repo/register/wire Go files. It only materializes raw source inputs and prepares the config consumed by the normal generator.
