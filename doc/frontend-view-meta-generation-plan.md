# xkit 前端视图元数据生成方案

本文定义 `xkit` 为 `admin-ui` 生成可复用前端视图元数据的方案，目标是生成标准 CRUD `form` 对话框和 `list` 列表页所需的 `VbenFormProps`、`VxeTableGridOptions` 与 i18n 文案资源，并由现有页面引用。

## 目标

- 保持 `admin-ui` 当前以前端页面为中心的组织方式。
- 让 `xkit` 基于 schema、proto、`admin-config/admin.yaml` 生成标准 CRUD `form` 对话框和 `list` 列表页的可复用前端元数据。
- 减少 `admin-ui` 中重复的表单项、表格列、i18n key 和默认排序配置。
- 保留复杂页面的手写能力，不追求第一阶段自动生成完整 `.vue` 页面。
- 为 `admin-v2` 和后续项目提供一条可批量复制的前端生成路径。

## 第一阶段边界

第一阶段明确只解决标准页面元数据，不解决所有前端问题。

第一阶段做：

- 生成列表页搜索表单元数据
- 生成列表页表格列元数据
- 生成标准 CRUD 弹窗表单元数据
- 生成对应页面 i18n 文案资源
- 生成少量默认排序和字段映射辅助常量

第一阶段不做：

- 不生成完整业务页面
- 不生成复杂 `slots` 的实现代码
- 不生成 `action` 列里的业务按钮
- 不生成复杂表单联动逻辑
- 不覆盖树、授权、Tabs、Drawer、复合布局等复杂页面

## 当前现状

`admin-ui` 当前页面普遍采用以下结构：

- 页面自己手写 `formOptions: VbenFormProps`
- 页面自己手写 `gridOptions: VxeTableGridOptions<T>`
- 页面自己手写 `$t('page.xxx.xxx')` 文案 key
- 页面自己调用 `src/api/admin/*.ts` 的 adapter

例如：

- `apps/web-antd/src/views/app/log/login-audit-log/index.vue`
- `apps/web-antd/src/views/app/log/api-audit-log/index.vue`
- `apps/web-antd/src/views/system/api/index.vue`
- `apps/web-antd/src/views/system/role/index.vue`

这些页面存在大量可复制的结构，但又夹杂少量页面特有逻辑。适合抽取“可生成的前端元数据”，不适合第一阶段直接替换成纯生成页面。

## 总体方案

`xkit` 新增一类“前端视图元数据生成”能力，直接把生成物写入 `admin-ui` 仓库，由现有页面薄壳引用。

### 生成物目录

生成物统一放到：

```text
admin-ui/apps/web-antd/src/views/generated/admin/
```

该目录对应一套独立的生成配置，不和手写页面目录混放。

### 建议的生成物形态

建议第一阶段生成：

```text
admin-ui/apps/web-antd/src/views/generated/admin/
  config.ts
  page_i18n.zh-CN.json
  page_i18n.en-US.json
  app/log/login-audit-log.meta.ts
  app/log/api-audit-log.meta.ts
  system/tenant.meta.ts
```

其中：

- `config.ts`
  - 导出该生成目录的全局配置
  - 包括目录前缀、默认行为、可选公共 helper
- `page_i18n.zh-CN.json`
  - 当前生成目录对应的中文页面文案集合
- `page_i18n.en-US.json`
  - 当前生成目录对应的英文页面文案集合
- `*.meta.ts`
  - 每个资源一个元数据模块
  - 负责导出列表搜索表单、列表列、标准 CRUD 表单配置

## 为什么 i18n 采用统一文件

在“每页一个 i18n 文件”和“生成目录统一一套 i18n 文件”之间，当前更推荐统一文件：

```text
admin-ui/apps/web-antd/src/views/generated/admin/page_i18n.zh-CN.json
admin-ui/apps/web-antd/src/views/generated/admin/page_i18n.en-US.json
```

原因：

- 更符合当前 `admin-ui` 的语言资源组织方式
- 更容易整体审查生成文案
- 更容易做统一 merge 或替换
- 减少生成文件数量
- 对生成器更简单

每页一个 i18n 文件的优点是资源隔离更强，但第一阶段收益不大，文件碎片反而更多。

因此本文结论是：

- 第一阶段采用“生成目录统一一套 i18n 文件”
- 每个 `meta.ts` 仍按统一 key 前缀引用，比如 `page.loginAuditLog.createdAt`

## admin-ui 侧的引用方式

页面改造成“手写薄壳 + 生成元数据”的模式。

### 推荐结构

