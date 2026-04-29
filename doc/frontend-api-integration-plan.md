# 前端 API 集成与重构交接

本文记录 `admin-02-ui` 前端接入 XAdmin 后端 API 的当前状态、上下文和后续重构建议。后续如果重新开启上下文，应先阅读本文，再查看文中列出的关键文件。

## 目标

- 以前端仓库 `https://github.com/vbenjs/vue-vben-admin.git` 作为后台管理前端起点。
- 前端代码落地到 Go 后端项目同级目录：`D:\GoProjects\XAdmin\admin-02-ui`。
- 选用 Vben monorepo 中的 `apps/web-antd` 作为实际应用。
- XAdmin 后端项目为：`D:\GoProjects\XAdmin\admin-02`。
- 后端 proto/schema 原始示例数据来自：`D:\GoProjects\XAdmin\xkit\examples\xadmin`。
- TypeScript API 生成代码写入 Vben 应用内部，但与手写业务适配层隔离。
- 后续目标是逐步把 Vben 的登录、用户、菜单、权限和资源 CRUD 接到 `admin-02` 后端，而不是继续使用 Vben mock 接口。

## 当前已完成状态

- `vue-vben-admin` 已复制到 `D:\GoProjects\XAdmin\admin-02-ui`，目标目录不保留上游 `.git`。
- `admin-02-ui` 是 Vben monorepo 结构，包含 `apps/web-antd`、`packages`、`internal`、`pnpm-workspace.yaml`、`pnpm-lock.yaml` 等。
- 只保留 Vue TypeScript 生成模板，已放弃 React 生成模板。
- `xkit init source` 会把 TypeScript Buf 插件输出路径重写到：

  ```text
  <typescript-project>/apps/web-antd/src/api/generated
  ```

- `admin-02/api/buf.vue.admin.typescript.gen.yaml` 当前输出到：

  ```text
  ../../admin-02-ui/apps/web-antd/src/api/generated
  ```

- 已生成的 TS 文件存在：

  ```text
  D:\GoProjects\XAdmin\admin-02-ui\apps\web-antd\src\api\generated\admin\service\v1\index.ts
  ```

- `xkit` 侧相关规则位于：

  ```text
  D:\GoProjects\XAdmin\xkit\internal\sourceimport\importer.go
  D:\GoProjects\XAdmin\xkit\internal\sourceimport\importer_test.go
  D:\GoProjects\XAdmin\xkit\examples\xadmin\generateAll.ps1
  D:\GoProjects\XAdmin\xkit\examples\xadmin\api\buf.vue.admin.typescript.gen.yaml
  ```

## 关键目录

```text
D:\GoProjects\XAdmin
  admin-02\                         <- 后端验证项目
    api\protos\                     <- proto 源码
    api\buf.vue.admin.typescript.gen.yaml
    cmd\server\assets\openapi.yaml  <- 生成的 OpenAPI 文档
    configs\server.yaml             <- REST 默认 :7788

  admin-02-ui\                      <- Vben 前端
    apps\web-antd\                  <- 选定前端应用
      src\api\                      <- 前端 API 接入重点目录
      src\router\                   <- 路由、菜单、权限接入点
      src\store\auth.ts             <- 登录流程入口
      vite.config.ts                <- dev proxy 配置
      .env.development              <- 开发环境 API URL 与 mock 开关

  xkit\
    doc\frontend-api-integration-plan.md
    examples\xadmin\
```

## Vben 基线信息

`apps/web-antd/package.json`：

- 包名：`@vben/web-antd`
- 运行：`pnpm -F @vben/web-antd run dev`
- 类型检查：`pnpm -F @vben/web-antd run typecheck`
- 构建：`pnpm -F @vben/web-antd run build`
- `imports` alias：`#/*` 指向 `apps/web-antd/src/*`

根 `package.json`：

- Vben 版本：`5.7.0`
- Node 要求：`^20.19.0 || ^22.18.0 || ^24.0.0`
- pnpm 要求：`>=10.0.0`
- packageManager：`pnpm@10.33.0`

当前沙箱里 `pnpm --version` 因无权限写入用户目录失败，尚未完成前端 `pnpm install` / `typecheck` 验证。后续在正常本机终端验证即可。

