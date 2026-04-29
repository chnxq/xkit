# Bootstrap Template and Generated Code Boundary

## Current Decision

Startup code is split into two ownership zones.

Template-owned code is copied by `xkit init template` and then treated as normal project code. It is created once, can be edited by hand, and should not be overwritten by `xkit gen all`.

Generated code is written by `xkit gen ...`. It must live in `*.gen.go` files or one-time extension files, so regeneration does not erase project-specific changes.

In this repository, "static" means code that becomes project-owned after initialization and may change with product requirements. "Dynamic" means code that can be mechanically derived from proto, generated API bindings, Ent schema, or xkit generation config.

## Template-Owned Startup Skeleton

The template repository owns stable startup structure:

- `cmd/server/main.go`
- `cmd/server/server.go`
- `internal/bootstrap/app.go`
- `internal/bootstrap/cleanup.go`
- `internal/bootstrap/factories.go`
- `internal/bootstrap/infra.go`
- `internal/bootstrap/hooks.go`
- `internal/server/asynq.go`
- `internal/server/http.go`
- `internal/server/grpc.go`
- `internal/server/http_options.go`
- `internal/server/grpc_options.go`
- `internal/server/asynq_options.go`
- `internal/server/sse_options.go`
- `internal/server/transport_config.go`
- `internal/server/options.go`
- `internal/server/sse.go`
- `internal/server/tls.go`
- `internal/data/bootstrap/data.go`
- `internal/data/bootstrap/resources.go`
- `configs/*.yaml`
- startup assets under `cmd/server/assets`

These files are platform skeleton code. They define command entry, config loading, logger/registry/tracer initialization, transport construction, and handwritten hooks.

Asynq and SSE are also treated as template-owned optional transports. The template builds them from `server.asynq` and `server.sse` config when present. Concrete task subscribers and domain-specific SSE handlers stay in handwritten hooks.

They are copied by `xkit init template`, not produced by `xkit gen all`.

## Transport Layout

All server lifecycle is delegated to `github.com/chnxq/xkitpkg/transport` packages:

- `transport/http` for REST/HTTP server lifecycle, generated REST route registration, manual routes, configured CORS filters, and pprof handlers.
- `transport/grpc` for gRPC server lifecycle and generated gRPC service registration.
- `transport/asynq` for task server/client/scheduler lifecycle, task publication, and subscriber registration.
- `transport/sse` for SSE lifecycle, stream management, and publish/notify APIs.

The template only maps project configuration and extension hooks onto those public transport options:

- `http_options.go` maps `server.rest` network/address/timeout/TLS/CORS/pprof and common middleware flags.
- `grpc_options.go` maps `server.grpc` network/address/timeout/TLS and common middleware flags.
- `asynq_options.go` maps `server.asynq` Redis, codec, queue, scheduler, shutdown, and TLS.
- `sse_options.go` maps `server.sse` network/address/path/codec/TLS/event options and default connection logging.
- `transport_config.go` contains shared config helpers and common middleware mapping.
- `options.go` is intentionally small and preserved. It is the project-owned hook file for business HTTP/gRPC middleware only.

`xkitpkg/server_utils` remains a useful reference for config-to-server mapping, but the template keeps direct `transport/*` construction so HTTP/gRPC/Asynq/SSE follow the same extension shape and projects can override options without replacing the constructor.

## Xkit-Generated Glue

`xkit gen bootstrap` is narrowed to dynamic glue:

- `internal/bootstrap/generated_servers.gen.go`
- `internal/data/bootstrap/ent_client.gen.go`

Resource generation remains in:

- `internal/service/*_service.gen.go`
- `internal/data/repo/*_repo.gen.go`
- `internal/server/rest_register.gen.go`
- `internal/server/grpc_register.gen.go`

The startup skeleton calls `NewGeneratedServers`, `RegisterGeneratedHTTPServices`, and `RegisterGeneratedGRPCServices`. Those functions are the generated boundary.

`internal/bootstrap/generated_servers.gen.go` is deliberately structured in small layers:

- `GeneratedData` owns generated repository construction and the Ent client cleanup returned by `NewEntClient`.
- `GeneratedServices` owns generated service construction.
- `GeneratedComponents` groups data and service objects.
- `GeneratedComponents.Servers` converts generated services into HTTP/gRPC transport servers.

This keeps template-owned startup code readable while allowing `xkit gen all` to replace the dynamic assembly safely.

`xkit gen bootstrap` does not write `configs/*.yaml`, `cmd/server/*`, `internal/bootstrap/app.go`, `internal/bootstrap/infra.go`, `internal/server/http.go`, or `internal/server/grpc.go`.

Wire provider sets are no longer part of the default `xkit gen all` flow. The current startup path uses explicit generated assembly in `internal/bootstrap/generated_servers.gen.go`, so `xkit gen all` does not create `internal/service/providers/wire_set.gen.go` or `internal/data/providers/wire_set.gen.go`. Existing Wire files are treated as manual migration leftovers and are not removed automatically. The standalone `xkit gen wire` command may remain as a legacy explicit command, but it is not used by the template startup path.

