# Bootstrap Template and Generated Code Boundary

## Current Decision

Startup code is split into two ownership zones.

Template-owned code is copied by `xkit init template` and then treated as normal project code. It is created once, can be edited by hand, and should not be overwritten by `xkit gen all`.

Generated code is written by `xkit gen ...`. It must live in `*.gen.go` files or one-time extension files, so regeneration does not erase project-specific changes.

## Template-Owned Startup Skeleton

The template repository owns stable startup structure:

- `cmd/server/main.go`
- `cmd/server/server.go`
- `internal/bootstrap/app.go`
- `internal/bootstrap/infra.go`
- `internal/bootstrap/hooks.go`
- `internal/server/http.go`
- `internal/server/grpc.go`
- `internal/server/manual.go`
- `internal/data/bootstrap/data.go`
- `configs/*.yaml`
- startup assets under `cmd/server/assets`

These files are platform skeleton code. They define command entry, config loading, logger/registry/tracer initialization, transport construction, and handwritten hooks.

## Xkit-Generated Glue

`xkit gen bootstrap` is narrowed to dynamic glue:

- `internal/bootstrap/generated_servers.gen.go`
- `internal/data/bootstrap/ent_client.gen.go`

Resource generation remains in:

- `internal/service/*_service.gen.go`
- `internal/data/repo/*_repo.gen.go`
- `internal/server/rest_register.gen.go`
- `internal/server/grpc_register.gen.go`
- `internal/service/providers/wire_set.gen.go`
- `internal/data/providers/wire_set.gen.go`

The startup skeleton calls `NewGeneratedServers`, `RegisterGeneratedHTTPServices`, and `RegisterGeneratedGRPCServices`. Those functions are the generated boundary.

## Handwritten Extension Points

The following files are reserved for later handwritten changes and are created once only:

- `internal/service/*_service_ext.go`
- `internal/data/repo/*_repo_ext.go`
- `internal/bootstrap/hooks.go`
- `internal/server/manual.go`
- `internal/data/bootstrap/data.go`
- `configs/*.yaml`

`internal/bootstrap/hooks.go` is for extra transport servers or lifecycle additions.

`internal/server/manual.go` is for custom HTTP/gRPC registrations that are not derived from proto resources.

`internal/data/bootstrap/data.go` is for shared data providers such as Redis, cache, object storage, queues, or domain-specific clients.

## Template Init Flow

The intended target flow is:

```text
xkit init template <template-path> --project <target> ...
go get -u all
xkit gen all <service> --project <target> --config <config>
```

The `go get -u all` step is now part of `xkit init template` after real copying. It is skipped for `--dry-run` and can be disabled with `--skip-go-get-update-all` when working offline.

## Old `pkg` Reuse Assessment

Reference package path:

```text
D:\GoProjects\chnxq\xadmin\pkg
```

Do not copy this directory wholesale into the template. Several packages contain hard-coded `xadmin/...` imports or strong business assumptions.

Good candidates for later extraction after import cleanup:

- `crypto`: general AES-GCM helpers.
- `eventbus`: general event manager and middleware ideas.
- `utils`: small generic helpers.
- `oss`: reusable if the target template standardizes object storage.

Reference-only for now:

- `metadata`
- `middleware/auth`
- `authorizer`
- `middleware/logging`
- `lua`

These contain useful design, but they depend on generated API packages, auth payloads, policy models, or project-specific behavior. They should be adapted through the handwritten hook files instead of copied directly.

Keep project-specific:

- `constants`
- `jwt`
- `utils/converter`
- `task`
- `entgo/viewer`

These encode xadmin domain defaults or concrete business data.

