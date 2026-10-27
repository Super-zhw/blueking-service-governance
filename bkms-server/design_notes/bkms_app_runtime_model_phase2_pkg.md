# 阶段 2 收尾：能力包划分（M3）

状态：已合并，PR #307。

对应 [里程碑](./bkms_app_runtime_model_milestones.md) M3。前置为 M2（PR #302）合入。本阶段移动能力声明并调整包名与引用，顺带把 L-2、L-4、L-7 的执行代码移入能力包；只移动，不改逻辑。

---

## 1. 决策

| 问题 | 结论 |
| --- | --- |
| 目录粒度 | 一个能力一个目录；北极星两个能力合并为 `polaris/` |
| 移动范围 | 移动能力声明（Spec、Props、常量）；执行代码中 L-2、L-4、L-7 随本阶段移入，其余留在原包，见 §5 |
| 执行函数入参 | 能力包内的执行函数接收 `mods.Of(app).<能力>`，不自行解析技术栈，避免与 `mods` 循环依赖 |
| 框架协议实现 | `pkg/framework/*` 不在本阶段拆分，单独在 M4.5 做 |
| 组装包 | `Set`、`Of`、`Resolve`、`Matrix` 保留在 `runtimemod/mods`，调用方继续使用 `mods.Of`；该目录只剩组装与测试 |
| 基础包 | `pkg/core/appruntime/runtimemod` 不动 |

---

## 2. 目标结构

```
pkg/core/appruntime/runtimemod/       基础：ModID、ReasonCode、Mod、Spec、查表工厂、对外结构（不变）
pkg/core/appruntime/runtimemod/mods/  组装：Set、Of、Resolve、Matrix、Public（保留）
pkg/capability/frameworkfile/         框架配置文件
pkg/capability/devmode/               开发模式
pkg/capability/platformbuild/         平台构建：声明 + DockerfileLanguage
pkg/capability/admincommand/          admin 命令
pkg/capability/apmservicename/        APM 服务名提取
pkg/capability/polaris/               servicecheck.go（服务名校验：声明 + CheckServiceName）、frameworkpatch.go（注册信息注入）
pkg/capability/bscpworkloadmeta/      BSCP 元信息自动填充
```

依赖方向（不允许反向）：

```
执行代码 ──▶ runtimemod/mods ──▶ pkg/capability/<能力> ──▶ runtimemod ──▶ appruntime
    └──────────────────────────────▶ pkg/capability/<能力>（取 Props、常量，调用执行函数）
```

- `pkg/capability/<能力>`：不依赖 `runtimemod/mods`。声明只依赖 `runtimemod`、`appruntime`；执行函数可依赖底层领域包（如 `polaris` 依赖 `core/app/appcfg`），但不能依赖引用它的业务包。
- `runtimemod/mods`：依赖各能力包、`runtimemod`、`appruntime`、`core/app`、`common/logging`。它虽然位于基础包目录下，但职责是组装，不属于基础包。
- 反向依赖会形成循环，编译期即报错，不再另加 import 约束测试。

---

## 3. 命名调整

包名已表达能力，去掉标识符中重复的能力前缀。只改名字，不改值。

| 原标识符（`mods.`） | 新标识符 |
| --- | --- |
| `FrameworkFile`、`FrameworkFileProps`、`FrameworkFileFormat`、`FrameworkFileYAML`、`FrameworkFileTAF` | `frameworkfile.Spec`、`Props`、`Format`、`FormatYAML`、`FormatTAF` |
| `DevMode`、`DevModeProps`、`DevModeScriptSet`、`DevModeScriptSetTRPC`、`DevModeScriptSetTAF` | `devmode.Spec`、`Props`、`ScriptSet`、`ScriptSetTRPC`、`ScriptSetTAF` |
| `DevModePathScriptsSubDir`、`DevModeTrpcWorkPath`、`DevModeTrpcMountPath`、`DevModeTrpcBinaryPath`、`DevModeTafWorkPath`、`DevModeTafMountPath` | `devmode.ScriptsSubDir`、`TrpcWorkPath`、`TrpcMountPath`、`TrpcBinaryPath`、`TafWorkPath`、`TafMountPath` |
| `PlatformBuild`、`PlatformBuildProps` | `platformbuild.Spec`、`Props` |
| `AdminCommand`、`AdminCommandProps`、`AdminProtocol`、`AdminProtocolTRPC`、`AdminProtocolTAF` | `admincommand.Spec`、`Props`、`Protocol`、`ProtocolTRPC`、`ProtocolTAF` |
| `TrpcAdminConfigLayout`、`TrpcAdminLayoutNested`、`TrpcAdminLayoutFlat`、`TrpcAdminLayoutNestedPrefixed` | `admincommand.TrpcConfigLayout`、`TrpcLayoutNested`、`TrpcLayoutFlat`、`TrpcLayoutNestedPrefixed` |
| `APMServiceName`、`APMServiceNameProps`、`APMConfigParser`、`APMConfigParserTrpcYAML`、`APMConfigParserTafXML` | `apmservicename.Spec`、`Props`、`Parser`、`ParserTrpcYAML`、`ParserTafXML` |
| `PolarisServiceCheck`、`PolarisServiceCheckProps` | `polaris.ServiceCheck`、`ServiceCheckProps` |
| `PolarisFrameworkPatch`、`PolarisFrameworkPatchProps` | `polaris.FrameworkPatch`、`FrameworkPatchProps` |
| `BSCPWorkloadMeta`、`BSCPWorkloadMetaProps` | `bscpworkloadmeta.Spec`、`Props` |

