# XAdmin 工作区目录结构图（2026-05-10 快照）

本文档用于补充当前阶段 `xkit` 与 `admin` 的目录结构认知，并新增依赖关系图，供后续开发快速定位代码归属与调用链路。  
路径基准：`D:\GoProjects\XAdmin`

## 1. `xkit` 目录结构（中文注释）

```text
xkit/
├─ cmd/
│  └─ xkit/                    # xkit CLI 入口（命令行驱动）
├─ developer/                  # 开发期辅助文件（如本地联调配置）
├─ doc/                        # 设计文档、交接文档、方案沉淀
├─ examples/
│  ├─ xadmin/                  # 后端示例（proto/schema/config 等示例资产）
│  └─ xadmin-web/              # 前端示例/素材（非当前 admin-ui 主线）
├─ internal/
│  ├─ binding/                 # 配置/模型绑定层
│  ├─ cli/                     # CLI 编排逻辑
│  ├─ codegen/
│  │  └─ template/             # 代码生成模板与渲染逻辑
│  ├─ config/                  # 生成器配置模型与加载
│  ├─ entschema/               # Ent Schema 解析/抽象
│  ├─ project/                 # 项目结构规划与生成计划
│  ├─ proto/                   # Proto 相关处理（解析/生成辅助）
│  ├─ scaffold/                # 脚手架能力
│  └─ sourceimport/            # 源码/模板导入能力
├─ NEXT_CONTEXT_HANDOFF_20260501.md
├─ NEXT_CONTEXT_HANDOFF_20260505.md
├─ NEXT_CONTEXT_HANDOFF_20260510.md
├─ README.md
├─ go.mod
└─ go.sum
```

## 2. `xkit-template` 目录结构（中文注释）

```text
xkit-template/
├─ cmd/
│  └─ server/                  # 模板项目服务启动入口
├─ configs/                    # 模板默认配置文件
├─ internal/
│  ├─ bootstrap/               # 模板级服务装配/启动编排
│  ├─ data/
│  │  └─ bootstrap/            # 模板级数据初始化能力
│  ├─ server/                  # 模板级 HTTP/GRPC/SSE 服务封装
│  └─ service/                 # 模板级 Service 基础骨架
├─ template.yaml               # 模板元信息配置
├─ README.md
├─ go.mod
└─ go.sum
```

## 3. `admin` 目录结构（中文注释）

```text
admin/
├─ api/
│  ├─ protos/                  # Proto 源定义（按业务域拆分）
│  │  ├─ admin/
│  │  ├─ audit/
│  │  ├─ authentication/
│  │  ├─ dict/
│  │  ├─ identity/
│  │  ├─ internal_message/
│  │  ├─ permission/
│  │  ├─ resource/
│  │  ├─ storage/
│  │  └─ task/
│  └─ gen/                     # 生成后的 Go 代码（pb/grpc/http/openapi 相关）
│     ├─ admin/
│     ├─ audit/
│     ├─ authentication/
│     ├─ dict/
│     ├─ identity/
│     ├─ internal_message/
│     ├─ permission/
│     ├─ resource/
│     ├─ storage/
│     └─ task/
├─ cmd/
│  └─ server/                  # 后端服务启动入口与嵌入资源
├─ configs/                    # 运行时配置（server/data/auth 等）
├─ docs/                       # 项目文档
├─ internal/
│  ├─ bootstrap/               # 服务组装/依赖注入/启动编排
│  ├─ data/
│  │  ├─ bootstrap/            # 数据层初始化（migrate、默认数据等）
│  │  ├─ ent/                  # Ent 生成产物（schema/client/query/runtime）
│  │  └─ repo/                 # Repo 层（gen + *_ext 手写扩展）
│  ├─ server/                  # HTTP/GRPC/SSE 服务器与中间件、手写路由服务
│  └─ service/                 # Service 层（gen + *_ext 业务扩展）
├─ logfile/                    # 本地日志输出目录
├─ sql/                        # 默认/演示 SQL 资源
├─ go.mod
└─ go.sum
```

## 4. 依赖关系图（中文注释）

### 4.1 仓库级依赖关系（构建与生成维度）

