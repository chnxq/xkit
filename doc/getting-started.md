# Getting Started: xkit 项目初始化与代码生成

本文面向第一次把 `xkit-template`、原始 `proto/schema` 和 `xkit` 生成器串起来的场景，目标是得到一个可以继续开发和验证的 XAdmin 风格 Go 服务工程。

示例路径以当前工作区为准：

```text
D:\GoProjects\XAdmin
  xkit\                  # 生成器
  xkit-template\         # 启动模板工程
  xadmin-web\            # 目标工程
```

## 一、先理解三个边界

`xkit-template` 负责项目启动骨架。它提供 `cmd/server`、`configs`、`internal/bootstrap`、`internal/server`、`internal/data/bootstrap` 等静态代码。这些文件在 `xkit init template` 后成为目标项目自己的代码，后续允许人工维护。

`xadmin-web/source` 负责原始业务输入。推荐结构如下：

```text
xadmin-web/source/
  api/protos/            # 原始 Protobuf 定义
  schema/                # 原始 Ent schema
  xadmin-web-config/     # xkit init source 生成的配置目录
```

`xkit` 负责动态代码。它读取目标项目中的 `api/protos`、`api/gen`、`internal/data/ent/schema`、Ent 生成代码和 `admin.yaml`，生成 `*.gen.go` 以及一次性创建的 `*_ext.go`。

## 二、前置条件

开始前需要确认：

- 已安装 Go，并且目标工程 `xadmin-web/go.mod` 存在。
- 本地存在 `xkit`、`xkit-template`、`xadmin-web` 三个目录。
- `xadmin-web/source/api/protos` 和 `xadmin-web/source/schema` 已准备好。
- 如需重新生成 API Go 代码，已安装并可执行 `buf` 以及 `buf.gen.yaml` 中使用到的 protoc 插件。
- 如需重新生成 Ent 代码，目标工程已有 Ent 生成入口，例如 `internal/data/ent/generate.go`。
- `xkit init template` 默认会在真实复制后执行 `go get -u all`，因此需要可访问 Go module 网络源；离线时使用 `--skip-go-get-update-all`。

建议所有会写文件的命令先加 `--dry-run` 观察计划，再去掉该参数执行。

## 三、初始化启动模板

先在 `xkit` 目录执行 dry-run：

```powershell
cd D:\GoProjects\XAdmin\xkit

go run ./cmd/xkit init template D:\GoProjects\XAdmin\xkit-template `
  --project D:\GoProjects\XAdmin\xadmin-web `
  --module xadmin-web `
  --app-name XAdmin `
  --command-name xadmin-web `
  --service-name admin `
  --dry-run
```

重点观察输出：

- `planned ... (template)`：将要复制或渲染的模板文件。
- `skipped ... (exists)`：已存在且被保护的文件，例如 `configs/*.yaml`、`internal/bootstrap/hooks.go`、`internal/server/options.go`。
- `planned remove ... (template)`：模板声明的过期文件，例如旧的 `cmd/server/wire.go`、`cmd/server/wire_gen.go`。

确认无误后执行真实初始化：

```powershell
go run ./cmd/xkit init template D:\GoProjects\XAdmin\xkit-template `
  --project D:\GoProjects\XAdmin\xadmin-web `
  --module xadmin-web `
  --app-name XAdmin `
  --command-name xadmin-web `
  --service-name admin
```

执行后应看到：

- 目标工程出现或更新 `cmd/server`、`configs`、`internal/bootstrap`、`internal/server`、`internal/data/bootstrap`。
- 被 `template.yaml` 标记为 preserve 的文件如果已存在，会继续保留。
- 命令末尾输出 `ran go get -u all (...)`。如果不希望更新依赖，执行时追加 `--skip-go-get-update-all`。

## 四、导入原始 proto/schema 并生成 xkit 配置

先观察 source import 计划：

```powershell
cd D:\GoProjects\XAdmin\xkit

go run ./cmd/xkit init source D:\GoProjects\XAdmin\xadmin-web\source `
  --project D:\GoProjects\XAdmin\xadmin-web `
  --service admin `
  --dry-run
```

重点观察输出：

