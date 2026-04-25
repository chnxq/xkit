# 启动骨架与公用 provider 模板仓库方案

## 结论

对于服务启动、基础设施 provider 装配、配置目录、公用中间件和通用 `pkg` 代码，这部分确实更适合从 `xkit` 的细粒度代码生成中拆出来，转成“模板仓库 + 少量变量替换 + 局部代码生成”的模式。

原因很直接：

1. 这部分代码主要受技术栈影响，不直接受 API 和 Ent schema 影响。
2. 它的变更节奏远慢于业务资源、service/repo/register 的生成代码。
3. 把这部分长期放在 `runner.go` 里，会导致模板数量越来越多、职责混杂、维护成本升高。
4. 启动骨架通常需要人工长期演进，模板仓库更适合承载“可读、可调、可二次开发”的代码。

建议后续采用：

- `xkit` 负责“资源相关的增量生成”。
- 模板仓库负责“项目脚手架、启动链路、公用能力、默认目录结构”。

## 当前问题

从现状看，`xkit` 里这几类文件已经偏向“项目模板”而不是“资源生成”：

- `cmd/server/main.go`
- `cmd/server/server.go`
- `internal/bootstrap/*.go`
- `internal/data/bootstrap/*.go`
- `internal/server/http.go`
- `internal/server/grpc.go`
- `configs/*.yaml`

这些文件有两个特点：

1. 它们并不随着 `user`、`position`、`org_unit` 之类资源数量变化而显著变化。
2. 它们更依赖统一技术方案，例如：
   - 日志实现
   - trace exporter
   - registry provider
   - 配置加载方式
   - server middleware 组合
   - Ent/Redis/Cache 客户端初始化

所以，把它们继续塞在 `xkit/internal/codegen/template/*.tmpl` 里，不是不能做，但会让 `xkit` 同时承担两种职责：

- 项目模板引擎
- 资源代码生成器

这两个职责应当分开。

## 目标架构

建议拆成三层：

### 第一层：模板仓库

单独建立一个模板仓库，例如：

- `xkit-template-xadmin-web`

它负责存放“项目初始化时就应该存在，且后续主要人工维护”的代码：

- `cmd/server/*`
- `internal/bootstrap/*`
- `internal/server/http.go`
- `internal/server/grpc.go`
- `internal/data/bootstrap/*`
- `internal/pkg/*` 或 `pkg/*`
- `configs/*`
- `Makefile`
- `Dockerfile`
- `.gitignore`
- `wire.go` / `wire_gen.go` 的基础骨架
- README、部署说明、环境变量样例

模板仓库不是 demo，而是“可直接运行的起始工程”。

### 第二层：xkit 资源生成器

`xkit` 继续负责和 API/schema 强相关的代码：

- `internal/service/*.gen.go`
- `internal/data/repo/*.gen.go`
- `internal/server/*_register.gen.go`
- `internal/*/providers/wire_set.gen.go`
- 资源相关的 ext 文件
- 与 proto / ent schema 绑定的局部 glue code

这部分仍然适合 `gen service`、`gen repo`、`gen register`、`gen wire` 这类命令。

### 第三层：模板同步器

在 `xkit` 中新增“模板项目初始化 / 升级”能力，例如：

- `xkit init project --template xadmin-web`
- `xkit sync template --template xadmin-web`

其中：

- `init`：从模板仓库复制一份初始工程。
- `sync`：只同步允许覆盖的模板文件。

## 目录建议

建议后续形成如下结构：

### 模板仓库目录

```text
xkit-template-xadmin-web/
  template.yaml
  cmd/server/
  internal/bootstrap/
  internal/server/
  internal/data/bootstrap/
  internal/pkg/
  configs/
  deploy/
  hack/
```

### `xkit` 内部目录

```text
xkit/
  internal/scaffold/
    template/
    sync/
    manifest/
  internal/codegen/
    service/
    repo/
    register/
    wire/
```

这里的关键点是：

- `internal/scaffold` 负责模板仓库相关逻辑。
- `internal/codegen` 只负责资源生成。

