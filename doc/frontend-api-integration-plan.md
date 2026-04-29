# xadmin-ui 前端 API 集成计划

本文记录 `xadmin-ui` 基于 Vben 接入 `xadmin` 后端 API 的实施约定、当前状态和分阶段计划。

## 当前口径

- 后端代码已经由 `xkit` 生成到 `D:\GoProjects\XAdmin\xadmin`。
- 前端代码位于 `D:\GoProjects\XAdmin\xadmin-ui`。
- 前端不使用 `xkit` 生成页面、store、router 或业务适配代码。
- 前端以 `vbenjs/vue-vben-admin` 为基线，在 `xadmin-ui` 仓库分支上适配 `xadmin` 后端 API。
- `xkit` 只负责后端代码、OpenAPI 和 TypeScript API 客户端生成。
- TypeScript API 生成代码允许写入 Vben 应用内部，但必须和手写业务适配层隔离。
- 如本文历史内容、旧路径或旧项目名与上述口径冲突，以上述口径为准。

## 目标

- 选用 Vben monorepo 中的 `apps/web-antd` 作为实际后台应用。
- 保持 Vben 的登录、权限、菜单、路由和基础布局机制，优先通过 API adapter 适配后端。
- 逐步把登录、用户信息、权限码、菜单和资源 CRUD 接入 `xadmin` 后端。
- 不再让 Vben mock API 作为 `src/api/core/*` 的主实现。
- 不直接修改 `src/api/generated` 下的生成文件。

## 关键目录

```text
D:\GoProjects\XAdmin
  xadmin\                          <- xkit 已生成的 Go 后端
    api\protos\                    <- proto 源码
    api\buf.vue.admin.typescript.gen.yaml
    cmd\server\assets\openapi.yaml <- 生成的 OpenAPI 文档
    configs\server.yaml            <- REST 默认 :7788

  xadmin-ui\                       <- Vben 前端仓库
    apps\web-antd\                 <- 选定前端应用
      src\api\generated\           <- Buf 生成的 TS API 客户端
      src\api\xadmin\              <- 手写后端适配层
      src\api\core\                <- Vben 现有 API 入口的薄封装
      src\router\                  <- 路由、菜单、权限接入点
      src\store\auth.ts            <- 登录流程入口
      vite.config.ts               <- dev proxy 配置
      .env.development             <- 开发环境 API URL 与 mock 开关

  xkit\
    doc\frontend-api-integration-plan.md
    examples\xadmin\
```

## 当前状态

- `xadmin-ui` 已切到集成分支：`xadmin-api-integration`。
- Vben 基线已经合入 `xadmin-ui`，保留 `xadmin-ui/.git`。
- 生成的 TS API 文件位于：

  ```text
  D:\GoProjects\XAdmin\xadmin-ui\apps\web-antd\src\api\generated\admin\service\v1\index.ts
  ```

- `xadmin/api/buf.vue.admin.typescript.gen.yaml` 输出路径为：

  ```text
  ../../xadmin-ui/apps/web-antd/src/api/generated
  ```

- `apps/web-antd` 包信息：
  - 包名：`@vben/web-antd`
  - 运行：`pnpm -F @vben/web-antd run dev`
  - 类型检查：`pnpm -F @vben/web-antd run typecheck`
  - 构建：`pnpm -F @vben/web-antd run build`
- 第一阶段前端适配已完成：
  - `apps/web-antd/.env.development` 已关闭 `VITE_NITRO_MOCK`。
  - `apps/web-antd/vite.config.ts` 已把 `/api` 代理到 `http://localhost:7788`。
  - `apps/web-antd/src/api/xadmin` 已新增 request handler、generated clients、auth、user、portal adapter。
  - `apps/web-antd/src/api/core/auth.ts`、`user.ts`、`menu.ts` 已改为转发到 xadmin adapter。
  - `apps/web-antd/src/preferences.ts` 已切到 `backend` access mode。
  - `apps/web-antd/src/views/_core/authentication/login.vue` 已去掉 Vben mock 账号下拉、滑块验证码和其他非当前登录方式。
  - `xadmin/internal/server` 已增加临时 manual HTTP service 注册，用于闭环验证 login、me、routes、perm-codes、initial-context。
  - `pnpm -F @vben/web-antd run typecheck` 已通过。
- 第一轮资源页重构已开始：
  - `apps/web-antd/src/api/xadmin/users.ts` 已封装 `UserService` 的 list/create/update/delete/password API。
  - `apps/web-antd/src/views/system/user/index.vue` 已新增用户管理页面，页面只依赖手写 xadmin adapter，不直接调用 generated client。

## 生成代码边界

- `apps/web-antd/src/api/generated` 只放 Buf 生成代码。
- 不手工修改 `src/api/generated` 下文件。
- 不在生成代码里写 Vben request、store、router 或 UI 逻辑。
- 手写适配层放在 `apps/web-antd/src/api/xadmin`。
- 现有 Vben 调用方继续从 `src/api/core/*` 或 `#/api` 导入，`core/*` 只做薄封装。
- Vue 页面组件不直接调用 generated client。

## 后端 API 事实

后端 REST 配置见 `xadmin/configs/server.yaml`：

```text
rest.addr: ":7788"
```

OpenAPI 文档位于：

```text
xadmin/cmd/server/assets/openapi.yaml
```

关键端点：

```text
POST /admin/v1/login
POST /admin/v1/logout
POST /admin/v1/refresh-token
GET  /admin/v1/me
GET  /admin/v1/routes
GET  /admin/v1/perm-codes
GET  /admin/v1/initial-context
```

生成客户端由 `protoc-gen-typescript-http` 输出，形态为：

