# xadmin-web 启动骨架生成方案

> 本文保留为历史阶段记录。当前启动模板与动态生成边界以项目根目录的 `README.md` 和 `doc/bootstrap-template-generated-boundary.md` 为准；`cmd/server`、`configs`、`internal/bootstrap`、`internal/server` 等启动骨架现在由 `xkit-template` 通过 `xkit init template` 落地，`xkit gen bootstrap` 只保留动态装配代码。

## 参考结论

参考 `D:\GoProjects\chnxq\XAdmin` 后，启动链路可以拆成三层：

1. `cmd/server`：命令入口，负责版本信息、命令参数、启动应用。
2. `internal/bootstrap`：负责装配配置、日志、注册中心、链路追踪、transport server 和生命周期。
3. `internal/{data,server,service}`：业务 provider、仓库 provider、HTTP/gRPC 注册与 middleware 组合。

## xadmin-web 当前状态

`xadmin-web` 已有：

- `api/gen/...` protobuf 生成代码。
- `internal/data/ent` Ent 代码。
- `internal/service/*.gen.go` 服务骨架。
- `internal/data/repo/*_repo.gen.go` 仓库骨架。
- `internal/server/*_register.gen.go` 聚合注册代码。

缺少：

- `cmd/server` 程序入口。
- `internal/bootstrap` 生命周期装配。
- 公共基础设施 provider：配置、日志、追踪、注册发现、数据库连接、缓存等。
- 与参考项目 `pkg` 对应的领域共享包。

## 生成策略

### 第一阶段：可编译启动骨架

已实现 `xkit gen bootstrap`，生成：

- `cmd/server/main.go`
- `cmd/server/server.go`
- `internal/bootstrap/app.go`
- `internal/data/data.go`
- `internal/server/server.go`

第一阶段保持零新增依赖，生成代码可编译，并在源码中保留 TODO，后续逐项接入：

- `xkitpkg/config` 本地与远程配置。
- `xkitpkg/logger` / `xkitmod/log` 统一日志。
- `xkitpkg/tracer` 链路追踪。
- `xkitpkg/registry` + `xkitmod/registry` 多实例注册发现。
- `xkitpkg/transport/http` 与 `xkitpkg/transport/grpc` 服务启动。

### 第二阶段：基础设施 provider

建议继续生成：

- `internal/data/ent_client.go`：Ent 客户端和迁移。
- `internal/bootstrap/config.go`：配置加载和远程配置开关。
- `internal/bootstrap/logger.go`：日志 provider。
- `internal/bootstrap/registry.go`：注册中心 provider。
- `internal/bootstrap/tracer.go`：链路追踪 provider。
- `internal/server/http.go` / `grpc.go`：组合 generated register 和通用 middleware。

### 第三阶段：pkg 处理策略

不建议直接复制参考项目完整 `pkg` 到每个目标项目。建议拆成两类：

1. 通用能力沉淀到独立模块或 xkit 模板项目：
   - crypto
   - metadata
   - middleware/logging
   - middleware/auth 基础接口
   - eventbus
   - oss 抽象
   - entgo viewer

2. 业务常量/默认数据留在目标项目或领域模板：
   - constants/default_data.go
   - permission/role/tenant 初始化数据
   - 与 XAdmin 业务强绑定的 converter、authorizer、jwt payload

后续可以新增一个 `xkit-template-xadmin-web` 或在 xkit 内建立 `template/project/xadmin-web` 模板目录。代码生成器负责把模板项目实例化，而不是把所有模板塞进 `runner.go`。

## 后续实施顺序

1. 生成 `internal/data/ent_client.go` 并编译。
2. 生成 `internal/server/http.go` / `grpc.go`，接入 `RegisterGeneratedHTTPServices` / `RegisterGeneratedGRPCServices`。
3. 生成 `internal/bootstrap/config/logger/registry/tracer.go`，替换 `app.go` 中 TODO。
4. 把参考 `pkg` 拆为“通用模板包”和“XAdmin 业务包”，先生成最小 authorizer/metadata/middleware 接口。
5. 再考虑独立模板项目，作为新项目 scaffold 的来源。

## 第二阶段进展

