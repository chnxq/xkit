# admin example

`examples/admin` 是 `xkit` 当前唯一保留的 source 样例。

它用于演示和验证以下链路：

- `xkit init source` 从原始 `proto/schema` 导入目标工程
- `buf generate` 生成 Go / OpenAPI / TypeScript API 代码
- `ent generate` 生成 Ent 数据层代码
- `xkit gen all` 生成 service / repo / register / bootstrap glue

目录说明：

- `api/`：原始 proto 与 Buf 配置
- `schema/`：原始 Ent schema
- `admin-config/admin.yaml`：当前可用的手工生成配置样例

推荐入口脚本：

```powershell
cd D:\GoProjects\XAdmin\xkit
.\examples\generateAll.ps1 -ProjectName admin -SkipGoGetUpdateAll
```

说明：

- 该样例以当前工作区 `admin` 项目的真实 `proto/schema` 为基线同步。
- `api/gen/` 不属于 source 样例输入，因此不会保留在本目录中。
