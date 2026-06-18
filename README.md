# xkit

`xkit` 是面向 XAdmin 风格 Go 服务工程的代码生成器。它把原始 `proto/schema`、已生成的 API/Ent 代码和服务配置组合起来，生成可重复覆盖的服务层、数据层、传输注册和启动装配代码。

当前建议把项目分成三类输入和代码：

- `xkit-template`：启动模板仓库，默认来自 `https://github.com/chnxq/xkit-template.git`。它负责 `cmd/server`、`configs`、`internal/bootstrap`、`internal/server`、`internal/data/bootstrap` 等项目骨架。
- `source`：目标项目中的原始输入目录，通常包含 `source/api/protos`、`source/schema` 和生成出来的 `<project>-config/<service>.yaml`。
- `xkit`：动态代码生成器，只覆盖 `*.gen.go`，并且只在缺失时创建 `*_ext.go`。

默认启动流程已经不依赖 Wire。`xkit gen all` 不生成、不清理 Wire provider set；`xkit gen wire` 仅作为显式执行的历史兼容命令保留。

## 前置条件

- Go 可用，目标工程已有 `go.mod`。
- 可执行 `git`。默认模板源是 GitHub；离线时可把 `xkit init template` 的第一个参数换成本地模板目录。
- 需要重新生成 API 时，目标工程可执行 `buf generate --template buf.gen.yaml`。
- 需要重新生成 Ent 时，目标工程可执行 Ent 生成命令；下面示例使用显式 `ent generate` 命令。
- 建议所有会写文件的 `xkit` 命令先执行 `--dry-run`，确认计划后再真实写入。

## 命令总览

```text
xkit init template [template-source] [--project <path>] [--module <module>] [--app-name <name>] [--command-name <name>] [--service-name <name>] [--force] [--dry-run] [--skip-go-get-update-all]
xkit init source <source-path> [--project <path>] [--service <name>] [--config <path>] [--typescript-project <path>] [--force] [--dry-run]
xkit gen service <service> [--project <path>] [--config <path>] [--domain <name>] [--dry-run]
xkit gen repo <service> [--project <path>] [--config <path>] [--domain <name>] [--dry-run]
xkit gen register <service> [--project <path>] [--config <path>] [--domain <name>] [--dry-run]
xkit gen bootstrap <service> [--project <path>] [--config <path>] [--domain <name>] [--dry-run]
xkit gen wire <service> [--project <path>] [--config <path>] [--domain <name>] [--dry-run]
xkit gen frontend-meta <service> [--project <path>] [--config <path>] [--domain <name>] [--typescript-project <path>] [--dry-run]
xkit gen all <service> [--project <path>] [--config <path>] [--domain <name>] [--typescript-project <path>] [--dry-run]
```

`xkit init template` 在真实复制模板后默认执行 `go get -u all`。实际项目初始化时，如果 API/Ent 代码还没有生成完整，建议先加 `--skip-go-get-update-all`，等完整生成后再在目标工程手动执行依赖更新。

## 快速开始

下面以当前工作区的 `xadmin-web` 和 `admin` 服务为例。其它项目只需要替换路径、module 和服务名。

1. 预览启动模板落地计划：

```powershell
cd D:\GoProjects\XAdmin\xkit

go run ./cmd/xkit init template `
  --project D:\GoProjects\XAdmin\xadmin-web `
  --module xadmin-web `
  --app-name XAdmin `
  --command-name xadmin-web `
  --service-name admin `
  --dry-run
```

2. 真实复制启动模板：

```powershell
go run ./cmd/xkit init template `
  --project D:\GoProjects\XAdmin\xadmin-web `
  --module xadmin-web `
  --app-name XAdmin `
  --command-name xadmin-web `
  --service-name admin `
  --skip-go-get-update-all
```

如果希望使用本地模板目录，把命令改成：

```powershell
go run ./cmd/xkit init template D:\GoProjects\XAdmin\xkit-template `
  --project D:\GoProjects\XAdmin\xadmin-web `
  --module xadmin-web `
  --app-name XAdmin `
  --command-name xadmin-web `
  --service-name admin `
  --skip-go-get-update-all
```

3. 导入原始 `proto/schema` 并生成 `admin.yaml`：

```powershell
go run ./cmd/xkit init source D:\GoProjects\XAdmin\xadmin-web\source `
  --project D:\GoProjects\XAdmin\xadmin-web `
  --service admin `
  --typescript-project D:\GoProjects\XAdmin\xadmin-web-ui `
  --dry-run

go run ./cmd/xkit init source D:\GoProjects\XAdmin\xadmin-web\source `
  --project D:\GoProjects\XAdmin\xadmin-web `
  --service admin `
  --typescript-project D:\GoProjects\XAdmin\xadmin-web-ui
