# xdev API Layer

本目录是用于验证 `xkit` 新增单个 CRUD 资源能力的最小 API 输入样例。

当前包含：

- `protos/device/service/v1/device.proto`
  - 设备领域 CRUD proto
- `protos/xdev/service/v1/i_device.proto`
  - `xdev` 聚合包装 service
- `buf*.yaml`
  - 与 `examples/admin/api` 对齐后改成独立 `xdev` 外壳的最小 buf 配置

这个样例当前目标不是直接代表一个完整项目，而是作为“向 `admin` 增加一个独立 CRUD 资源”的受控输入集合。