## 当前 Vben API 接入点

现有 Vben API 文件：

```text
apps/web-antd/src/api/index.ts
apps/web-antd/src/api/request.ts
apps/web-antd/src/api/core/auth.ts
apps/web-antd/src/api/core/user.ts
apps/web-antd/src/api/core/menu.ts
```

当前调用关系：

- `src/store/auth.ts` 调用 `loginApi`、`logoutApi`、`getUserInfoApi`、`getAccessCodesApi`。
- `src/router/access.ts` 调用 `getAllMenusApi` 获取菜单。
- `src/views/_core/profile/base-setting.vue` 调用 `getUserInfoApi`。
- `src/api/request.ts` 使用 `refreshTokenApi` 处理刷新 token。

Vben 当前默认 API 期望：

- `loginApi` 返回 `{ accessToken: string }`。
- `getUserInfoApi` 返回 `@vben/types` 的 `UserInfo`。
- `getAccessCodesApi` 返回 `string[]`。
- `getAllMenusApi` 返回 `RouteRecordStringComponent[]`。
- `requestClient` 默认用 `codeField: 'code'`、`dataField: 'data'`、`successCode: 0` 解包响应。

`UserInfo` 关键字段来自 `packages/types/src/user.ts` 和 `packages/@core/base/typings/src/basic.d.ts`：

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

## 当前 Vben 运行配置

`apps/web-antd/.env.development` 当前仍是 Vben mock 风格：

```text
VITE_GLOB_API_URL=/api
VITE_NITRO_MOCK=true
```

`apps/web-antd/vite.config.ts` 当前 proxy：

```text
/api -> http://localhost:5320/api
rewrite: remove /api prefix
```

接入 XAdmin 后端时建议调整为：

```text
VITE_NITRO_MOCK=false
VITE_GLOB_API_URL=/api
/api -> http://localhost:7788
rewrite: remove /api prefix
```

理由：

- 后端 REST 地址见 `admin-02/configs/server.yaml`，默认 `rest.addr: ":7788"`。
- 生成 TS 客户端路径不带前导 `/`，例如 `admin/v1/login`、`admin/v1/routes`。
- 前端用 `/api` 作为浏览器同源代理前缀，proxy rewrite 后实际访问后端 `/admin/v1/...`。

## 生成代码边界

只运行 Vue TypeScript 生成：

```powershell
cd D:\GoProjects\XAdmin\admin-02\api
buf generate --template buf.vue.admin.typescript.gen.yaml
```

输出目录：

```text
D:\GoProjects\XAdmin\admin-02-ui\apps\web-antd\src\api\generated
```

边界规则：

- `src/api/generated` 只放 Buf 生成代码。
- 不手工修改 `src/api/generated` 下文件。
- 不在生成代码里写 Vben request、store、router 或 UI 逻辑。
- 手写适配层放在 `src/api/xadmin` 或 `src/api/admin`。
- 前端页面、表格、表单、枚举展示、字段显示转换都属于手写业务代码。

建议使用 `src/api/xadmin`，避免和后端生成包名 `admin` 混淆。

## 生成 TS 客户端形态

当前生成文件：

```text
apps/web-antd/src/api/generated/admin/service/v1/index.ts
```

生成器为 `protoc-gen-typescript-http`，生成文件特点：

- 顶部有 `// @ts-nocheck`。
- 导出大量 `type`。
- 每个 service 导出一个接口和 `create<Service>Client(handler)`。
- 生成客户端不直接依赖 axios，也不直接依赖 Vben `requestClient`。
- 需要前端提供一个 `RequestHandler` 风格的桥接函数：

  ```ts
  type RequestType = {
    path: string;
    method: string;
    body: string | null;
  };
  ```

示例生成服务：

```text
createAuthenticationServiceClient(handler)
createAdminPortalServiceClient(handler)
createUserProfileServiceClient(handler)
createUserServiceClient(handler)
createMenuServiceClient(handler)
createRoleServiceClient(handler)
createPermissionServiceClient(handler)
createTaskServiceClient(handler)
```

关键生成端点：

```text
POST admin/v1/login
POST admin/v1/logout
POST admin/v1/refresh-token
GET  admin/v1/routes
GET  admin/v1/perm-codes
GET  admin/v1/initial-context
```

