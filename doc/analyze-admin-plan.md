问题陈述

admin 项目由 xkit 生成器（基于 xkit-template、ent schema、proto）产生。后期 Codex/人工修改导致生成器模板与实际项目出现漂移（regression/drift）。目标是：分析 admin 项目、识别 Codex 产生或手工修改的代码，规划并把可回归的改动合并回 xkit-template，减少未来生成/再生成时的漂移。

总体方法

1. 快速审查：audit admin 代码、xkit/doc、NEXT_CONTEXT_HANDOFF 文件，定位近期重要变更和手工补丁。
2. 区分类型：把改动分为（A）应回归到模板的生成逻辑/模板文件；（B）是环境/部署/seed 数据等仅项目专有内容；（C）临时修复或实验性变更不合并。
3. 实施分步：提取候选补丁 → 在 xkit-template 中实现模板改动 → 生成样本代码并对比 → 运行单元/集成测试 → 提交模板 PR 并同步到 xkit/xkitmod。

里程碑与任务

- analyze-admin: 分析 admin 源代码，列出所有明显由生成器产生但后续改动的文件和关联变更（proto/ent/gen、internal/data、api 等）。
- extract-codex-diffs: 从 git 历史提取 candidate diff 列表（最近 3-6 周重点）。
- identify-template-candidates: 确定哪些改动应放回 xkit-template（列出具体模板文件与变更摘要）。
- implement-template: 在 xkit-template 中实现变更（含测试/样例）。
- regen-verify: 使用更新后的模板生成代码到临时目录，运行 go test / 前端 typecheck，验证无回归。
- submit-prs: 为 xkit-template 与相关 tooling（xkit, xkitmod）分别提交 PR 并记录变更说明。

验证方法

- 后端：cd admin && go test ./...
- 前端：cd admin-ui && pnpm -F @vben/web-antd run typecheck
- 生成校验：生成后与当前 admin repo 做 diff，检查不必要的变更。

风险与注意事项

- 某些变更是环境或部署相关（configs/seed），不应回写模板。
- 回归模板前需保证向后兼容，必要时以 feature-flag 或模板参数化方式实现。

下一步

1. 先执行 analyze-admin 与 extract-codex-diffs。
2. 根据输出列出 template-candidates 并征求团队同意（PR 草案）。

附：关键验证命令与参考文件
- D:\GoProjects\XAdmin\xkit\NEXT_CONTEXT_HANDOFF_20260510.md
- D:\GoProjects\XAdmin\xkit\doc
- go test ./...
- pnpm -F @vben/web-antd run typecheck