```

默认配置输出路径：

```text
D:\GoProjects\XAdmin\xadmin-web\source\xadmin-web-config\admin.yaml
```

`source/api` 根目录下的普通文件会同时复制到目标项目的 `api/` 目录，供后续 `buf generate --template buf.gen.yaml` 等步骤使用。导入 `buf*.gen.yaml` / `buf*.gen.yml` 时，`managed.override` 中的 `go_package_prefix` 和 `go_package` 会按目标项目 module 强制校正为 `<module>/api/gen...`；导入 schema `.go` 文件时，本地 proto 对应的 `*/api/gen/<domain>/...` import 也会校正到目标 module，避免从旧项目复制来的 Go 包路径残留。

如果目标位置已有旧的 proto、schema 或 config，默认会跳过。确认需要覆盖时再追加 `--force`。

4. 在目标项目生成 API Go 代码：

```powershell
cd D:\GoProjects\XAdmin\xadmin-web\api
buf generate --template buf.gen.yaml
buf generate --template buf.admin.openapi.gen.yaml
buf generate --template buf.vue.admin.typescript.gen.yaml
```

应看到 `api/gen/<domain>/v1/*.pb.go`、`*_grpc.pb.go`、`*_http.pb.go`。

5. 在目标项目生成 Ent 代码：

```powershell
cd D:\GoProjects\XAdmin\xadmin-web
go run -mod=mod entgo.io/ent/cmd/ent generate ./internal/data/ent/schema --feature privacy,sql/upsert,sql/versioned-migration
```

应看到 `internal/data/ent/client.go`、entity、query、create、update、predicate 等代码。

如果目标工程维护了可靠的 `internal/data/ent/generate.go`，也可以按项目约定执行 `go generate ./internal/data/ent`。

6. 预览并执行 `xkit` 动态代码生成：

```powershell
cd D:\GoProjects\XAdmin\xkit

go run ./cmd/xkit gen all admin `
  --project D:\GoProjects\XAdmin\xadmin-web `
  --config D:\GoProjects\XAdmin\xadmin-web\source\xadmin-web-config\admin.yaml `
  --typescript-project D:\GoProjects\XAdmin\xadmin-web-ui `
  --dry-run

go run ./cmd/xkit gen all admin `
  --project D:\GoProjects\XAdmin\xadmin-web `
  --config D:\GoProjects\XAdmin\xadmin-web\source\xadmin-web-config\admin.yaml `
  --typescript-project D:\GoProjects\XAdmin\xadmin-web-ui
```

`gen all` 会生成 service、repo、HTTP/gRPC register 和 bootstrap glue。它会覆盖 `*.gen.go`，但不会覆盖已有的 `*_ext.go`。

如果目标前端是 `admin-ui` 这种基于生成元数据的页面体系，还可以单独执行 `frontend-meta` 生成，将搜索、列表、表单元数据以及页面 I18n 输出到前端项目：

```powershell
cd D:\GoProjects\XAdmin\xkit

go run ./cmd/xkit gen frontend-meta admin `
  --project D:\GoProjects\XAdmin\admin-ui `
  --config D:\GoProjects\XAdmin\xkit\examples\admin\admin-target-config\admin.yaml `
  --typescript-project D:\GoProjects\XAdmin\admin-ui `
  --dry-run

go run ./cmd/xkit gen frontend-meta admin `
  --project D:\GoProjects\XAdmin\admin-ui `
  --config D:\GoProjects\XAdmin\xkit\examples\admin\admin-target-config\admin.yaml `
  --typescript-project D:\GoProjects\XAdmin\admin-ui
```

当前推荐以 `xkit/examples/admin/admin-target-config/admin.yaml` 作为真实迭代基线，等验证通过后再同步回：

- `xkit/examples/admin/admin-config/admin.yaml`
- `xkit/examples/admin-v2/admin-config/admin.yaml`

`frontend-meta` 当前输出路径不再从 `admin.yaml` 顶层配置读取，而是和 TypeScript API 一样从 `TypeScriptRoot` 推算。对于当前 `admin-ui`，输出到：

```text
admin-ui/apps/web-antd/src/views/generated/admin/
```

主要包含：

- `**/*.meta.ts`
- `page_i18n.zh-CN.json`
- `page_i18n.en-US.json`
- `langs/zh-CN/*.json`
- `langs/en-US/*.json`
- `config.ts`

当前生成的元数据模块约定导出：

- `buildSearchFormOptions`
- `buildListGridColumns`
- `buildFormOptions`

其中非标准 CRUD 弹窗页面可以只接入搜索和列表元数据，把业务特有的 slot、动作列、请求参数映射、复杂弹窗逻辑继续保留在手写 `index.vue` 中。

7. 更新依赖、测试并启动：

```powershell
cd D:\GoProjects\XAdmin\xadmin-web
go get -u all
go mod tidy
go test ./...
go run ./cmd/server server -config_path ./configs
```

如果目标环境的数据库、Redis、注册中心、trace exporter 等外部依赖还没有准备好，启动失败不一定代表生成失败；先以 `go test ./...` 和编译错误为准。

如果希望复用 `examples/generateAll.ps1` 走一键流程，但默认的 `examples/admin/<project>-config/<service>.yaml` 约定不适合当前项目，可以显式传入：

- `-ConfigPath <生成配置文件路径>`
- `-CanonicalConfigPath <用于覆盖的基线配置路径>`

例如使用 `admin-v2` 的配置作为基线：

```powershell
& 'D:\GoProjects\XAdmin\xkit\examples\generateAll.ps1' `
  -WorkspaceRoot 'D:\GoProjects\XAdmin' `
  -ProjectName 'qadmin' `
  -Module 'qadmin' `
  -AppName 'QAdmin' `
  -ServiceName 'admin' `
  -TypeScriptRoot 'D:\GoProjects\XAdmin\qadmin\.generated-ui' `
  -ConfigPath 'D:\GoProjects\XAdmin\xkit\examples\admin\qadmin-config\admin.yaml' `
  -CanonicalConfigPath 'D:\GoProjects\XAdmin\xkit\examples\admin-v2\admin-config\admin.yaml' `
  -SkipGoGetUpdateAll
```

## 文件归属

| 类型 | 典型文件 | 维护方式 |
| --- | --- | --- |
| 启动模板代码 | `cmd/server/*`、`configs/*`、`internal/bootstrap/app.go`、`internal/server/*`、`internal/data/bootstrap/*` | `xkit init template` 复制后归目标项目维护 |
| 动态生成代码 | `internal/service/*_service.gen.go`、`internal/data/repo/*_repo.gen.go`、`internal/server/*_register.gen.go`、`internal/bootstrap/generated_servers.gen.go`、`internal/bootstrap/generated_data_providers.gen.go`、`internal/data/bootstrap/ent_client.gen.go` | `xkit gen ...` 可重复覆盖 |
| 手写扩展代码 | `*_ext.go`、`internal/bootstrap/hooks.go`、`internal/server/options.go`、`internal/server/manual_http_data.go`、`internal/data/bootstrap/data.go`、`internal/data/bootstrap/resources.go` | 只在缺失时创建或由模板 preserve，后续人工维护 |
| 前端生成资源 | `<TypeScriptRoot>/apps/web-antd/src/views/generated/admin/**/*` | `xkit gen frontend-meta` / `xkit gen all` 可重复覆盖，页面层按需引用 |

`internal/server/manual_http.go` 属于模板基线文件，不再作为 project preserve 扩展点保留；需要 `GeneratedData` 的项目手写 HTTP 逻辑应放在 `internal/server/manual_http_data.go`。

## 常见问题

`proto service not found`：确认已执行 `xkit init source`，并且 `admin.yaml` 中的 `proto_service` 与 proto 文件一致。

找不到 HTTP/gRPC binding 或注册函数：确认已在目标项目 `api` 目录执行 `buf generate --template buf.gen.yaml`。

找不到 Ent 类型、predicate、字段或 client：确认已重新执行 Ent 生成命令，并且 `internal/data/ent/schema` 与 Ent 生成代码一致。

`go.sum` 或依赖缺失：完整生成 API、Ent 和 `xkit` 代码后，在目标项目执行 `go get -u all` 与 `go mod tidy`。

`skipped resources without matching proto service`：通常是关系表、详情表或嵌入实体没有独立 service，属于正常提示；只有需要生成 service/repo 的资源才必须补齐 proto service。

前端生成后如果所见变更很多，先区分是否属于真实语义差异，还是仅仅是 `page_i18n`、meta 文件格式或时间戳刷新。

如果改了 `frontend-meta` 规则，可先执行：

```powershell
cd D:\GoProjects\XAdmin\xkit
go test ./internal/codegen -count=1