这些端点源自：

```text
admin-02/api/protos/admin/v1/i_authentication.proto
admin-02/api/protos/admin/v1/i_admin_portal.proto
```

OpenAPI 文档位置：

```text
admin-02/cmd/server/assets/openapi.yaml
```

## 建议的手写适配层结构

建议新增：

```text
apps/web-antd/src/api/xadmin/
  request-handler.ts       <- 生成客户端到 Vben requestClient 的桥接
  clients.ts               <- 统一创建 generated service clients
  auth.ts                  <- login/logout/refresh/access-codes adapter
  user.ts                  <- UserInfo adapter
  portal.ts                <- menus/routes/permission codes/initial context adapter
  crud.ts                  <- 可选：通用 CRUD 帮助函数
  index.ts                 <- 手写导出入口
```

随后让现有 core API 薄封装到新 adapter：

```text
src/api/core/auth.ts  -> 调用 src/api/xadmin/auth.ts
src/api/core/user.ts  -> 调用 src/api/xadmin/user.ts
src/api/core/menu.ts  -> 调用 src/api/xadmin/portal.ts
src/api/index.ts      -> 继续导出 ./core，必要时额外导出 ./xadmin
```

这样可以降低改动面，`src/store/auth.ts`、`src/router/access.ts`、`profile` 页面可以先保持现有 import 不变。

## RequestHandler 设计建议

桥接函数职责：

- 接收生成客户端传入的 `{ path, method, body }`。
- 将 `path` 规范化为带前导 `/` 的 URL。
- 将 `body` 从 JSON string 解析为对象，或者在需要时直接传字符串。
- 调用 Vben `requestClient` 或 `baseRequestClient`。
- 保留 Vben 统一 token header、语言 header、错误提示和响应拦截器。

需要注意：

- 普通业务接口优先使用 `requestClient`，让它负责 `code/data` 解包。
- 刷新 token 如果要绕开自动刷新拦截器，可以继续使用 `baseRequestClient`。
- 生成客户端返回的是 Promise 类型，实际数据形态受 Vben response interceptor 影响，必须用真实后端响应确认。
- 当前后端响应到底是裸 JSON，还是 `{ code, data, message }` 包裹，需要用 `openapi.yaml` 和实际请求确认；这会决定是否继续保留 `defaultResponseInterceptor` 当前配置。

建议实现时先写一个最小桥接，不要提前做复杂抽象。

## 认证匹配重点

Vben 当前登录参数：

```text
username?: string
password?: string
```

XAdmin 生成的 `LoginRequest`：

```text
grant_type: "password" | "client_credentials" | "authorization_code" | "refresh_token" | "implicit"
client_type?: "admin" | "app"
username?: string
password?: string
refresh_token?: string
...
```

XAdmin 生成的 `LoginResponse`：

```text
token_type
access_token
expires_in
refresh_token
refresh_expires_in
scope
id_token
```

适配要求：

- `loginApi` 仍返回 Vben 期望的 `{ accessToken: string }`。
- 调用后端时补齐 `grant_type: "password"`，建议补齐 `client_type: "admin"`。
- 将后端 `access_token` 映射为 Vben `accessToken`。
- 如果启用刷新 token，需要保存和使用 `refresh_token`，并把 `preferences.app.enableRefreshToken` 改为 `true`。
- 当前 Vben 默认 `enableRefreshToken: false`，可以第一阶段先不启用刷新 token，只完成登录、登出、用户信息和菜单。

## 用户信息匹配重点

Vben 需要 `UserInfo`，后端可优先考虑：

```text
UserProfileService.GetUser
GET admin/v1/me
```

该路径源自 `admin-02/api/protos/admin/v1/i_user_profile.proto`，生成在 `createUserProfileServiceClient` 中。适配层输出必须满足：

```text
userId
username
realName
avatar
roles
desc
homePath
token
```

建议策略：

- `userId` 统一转字符串。
- `realName` 优先使用后端昵称、姓名或 username。
- `avatar` 为空时使用 Vben 默认头像即可，也可以返回空字符串让布局兜底。
- `roles` 第一阶段可从权限/角色 API 或用户 DTO 中提取；没有后端字段时先返回空数组，权限以 code/menu 为主。
- `homePath` 默认使用 `/analytics`，或改为后端菜单中的第一个可访问页面。