## 模板仓库中的文件分类

模板仓库中的文件建议分成三类。

### A. 永久模板文件

直接来自模板仓库，后续允许项目自行长期修改，不再由 `xkit gen` 覆盖。

例如：

- `cmd/server/main.go`
- `internal/bootstrap/infra.go`
- `configs/logger.yaml`
- `configs/trace.yaml`

这类文件适合人工维护。

### B. 模板生成文件

来自模板仓库，但允许按变量渲染一次，例如：

- 模块名
- 服务名
- app id
- 默认端口
- 默认 registry key

例如：

- `internal/bootstrap/app.go`
- `configs/server.yaml`
- `configs/registry.yaml`

这类文件适合在 `init` 阶段渲染。

### C. 模板预留挂点文件

模板仓库预留一些“生成器会写入或引用”的聚合点，但不直接维护业务代码。

例如：

- `internal/bootstrap/app.go` 中只保留固定启动逻辑和调用挂点。
- `internal/server/http.go` / `grpc.go` 中调用 `RegisterGeneratedHTTPServices`。
- `internal/bootstrap/providers.go` 中预留 provider set 入口。

这类文件的原则是：只依赖固定接口，不直接依赖具体资源名。

## 推荐的责任边界

建议按下面的边界来收敛。

### 放到模板仓库的内容

- 服务入口
- 配置系统
- 日志初始化
- trace 初始化
- registry 初始化
- 远程配置初始化
- HTTP/gRPC server 基础构造
- middleware 链定义
- EntClient / Redis / Cache / MQ / OSS 等基础 provider
- 通用错误处理
- 通用 metadata / auth / operator / tenant 上下文
- 项目级 README / 部署脚本 / 容器化文件

### 留在 `xkit` 生成器中的内容

- 每个资源的 Service 实现骨架
- 每个资源的 Repo 实现骨架
- 每个资源的 HTTP/gRPC register 聚合
- 与资源相关的 wire provider set
- Service method 的定制 body 注入
- schema/proto 推导出的 filters / exists / CRUD / list 逻辑

### 不建议放进模板仓库的内容

- `user`、`role`、`position` 这类具体资源生成文件
- 依赖 proto schema 结构的 method body
- 依赖 Ent 字段列表推导出的 repo 逻辑

这些文件天然属于生成器，不属于模板仓库。

## 一个更稳妥的实现方式

不建议一上来就把现在所有 bootstrap 模板全部挪出 `xkit`。更稳妥的做法是分两步。

### 第一步：先抽“静态骨架”

先把几乎不随资源变化的部分迁到模板仓库：

- `cmd/server/*`
- `internal/bootstrap/infra.go`
- `internal/server/http.go`
- `internal/server/grpc.go`
- `configs/*.yaml`
- `internal/data/bootstrap/data.go`

这些文件迁出后，`xkit gen bootstrap` 只保留最少的动态部分，例如：

- `internal/bootstrap/app.go`
- `internal/data/bootstrap/ent_client.go`

### 第二步：再抽“启动聚合”

等模板仓库稳定后，再把 `internal/bootstrap/app.go` 也收敛成模板仓库中的固定骨架，只给 `xkit` 留一个小型生成文件，例如：

- `internal/bootstrap/generated_servers.gen.go`
- `internal/bootstrap/generated_providers.gen.go`

这样 `app.go` 不再知道有哪些资源，只依赖：

- `NewGeneratedHTTPServices`
- `NewGeneratedGRPCServices`
- `NewGeneratedServiceProviders`

真正随资源变化的部分，全部收敛到 `*.gen.go`。

这比当前把 repo/service/http/grpc 聚合逻辑直接拼进 `app.go` 更清晰。

## 推荐的最终落地形态

建议最终把启动链路拆成以下结构：

```text
cmd/server/
  main.go
  server.go

internal/bootstrap/
  app.go
  infra.go
  generated_providers.gen.go
  generated_servers.gen.go

internal/server/
  http.go
  grpc.go
  rest_register.gen.go
  grpc_register.gen.go

internal/data/bootstrap/
  ent_client.go
  redis.go
  cache.go
```