- `planned ... (source)`：将复制到目标项目的 proto/schema 文件。
- `skipped ... (exists)`：目标位置已有同名文件，默认不会覆盖。
- `skipped resources without matching proto service: ...`：Ent schema 没有匹配的 proto service，常见于关联表、详情表或纯嵌入资源，通常是正常现象。
- `config ...\source\xadmin-web-config\admin.yaml`：本次生成器配置的目标路径。

确认无误后执行真实导入：

```powershell
go run ./cmd/xkit init source D:\GoProjects\XAdmin\xadmin-web\source `
  --project D:\GoProjects\XAdmin\xadmin-web `
  --service admin
```

执行后应看到：

```text
xadmin-web/api/protos/                 # 从 source/api/protos 复制
xadmin-web/internal/data/ent/schema/   # 从 source/schema 复制
xadmin-web/source/xadmin-web-config/admin.yaml
```

如果确实要覆盖已有 proto/schema/config，再追加 `--force`。

## 五、生成 API 绑定代码

`xkit gen all` 会读取 `api/gen` 中的 gRPC/HTTP 绑定信息，所以当 proto 有变化时，需要先重新生成 API Go 代码。

```powershell
cd D:\GoProjects\XAdmin\xadmin-web\api

buf generate --template buf.gen.yaml
```

重点观察：

- `api/gen/<domain>/v1/*.pb.go`
- `api/gen/<domain>/v1/*_grpc.pb.go`
- `api/gen/<domain>/v1/*_http.pb.go`

如果 `xkit gen all` 报找不到 binding 或注册函数，优先检查这一步是否执行成功。

## 六、生成 Ent 代码

`xkit gen all` 不只读取 Ent schema，也会引用 Ent 已生成的 entity、predicate、client 等代码。schema 有变化时，需要按目标工程的 Ent 流程重新生成。

当前 `xadmin-web` 可在工程根目录执行：

```powershell
cd D:\GoProjects\XAdmin\xadmin-web

go generate ./internal/data/ent
```

重点观察：

- `internal/data/ent/client.go`
- `internal/data/ent/<entity>.go`
- `internal/data/ent/<entity>_query.go`
- `internal/data/ent/<entity>_create.go`
- `internal/data/ent/<entity>_update.go`
- `internal/data/ent/predicate`

如果 repo 生成代码编译时报 Ent 类型、predicate 或字段不存在，优先检查 Ent 代码是否重新生成。

## 七、生成 xkit 动态代码

先 dry-run：

```powershell
cd D:\GoProjects\XAdmin\xkit

go run ./cmd/xkit gen all admin `
  --project D:\GoProjects\XAdmin\xadmin-web `
  --config D:\GoProjects\XAdmin\xadmin-web\source\xadmin-web-config\admin.yaml `
  --dry-run
```

重点观察：

- `planned ... (all)`：将被生成或覆盖的 `*.gen.go`。
- `skipped ... (exists)`：已存在的一次性扩展文件，例如 `*_ext.go`，正常情况下不会覆盖。

确认无误后执行真实生成：

```powershell
go run ./cmd/xkit gen all admin `
  --project D:\GoProjects\XAdmin\xadmin-web `
  --config D:\GoProjects\XAdmin\xadmin-web\source\xadmin-web-config\admin.yaml
```

执行后重点检查这些目录：

```text
xadmin-web/internal/service/
  *_service.gen.go
  *_service_ext.go
  providers/wire_set.gen.go

xadmin-web/internal/data/repo/
  *_repo.gen.go
  *_repo_ext.go

xadmin-web/internal/data/providers/
  wire_set.gen.go

xadmin-web/internal/server/
  rest_register.gen.go
  grpc_register.gen.go

xadmin-web/internal/bootstrap/
  generated_servers.gen.go

xadmin-web/internal/data/bootstrap/
  ent_client.gen.go