```text
src/views/app/log/login-audit-log/index.vue
src/views/app/log/api-audit-log/index.vue
src/views/system/tenant/index.vue

src/views/generated/admin/
  config.ts
  page_i18n.zh-CN.json
  page_i18n.en-US.json
  app/log/login-audit-log.meta.ts
  app/log/api-audit-log.meta.ts
  system/tenant.meta.ts
```

### 页面职责划分

手写页面继续负责：

- 调用 `useVbenVxeGrid`
- 调用 `src/api/admin/*.ts`
- 组装 query 参数
- 定义 `slots`
- 定义 `action` 列
- 定义权限码和页面特有逻辑

生成的 `meta.ts` 负责：

- 标准搜索项
- 标准数据列
- 标准 CRUD `form` 对话框字段
- 默认宽度
- 默认排序字段
- 基础 formatter 标记
- i18n key 命名约定

### 典型引用形式

```ts
import { $t } from '#/locales';
import {
  buildSearchFormOptions,
  buildListGridColumns,
  buildFormOptions,
  defaultSortField,
} from '#/views/generated/admin/app/log/login-audit-log.meta';
```

这里不要求 `meta.ts` 直接产出完整不可变的 `gridOptions`。更推荐生成细粒度部件，让页面去组合。

## 生成模块的建议接口

建议每个 `meta.ts` 暴露最小稳定接口：

```ts
export function buildSearchFormOptions(t: TranslateFn): VbenFormProps;

export function buildListGridColumns(
  t: TranslateFn,
): VxeTableGridOptions<Row>['columns'];

export function buildFormOptions(t: TranslateFn): VbenFormProps;

export const defaultSortField = 'created_at';
export const defaultSortDirection = 'DESC';
```

如果资源没有标准 CRUD 弹窗，例如纯日志页，可不导出 `buildFormOptions`，或导出空配置。

## xkit 的输入来源

前端视图元数据不能只靠 schema 推导，需要组合多源信息，但第一阶段应尽量少配。

### 1. schema

schema 是默认来源，负责提供：

- 字段集合
- 基础类型
- 是否适合搜索 / 排序 / 编辑的基础判断

生成器应优先从 schema 推导默认值，避免把大量本可推导的信息重新写回配置。

### 2. proto / DTO

proto / DTO 负责提供：

- 前端最终可见字段名
- 枚举、时间、嵌套对象等类型信息
- 某些页面真实使用的 camelCase 字段

### 3. `admin-config/admin.yaml`

`admin-config/admin.yaml` 只负责补充“推导不出来或需要明确覆盖”的部分。

第一阶段不建议把配置设计得过细。原则是：

- 能从 schema 推导的，不配置
- 能从 proto 推导的，不配置
- 只有展示顺序、页面路径、少量组件类型、少量覆盖项需要配置

## 建议的最小配置模型

第一阶段建议只增加一个较薄的 `frontend` 段，而不是引入大量新字段。

示意：

```yaml
resources:
  - name: login_audit_log
    entity: LoginAuditLog
    frontend:
      view_path: app/log/login-audit-log
      i18n_prefix: page.loginAuditLog
      list:
        columns:
          - [createdAt, Created At, 创建时间]
          - [status, Status, 状态]
          - [username, Username, 用户名]
          - [actionType, Action Type, 动作类型]
          - [riskLevel, Risk Level, 风险等级]
          - [platformSummary, Platform, 平台]
          - [geoLocationSummary, Geo Location, 地理位置]
          - [ipAddress, IP Address, IP地址]
          - [loginMethod, Login Method, 登录方式]
        filters:
          username: Input
          ipAddress: Input
          actionType: Select
          riskLevel: Select
          status: Select
          createdAt: RangePicker
      form:
        enabled: false
```

对于普通 CRUD 资源，可再加：

```yaml
resources:
  - name: tenant
    entity: Tenant
    frontend:
      view_path: system/tenant
      i18n_prefix: page.tenant
      list:
        columns:
          - [name, Tenant, 租户]
          - code
          - status
          - createdAt
      form:
        fields:
          - [name, Tenant, 租户]
          - code
          - description
          - status
```

这里的设计意图是：

- `columns` 可以只写字段名，也可以写 `[field, enTitle, cnTitle]`
- `filters` 保持最简单的 `field: component` 表达；标题默认复用同字段 `columns` 的标题，找不到时再回退到 schema field/comment
- `form.fields` 与 `list.columns` 使用相同结构，可写字段名，也可写 `[field, enTitle, cnTitle]`
- 宽度、排序能力、时间格式、布尔/枚举默认行为都尽量由生成器推导

## 推导优先级

建议生成器采用以下优先级：

