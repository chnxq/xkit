# xdev example

`examples/xdev` 用于验证 `xkit` 在已有 `admin` 宿主体系中新增单个 CRUD 模块时的行为。

当前样例围绕设备域三张表：

- `dev_info`
- `dev_model`
- `dev_model_type`

目标是：

- 用最小 schema/proto/config 输入验证模块生成链路
- 尽量贴近真实 `admin` 宿主 + `modules/<module>` 目录约定
- 保留中文注释、`json_name`、`tenant_id` 与关联关系
- 沿用 `config -> target-config` 的外部配置工作流，而不是把配置长期放进模块源码目录

当前目录结构：

- `schema/*.go`
- `api/protos/device/v1/*.proto`
- `langs/zh-CN/*.json`
- `langs/en-US/*.json`
- `xdev-config/xdev.yaml`
  - 规范基线配置
- `xdev-target-config/xdev.yaml`
  - 当前用于真实 `admin/modules/xdev` 生成验证的工作配置

当前约定：

- `xkit init module` 默认将配置写到
  - `examples/xdev/xdev-target-config/xdev.yaml`
- `xkit gen module` 默认优先读取
  - `examples/xdev/xdev-target-config/xdev.yaml`
- 以后的生成与重构都以 `xdev-target-config/xdev.yaml` 为准
- 只有当配置变更被验证和认可后，才手工同步回 `xdev-config/xdev.yaml`
- `admin/modules/xdev` 目录只放模块源码、proto、schema 和生成物，不再依赖长期保留的模块内 yaml 配置
- `langs/*` 下的页面与枚举文案由样例侧维护，`xkit` 只负责同步到目标前端代码中

示例命令：

```powershell
& 'D:\GoProjects\XAdmin\xkit\examples\generateModule.ps1' `
  -WorkspaceRoot 'D:\GoProjects\XAdmin' `
  -ModuleName 'xdev' `
  -ServiceName 'xdev' `
  -HostProject 'D:\GoProjects\XAdmin\admin' `
  -TypeScriptRoot 'D:\GoProjects\XAdmin\admin-ui' `
  -ConfigPath 'D:\GoProjects\XAdmin\xkit\examples\xdev\xdev-target-config\xdev.yaml' `
  -CanonicalConfigPath 'D:\GoProjects\XAdmin\xkit\examples\xdev\xdev-config\xdev.yaml'
```
