# 字典域重构方案（2026-05-30）

本文档用于替换当前 `dict_type / dict_entry / *_i18n` 的旧设计，作为后续从 `xkit/examples/admin` 到 `admin`、`admin-ui` 全链路重构的实施基线。

## 1. 当前问题

当前字典设计的核心问题不是租户边界，而是领域模型本身失真：

1. `dict_type` 混合了承载“分类”和“业务字典类型”的双重职责，但结构上只有一级。
2. `dict_entry` 既承担“业务实体标签”又承担“枚举值”的职责，语义过宽。
3. `dict_type.entries`、`dict_type.i18ns`、`dict_entry.i18ns` 被声明为 `Required()`，导致基础 CRUD 生成代码无法正常创建空父对象。
4. `dict_entry.i18n` 被建成 `map<string, DictEntryI18n>`，但底层表结构并没有形成稳定的“主对象 + 多语言值”管理模式。
5. 当前 proto/API 只适合“类型 + 条目”的简单字典，不适合页面、菜单、提示、设备、业务标签等多语言内容管理。
6. 前端 CRUD 界面仍是典型双表字典页，无法自然支撑“分类 -> 实体标签 -> 多语言值”的层级管理。

结论：后续不应继续在现有 `dict_type / dict_entry` 上修补，应先重构字典域模型，再重生生成代码与 UI。

## 2. 目标模型

新的字典域采用三层主从结构：

1. 分类 `DictCategory`
2. 实体标签 `DictLabel`
3. 语言值 `DictLabelI18n`

其中分类本身分为两级：

1. 主类
2. 子类

示例：

- 主类：`page`、`menu`、`prompt`、`device`、`other`
- 子类：`platform_feature`、`user_management`、`device_management`

最终可表达为：

- 主类：页面
- 子类：用户管理
- 标签：`user.list.title`
- 多语言：
  - `zh-CN = 用户列表`
  - `en-US = User List`

这套模型要同时满足：

- 系统页面文案
- 菜单显示文本
- 提示信息模板
- 设备/状态/业务标签
- 业务模块初始化数据
- 后续 UI 低成本 CRUD 与导入导出

## 3. 新表结构建议

### 3.1 `sys_dict_categories`

职责：

- 承载分类树，不直接承载语言值
- 支持主类/子类两级，也保留未来扩展更多层的可能

建议字段：

- `id`
- `parent_id`
- `category_key`
- `category_name`
- `category_level`
- `scene`
- `is_builtin`
- `is_enabled`
- `sort_order`
- `tenant_id`
- `remark`
- `created_* / updated_* / deleted_*`

说明：

- `category_key` 全局稳定，供初始化脚本和前端引用
- `category_level` 至少区分 `ROOT / CHILD`
- `scene` 用于区分 `page / menu / prompt / device / business`
- `is_builtin` 标记系统内置分类

唯一约束建议：

- `(tenant_id, parent_id, category_key)` 唯一

### 3.2 `sys_dict_labels`

职责：

- 承载真正的“业务标签/文案实体”
- 一个标签归属于一个子类分类
- 标签本身不保存多语言文本，只保存稳定键和非语言属性

建议字段：

- `id`
- `category_id`
- `label_key`
- `label_code`
- `label_kind`
- `default_text`
- `status`
- `payload_json`
- `is_builtin`
- `is_enabled`
- `sort_order`
- `tenant_id`
- `remark`
- `created_* / updated_* / deleted_*`

说明：

- `label_key` 是跨模块稳定主键，例如 `page.user.list.title`
- `label_code` 用于兼容旧业务的枚举/机器值
- `label_kind` 区分 `text / menu / message / enum / hint / badge`
- `default_text` 作为缺省显示值，避免语言包缺失时全空
- `payload_json` 承载附加元数据，如图标、颜色、路由、参数模板

唯一约束建议：

- `(tenant_id, category_id, label_key)` 唯一
- `(tenant_id, label_code)` 可选唯一，按场景决定

### 3.3 `sys_dict_label_i18n`

职责：

- 语言值从表
- 一条标签可有多条语言值

建议字段：

