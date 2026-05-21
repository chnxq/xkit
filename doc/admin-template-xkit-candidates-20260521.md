# admin / xkit-template / xkit 可回归部分分析

## 1. 范围

本次只分析三部分：

- `D:\GoProjects\XAdmin\xkit`
- `D:\GoProjects\XAdmin\xkit-template`
- `D:\GoProjects\XAdmin\admin`

目标是判断 `admin` 中由 Codex 后续补上的代码，哪些已经稳定到足以：

1. 回归到静态模板 `xkit-template`
2. 回归到生成器 `xkit`
3. 暂时继续留在目标项目 `admin`

这里的判断标准不是“现在能不能复制”，而是：

- 是否和资源数量、proto/ent schema 直接相关
- 是否属于启动骨架、平台基础设施、通用后台能力
- 是否已经在多个点重复出现，继续手改会持续漂移
- 是否会污染模板边界，导致模板重新变成业务仓库

## 2. 结论总览

### 2.1 应优先回归到 `xkit-template`

这类内容的共同点是：

- 不是资源级 CRUD 代码
- 不是由 proto / schema 自动推导出来的
- 明显属于后台项目的稳定基础设施或稳定手写骨架

建议优先回归：

1. 登录认证与验证码骨架
2. viewer 注入与 token 校验中间件
3. 数据库审计日志写入中间件骨架
4. 手写 HTTP 服务中的“稳定公共部分”
5. HTTP / gRPC server options 中已经稳定的鉴权装配

### 2.2 应优先回归到 `xkit`

这类内容的共同点是：

- 已经出现“生成后再手改”的漂移
- 逻辑具有清晰生成规律
- 如果继续留在 `admin`，每次重新生成都会再次漂移

建议优先回归：

1. `bootstrap_generated_servers.tmpl` 的当前落后问题
2. `GeneratedData` 的 provider/accessor 生成
3. repo 层时间过滤能力（尤其 `created_at` 和 `BETWEEN`）
4. generated bootstrap 的扩展 hook

### 2.3 暂时继续留在 `admin`

这类内容的共同点是：

- 强业务语义
- 依赖当前后台产品的菜单、权限、分析口径、默认数据
- 即使看起来“通用”，实际上在别的项目里也很容易失真

建议继续留在 `admin`：

1. 默认租户/组织/岗位/角色/用户 seed 数据
2. 当前分析页统计口径
3. 权限审计日志的快照结构和 enrich 逻辑
4. 当前 fallback 菜单树

## 3. 明显存在的漂移点

### 3.1 `manual_http.go` 已经不是“保留给项目自己写一点点”

`xkit-template` 当前文件：

- `xkit-template/internal/server/manual_http.go`

内容基本只有：

- `RegisterManualHTTPServices(...)`
- 空实现

而 `admin` 当前文件：

- `admin/internal/server/manual_http.go`

已经承载了大量稳定能力：

- 登录
- 刷新 token
- 注销
- 验证码获取
- 用户资料
- 修改密码
- 导航加载
- 权限码加载
- 初始上下文
- 分析页数据
- 菜单同步入口

这说明当前模板边界过粗。`manual_http.go` 被 `template.yaml` 标成 `preserve`，导致：

- 模板端后续演进无法同步到目标项目
- 所有公共手写 HTTP 能力都会继续堆在 `admin`
- `xkit-template` 永远只有占位符，`admin` 永远是事实模板

### 3.2 `xkit-template` 的 server options 仍落后于 `admin`

当前差异非常明确：

- `xkit-template/internal/server/http_options.go`
- `admin/internal/server/http_options.go`

`admin` 已经加入：

- `authViewerMiddleware(data)`
- db logging 装配后的完整链路

同样：

- `xkit-template/internal/server/grpc_options.go`
- `admin/internal/server/grpc_options.go`

`admin` 已经把鉴权 viewer middleware 插到 gRPC middleware 链。

这类差异不是业务逻辑差异，而是稳定基础设施差异，应回归模板。

### 3.3 `xkit` 的 `bootstrap_generated_servers.tmpl` 已落后于 `admin` 当前真实需要

当前模板：

- `xkit/internal/codegen/template/bootstrap_generated_servers.tmpl`

仍然缺少几个 `admin` 已经证明必要的点：

1. `GeneratedData` 上保存 `AppContext`
2. `NewGRPCServer(appCtx, services, data)` 这类带 `data` 的调用方式
3. generated data 初始化后的扩展 hook

