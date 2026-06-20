# generateModule.ps1

`examples/generateModule.ps1` 用于验证模块模式下的一整套生成链路。

它对标 `examples/generateAll.ps1`，但目标是：

- 将 `examples/<module>` 的 proto/schema/source 导入到宿主项目
- 生成到 `admin/modules/<module>` 这样的模块目录
- 使用外部模块配置栈，而不是模块目录内 yaml

## 配置约定

每个模块使用两份外部配置：

- `examples/<module>/<module>-config/<module>.yaml`
  - 规范基线配置
- `examples/<module>/<module>-target-config/<module>.yaml`
  - 当前真实生成、调试、重构使用的工作配置

当前规则：

- 后续 `xkit gen module` 默认只认 `*-target-config`
- 验证通过之前，不回写 `*-config`
- 只有当配置变更被确认后，才手工同步回 `*-config`
- 不再把 `admin/modules/<module>/<module>.yaml` 视为工作配置

## 常用命令

显式传参运行：

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

只传部分参数，交互输入剩余参数：

```powershell
& 'D:\GoProjects\XAdmin\xkit\examples\generateModule.ps1' `
  -WorkspaceRoot 'D:\GoProjects\XAdmin' `
  -ModuleName 'xdev'
```

跳过 dry-run：

```powershell
& 'D:\GoProjects\XAdmin\xkit\examples\generateModule.ps1' `
  -WorkspaceRoot 'D:\GoProjects\XAdmin' `
  -ModuleName 'xdev' `
  -SkipDryRun
```

跳过 TypeScript API 生成：

```powershell
& 'D:\GoProjects\XAdmin\xkit\examples\generateModule.ps1' `
  -WorkspaceRoot 'D:\GoProjects\XAdmin' `
  -ModuleName 'xdev' `
  -SkipTypeScript
```