## Handwritten Extension Points

The following files are reserved for later handwritten changes and are created once only:

- `internal/service/*_service_ext.go`
- `internal/data/repo/*_repo_ext.go`
- `internal/bootstrap/hooks.go`
- `internal/server/options.go`
- `internal/data/bootstrap/data.go`
- `internal/data/bootstrap/resources.go`
- `configs/*.yaml`

`internal/bootstrap/hooks.go` is for extra transport servers or lifecycle additions.

`internal/server/options.go` is for project-specific HTTP/gRPC business middleware. Default config mapping and framework transport options belong in the template-maintained `*_options.go` files.

`internal/data/bootstrap/data.go` is for shared data providers such as Redis, cache, object storage, queues, or domain-specific clients.

`internal/data/bootstrap/resources.go` is for shared data resource lifecycle wiring.

## Template Init Flow

The intended target flow is:

```text
xkit init template [template-source] --project <target> ...
xkit init source <source-path> --project <target> --service <service>
xkit gen all <service> --project <target> --config <config>
go get -u all
go mod tidy
```

The default template source is:

```text
https://github.com/chnxq/xkit-template.git
```

Local template directories are still accepted for offline development.

The `go get -u all` step is part of `xkit init template` after real copying. It is skipped for `--dry-run` and can be disabled with `--skip-go-get-update-all` when working offline or when API/Ent generated code is not complete yet. In that practical flow, run `go get -u all` and `go mod tidy` after `buf generate`, Ent generation, and `xkit gen all`.

Running `xkit gen all` before template initialization is intentionally incomplete: static startup/config files are no longer generated by `xkit`.

`xkit init source` is the bridge between raw API/schema inputs and generated glue. It is not a template command and does not write startup skeleton files. It copies proto/schema source files into the target project's active generation locations and creates the YAML consumed by `xkit gen all`.

Default source import behavior:

- source proto root: `<source>/api/protos`
- source schema root: `<source>/schema`
- target proto root: `<target>/api/protos`
- target schema root: `<target>/internal/data/ent/schema`
- default config path: `<source>/<project-name>-config/<service>.yaml`

For `xadmin-web`, the default admin config path is:

```text
D:\GoProjects\XAdmin\xadmin-web\source\xadmin-web-config\admin.yaml
```

This keeps the pipeline explicit:

1. `xkit init template` creates project-owned startup code.
2. `xkit init source` materializes raw proto/schema inputs and derives generation config.
3. `xkit gen all` writes the dynamic resource glue.

## Old `pkg` Reuse and Integration Plan

Reference package path:

```text
D:\GoProjects\chnxq\xadmin\pkg
```

Do not copy this directory wholesale into the template. Several packages contain hard-coded `xadmin/...` imports or strong business assumptions. The template should first expose stable hooks and adapters, then migrate generic pieces selectively.

Transport and middleware integration now follows this rule:

- Framework middleware controlled by `conf.Middleware` belongs in the template and should use `github.com/chnxq/xkitpkg/middleware` directly. The current template maps recovery, tracing, validation, metadata, rate limit, circuit breaker, and standard request logging for HTTP/gRPC.
- Business middleware from old `pkg/middleware`, such as auth, authorization, and audit logging, remains handwritten. It should be plugged through `HTTPMiddlewares`, `GRPCMiddlewares`, and `GRPCStreamMiddlewares` after its generated API/domain dependencies are available.
- Asynq task infrastructure belongs in the template through `transport/asynq`; task types, task payloads, service callbacks, and scheduled task startup belong in project service extension files when needed.
- SSE infrastructure belongs in the template through `transport/sse`; stream IDs, domain events, and publish policy belong in project service extension files when needed.
- OSS provider wiring should start in `internal/data/bootstrap` using `github.com/chnxq/xkitpkg/oss/minio` once a target project needs object storage. The old `pkg/oss` service methods depend on generated storage APIs, so only generic URL/object-name helpers should be considered for extraction.

Good candidates for later extraction after import cleanup:

- `crypto`: general AES-GCM helpers.
- `eventbus`: general event manager and middleware ideas.
- `utils`: small generic helpers.
- generic parts of `oss`: object name helpers, content-type helpers, host replacement, and bucket helpers after removing generated API error types.

Reference-only for now:

- `metadata`
- `middleware/auth`
- `authorizer`
- `middleware/logging`
- `task`
- `lua`

These contain useful design, but they depend on generated API packages, auth payloads, policy models, audit log repositories, task services, or project-specific behavior. They should be adapted through the handwritten hook files instead of copied directly.

Keep project-specific:

- `constants`
- `jwt`
- `utils/converter`
- `entgo/viewer`

These encode xadmin domain defaults or concrete business data.