而 `admin/internal/bootstrap/generated_servers.gen.go` 已经明显不是当前模板直接生成出来的结果。

这类问题不应继续靠手改 generated 文件解决，应该回到 `xkit`。

### 3.4 审计日志 repo 的“时间过滤补丁”暴露出生成器能力缺口

当前 `admin` 中有两组非常典型的补丁：

- `admin/internal/data/repo/audit_log_time_filter_ext.go`
- `admin/internal/data/repo/audit_log_repo_wrappers_ext.go`

其核心作用是：

- 处理 `created_at`
- 处理 `GT / GTE / LT / LTE / BETWEEN`
- 对 generated repo 的 `List` 做一次前置包装

这不是某个业务独有逻辑，而是 repo 生成器暂时不支持 `Time` 类型过滤与区间过滤造成的补丁。

因此它更适合回归 `xkit`，而不是长期留在 `admin`。

## 4. 适合回归到 `xkit-template` 的内容

## 4.1 登录、token、验证码、viewer 鉴权

建议回归文件：

- `admin/internal/server/auth_support_ext.go`
- `admin/internal/server/auth_password_ext.go`
- `admin/internal/server/captcha_ext.go`
- `admin/internal/server/viewer_auth.go`

理由：

- 这些代码属于后台管理系统的稳定基础设施
- 它们不依赖资源数量，也不依赖某个资源的 schema
- 它们直接影响 `http_options.go` / `grpc_options.go` / `manual_http.go` 的装配方式
- 继续留在 `admin`，模板就无法提供完整可用的登录骨架

但要注意：

- 这更像“admin 模板能力”，不是完全通用的服务模板能力
- 如果未来 `xkit-template` 想同时支持“普通服务模板”和“后台管理模板”，最好拆模板变体，而不是继续让一个模板承载全部场景

建议落地方式：

1. 把这部分代码放回 `xkit-template/internal/server/`
2. 同步修改：
   - `xkit-template/internal/server/http_options.go`
   - `xkit-template/internal/server/grpc_options.go`
   - `xkit-template/internal/server/http.go`
   - `xkit-template/internal/server/grpc.go`
3. 让模板默认具备：
   - token 解析/签名
   - viewer 构建
   - captcha
   - 公共登录接口骨架

## 4.2 数据库审计日志中间件骨架

建议回归文件：

- `admin/internal/bootstrap/db_logging_ext.go`

建议继续保留接口挂点：

- `internal/server/db_logging.go`

理由：

- 这是明显的平台基础设施
- 它和当前 `rest-enable_db_logging` 配置是同一层
- 其依赖的是审计 log repo writer、HTTP transport、viewer、请求头解析、IP/Geo 信息
- 这些都属于后台平台通用能力，不属于具体业务资源

需要拆分的部分：

- 统计/字段填充的公共部分可以模板化
- 某些非常具体的日志字段口径，如果未来出现项目差异，可以继续留 hook

## 4.3 手写 HTTP 服务中的稳定公共部分

`admin/internal/server/manual_http.go` 不应该继续作为一个大杂烩保留文件。

建议拆分成两类：

### 模板拥有

- 登录服务骨架
- 用户资料服务骨架
- 获取 captcha
- 初始上下文装配骨架
- 基于用户角色/权限取菜单与权限码的基础流程

### 项目保留

- `defaultNavigationRoutes()` 这种当前项目菜单 fallback
- `GetAnalyticsDashboard()` 中当前分析页统计口径
- 当前专门的菜单同步 HTTP 入口是否保留

建议模板层把 `manual_http.go` 改成：

- 模板拥有的固定文件，例如：
  - `internal/server/auth_http.go`
  - `internal/server/profile_http.go`
  - `internal/server/portal_http.go`
- 项目保留的 hook 文件，例如：
  - `internal/server/manual_http_ext.go`

也就是说，不能继续 `preserve` 整个 `manual_http.go`。

## 4.4 `template.yaml` 的 preserve 边界需要收缩

当前：

- `xkit-template/template.yaml`

把这些文件设成 preserve：

- `internal/server/manual_http.go`
- `internal/server/options.go`
- `internal/data/bootstrap/data.go`
- `internal/data/bootstrap/resources.go`

其中问题最大的是：

- `manual_http.go`

建议调整方向：

