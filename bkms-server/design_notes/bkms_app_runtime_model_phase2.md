# 阶段 2 实现：能力执行端切换（M2）

状态：已合并，PR #302。

对应总体方案阶段 2 与里程碑 M2。阶段 2 分两步：[M1 能力声明](./bkms_app_runtime_model_phase1.md) 建立声明与查询路径，本文件记录第二步——把各能力的执行端切换到这份声明上。

---

## 1. 目标与边界

**目标：** 能力是否支持、支持时用什么参数，统一从 `mods.Of(app)` 读取；各业务模块删除 `AppType` / `Workload.Type` / `TrpcSpec.Language` 类型的判断分支。

**范围内：** 8 个能力，加上两处基座判断。

**范围外：**

- 渲染插件选择（按 framework 分派）→ M4。
- 创建编排合并、新字段写入 → M5。
- 一键构建部署的两个接口合并 → M6。
- 能力代码目录划分 → M3。

---

## 2. 逐能力改动

| 能力 | 改动位置 | 改动内容 | 行为变化 |
| --- | --- | --- | --- |
| devMode | `extension/component/devmode/{devmode,builder,constants}.go`、`devmode/handler/preflight.go`、`workload/appmodelcore/workload/builder.go`、`appspec/sections/devmode/spec.go`、`appspec/serializer/dev_mode.go`、`appspec/handler/dev_mode.go`、`appdefaults/resolve.go`、`trpc/trpc.go`、`taf/taf.go` | `Config` 持有 `DevModeProps`；脚本集、路径、二进制目录来自参数；`New` 不再按类型覆盖路径；`SetPathsByAppType` 改为 `SetPaths(workPath, mountPath)`；能力参数由 `mods.Of(app)` 解析 | TAF 且 `Workload.Type` 为空时改用 TAF 脚本（修复）；blank 开启 devMode 在渲染时返回能力错误；AppSpec 默认规则对不支持的技术栈跳过 devMode |
| platformBuild | `build/build/service.go`、`build/image/build.go` | 可用性用 `PlatformBuild.Require()`；Dockerfile 语言来自参数 | 错误文案统一为能力错误 |
| adminCommand | `trpc/admincmd/admin_cmd.go`、`taf/admincmd/admin_cmd.go` | 服务持有 `AdminCommandProps`；协议与配置键布局来自参数；删除语言常量 | 无（数据一致时） |
| apmServiceName | `observability/bkmonitor/apm.go`、`bkintegrations/handler/bkmonitor.go`、`appmodeldeploypoll/common.go` | 入口改为接收 `Application`，解析器来自参数 | 无 |
| polarisServiceCheck | `extension/addon/polaris/trpc_validate.go` | 改为加载 Application 并按能力判断 | `Workload.Type` 为空的 tRPC 应用开始收到校验警告（修复） |
| polarisFrameworkPatch | `trpc/plugin.go` | 渲染补丁链按能力判断 | 无 |
| bscpWorkloadMeta | `extension/bscpcfg/handler/metadata.go`、`cmd/migration/app_bscpcfg_mgr.go` | 是否自动填充按能力判断 | 无 |
| frameworkFile | `core/app/appcfg/config_kind_policy.go`、`appcfg/taf.go` | 内容校验按格式选择 YAML / TAF 解析器 | 合法 TAF 配置不再被 YAML 校验误拦（修复） |

基座判断：

- `appcfg/appcfgfiledef/handler.go`：plain 配置文件仅允许 AppModel 类应用创建。

---

## 3. 关键实现约定

**技术栈来源。** 所有能力解析统一走 `mods.Of(app)`，内部用 `appruntime.CapabilityStack` 从 Application 快照取技术栈：框架取 `Application.Framework`，为空时由旧 `type`（trpc / taf）推出；语言取 `Application.Language`，为空时取 `TrpcSpec.Language`。不读取 AppModel，因此只记录在 `TrpcConfig.Language` 的语言不参与解析；存量 TAF 没有语言属于正常情况。

**错误语义。** `Of` 不返回 error：技术栈非法（类型与框架冲突、未知语言等）时记录告警日志，返回全部能力不支持、原因为 `invalid_stack` 的能力集合。能力是否可用统一由 `Require()` 判断，不可用时返回 `ErrModUnsupported`，渲染和构建链路直接上抛。

**TAF 内容校验。** `appcfg` 内用 TAF 解析器校验 TAF 格式内容。该解析器较宽松，本次只解决"把合法 TAF 当 YAML 拒绝"的问题，不引入更严格的语法校验。

**BSCP 工作负载类型。** 仍写入 `GameDeployment`。BSCP 元信息是应用级、创建时初始化，而实际渲染类型随环境变化（联邦环境渲染 Deployment），且 `WorkloadKind` 目前不在注入链路中被消费。按环境覆盖 kind 需要先定义应用级元信息的表达方式；目前只影响展示，暂不处理，见遗留项 L-9。

**devMode 的启用判断。** 能力判断集中在 `devmode.RequireProps(app)`，配置写入、预检、重启、渲染四个入口都经过它：技术栈不支持开发模式时，写入 `enabled=true` 返回参数错误，预检与重启返回请求错误，渲染返回能力错误。是否启用、环境是否允许仍由 AppSpec 与 devmode 模块负责。

---

## 4. 上线前检查

在只读副本执行 `scan_app_runtime_model`，重点关注：

1. **语言冲突**：`Application.Language`、`TrpcSpec.Language`、`TrpcConfig.Language` 三处不一致的应用。M2 后 `CapabilityStack` 会直接报错，这些应用将不可用，需先修数据。
2. **TAF 且 `Workload.Type` 为空**：M2 后脚本集由 tRPC 变为 TAF，需确认影响范围符合预期。
3. **目录外语言**：若存在 java / nodejs / python 应用，切换后会被技术栈校验拒绝。

---

## 5. 测试

- 新增或更新：devMode 无能力参数报错、TAF 内容按格式校验、admin 各语言布局、polaris 校验加载 Application、platform build 错误语义、appdefaults 按能力参数生成路径。
- `go build ./...`、`make lint` 通过；extension、workload、build、observability、core/app、server、bkintegrations 全量测试通过。
- `pkg/extension/bscpcfg/service` 的 4 个失败依赖外部 BSCP 域名解析，在 base/main 上同样失败，属既有环境问题。

---

## 6. 后续

| 里程碑 | 与本阶段的关系 |
| --- | --- |
| M3 | 按能力划分 pkg，纯移动；顺带处理 `polaris.go` 的文件名问题 |
| M4 | 渲染插件按 framework 分派 |
| M5 | 创建编排合并，写入新字段 |
| M6 | 合并一键构建部署接口 |