已继续生成并验证：

- `internal/data/ent_client.go`：提供 `NewEntClient` provider 占位，返回明确错误，后续接入数据库配置与 `xkitpkg/orm/ent`。
- `internal/server/http.go`：创建 `xkitpkg/transport/http.Server`，读取 `cfg.Server.Rest.GetAddr()`，并调用 `RegisterGeneratedHTTPServices`。
- `internal/server/grpc.go`：创建 `xkitpkg/transport/grpc.Server`，读取 `cfg.Server.Grpc.GetAddr()`，并调用 `RegisterGeneratedGRPCServices`。

当前 `EntClient` 仍是占位实现，目的是保证 provider 形态先稳定、生成代码可编译。下一步需要接入 `xkitpkg/orm/ent` 或直接使用 Ent SQL driver，并从 `conf.ServerConfig.Data.Database` 读取驱动、DSN、迁移开关。

## data 目录拆分

已插入目录拆分步骤，避免 repo 生成代码与 bootstrap 基础设施代码混在 `internal/data` 根目录：

- repo 生成代码：`internal/data/repo/*.gen.go`
- bootstrap glue：`internal/bootstrap/generated_servers.gen.go` 直接装配 `internal/data/repo` 与 `internal/service`
- bootstrap data 代码：`internal/data/bootstrap/*.go`

服务层现在导入 `xadmin-web/internal/data/repo`，构造函数使用 `repo.UserRepo`、`repo.UserCredentialRepo` 等接口类型。


## EntClient 接入进展

`internal/data/bootstrap/ent_client.go` 已由占位实现升级为真实 Ent 初始化骨架：

- 从 `ctx.GetConfig().Data.Database` 读取 `driver`、`source`、连接池和 `migrate` 配置。
- 使用 `entgo.io/ent/dialect/sql.Open` 创建 Ent SQL driver。
- 创建 `ent.NewClient(ent.Driver(drv))`。
- 包装为 `entCrud.NewEntClient[*ent.Client]`，供 repo 构造函数使用。
- 支持 `migrate.WithForeignKeys(true)` 自动迁移。

注意：模板不主动 blank import 具体数据库驱动。目标项目需要在主程序或独立 driver 文件中导入实际驱动，例如 MySQL、PostgreSQL 或 SQLite。

## 配置系统生成规划与进展

参考 `D:\GoProjects\chnxq\XAdmin\app\admin\service\configs`，配置文件生成到目标项目根目录的 `configs/`：

- `configs/server.yaml`：REST/gRPC 协议监听、超时、中间件、CORS、pprof、swagger。
- `configs/data.yaml`：数据库、Redis、迁移、连接池、数据库 tracing/metrics 开关。
- `configs/logger.yaml`：统一日志 provider 配置，默认 zap。
- `configs/trace.yaml`：OTLP trace exporter、采样率、batcher、trace context。
- `configs/registry.yaml`：etcd 服务注册发现配置，用于多实例。
- `configs/client.yaml`：gRPC client 超时与中间件配置。
- `configs/remote_config.yaml`：远程配置中心配置。

`internal/bootstrap/app.go` 已调用 `xkitpkg/config.LoadServerConfig(opts.ConfigPath)` 加载配置目录，并将 `config.GetServerConfig()` 注入 `app.AppCtx`。日志、注册中心、链路追踪 provider 仍保留 TODO，后续按配置逐项接入。

## 基础设施 provider 接入进展

已新增 `internal/bootstrap/infra.go`：

- `NewLogger`：当前返回 `xkitmod/log.DefaultLogger`，后续在目标项目加入 `xkitpkg/logger` 模块后按 `configs/logger.yaml` 初始化。
- `NewRegistrar`：已读取 `ServerConfig.Registry` 并调用 `xkitpkg/registry.NewRegistrar`，用于多实例服务注册发现。
- `NewTracer`：当前保留 TODO，后续在目标项目加入 `xkitpkg/tracer` 模块后按 `configs/trace.yaml` 初始化。

`internal/bootstrap/app.go` 已按顺序初始化 logger、registrar、tracer，并将 logger/registrar 注入 `app.NewAppCtx`。