- 保留更小粒度的 hook 文件
- 模板重新接管稳定公共实现

否则回归模板没有意义，因为模板永远无法把公共能力再同步回项目。

## 5. 适合回归到 `xkit` 的内容

## 5.1 `bootstrap_generated_servers.tmpl` 需要补齐当前真实生成边界

建议修改：

- `xkit/internal/codegen/template/bootstrap_generated_servers.tmpl`

至少补齐：

1. `GeneratedData` 增加 `AppContext *app.AppCtx`
2. `NewGeneratedData()` 初始化 `AppContext`
3. `components.Servers()` 调用：
   - `server.NewHTTPServer(appCtx, components.Services.HTTP(), components.Data)`
   - `server.NewGRPCServer(appCtx, components.Services.GRPC(), components.Data)`
4. 增加 generated data 初始化后的 hook

推荐形式：

```go
data.afterInit()
```

并配套生成一个一次性扩展文件，例如：

- `internal/bootstrap/generated_data_ext.go`

默认 no-op：

```go
func (data *GeneratedData) afterInit() {}
```

这样以后再有 repo wrapper、额外 provider 注入，不需要再碰 generated 文件。

## 5.2 生成 `GeneratedData` 的 repo provider/accessor

当前 `admin/internal/bootstrap/generated_data_ext.go` 的大量内容其实是在补模板与 generated data 之间的桥：

- `UserRepoProvider()`
- `UserCredentialRepoProvider()`
- `ApiAuditLogRepoProvider()`
- `GetAppCtx()`
- 以及若干 repo 向窄接口的适配

其中最适合回归 `xkit` 的部分是：

- 为每个 generated repo 自动生成 provider 方法
- 自动生成 `GetAppCtx()`

这样模板里的手写公共代码就可以依赖稳定命名规则，而不必在每个目标项目里再写一层重复 accessor。

建议生成结果类似：

```go
func (data *GeneratedData) UserRepoProvider() repo.UserRepo { ... }
func (data *GeneratedData) MenuRepoProvider() repo.MenuRepo { ... }
func (data *GeneratedData) GetAppCtx() *app.AppCtx { ... }
```

之后像 `MenuNavigationReader`、`RolePermissionReader`、`PermissionCodeReader` 这类更窄的语义接口，可以在模板公共代码里再做 type assertion，不一定要每个项目手写一个 semantic accessor。

## 5.3 repo 生成器增加 `Time` 过滤能力

建议修改：

- `xkit/internal/codegen/runner.go`
- `xkit/internal/codegen/template/repo_file.tmpl`

当前生成器已经支持：

- `String`
- `Enum`
- `Uint`
- `Int`

但对 audit log 类 repo 缺少：

- `Time`
- `BETWEEN`
- `GT/GTE/LT/LTE`

这直接导致 `admin` 需要补：

- `audit_log_time_filter_ext.go`
- `audit_log_repo_wrappers_ext.go`

这类补丁很典型，说明生成器缺能力，而不是项目有特殊性。

建议优先让 generated repo 对 `created_at` 这类时间字段原生支持：

- 单值比较
- 区间比较
- RFC3339 / 日期 / 时间戳解析

这样 audit log repo 的 ext 代码会明显变少。

## 5.4 给 generated bootstrap 留正式扩展点

当前 `admin` 已经出现这种需求：

- generated repo 初始化后，要包一层 wrapper
- generated data 初始化后，要暴露一些额外能力

如果 `xkit` 不提供正式 hook，后续还会继续出现“生成后手改 generated 文件”。

建议 `xkit` 提供两类扩展点：

1. data 初始化后 hook
2. services 初始化后 hook

例如：

```go
data.afterInit()
services.afterInit(data)
```

扩展文件仍然一次性生成、后续不覆盖。

这样 `WrapAuditLogRepos()` 之类逻辑就不必继续靠漂移实现。

## 6. 暂时建议继续留在 `admin` 的内容

## 6.1 `default_data_ext.go`

文件：

- `admin/internal/data/bootstrap/default_data_ext.go`

虽然它很大，也很稳定，但当前不建议回归到 `xkit`。

原因：

- 默认租户、组织、岗位、角色、用户、默认密码，本质上是业务初始化数据
- 它依赖当前后台产品的角色模型、权限模型、组织模型
- 它包含当前项目自己的默认值与默认语义

更合理的边界是：

