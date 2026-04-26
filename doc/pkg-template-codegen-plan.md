# pkg 模板代码迁移到 xkit 生成器方案

## 背景

`xkit-template/pkg` 当前承载了从参考项目拆出的公共能力和一部分 XAdmin 业务能力。它不适合长期放在模板仓库里直接复制，原因是：

1. `pkg` 中存在大量硬编码模块路径，例如 `xkit-template-v01/pkg/...`。
2. 部分文件直接依赖目标项目的 `api/gen/...`，并不是通用模板代码。
3. 后续目标项目会修改 `pkg` 代码，生成器必须避免反复覆盖人工修改。
4. 模板仓库应主要保留稳定启动骨架；可选公共包更适合作为 `xkit` 的初始化生成目标。

因此，本阶段将 `xkit-template/pkg` 按依赖边界拆分，第一批迁移为 `xkit gen pkg` 生成的代码。

## 目标

新增独立生成目标：

```text
xkit gen pkg <service> [--project <path>] [--config <path>] [--domain <name>] [--dry-run]
```

并纳入：

```text
xkit gen all <service> ...
```

生成策略：

- 输出目录为目标项目根目录下的 `pkg/`。
- 默认只创建不存在的文件，不覆盖已有 `pkg` 文件。
- 生成时将模板中的 `xkit-template-v01` 模块路径替换为目标项目 `go.mod` 的 module。
- 先迁移不直接依赖 `api/gen` 的通用包。
- 依赖具体业务 proto 的包保留在待迁移清单，后续通过领域模板或资源生成器生成。

## 第一批迁移范围

第一批迁移“通用能力包”，这些包不直接绑定目标项目的 protobuf 类型：

- `pkg/authorizer`
- `pkg/crypto`
- `pkg/eventbus`
- `pkg/entgo/viewer/system_viewer.go`
- `pkg/task`
- `pkg/utils/slice.go`
- `pkg/utils/converter/api.go`
- `pkg/README.md`

说明：

- `authorizer` 依赖权限引擎抽象和外部策略引擎，属于平台能力，可随项目初始化生成。
- `crypto`、`eventbus`、`task`、`utils` 属于通用工具。
- `entgo/viewer/system_viewer.go` 可作为系统任务 viewer 的基础实现。
- `utils/converter/api.go` 只依赖 OpenAPI 路径和 operationID 字符串，不依赖具体业务 proto。

## 暂缓迁移范围

以下文件或目录直接依赖具体 `api/gen`，第一阶段不进入通用 `pkg` 生成：

- `pkg/constants`
- `pkg/jwt`
- `pkg/metadata`
- `pkg/middleware/auth`
- `pkg/middleware/ent`
- `pkg/middleware/logging`
- `pkg/oss`
- `pkg/entgo/viewer/user_viewer.go`
- `pkg/utils/converter/menu.go`
- `pkg/lua`

处理策略：

- 强业务常量和默认数据后续放到 `xkit gen xadmin-domain` 或类似领域生成目标。
- 认证、审计、metadata、viewer、OSS 这类依赖具体 proto 的包，后续由资源配置或领域配置驱动生成。
- `lua` 依赖 `eventbus`、`oss` 等多个包，且 `oss` 当前绑定 storage proto，因此等 `oss` 解耦后再迁移。

## 生成器设计

在 `xkit` 中新增包：

```text
internal/codegen/pkgtemplate/
  files/
    pkg/...
  pkgtemplate.go
```

职责：

- 通过 `embed.FS` 持有第一批 `pkg` 模板文件。
- 返回按相对路径排序后的模板文件列表。
- `Runner.generatePkgFiles` 负责渲染模块路径并写入目标项目。

渲染规则：

- `xkit-template-v01` -> 目标项目 module。
- `xkit-template/pkg` -> `<module>/pkg`。
- Go 文件经过 `go/format`。
- 非 Go 文件原样写出，只做模块路径替换。

## 覆盖策略

`pkg` 文件默认使用“只创建不覆盖”：

- 目标项目不存在对应文件时写入。
- 目标项目已有文件时跳过并在结果中标记 `skipped`。

这样可以把 `pkg` 作为初始化能力，而不是长期覆盖业务项目的手写代码。

## 后续步骤

1. 将 `xkit-template/pkg` 第一批通用文件迁移到 `xkit/internal/codegen/pkgtemplate/files/pkg`。
2. 新增 `xkit gen pkg` 目标并纳入 `gen all`。
3. 用 `--dry-run` 验证 `xadmin-web` 的生成计划。
4. 后续再拆第二批领域包，优先处理 `metadata`、`middleware/auth`、`middleware/logging` 的 proto 依赖参数化问题。
