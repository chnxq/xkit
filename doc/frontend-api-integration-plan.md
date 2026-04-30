# xadmin-ui 前端 API 集成计划

本文记录 `xadmin-ui` 基于 Vben 接入 `xadmin` 后端 API 的实施约定、当前状态和分阶段计划。

## 当前口径

- 后端代码已经由 `xkit` 生成到 `D:\GoProjects\XAdmin\xadmin`。
- 前端代码位于 `D:\GoProjects\XAdmin\xadmin-ui`，当前集成分支为 `xadmin-api-integration`。
- 前端以 `vbenjs/vue-vben-admin` 的 `apps/web-antd` 为基线，在 Vben 上适配 `xadmin` 后端 API。
- 前端不使用 `xkit` 生成页面、store、router 或业务适配代码。
- `xkit` 只负责后端代码、OpenAPI 和 TypeScript API 客户端生成。
- 生成的 TypeScript API 客户端位于 `apps/web-antd/src/api/generated/admin/service/v1`，该目录只允许生成器写入，不手工修改。
- 手写适配层位于 `apps/web-antd/src/api/xadmin`，页面只能依赖手写 adapter，不直接调用 generated client。
- 如本文历史内容与以上口径冲突，以上述口径为准。

## 目标

- 保持 Vben 的登录、权限、菜单、路由和基础布局机制。
- 通过 `src/api/xadmin` 适配 `xadmin` 后端，逐步替换 Vben mock API。
- 先完成登录、用户信息、权限码、菜单和基础资源 CRUD，再扩展到组织、岗位、租户等资源页面。
- 保持 `generated TS client -> xadmin adapter -> Vben core/pages` 的边界；临时绕过 generated client 的逻辑必须封装在 adapter 内，不能进入页面。

## 关键目录

```text
D:\GoProjects\XAdmin
  xadmin\                          <- xkit 已生成的 Go 后端
    api\protos\
    api\buf.vue.admin.typescript.gen.yaml
    cmd\server\assets\openapi.yaml
    configs\server.yaml            <- REST 默认 :7788

  xadmin-ui\                       <- Vben 前端仓库
    apps\web-antd\
      src\api\generated\           <- Buf 生成的 TS API 客户端
      src\api\xadmin\              <- 手写后端适配层
      src\api\core\                <- Vben 现有 API 入口薄封装
      src\views\system\            <- 用户、角色、菜单页面
      src\views\app\permission\    <- 权限点、角色路由兼容页面
      vite.config.ts
      .env.development

  xkit\
    doc\frontend-api-integration-plan.md
```

## 当前状态

- `apps/web-antd/.env.development` 已关闭 mock，并使用 `/api` 作为前端 API URL。
- `apps/web-antd/vite.config.ts` 已把 `/api` 代理到 `http://localhost:7788`。
- `apps/web-antd/src/preferences.ts` 已切到 `backend` access mode。
- `apps/web-antd/src/views/_core/authentication/login.vue` 已去掉当前无关的 mock 登录辅助 UI。
- `apps/web-antd/src/api/core/auth.ts`、`menu.ts`、`user.ts` 已转发到 `xadmin` adapter。
- `apps/web-antd/src/api/xadmin` 已建立 request handler、generated clients、auth、user、portal、users、roles、menus、permissions 等 adapter。
- 已完成基础页面：
  - `apps/web-antd/src/views/system/user/index.vue`
  - `apps/web-antd/src/views/system/role/index.vue`
  - `apps/web-antd/src/views/system/menu/index.vue`
  - `apps/web-antd/src/views/app/permission/permission/index.vue`
  - `apps/web-antd/src/views/app/permission/role/index.vue`
- 前端类型检查已通过：`pnpm -F @vben/web-antd run typecheck`。

## 适配层结构

```text
apps/web-antd/src/api/xadmin/
  request-handler.ts       <- generated client 到 Vben requestClient 的桥接
  clients.ts               <- 统一创建 generated service clients
  paging.ts                <- PagingRequest/filterExpr 查询参数构造与列表请求封装
  auth.ts                  <- login/logout/refresh/access-codes adapter
  user.ts                  <- UserInfo adapter
  users.ts                 <- UserService CRUD adapter
  roles.ts                 <- RoleService CRUD adapter
  menus.ts                 <- MenuService CRUD adapter
  permissions.ts           <- Permission/PermissionGroup CRUD adapter
  portal.ts                <- menus/routes/permission codes adapter
  index.ts                 <- 手写导出入口
```