cd D:\GoProjects\XAdmin\admin-ui
pnpm exec eslint apps/web-antd/src/views/generated/admin --fix
pnpm -F @vben/web-antd run typecheck
```

对于 zh-CN 占位文案，当前规则应避免出现：

- `Search ...`
- `Select ...`
- `... Range`

这类中英混杂文本，优先回退到字段标题。

## 进一步文档

- [启动模板与生成代码边界](doc/bootstrap-template-generated-boundary.md)
- [Source import 命令](doc/source-import-command.md)
- [前端 API 集成计划](doc/frontend-api-integration-plan.md)
- [前端 view meta 生成方案](doc/frontend-view-meta-generation-plan.md)
- [服务代码生成规格](doc/service-codegen-spec.md)
- [模板仓库方案记录](doc/template-repo-bootstrap-solution.md)：历史方案和决策过程，以 README 与边界文档为当前准则。
- [xkit-helper 技能主文档](xkit-helper/SKILL.md)：面向 Codex 的三阶段操作说明，覆盖生成、对齐与功能差异复盘。
- [xkit-helper 第一阶段生成说明](xkit-helper/references/phase-1-generate.md)：对应 `examples/generateAll.ps1` 的参数、步骤与校验点。
- [xkit-helper 第二阶段修补热点](xkit-helper/references/phase-2-repair-hotspots.md)：生成后对齐参考项目时的常见差异与修补清单。
- [xkit-helper 第三阶段功能差异复盘](xkit-helper/references/phase-3-feature-gap-review.md)：对比新项目与参考项目的剩余差异，以功能差异而非原始文件差异为主导，并要求将结果落盘到目标仓库文档。