## 菜单与权限匹配重点

Vben 路由模式由 `preferences.app.accessMode` 决定，默认在 `packages/@core/preferences/src/config.ts` 中是：

```text
accessMode: "frontend"
```

项目覆盖文件：

```text
apps/web-antd/src/preferences.ts
```

若使用后端菜单，应在 `overridesPreferences` 中覆盖：

```text
app.accessMode = "backend" 或 "mixed"
```

相关 Vben 文件：

```text
apps/web-antd/src/router/access.ts
packages/utils/src/helpers/generate-routes-backend.ts
packages/@core/base/typings/src/vue-router.d.ts
```

后端已有 portal API：

```text
GET admin/v1/routes           -> ListRouteResponse { items: MenuRouteItem[] }
GET admin/v1/perm-codes       -> ListPermissionCodeResponse { codes: string[] }
GET admin/v1/initial-context  -> InitialContextResponse { menus, permissions }
```

建议：

- 第一阶段可继续用 Vben frontend routes，只接登录和用户信息，降低变量。
- 第二阶段切到 `backend` 或 `mixed`，用 `GetNavigation` 返回菜单路由。
- 如果使用 `GetInitialContext`，可以一次请求拿菜单和权限码，减少登录后的并发请求。
- 后端 `MenuRouteItem.component` 必须能匹配 Vben `pageMap` 或 `layoutMap`：
  - 布局组件支持 `BasicLayout`、`IFrameView`。
  - 页面组件路径会通过 `normalizeViewPath` 匹配 `apps/web-antd/src/views/**/*.vue`。
  - 不匹配时会落到 not-found，并在 console 输出错误。

## CRUD 与资源页面策略

XAdmin 已生成多个 admin service client，例如 User、Role、Menu、Permission、Task、Dict、File、AuditLog 等。后续做资源页面时建议：

- 每个资源页面先写一个独立 adapter，不直接在 Vue 组件里调用 generated client。
- 列表请求统一处理分页参数：
  - Vben 表格页通常有 current/pageSize。
  - 后端 `pagination_PagingRequest` 支持 `page`、`pageSize`、`query`、`filter`、`orderBy`、`sorting` 等。
- 列表响应通常是 `ListXxxResponse { items, totalSize/total/... }`，需要按实际生成类型确认总数字段。
- 删除、创建、更新返回 `google.protobuf.Empty` 时，前端只关心成功或失败。
- 字段命名以生成 DTO 为准，UI 层再做 camelCase、标签、枚举展示转换。
- 不要把 table/form 的 UI 状态写进 API adapter。

## 文件上传与下载

后端有 `FileTransferService` 和 `OssService` 相关 proto。生成的普通 JSON handler 未必适合二进制上传下载。

后续处理建议：

- JSON CRUD 继续走 generated client + request handler。
- 上传、下载、预签名 URL 走单独 adapter。
- 对 `google.api.HttpBody`、stream、Blob、FormData 的接口，不要强行复用 JSON body 逻辑。
- 需要根据后端实际 REST 路由和 OpenAPI 文档单独实现。

## Mock 与真实后端切换

接真实后端前，需要处理：

```text
apps/web-antd/.env.development
apps/web-antd/vite.config.ts
```

建议第一阶段：

```text
VITE_NITRO_MOCK=false
VITE_GLOB_API_URL=/api
proxy target = http://localhost:7788
rewrite /api -> ""
```

如果后端临时不可启动，可保留 mock 分支，但不要让 mock API 继续作为 `src/api/core/*` 的主实现。更好的方式是：

- `core/*` 永远调用 `xadmin/*` adapter。
- adapter 内部可以在开发期临时兜底 mock，但需要显式标注并后续移除。

## 后端启动与 API 验证上下文

后端启动命令：

```powershell
cd D:\GoProjects\XAdmin\admin-02
go run ./cmd/server server -config_path ./configs
```

REST 默认端口：

```text
http://localhost:7788
```

Swagger/OpenAPI：

```text
admin-02/cmd/server/assets/openapi.yaml
```

`server.yaml` 中已有：