- `id`
- `label_id`
- `language_code`
- `text_value`
- `short_text`
- `description`
- `tenant_id`
- `created_* / updated_* / deleted_*`

唯一约束建议：

- `(label_id, language_code)` 唯一

### 3.4 可选：`sys_dict_category_i18n`

如果分类名称也要完全多语言，则增加：

- `category_id`
- `language_code`
- `display_name`
- `description`

如果短期内分类名只在后台维护端显示，可先保留 `category_name` 单列，不急于引入分类 i18n。

## 4. 取代关系

现有对象与新对象的映射：

- `dict_type` -> 拆分为 `dict_category` 与部分 `dict_label`
- `dict_entry` -> 迁移为 `dict_label`
- `dict_entry_i18n` -> 迁移为 `dict_label_i18n`
- `dict_type_i18n` -> 大概率删除，必要时改为 `dict_category_i18n`

因此不是“表字段调整”，而是“领域对象重命名 + 职责拆分”。

## 5. 租户模型建议

新的字典域建议继续保留 Hybrid 模型，但细分规则应调整为：

- `dict_category`
  - 平台可维护全局分类
  - 租户可见全局分类
  - 是否允许租户自建分类需要单独开关，默认不开放或只开放叶子分类
- `dict_label`
  - 平台可维护全局标签
  - 租户可读全局标签 + 自有标签
  - 租户只能改自有标签
- `dict_label_i18n`
  - 必须继承所属 `label` 的 tenant 归属

推荐策略：

- 分类树以平台全局为主
- 标签层允许租户扩展
- 多语言值严格跟随标签归属

## 6. proto / API 重构建议

旧接口：

- `DictTypeService`
- `DictEntryService`

建议替换为三组资源接口：

1. `DictCategoryService`
2. `DictLabelService`
3. `DictLabelValueService` 或直接将 i18n 作为 `DictLabel` 的聚合字段

推荐 API 形态：

### 6.1 `DictCategory`

- `List`
- `Get`
- `Create`
- `Update`
- `Delete`
- `ListTree`

DTO 字段建议：

- `id`
- `parent_id`
- `category_key`
- `category_name`
- `category_level`
- `scene`
- `is_builtin`
- `is_enabled`
- `sort_order`
- `tenant_id`
- `children`

### 6.2 `DictLabel`

- `List`
- `Get`
- `Create`
- `Update`
- `Delete`
- `BatchUpsert`
- `ListByCategoryKey`
- `ResolveBySceneAndKey`

DTO 字段建议：

- `id`
- `category_id`
- `category_key`
- `label_key`
- `label_code`
- `label_kind`
- `default_text`
- `payload_json`
- `tenant_id`
- `tenant_name`
- `translations`
- `current_translation`

### 6.3 `DictLabelTranslation`

如果单独建 service：

- `ListByLabelId`
- `BatchReplace`

如果聚合进 `DictLabel`：

- 直接在 `Create/Update` 中维护 `translations`

推荐：聚合进 `DictLabel`，减少前端编辑复杂度。

### 6.4 初始化与运行时读取

增加查询接口：

- `ListLabelsByCategoryKeys`
- `ResolveLabels`
- `BatchResolveI18n`

典型用途：

- 页面启动时批量取页面标题/按钮文案
- 菜单渲染时按分类批量取 label
- 提示模板按 key 和 locale 解析

## 7. xkit 生成配置改造

当前 `xkit/examples/admin/admin-config/admin.yaml` 里有：

- `dict_type`
- `dict_entry`

应调整为：

- `dict_category`
- `dict_label`

如需要独立 CRUD：

- `dict_label_i18n`

但更推荐：

- `dict_label_i18n` 不单独暴露 CRUD 资源
- 由 `dict_label` 聚合写入

生成配置需要支持的新能力：

1. 聚合子表写入
2. 树形父子资源
3. 运行时解析型查询接口
4. i18n map/list 与从表之间的双向映射
5. 初始化种子导入

因此不仅是 `admin.yaml` 改资源名，还要补 xkit 的生成能力。

## 8. xkit 模板/生成器改造点

需要重点检查和改造：