1. 显式配置
2. proto / DTO 信息
3. schema 信息
4. 生成器内置默认规则

例如：

- 列宽：优先显式配置，否则按字段类型给默认宽度
- 组件类型：优先显式配置，否则按字段类型推导
- 是否可排序：优先显式配置，否则按字段类型和字段语义推导
- i18n key：优先用 `i18n_prefix` 拼装

## 适合自动生成的内容

- 简单 `Input` / `InputNumber` / `Select` / `RangePicker`
- 普通文本列
- 时间列
- 布尔列
- 简单枚举列
- 标准 CRUD 对话框字段
- 默认排序
- 默认 placeholder / label key

## 不适合第一阶段自动生成的内容

- `action` 列
- 依赖运行时状态的列显隐
- 复杂 slot
- 跨字段拼装列
- 权限驱动按钮
- 复杂表单联动
- Tree / Tabs / Drawer / Modal 复合页面

## xkit 侧实现建议

### 建议新增的内部模块

```text
xkit/internal/frontendmeta/
  model.go
  infer.go
  render.go
  writer.go

xkit/internal/codegen/template/
  frontend_view_meta.tmpl
  frontend_view_config.tmpl
  frontend_page_i18n_zh.tmpl
  frontend_page_i18n_en.tmpl
```

### 建议新增的生成步骤

建议增加独立生成命令或独立阶段：

```text
xkit gen frontend-meta
```

也可以挂入 `gen all`，但要能单独开关，避免每次都写前端仓库。

### 生成流程

1. 读取 `admin-config/admin.yaml`
2. 读取 schema
3. 读取 proto / DTO 元信息
4. 推导资源的前端视图模型
5. 生成 `config.ts`
6. 生成 `page_i18n.zh-CN.json`
7. 生成 `page_i18n.en-US.json`
8. 为每个资源生成 `*.meta.ts`

## 与现有 generated TS client 的关系

该方案不替代：

- `apps/web-antd/src/api/generated/admin/service/v1/index.ts`

而是建立在它之上。

推荐关系：

```text
generated TS client
  -> 手写 adapter（src/api/admin）
    -> 手写页面（src/views/.../index.vue）
      -> 生成视图元数据（src/views/generated/admin/*.meta.ts）
```

也就是说：

- TS generated client 仍是接口类型来源
- 手写 adapter 仍是后端接入边界
- 生成视图元数据只解决页面配置重复问题

## 第一阶段 PoC 范围

建议先做 3 个页面：

### 1. `login-audit-log`

原因：

- 列表结构简单
- 搜索项清晰
- 几乎没有标准 CRUD 表单，适合验证“列表页元数据”

### 2. `api-audit-log`

原因：

- 与 `login-audit-log` 相近
- 可以验证第二个日志页模板

### 3. `tenant`

原因：

- 同时具备标准列表页和标准 CRUD 对话框
- 适合验证 `form + list` 一起生成

不建议首批就做：

- `role`
- `permission`
- `menu`
- `dict`

因为这些页面通常带有树、授权、组合资源或复杂交互。

## 页面改造策略

建议每个页面按两步走。

### 第一步

只替换：

- 搜索表单配置
- 基础数据列
- 标准 CRUD 对话框表单配置

保留：

- `proxyConfig.ajax.query`
- `slots`
- `action` 列
- `toolbarConfig`

### 第二步

再评估是否把：

- 默认排序字段
- 枚举渲染
- 部分 formatter

也完全收敛到生成元数据中。

## 第一版验收标准

满足以下条件即可认为第一版成立：

1. `xkit` 能为 `login-audit-log`、`api-audit-log`、`tenant` 生成 `meta.ts`
2. `xkit` 能生成 `views/generated/admin/config.ts`
3. `xkit` 能生成统一的 `page_i18n.zh-CN.json` / `page_i18n.en-US.json`
4. `admin-ui` 页面能引用生成的列表表单、基础列和标准 CRUD 表单
5. 现有 slot 渲染不受影响
6. `pnpm -F @vben/web-antd run typecheck` 通过

## 最终结论

对当前 `admin` / `admin-ui` / `xkit` 体系来说，适合落地的路线是：

- `xkit` 生成前端可直接引用的视图元数据源码
- 生成范围覆盖标准 CRUD `form` 对话框和 `list` 列表页
- 生成物统一进入 `admin-ui/apps/web-antd/src/views/generated/admin/`
- i18n 资源在该生成目录下统一输出
- 配置项尽量保持精简，以 schema 和 proto 推导为主，以少量显式覆盖为辅

这条路线边界清晰、实现可控，也更符合 `xkit` 作为代码生成器的定位。