保持不变：`mods.Set` / `Of` / `Resolve` / `Matrix`、`Set` 的字段名（`DevMode`、`PlatformBuild` 等）、`runtimemod.ModXxx` 常量、对外 JSON。

**同名包导入约定。** 引用方已有同名包时，能力包统一以 `<名称>cap` 为别名导入。已知冲突：

| 能力包 | 冲突的引用方 | 别名 |
| --- | --- | --- |
| `capability/devmode` | `extension/component/devmode`、`appspec/sections/devmode` 及引用它们的文件 | `devmodecap` |
| `capability/polaris` | `extension/addon/polaris` 等 5 个 `polaris` 包 | `polariscap` |

---

## 4. 改动清单

| 类别 | 文件 | 改动 |
| --- | --- | --- |
| 新增 | `pkg/capability/<能力>/`（8 个声明文件） | 从 `runtimemod/mods` 移入并按 §3 改名 |
| 删除 | `runtimemod/mods` 下 8 个声明文件 | 只保留 `set.go` 与测试 |
| 组装与测试 | `mods/set.go`、`mods/set_test.go`、`mods/declaration_test.go` | 改为引用各能力包，只改 import 与标识符 |
| 引用方 | `pkg/extension/component/devmode/{constants,devmode}.go`、`pkg/workload/appmodelcore/appdefaults/resolve.go`、`pkg/workload/appmodelcore/appspec/handler/dev_mode.go`、`pkg/workload/appmodelcore/{trpc,taf}/admincmd/admin_cmd.go`、`pkg/observability/bkmonitor/apm.go` | 只改引用 Props、常量的 import 与标识符；只调用 `mods.Of` 的文件不改 |
| 引用方测试 | `pkg/extension/component/devmode/helpers_test.go`、`pkg/workload/appmodelcore/appdefaults/resolve_test.go`、`pkg/workload/appmodelcore/trpc/admincmd/admin_cmd_test.go` | 同上 |
| L-2 | `extension/component/devmode/constants.go`；引用方 `appspec/sections/devmode/spec.go` 及若干测试 | 删除路径常量别名与重复的 `ScriptsSubDir`，引用方直接用 `capability/devmode` |
| L-4 | `build/image/build.go`、`build/build/service.go` | `getPlatformDockerfileLanguage` 移入 `platformbuild.DockerfileLanguage`，构建前置校验与实际构建共用 |
| L-7 | `extension/addon/polaris/trpc_validate.go` → `config_warnings.go`；`core/app/appcfg/trpc_utils.go` 删除 | 逐环境校验、提示文案、tRPC 服务名解析移入 `polaris.CheckServiceName`；addon 只保留 immediate 判断与应用加载；解析测试随迁 |
| 文档 | 总体方案 §4.5、phase1 §6、里程碑 M2 / M3 状态 | 同步新路径 |

实施顺序：

1. 新建 `pkg/capability/<能力>/`，移入声明并改名，补全导出注释。
2. `mods/set.go` 与测试改为引用各能力包。
3. 批量替换引用方 import 与标识符，按 §3 处理别名。
4. 删除 `mods` 下的声明文件，`go build ./...`、`make lint`、`make test`。
5. 更新文档。

---

## 5. 遗留项

本阶段处理了 L-2、L-4、L-7，其余执行代码仍在原包。下表登记所有遗留项，后续每个里程碑结束前按"检查方式"复核。