- `xkit/internal/codegen/template/repo_file.tmpl`
- `xkit/internal/codegen/template/service_file.tmpl`
- `xkit/internal/codegen/template/repo_ext.tmpl`
- `xkit/internal/codegen/template/service_ext.tmpl`

建议新增能力：

1. `aggregate_relations`
   - 允许生成父资源 CRUD 时自动处理子表
2. `tree_resource`
   - 允许生成父子树形 list/query
3. `manual_methods`
   - 为 `ListTree / ResolveByKey / BatchUpsert` 预留扩展点
4. `i18n_relation`
   - 定义主对象和语言值对象的映射规则

## 9. 后端实施顺序

### Phase 1：xkit example 领域建模

修改：

- `xkit/examples/admin/schema/dict_type.go`
- `xkit/examples/admin/schema/dict_entry.go`
- `xkit/examples/admin/schema/dict_type_i18n.go`
- `xkit/examples/admin/schema/dict_entry_i18n.go`

目标：

- 删除旧模型
- 引入 `dict_category / dict_label / dict_label_i18n`

### Phase 2：proto 重新设计

修改：

- `xkit/examples/admin/api/protos/dict/v1/*.proto`
- `xkit/examples/admin/api/protos/admin/v1/i_dict_*.proto`

目标：

- 以新资源和聚合 DTO 重写字典接口

### Phase 3：生成配置改造

修改：

- `xkit/examples/admin/admin-config/admin.yaml`
- 必要时改 `admin-target-config/admin.yaml`

目标：

- 用新资源替换旧资源
- 明确聚合关系和树形关系

### Phase 4：xkit 模板/代码生成增强

目标：

- 支撑树形资源 + 聚合 i18n 子表

### Phase 5：在 `admin` 重生后端代码

目标：

- 重新生成 ent / proto / repo / service / server register
- 补手写扩展和迁移脚本

### Phase 6：数据迁移与初始化

目标：

- 从旧 `dict_type / dict_entry / *_i18n` 迁移到新结构
- 建立系统默认分类和标签初始化机制

### Phase 7：前端重构

目标：

- 页面从“双表 dict”改为“三栏/分层管理”

## 10. 前端页面和风格建议

旧 UI 不再适合新模型。新页面建议采用三栏结构：

1. 左侧：分类树
2. 中间：标签列表
3. 右侧：语言值编辑器

### 页面形态

- 顶部：按 `scene` 切换，如 页面 / 菜单 / 提示 / 设备 / 其它
- 左栏：主类/子类树，可增删改排序
- 中栏：标签表格，字段如 `label_key / default_text / kind / 内置 / 启用`
- 右栏：翻译表单，按语言 tabs 或 stack card 编辑

### 风格方向

- 不要沿用普通后台“两个表格 + 弹窗”的弱信息结构
- 用稳定的分栏工作台式布局
- 分类树和语言编辑器需要明显的信息层级
- 支持批量初始化和导入导出入口

## 11. 对旧页面/旧接口的兼容策略

建议分两阶段：

1. 新接口与旧接口短期并存
2. 前端切到新页面后，再移除旧 `dict_type / dict_entry`

如果想一次性切换，也可以，但成本更高，且迁移窗口更陡。

## 12. 明确的下一步

下一步不建议直接修改 `admin/internal/data/repo/...`，而应先做下面三件事：

1. 在 `xkit/examples/admin/schema/` 中把旧字典四个 schema 替换成新模型。
2. 在 `xkit/examples/admin/api/protos/dict/v1/` 中重写字典 proto。
3. 在 `xkit/examples/admin/admin-config/admin.yaml` 中把 `dict_type / dict_entry` 资源替换为新资源定义。

完成这三步后，再决定是否需要先增强 xkit 模板，还是先在 `admin` 做一次手工验证版生成。

## 13. 当前判断

这次重构属于“领域模型重做”，不是普通 CRUD 修订。

如果继续沿用旧资源名做局部修复，会产生两个问题：

1. 生成代码会持续围绕错误对象名扩散。
2. 前端页面会继续被迫套在错误的数据结构上。

因此后续实现应以 `xkit/examples/admin` 为源头，从模型、proto、生成配置开始重构，再向 `admin` 和 `admin-ui` 下游刷新。