当前列表查询统一使用 `filterExpr`，不再使用旧 `query` JSON 字符串。由于当前 generated TS client 对 repeated `filterExpr.conditions` 的序列化有缺陷，列表请求暂由 `paging.ts` 在手写 adapter 层直接构造查询参数；页面仍不直接访问后端，也不修改 generated 文件。生成器修复后，只需要回收 `paging.ts` 的临时直连列表逻辑。

## 后端 API 事实

后端 REST 默认地址：

```text
http://localhost:7788
```

开发环境前端访问：

```text
/api/admin/v1/...
```

Vite 代理会转发为：

```text
http://localhost:7788/admin/v1/...
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
GET  /admin/v1/users
GET  /admin/v1/roles
GET  /admin/v1/menus
GET  /admin/v1/permissions
GET  /admin/v1/permission-groups
```

## 菜单与路由

当前后端菜单数据需要稳定匹配 Vben 视图路径，例如：

```text
/dashboard/analytics/index        -> apps/web-antd/src/views/dashboard/analytics/index.vue
/system/user/index                -> apps/web-antd/src/views/system/user/index.vue
/system/role/index                -> apps/web-antd/src/views/system/role/index.vue
/system/menu/index                -> apps/web-antd/src/views/system/menu/index.vue
app/permission/permission/index   -> apps/web-antd/src/views/app/permission/permission/index.vue
app/permission/role/index         -> apps/web-antd/src/views/app/permission/role/index.vue
```

实际数据库菜单已经呈现 `/system/user`、`/system/role`、`/system/menu`，并包含权限管理路由：

```text
PermissionManagement        /permission      BasicLayout
PermissionPointManagement   codes            app/permission/permission/index.vue
RoleManagement              roles            app/permission/role/index.vue
```

## 分阶段实施

1. 已完成：建立 `xadmin-ui` 的 Vben 基线和 `xadmin-api-integration` 分支。
2. 已完成：调整 `.env.development` 和 `vite.config.ts`，关闭 mock 并代理到后端。
3. 已完成：新增 `src/api/xadmin/request-handler.ts` 和 `clients.ts`。
4. 已完成：新增 `auth.ts`、`user.ts`、`portal.ts`。
5. 已完成：将 `src/api/core/auth.ts`、`user.ts`、`menu.ts` 改为调用 xadmin adapter。
6. 已完成：切换 `preferences.app.accessMode` 到 `backend`。
7. 已完成：验证登录、用户信息、权限码、菜单端点。
8. 已完成：新增 `users.ts` 与 `views/system/user/index.vue`。
9. 已完成：新增 `roles.ts` 与 `views/system/role/index.vue`。
10. 已完成：新增 `menus.ts` 与 `views/system/menu/index.vue`。
11. 已完成：修复后端菜单适配到 Vue Router 时空 `alias` 导致的 `aliases is not iterable`。
12. 已完成：修复后端 enum 映射导致状态值落到默认值的问题。
13. 已完成：新增 `permissions.ts`，接入 Permission 与 PermissionGroup adapter。
14. 已完成：新增权限点管理页面与 `/permission/roles` 路由兼容页面。
15. 已完成：列表筛选从旧 `query` JSON 切换到 `filterExpr`。
16. 下一步：继续 OrgUnit、Position、Tenant 等资源页面。
17. 后续：补权限分配、角色授权、菜单/API 绑定等更完整交互。

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
pnpm -F @vben/web-antd run typecheck
pnpm -F @vben/web-antd run dev
```

## 待确认问题

- 后端当前运行实例是否已更新到支持 `filterExpr.conditions` repeated message 绑定的 `xkitmod` 版本。
- 后端权限码命名是否和 Vben `v-access` 使用规则完全一致。
- 真实菜单 `component` 是否会持续保持与 Vben `views` 路径一致。
- 文件上传下载最终采用 JSON、预签名 URL 还是二进制流。

## 不做事项

- 不使用 `xkit` 生成前端业务代码。
- 不手工修改 `apps/web-antd/src/api/generated` 下的文件。
- 不恢复旧 React TypeScript 生成模板。
- 不把新代码写到旧路径 `admin-02-ui`。
- 不一次性切换全部资源页面，按模块逐步重构和验证。