> 检查时注意：仓库 `.gitignore` 的 `build/` 会让 `rg` 跳过 `pkg/build`，检索需加 `--no-ignore-vcs`。

| 编号 | 能力 / 事项 | 执行代码当前位置 | 目标 | 处理时机 | 检查方式 |
| --- | --- | --- | --- | --- | --- |
| L-1 | devMode 执行代码 | `extension/component/devmode`（含 handler、脚本）、`appspec/sections/devmode`、`appspec/handler/dev_mode.go`、`appdefaults/resolve.go`、`workload/builder.go` | 收敛到 `capability/devmode` | 待排期，不晚于 M9 | `rg --no-ignore-vcs -l 'capability/devmode"' pkg` 结果只剩该目录 |
| L-2 ✅ | devMode 路径常量重复 | 已处理：别名常量删除，引用方直接用 `capability/devmode` | — | M3 | `rg -n 'TrpcWorkPath =' pkg/extension` 无结果 |
| L-3 | devMode 渲染仍读 AppSpec 持久化路径 | `workload/builder.go` 用 AppSpec 的 `workPath` / `mountPath` 覆盖能力参数 | 只用能力参数；AppSpec 中两字段删除 | M9 | `rg -n 'DevMode.WorkPath' pkg/workload/appmodelcore/workload` 无结果 |
| L-4 ✅ | platformBuild 执行代码 | 已处理：取 Dockerfile 语言移入 `platformbuild.DockerfileLanguage`；其余构建流程与技术栈无关，留在 `pkg/build` | — | M3 | `rg --no-ignore-vcs -n 'PlatformBuild\.Require' pkg` 无结果 |
| L-5 | adminCommand 执行代码 | `appmodelcore/{trpc,taf}/admincmd`、`workload/instance/handler/admin_cmd.go` | 能力壳进 `capability/admincommand`，协议实现进 `pkg/framework/*` | 协议实现随 M4.5；接口合并随 M6 | `ls pkg/workload/appmodelcore/*/admincmd` 不存在 |
| L-6 | apmServiceName 执行代码 | `observability/bkmonitor/{apm,telemetry_config}.go` | 解析器分派进 `capability/apmservicename`，解析实现进 `pkg/framework/*` | 随 M4.5 | `rg -n 'APMServiceName.Require' pkg/observability` 无结果 |
| L-7 ✅ | 北极星服务名校验 | 已处理：校验与 tRPC 服务名解析移入 `polaris.CheckServiceName`；addon 只保留 immediate 判断与应用加载 | 解析可随 M4.5 再移入 `pkg/framework/trpc` | M3 | `rg -n 'parseTrpcServiceNames' pkg` 只出现在 `capability/polaris` |
| L-8 | 北极星注册信息注入 | `appmodelcore/trpc/plugin.go`、`appmodelcore/trpc/patcher` | 开关进 `capability/polaris`，补丁器进 `pkg/framework/trpc` | 随 M4.5 | `rg -l 'package patcher' pkg/workload` 无结果 |
| L-9 | BSCP 元信息自动填充 | `extension/bscpcfg/handler/metadata.go`、`cmd/migration/app_bscpcfg_mgr.go` | `capability/bscpworkloadmeta`；工作负载类型按渲染结果取值 | 有使用方时，不晚于 M9（目前只影响展示） | `rg -n 'k8skind.GameDeploy' pkg/extension/bscpcfg cmd` 无结果 |
| L-10 | frameworkFile 无执行端读取 | 内容校验按存储的 `FileFormat` 分派（`core/app/appcfg`），未读能力声明 | 前端改读能力集合；后端校验按需接入 | M7 | `rg --no-ignore-vcs -n '\.FrameworkFile\.' pkg` 有消费方 |
| L-11 | 框架协议实现未拆分 | `appmodelcore/{trpc,taf}` | `pkg/framework/{trpc,taf,blank}` | M4.5 | `ls pkg/framework` 存在且 `appmodelcore/{trpc,taf}` 已移除 |

遗留项同步登记到里程碑文档的"遗留项"一节，处理后在两处同时标记完成。

---

## 6. 验收

- diff 只包含文件与函数移动、import 与标识符替换、注释；无逻辑改动。
- `runtimemod/mods` 下只剩 `set.go` 与测试，能力声明全部位于 `pkg/capability/<能力>/`。
- 对外 JSON 不变：golden 测试原样通过。
- `go build ./...`、`make lint`、`make test` 通过（`bkintegrations/bkci` 等既有环境相关失败除外）。
- 文档中的包路径与代码一致；遗留项已登记到里程碑文档。