- 导出类型和 `create<Service>Client(handler)`。
- 生成客户端不直接依赖 axios 或 Vben `requestClient`。
- 前端需要提供 `{ path, method, body }` 到 Vben request 的桥接函数。

## 前端运行配置

第一阶段接真实后端：

```text
VITE_GLOB_API_URL=/api
VITE_NITRO_MOCK=false
```

Vite dev proxy：

```text
/api -> http://localhost:7788
rewrite: remove /api prefix
```

原因：

- 浏览器访问 `/api/admin/v1/...`。
- Vite rewrite 后实际请求 `http://localhost:7788/admin/v1/...`。
- generated client 的 path 形如 `admin/v1/login`，adapter 会补前导 `/`。

## 适配层结构

新增或维护：

```text
apps/web-antd/src/api/xadmin/
  request-handler.ts       <- generated client 到 Vben requestClient 的桥接
  clients.ts               <- 统一创建 generated service clients
  auth.ts                  <- login/logout/refresh/access-codes adapter
  user.ts                  <- UserInfo adapter
  users.ts                 <- UserService CRUD adapter
  portal.ts                <- menus/routes/permission codes adapter
  index.ts                 <- 手写导出入口
```

现有 Vben API 入口改为薄封装：

```text
src/api/core/auth.ts  -> 调用 src/api/xadmin/auth.ts
src/api/core/user.ts  -> 调用 src/api/xadmin/user.ts
src/api/core/menu.ts  -> 调用 src/api/xadmin/portal.ts
src/api/index.ts      -> 继续导出 ./core，可额外导出 ./xadmin
```

## 认证适配

Vben 登录仍接收：

```text
username?: string
password?: string
```

调用后端时转换为：

```text
grant_type: "password"
client_type: "admin"
username
password
```

后端 `access_token` 映射为 Vben 需要的：

```text
{ accessToken: string }
```

刷新 token 第一阶段先不启用；`preferences.app.enableRefreshToken` 暂保持 Vben 默认配置。后续确认 refresh token 存储和刷新策略后再开启。

## 用户信息适配

Vben 需要 `@vben/types` 的 `UserInfo`：

```text
userId: string
username: string
realName: string
avatar: string
roles?: string[]
desc: string
homePath: string
token: string
```

后端优先使用：

```text
UserProfileService.GetUser
GET /admin/v1/me
```

映射策略：

- `userId` 使用后端 `id` 转字符串。
- `realName` 优先 `nickname`、`realname`、`username`。
- `avatar` 后端为空时返回空字符串。
- `roles` 优先使用后端 `roles`，否则返回空数组。
- `desc` 使用 `description` 或 `remark`。
- `homePath` 第一阶段使用 Vben 默认首页 `/analytics`。
- `token` 使用当前 access token。

## 菜单与权限

当前已切到 Vben `backend` access mode，并接入：

```text
GET /admin/v1/perm-codes
GET /admin/v1/routes
```

后端下发菜单的 `component` 必须稳定匹配 `apps/web-antd/src/views/**/*.vue`，例如：

```text
/dashboard/analytics/index -> apps/web-antd/src/views/dashboard/analytics/index.vue
/system/user/index         -> apps/web-antd/src/views/system/user/index.vue
```

当前 `xadmin/internal/server/manual_http.go` 中的 dashboard 与 system/user 菜单只是临时闭环实现；后续应替换为真实菜单数据。

## 分阶段实施

1. 已完成：建立 `xadmin-ui` 的 Vben 基线和集成分支。
2. 已完成：调整 `.env.development` 和 `vite.config.ts`，关闭 mock 并代理到 `http://localhost:7788`。
3. 已完成：新增 `src/api/xadmin/request-handler.ts` 和 `clients.ts`。
4. 已完成：新增 `src/api/xadmin/auth.ts`、`user.ts`、`portal.ts`。
5. 已完成：将 `src/api/core/auth.ts`、`user.ts`、`menu.ts` 改为调用 xadmin adapter。
6. 已完成：类型检查前端 `pnpm -F @vben/web-antd run typecheck`。
7. 已完成：启动后端和前端，验证登录、用户信息、权限码、菜单端点。
8. 已完成：将 `preferences.app.accessMode` 切到 `backend`，后端临时菜单下发 dashboard 与 system/user。
9. 已完成：新增 `src/api/xadmin/users.ts` 与 `views/system/user/index.vue`，开始 User CRUD 页面重构。
10. 下一步：用真实后端数据验证 User CRUD；再按 Role、Menu、Permission 等资源逐个实现页面。

## 验证命令

后端：

```powershell
cd D:\GoProjects\XAdmin\xadmin
go test ./...
go run ./cmd/server server -config_path ./configs
```

前端：

```powershell
cd D:\GoProjects\XAdmin\xadmin-ui
pnpm install
pnpm -F @vben/web-antd run typecheck
pnpm -F @vben/web-antd run dev
```

## 待确认问题

- 后端运行环境依赖是否已就绪，例如 PostgreSQL、Redis、OSS。
- 登录响应是否始终为 OpenAPI 中的裸 `LoginResponse`，还是运行时会被统一包装为 `{ code, data, message }`。
- `Logout` 是否只依赖 Authorization header，还是需要 refresh token。
- 真实后端菜单 `component` 是否已经和 Vben views 路径一致；当前 manual 菜单仅覆盖 dashboard 与 system/user。
- 后端权限码命名是否和 Vben `v-access` 使用规则一致。
- 文件上传下载走 JSON、预签名 URL 还是二进制流。

## 不做事项

- 不使用 `xkit` 生成前端业务代码。
- 不修改 generated TS 文件。
- 不恢复 React TypeScript 生成模板。
- 不把新代码写到旧路径 `admin-02-ui`。
- 不一次性切换登录、菜单、权限和全部 CRUD。