```text
rest.enable_swagger: true
rest.enable_pprof: true
rest.cors.origins: ["*"]
```

注意：后端实际启动可能依赖 PostgreSQL、Redis、认证/权限组件、OSS 等外部服务。前端重构时如果后端启动失败，应先区分是“生成代码问题”还是“运行环境依赖未就绪”。

## 生成与验证流程

后端 API 变化后：

```powershell
cd D:\GoProjects\XAdmin\admin-02\api
buf generate --template buf.gen.yaml
buf generate --template buf.admin.openapi.gen.yaml
buf generate --template buf.vue.admin.typescript.gen.yaml

cd D:\GoProjects\XAdmin\admin-02
go test ./...
```

前端验证：

```powershell
cd D:\GoProjects\XAdmin\admin-02-ui
pnpm install
pnpm -F @vben/web-antd run typecheck
pnpm -F @vben/web-antd run dev
```

如果只验证选定应用，也可以使用根脚本：

```powershell
pnpm run dev:antd
pnpm run build:antd
```

## 推荐实施顺序

1. 调整 `apps/web-antd/.env.development` 和 `vite.config.ts`，关闭 mock，指向 `http://localhost:7788`。
2. 新增 `src/api/xadmin/request-handler.ts` 和 `clients.ts`，让 generated client 能通过 Vben `requestClient` 发请求。
3. 新增 `src/api/xadmin/auth.ts`，完成 `loginApi`、`logoutApi`、`refreshTokenApi`、`getAccessCodesApi` 适配。
4. 将 `src/api/core/auth.ts` 改成调用 `xadmin/auth.ts`，保持 store 层不动。
5. 新增 `src/api/xadmin/user.ts`，完成 `getUserInfoApi` 适配。
6. 将 `src/api/core/user.ts` 改成调用 `xadmin/user.ts`。
7. 先用 frontend access mode 验证登录和用户信息。
8. 再接 `AdminPortalService`，实现权限码和菜单。
9. 根据后端菜单质量决定 `accessMode` 切到 `backend` 还是 `mixed`。
10. 最后开始逐个资源页面接 CRUD。

## 待确认问题

- 后端响应是否统一包裹为 `{ code, data, message }`，还是部分接口直接返回业务 JSON。
- 登录成功时是否需要前端持久化 `refresh_token`，以及刷新 token 的请求体格式。
- `Logout` 是否要求携带 refresh token 或只依赖 Authorization header。
- `UserProfileService.GetUser` 返回字段是否足够映射 Vben `UserInfo`。
- 后端菜单中的 `component` 是否已经与 `apps/web-antd/src/views` 路径一致。
- 后端权限码与 Vben `v-access` / `accessStore.setAccessCodes` 的命名规则是否一致。
- 文件上传下载是否使用 JSON API、预签名 URL，还是二进制流。
- 是否需要多租户切换入口；Vben 当前 preferences extension 已有 `tenantMode` 字段，但尚未接后端。

## 不建议做的事

- 不要修改 `src/api/generated` 下的生成代码。
- 不要恢复 `buf.react.*.typescript.gen.yaml`。
- 不要继续把新代码写到旧的 `apps/admin` 或 `src/generated/api` 路径。
- 不要在 Vue 页面组件里直接调用 generated client。
- 不要为了适配第一个接口就重写 Vben store、router 和 request 包；先用 adapter 保持边界。
- 不要一次性切换登录、菜单、权限和全部 CRUD；应先让登录链路闭环。

## 新上下文启动提示

如果后续重新开启上下文，可从以下检查开始：

```powershell
cd D:\GoProjects\XAdmin
Test-Path admin-02-ui\apps\web-antd
Test-Path admin-02-ui\apps\web-antd\src\api\generated\admin\service\v1\index.ts
Get-Content admin-02-ui\apps\web-antd\src\api\request.ts
Get-Content admin-02-ui\apps\web-antd\src\api\core\auth.ts
Get-Content admin-02-ui\apps\web-antd\vite.config.ts
```

然后优先处理：

```text
apps/web-antd/src/api/xadmin/request-handler.ts
apps/web-antd/src/api/xadmin/auth.ts
apps/web-antd/src/api/core/auth.ts
```