- 模板保留数据初始化 hook
- 具体 seed 内容继续由项目自己持有

如果未来明确所有 admin-family 项目都需要同一套默认数据，再考虑回归模板，而不是回归生成器。

## 6.2 当前分析页统计口径

文件：

- `admin/internal/server/manual_http.go`
- `admin/internal/data/repo/api_audit_log_repo_ext.go`

当前这套统计口径包括：

- TotalAccesses
- CurrentAccesses
- TotalDownloads
- CurrentDownloads
- TotalUsages
- CurrentActives
- 访问趋势
- 月访问分布
- 来源分布
- 业务分布
- 平台分布

这类逻辑和当前前端分析页强绑定，且带明显产品口径，不建议进入 `xkit`。

是否进入模板要看模板是否明确定位为“带 dashboard 的 admin 模板”。在这个定位没有收紧前，建议继续留在 `admin`。

## 6.3 权限审计日志的 enrich/snapshot 逻辑

文件：

- `admin/internal/data/repo/permission_audit_log_repo_ext.go`

其中这些逻辑业务味很重：

- old/new value snapshot 结构
- target name 推断
- operator name enrich
- target type 到实体的映射

这不是生成器层面的稳定规律，也不是纯模板骨架，更适合继续留在 `admin`。

## 6.4 fallback 菜单树

文件：

- `admin/internal/server/manual_http.go`

其中的：

- `defaultNavigationRoutes()`

直接写死了当前前端路由、图标、布局、页面路径。

这类内容不适合进入 `xkit`。

如果未来仍需要默认菜单能力，更合理的做法是：

- 由 `menu` 数据和 `SyncDefaultNavigation()` 提供默认菜单
- HTTP 层只负责取菜单，不再内嵌整棵 fallback 树

## 7. 优先级建议

## P0：先改边界，不先搬大文件

先做：

1. `xkit-template/template.yaml`
   - 收缩 `manual_http.go` 的 preserve 边界
2. `xkit/internal/codegen/template/bootstrap_generated_servers.tmpl`
   - 补齐 `AppContext`
   - 补齐 `NewGRPCServer(..., data)`
   - 增加 after-init hook
3. `xkit` repo 生成器
   - 加 `Time` filter 支持

原因：

- 这三步改完，后续才有条件“回归模板/回归生成器”
- 否则只是把文件搬回去，下一次又会漂移

## P1：回归模板的基础设施

接着做：

1. 认证/验证码/token/viewer middleware
2. HTTP/gRPC 鉴权 options 装配
3. db logging middleware 基础骨架

## P2：拆分 `manual_http.go`

最后再做：

1. 把稳定公共能力拆进模板
2. 把项目特有能力留在 `manual_http_ext.go`
3. 逐步移除 `defaultNavigationRoutes()` 这类 fallback

## 8. 最终建议

当前最值得执行的不是“把 admin 里所有手写代码都回归”，而是先承认并固化下面这个事实：

- `xkit-template` 应负责后台项目的稳定手写骨架
- `xkit` 应负责真正可推导的 generated glue
- `admin` 只保留业务初始化数据、业务统计口径、业务特化查询逻辑

按这个边界看，本轮最明确的回归候选是：

### 回归 `xkit-template`

- `auth_support_ext.go`
- `auth_password_ext.go`
- `captcha_ext.go`
- `viewer_auth.go`
- `db_logging_ext.go`
- `http_options.go` / `grpc_options.go` 的稳定鉴权装配
- `manual_http.go` 中的稳定公共部分

### 回归 `xkit`

- `bootstrap_generated_servers.tmpl`
- generated data accessors/provider methods
- generated bootstrap hook
- repo `Time` filter / `BETWEEN`

### 保留 `admin`

- `default_data_ext.go`
- `api_audit_log_repo_ext.go` 中分析统计口径
- `permission_audit_log_repo_ext.go`
- `defaultNavigationRoutes()`

## 9. 下一步落地顺序

建议后续按下面顺序实施：

1. 先改 `xkit`：
   - `bootstrap_generated_servers.tmpl`
   - repo time filters
2. 再改 `xkit-template`：
   - 调整 `template.yaml` preserve
   - 回归 auth/viewer/db logging 稳定骨架
3. 最后收缩 `admin`：
   - 删除已被模板吸收的重复文件
   - 把 `manual_http.go` 拆成模板部分和项目扩展部分