其中：

- `app.go`、`infra.go`、`http.go`、`grpc.go` 来自模板仓库。
- `generated_providers.gen.go`、`generated_servers.gen.go` 来自 `xkit`。
- `rest_register.gen.go`、`grpc_register.gen.go` 已经属于生成器，继续保留。

## 模板仓库的元信息

建议模板仓库增加一个描述文件，例如 `template.yaml`：

```yaml
name: xadmin-web
kind: service-template
version: 0.1.0

variables:
  - module
  - app_name
  - service_name
  - app_id

sync:
  overwrite:
    - cmd/server/main.go
    - cmd/server/server.go
    - internal/server/http.go
    - internal/server/grpc.go
  preserve:
    - configs/server.yaml
    - configs/data.yaml
    - internal/bootstrap/infra.go
```

这样 `xkit` 可以明确哪些文件：

- 初始化时生成
- 后续允许同步覆盖
- 后续只创建不覆盖

## `xkit` 命令建议

建议新增两组命令。

### 脚手架命令

```text
xkit init template xadmin-web --project D:\GoProjects\XAdmin\xadmin-web
xkit sync template xadmin-web --project D:\GoProjects\XAdmin\xadmin-web
```

### 资源命令

```text
xkit gen service admin --config ...
xkit gen repo admin --config ...
xkit gen register admin --config ...
xkit gen wire admin --config ...
xkit gen all admin --config ...
```

后续不建议再把 `bootstrap` 作为一个大而全的生成目标继续扩张。更合理的是：

- `bootstrap` 逐步缩小为“少量动态聚合代码”
- 大部分启动骨架转移到模板仓库

## 与参考项目 `pkg` 的关系

参考 `D:\GoProjects\chnxq\XAdmin` 的 `pkg`，建议拆成两部分处理。

### 通用包

适合进入模板仓库或独立基础模块：

- metadata
- operator
- auth context
- middleware helper
- errors
- response helper
- trace/log helper

### 业务包

仍然保留在具体项目中：

- 权限初始化数据
- 特定业务 converter
- 强业务语义的 authorizer
- 特定领域常量

不要把整个参考项目 `pkg` 原样复制进模板仓库，否则模板仓库会再次变成业务仓库。

## 升级策略

模板仓库模式最大的风险不是初始化，而是后续升级。建议一开始就明确：

1. 模板仓库只同步少量“平台层文件”。
2. 项目中凡是需要大量人工修改的文件，默认只初始化一次，不做覆盖升级。
3. 真正易变的业务聚合代码继续交给 `*.gen.go`。

这样可以避免每次升级模板时和业务修改发生大面积冲突。

## 推荐实施顺序

建议按下面顺序推进：

1. 新建模板仓库 `xkit-template-xadmin-web`。
2. 先迁出 `cmd/server`、`configs`、`internal/server/http.go`、`internal/server/grpc.go`。
3. 在 `xkit` 中新增 `init template` 命令。
4. 将 `gen bootstrap` 缩减为只生成动态聚合代码。
5. 再决定是否把 `internal/bootstrap/app.go` 也改成“模板固定文件 + generated glue”模式。

## 对当前工程的直接建议

对于现在的 `xkit` / `xadmin-web`，建议立即采用这个方向，但不建议一次性重构完。

最合适的下一步是：

1. 在 `kit/doc` 保留本方案文档。
2. 新建模板仓库目录或临时在 `xkit/templates/xadmin-web-base` 做第一版验证。
3. 先迁出稳定文件：
   - `cmd/server/*`
   - `internal/server/http.go`
   - `internal/server/grpc.go`
   - `configs/*`
4. 把 `gen bootstrap` 收缩为：
   - `internal/bootstrap/generated_servers.gen.go`
   - `internal/bootstrap/generated_providers.gen.go`
   - `internal/data/bootstrap/ent_client.go`

这个边界最清晰，也最便于后续维护。