```text
xkit
├─ 直接依赖: yaml.v3（配置解析）
└─ 作用: 代码生成器，输出/维护 admin 示例与生成规则

admin
├─ 直接依赖:
│  ├─ x-crud(api/entgo/viewer)      # CRUD抽象、分页过滤、viewer上下文
│  ├─ xkitpkg(app/conf/middleware/transport 等)  # 服务基础设施
│  ├─ xkitmod(errors/log/selector 等)            # 基础模块增强
│  ├─ x-utils(geoip/mapper/copier/id 等)         # 通用工具库
│  └─ x-swagger                                  # Swagger/OpenAPI 支持
└─ 作用: 主业务后端服务（REST/GRPC/SSE）

admin-ui
└─ 通过 HTTP 调用 admin（主要入口: /admin/v1/*）

关系总结:
xkit -> 生成/演进 admin 相关规则
admin-ui -> 调用 admin API
admin -> 运行时依赖 x-crud/xkitpkg/xkitmod/x-utils/x-swagger
```

### 4.2 运行时调用链（请求处理维度）

```text
浏览器(admin-ui)
  -> HTTP /admin/v1/*
    -> admin/internal/server (路由 + 中间件)
      -> admin/internal/service (业务服务层)
        -> admin/internal/data/repo (仓储层，含 *_ext 扩展)
          -> admin/internal/data/ent (ORM/SQL)
            -> MySQL/PostgreSQL
```

### 4.3 `admin` 代码逻辑架构依赖图（全局）

```text
启动与装配层
cmd/server/server.go
  -> internal/bootstrap.Initialize
    -> NewGeneratedServers + NewManualServers
      -> internal/server.NewHTTPServer / NewGRPCServer / (可选)NewSSEServer

传输与接入层
HTTP(REST) + GRPC + SSE
  -> generated register (rest_register.gen.go / grpc_register.gen.go)
  -> manual register (manual_http.go)

中间件与安全层
commonServerMiddlewares + authViewerMiddleware + defaultViewerMiddleware
  -> token解析/签名校验(auth_support_ext.go)
  -> viewer上下文注入(viewer_auth.go)
  -> 可选DB日志中间件(rest-enable_db_logging)

应用服务层
internal/service/* (gen + *_ext)
  -> 组合 repo 能力
  -> 对外提供 protobuf service 接口实现

数据访问层
internal/data/repo/* (gen + *_ext)
  -> Ent Client 查询/写入
  -> 审计仓库包装(WrapAuditLogRepos)

持久化层
internal/data/ent/*
  -> migrate + schema + runtime
  -> MySQL / PostgreSQL
```

主要代码锚点：

- `admin/cmd/server/server.go`：进程入口与版本信息注入。
- `admin/internal/bootstrap/app.go`：应用上下文、数据资源、传输服务总装配。
- `admin/internal/bootstrap/generated_servers.gen.go`：Repo/Service 生成组装与 server 注册桥接。
- `admin/internal/server/http_options.go`、`grpc_options.go`：协议级中间件、TLS、过滤器、DB 日志开关接入点。
- `admin/internal/server/manual_http.go`：登录、导航、分析等手写业务入口。

### 4.4 横切能力依赖图（配置、安全、可观测、初始化）

```text
configs/*
  -> appCtx.GetConfig()
    -> server options(http/grpc/sse)
    -> data bootstrap(ent_client + migrate + default data)
    -> auth(jwt method/key)

鉴权与权限
Authorization Bearer
  -> parseAndValidateToken
    -> user/role/permission repo 解析权限码
      -> viewer.Context 注入请求上下文

可观测与审计
logger/tracer/middleware
  -> transport operation 标注
  -> 错误码封装(xkitmod/errors)
  -> (可选)db logging 写入审计表

生成与手写协同
api/protos + xkit 生成物
  -> api/gen + internal/*/*.gen.go
    -> *_ext.go / manual_http.go 按扩展点覆盖业务差异
```

### 4.5 审计日志与分析页数据链路（业务示例，不代表全局全貌）

```text
请求进入 admin REST
  -> 中间件（含 database logging）
    -> 写入 sys_api_audit_logs / 其它审计表
      -> 分析页接口 /admin/v1/dashboard/analytics 聚合查询
        -> admin-ui dashboard/analytics 页面展示
```

说明：

- 分析页当前为普通 HTTP 拉取，不依赖 SSE。
- 若分析页访问量异常（如为 0），优先排查日志写入链路与审计仓库包装透传是否正常。

## 5. 使用建议（简版）

- 排查“生成逻辑问题”优先看：
  `xkit/internal/codegen`、`xkit/internal/entschema`、`xkit/internal/proto`
- 排查“后端运行行为”优先看：
  `admin/internal/server`、`admin/internal/data/repo/*_ext.go`
- 排查“模型与枚举不一致”优先看：
  `admin/api/protos/**` 与 `admin/internal/data/ent/schema/**`
- 需要跨阶段上下文时，优先阅读：
  `xkit/NEXT_CONTEXT_HANDOFF_20260510.md`