```

生成规则是：`*.gen.go` 可重复覆盖，`*_ext.go` 只在不存在时创建，手写文件不应被自动覆盖。

## 八、验证和启动

先验证生成器自身：

```powershell
cd D:\GoProjects\XAdmin\xkit
go test ./...
```

再验证目标工程：

```powershell
cd D:\GoProjects\XAdmin\xadmin-web
go test ./...
```

最后尝试启动：

```powershell
go run ./cmd/server server -config_path ./configs
```

启动观察重点：

- 配置目录是否被正确加载。
- 数据库、Redis、注册中心、trace exporter 等外部依赖是否可用。
- HTTP/gRPC/Asynq/SSE 中启用的 transport 是否按配置启动。
- `internal/bootstrap/hooks.go`、`internal/server/options.go`、`internal/data/bootstrap/data.go` 中的手写扩展是否需要补齐。

如果只是验证代码生成链路，外部依赖未准备好导致启动失败并不一定代表生成失败；先以 `go test ./...` 和编译错误为准。

## 九、中间结果观察表

| 阶段 | 命令 | 应看到的结果 |
| --- | --- | --- |
| 模板 dry-run | `xkit init template ... --dry-run` | 输出 `planned`、`skipped`、`planned remove`，目标项目不变 |
| 模板初始化 | `xkit init template ...` | 复制启动骨架，保留 preserve 文件，默认执行 `go get -u all` |
| source dry-run | `xkit init source ... --dry-run` | 输出 proto/schema 写入计划和 config 路径 |
| source 导入 | `xkit init source ...` | 写入 `api/protos`、`internal/data/ent/schema`、`source/xadmin-web-config/admin.yaml` |
| API 生成 | `buf generate --template buf.gen.yaml` | 写入 `api/gen` 下的 pb、grpc、http 代码 |
| Ent 生成 | `go generate ./internal/data/ent` | 写入 Ent client、entity、query、mutation、predicate 代码 |
| xkit dry-run | `xkit gen all ... --dry-run` | 输出将覆盖的 `*.gen.go` 和跳过的 `*_ext.go` |
| xkit 生成 | `xkit gen all ...` | 写入 service、repo、register、wire、bootstrap glue |
| 验证 | `go test ./...` | 编译和测试通过，或暴露需要补齐的手写扩展/依赖问题 |
| 启动 | `go run ./cmd/server server -config_path ./configs` | 服务按配置启动，或提示外部依赖缺失 |

## 十、常见问题定位

`config module "..." does not match target project module "..."`

检查 `admin.yaml` 中的 `module` 是否和 `xadmin-web/go.mod` 的 module 一致。

`proto service not found`

检查目标项目 `api/protos` 是否已经由 `xkit init source` 导入，并确认 `admin.yaml` 中的 `proto_service` 名称和 proto 文件一致。

找不到 HTTP/gRPC 注册函数或 binding

检查 `xadmin-web/api/gen` 是否已由 `buf generate --template buf.gen.yaml` 生成，尤其是 `*_grpc.pb.go` 和 `*_http.pb.go`。

找不到 Ent 类型、predicate、字段或 client

检查是否执行了 `go generate ./internal/data/ent`，并确认 schema 和生成代码一致。

`skipped resources without matching proto service`

通常是关联表、详情表、纯嵌入资源没有独立 proto service。只要这些资源不需要独立 service/repo 生成，就可以接受。

模板初始化时依赖更新失败

如果网络或私有源暂时不可用，可以先执行：

```powershell
go run ./cmd/xkit init template D:\GoProjects\XAdmin\xkit-template `
  --project D:\GoProjects\XAdmin\xadmin-web `
  --module xadmin-web `
  --app-name XAdmin `
  --command-name xadmin-web `
  --service-name admin `
  --skip-go-get-update-all
```

之后在目标工程网络可用时再手动执行：

```powershell
cd D:\GoProjects\XAdmin\xadmin-web
go get -u all
go mod tidy
```

## 示例

执行结果

- xkit init template：模板已落到 admin-01。初次默认 go get -u all 因 source/schema 里残留 xadmin-web/api/gen/... 失败，后续改为先 --skip-go-get-update-all 收口，完整生成后再跑依赖更新。
- xkit init source：成功复制 api/protos、internal/data/ent/schema，生成 admin-01/source/admin-01-config/admin.yaml。
- buf generate --template buf.gen.yaml：成功，生成 59 个 *_grpc.pb.go。
- Ent 生成：成功，生成 38 个 *_query.go，client.go/predicate 等均存在。
- xkit gen all admin：成功，生成 24 个 service、23 个 repo，并跳过已有 *_ext.go。
- go get -u all + go mod tidy：最终成功。
- go test ./...：
    - xkit 通过。
    - admin-01 通过。
- 启动验证：go run ./cmd/server server -config_path ./configs 运行 30 秒未退出，未出现即时配置/编译错误；超时后确认没有遗留 admin-01 进程。
