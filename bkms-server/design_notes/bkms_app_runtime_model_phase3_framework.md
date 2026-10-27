# 阶段 3 第二步：框架协议实现拆分（M4.5）

状态：已实现，待提 PR。

对应 [里程碑](./bkms_app_runtime_model_milestones.md) M4.5，处理遗留项 L-5（协议实现部分）、L-6、L-8、L-11。前置为 M4（PR #309）合入。除删除 standard 插件（§5）外只移动代码，不改逻辑。

---

## 1. 决策

| 问题 | 结论 |
| --- | --- |
| 目录位置 | 不在 `pkg` 下新增顶级目录：框架实现进 `appmodelcore/framework`，能力包（M3 的 `pkg/capability`）移到 `core/appruntime/capability`，见 §2 |
| 拆分方式 | 按原包整体搬，包边界不变。创建 / 更新服务（`trpc.go`、`taf.go`）先一起进 `appmodelcore/framework`，M5 合并编排时再拆出公共流程 |
| APM 服务名（L-6） | 解析实现进 `framework/{trpc,taf}/*conf`；两个错误进 `capability/apmservicename`；按解析器分派的入口留在 `bkmonitor` |
| 北极星服务名解析（L-7 后续） | 一起移入 `framework/trpc/trpcconf`，tRPC 配置结构只在一处维护 |
| 北极星服务名校验执行函数 | 从 `capability/polaris` 移到唯一调用方 `extension/addon/polaris`（包内不导出）；能力包只保留 `ServiceCheck` 声明。否则 `core/appruntime/capability` 要依赖 `workload` 下的 `trpcconf`，基础层反向依赖业务层 |
| standard 插件 | 删除，blank 不走插件，见 §5 |
| 执行代码去向 | 撤回 M3"执行代码逐步移入能力包"的规划，回到总体方案 §1 第 3 条与 §4.1：执行留在原模块，只通过能力 props 工作。L-1 按新目标关闭，L-5、L-9 目标同步修改 |

**为什么 APM 分派不进能力包：** 分派要调用 `trpcconf`，而 `trpcconf` 要返回能力约定的错误（调用方用 `errors.Is` 判断），能力包与 `trpcconf` 会互相引用。

**为什么解析单独成 `*conf` 子包：** `framework/trpc` 根包依赖 `mods`，`mods` 依赖 `capability/polaris`；若 `capability/polaris` 直接引用根包会成环。`*conf` 只依赖能力包和 `core/render`。

---

## 2. 目标结构

`pkg` 下不新增顶级目录：能力声明属于运行时模型，放在 `core/appruntime` 下；框架实现依赖 AppModel，放在 `appmodelcore` 下。

```
pkg/core/appruntime/
  runtimemod/              Mod、Spec、查表工厂（基础）
  runtimemod/mods/         Set、Of(app)（组装）
  capability/<能力>/        能力声明、Props、常量与错误约定（原 pkg/capability，M3 引入）
pkg/workload/appmodelcore/framework/
  trpc/                    渲染插件、frameworkConfig 解码、创建/更新服务（原 appmodelcore/trpc）
  trpc/admincmd/           HTTP admin 协议（原 appmodelcore/trpc/admincmd）
  trpc/patcher/            北极星注册信息补丁（原 appmodelcore/trpc/patcher）
  trpc/trpcconf/           tRPC 配置解析：APM 服务名、服务列表（原 bkmonitor、capability/polaris）
  taf/                     渲染插件、frameworkConfig 解码、创建/更新服务（原 appmodelcore/taf）
  taf/admincmd/            TAF 私有 admin 协议（原 appmodelcore/taf/admincmd）
  taf/tafconf/             TAF 配置解析：APM 服务名（原 bkmonitor）
```

依赖方向：框架包可以依赖能力包（取 Props、错误约定）；能力包只依赖 `appruntime`、`runtimemod`，不依赖 workload、extension 等业务层。需要框架知识的执行逻辑放在使用方：APM 分派在 `observability/bkmonitor`，北极星服务名校验在 `extension/addon/polaris`。

---

## 3. 命名调整

包名已表达框架，去掉重复前缀。只改名字，不改逻辑。

| 原标识符 | 新标识符 |
| --- | --- |
| `bkmonitor.GetTrpcApmServiceName` / `GetTafApmServiceName` | `trpcconf.APMServiceName` / `tafconf.APMServiceName` |
| `bkmonitor.TrpcServerConfig`、`ParseTrpcServerConfig` | `trpcconf.ServerConfig`、`ParseServerConfig` |
| `bkmonitor.ParseTelemetryConfig` 及 telemetry 相关类型 | `trpcconf.` 同名 |
| `bkmonitor.TafConfig`、`TafApplication`、`TafServer`、`ParseTafConfig` | `tafconf.Config`、`Application`、`Server`、`Parse` |
| `bkmonitor.ErrAPMConfigMissing` / `ErrAPMConfigParse` | `apmservicename.ErrConfigMissing` / `ErrConfigParse` |
| `polaris.parseTrpcServiceNames`（未导出） | `trpcconf.ParseServiceNames` |

---

## 4. 改动清单

| 类别 | 改动 |
| --- | --- |
| 目录移动 | `appmodelcore/{trpc,taf}` 及子包整体移到 `appmodelcore/framework/{trpc,taf}`；`pkg/capability` 整体移到 `core/appruntime/capability`；引用方只改 import 路径 |
| APM | `bkmonitor/telemetry_config.go` 拆入 `trpcconf/telemetry.go`、`tafconf/config.go`；`apm.go` 只保留 `GetApmServiceName` 分派 |
| 北极星 | 服务列表解析移入 `trpcconf/service.go`；校验执行函数从 `capability/polaris` 移到 `extension/addon/polaris/service_check.go`，改为包内不导出 |
| 测试 | 按被测代码随迁并拆分，用例数不变（bkmonitor 41 → bkmonitor 4 + trpcconf 25 + tafconf 12；polaris 解析 4 → trpcconf 4） |

---

## 5. 删除 standard 插件

| 事实 | 说明 |
| --- | --- |
| 来历 | 首次提交即存在，注释为"内部类型，提供通用渲染能力，可以用来测试" |
| 行为 | `Storage`、`ExtraResources`、`InitContainers` 全部返回空 |
| 生产使用 | M4 前按 `workload.type` 选插件，测试、正式环境没有 `workload.type=standard` 的数据，从未被选中 |
| 与 blank 的关系 | M4 把 blank 映射到它，只因两者都"不加资源"。blank 是产品框架；standard 只是 `workload.type` 的一个取值，M9 随该字段删除，两者不对等 |

**处理：**

- `GetFrameworkPlugin` 对 blank 返回 nil，builder 跳过插件步骤，plain 文件、devmode 等公共步骤照常执行。
- 删除 `appmodelcore/standard` 与 `appmodel.WorkloadTypeStandard`。

**影响：** blank 渲染结果不变（原插件不产生任何资源）；存量没有 blank 应用，也没有 `workload.type=standard` 的数据。

---

## 6. 验收

- `pkg` 下没有 `capability`、`framework` 顶级目录；`appmodelcore/{trpc,taf}` 已移除；`bkmonitor` 不再含 tRPC / TAF 配置解析。
- 除 §5 外，diff 只有文件移动、拆分、import 与标识符替换、注释。
- 代码中不再有 `standard` 插件与 `WorkloadTypeStandard`；blank 渲染不经过插件。
- `go build ./...`、`make lint`、`make test` 通过。
