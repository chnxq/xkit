# xdev example

`examples/xdev` 用于验证 `xkit` 在已有 `admin` 体系中新增单个 CRUD 功能时的行为。

当前样例围绕 `dev_info` 表构造，目标是：

- 用最小 schema/proto/config 输入验证生成链路
- 尽量贴近 `admin` 当前生成约定
- 在保留业务字段特征的同时，显式补入 `tenant_id`
- 尽量复用 mixin，减少样例层面的手工噪音

当前目录结构：

- `schema/dev_info.go`
- `api/protos/device/service/v1/device.proto`
- `api/protos/xdev/service/v1/i_device.proto`
- `xdev-config/xdev.yaml`
