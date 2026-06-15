# admin-v2-config

`admin-v2-config/admin.yaml` 以 `examples/admin/admin-config/admin.yaml` 为基线，
但目标不是“复制一份旧配置”，而是尽量让配置与当前 `schema` 和 `api/protos` 的资源现状对齐，
从而让 `xkit` 在新的 admin 项目中自动生成更完整的框架代码。

当前版本的设计原则：

1. 保留旧 `admin-config` 中已经验证过的手工增强配置。
   例如 `task`、`task_group`、`internal_message_recipient`、`user.EditUserPassword` 等。

2. 对现有 `admin.service.v1` wrapper 已覆盖的资源，补齐更丰富的过滤字段、树结构、聚合关系和 service-only 入口。

3. 对纯 API 入口型服务（如 `authentication`、`oauth`、`social_auth`、`user_profile`），
   先纳入 config，用于生成 service/register 框架，不强制启用 repo CRUD。

4. 对“schema 已有，但 proto 侧没有可直接生成的 service”的资源，当前不强行纳入。
   这类资源后续应优先补齐 proto service / admin wrapper，再回补到 `admin-v2-config`。

当前仍未自动纳入的典型对象：

- `dict_category_i18n`
- `membership`
- `membership_org_unit`
- `membership_position`
- `membership_role`
- `permission_api`
- `permission_menu`
- `permission_policy`
- `role_metadata`
- `role_permission`
- `user_org_unit`
- `user_position`
- `user_role`

原因不是 schema 缺失，而是当前 `examples/admin/api/protos` 中缺少与之对应、可被 `xkit` 直接绑定的 service 定义。

后续建议：

1. 先补这批资源的 proto service 与 admin wrapper。
2. 再把它们增量回填到 `admin-v2-config/admin.yaml`。
3. 最后再评估是否需要同步更新 `examples/generateAll.ps1` 和 `xkit-helper` 对 `admin-v2-config` 的支持说明。
