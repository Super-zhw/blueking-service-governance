# 阶段 3 第一步：运行链路识别 bkmsApp（M4）

状态：已实现，待提 PR。

对应 [里程碑](./bkms_app_runtime_model_milestones.md) M4，总体方案阶段 3 的 3A、3C、3E。前置为 M3（PR #307）合入。

---

## 1. 目标与边界

**目标：** 运行链路不再比较 `type == "trpc"` 这类字面值，改为按管理方式（`IsAppModelType` / `IsHelmBasedType`）或 framework 判断。只改判断逻辑，不改数据。

**原因：** blank 只能是 `bkmsApp`，M7 后所有新应用也是 `bkmsApp`，现有的字面值比较会全部落空。

**范围外：**

| 事项 | 去向 |
| --- | --- |
| `appmodelcore/{trpc,taf}` 拆到 `pkg/framework/*`（L-5、L-6、L-8、L-11） | M4.5，纯目录移动单独成里程碑 |
| 集群 addon 改为按管理方式表达 | M5，随 blank 后端闭环 |
| BSCP 元信息中的工作负载类型（L-9） | 暂不处理，见 §5 |
| 创建流程中的字面值（`handler/app.go`、`serializer/validate.go`） | M5 |
| 旧 API 输出投影（`serializer/app.go` 按 `Workload.Type` 输出 trpcSpec / tafSpec） | M6 / M9 |
| 指标归一中的 `"trpc"`、`"taf"`（`metrics.go`） | 不改：调用方都传 `DeployKindAppModel`，两个值只是兼容别名 |
| 组件引用查询 `OfType(AppTypeTRPC)`（`comp_ref.go`） | 不改：`OfType` 已同时匹配 bkmsApp + trpc |

---

## 2. 改动

**framework 的来源：** 统一取 `mods.Of(app).Stack.Framework`，与能力集合同一来源。技术栈非法时为空，按"不匹配"处理。`appcfg` 不能依赖 `mods`（`mods` → `capability/polaris` → `appcfg` 会成环），由调用方取值后传入。

| 事项 | 位置 | 现状 | 改为 |
| --- | --- | --- | --- |
| 渲染插件选择 | `workload/builder.go`、`workload/plugin/plugin.go` | 按 `AppModel.Workload.Type` 选插件，空值默认 trpc | `GetWorkloadPlugin(workloadType)` 改为 `GetFrameworkPlugin(framework)`：trpc → trpc 插件，taf → taf 插件，blank → standard 插件；去掉空值兜底；取不到 framework 时渲染报错 |
| 部署预检（泳道） | `deploy/pre_checker.go` | `switch` 只认 trpc / taf，bkmsApp 落入 default 报错 | 按 `IsHelmBasedType` / `IsAppModelType` 分支 |
| 容器名内置变量 | `envvars/listers.go` | 只给 trpc / taf 注入 | 按 `IsAppModelType` |
| 旧类型专属接口的类型校验 | `deploy/handler/appmodel_deploy_steps.go`（部署预检、部署状态）、`build/autodeploy/handler/build_auto_deploy.go`、`core/app/handler/app_trpc.go`、`app_taf.go` | 比较 `app.Type` 与接口期望的类型 | 比较 framework：trpc 接口要求 framework=trpc，taf 同理；错误码与提示文案不变（Bruno 用例断言 `invalid app type`） |
| 框架文件的名称、目录能否修改 | `appcfg/config_kind_policy.go` 及调用方 | `AllowNameUpdate` / `AllowMountDirUpdate` 比较 `"trpc"`、`"taf"` | 入参由 appType 改为 framework，取值规则不变；`CreateCfgFileParams` 增加 `Framework`，`FileDefUpdate`、`HelmFileUpdate` 的 `AppType` 改为 `Framework`（这两处只用于该判断），由 handler 与 trpc / taf 服务传入 |
| 集群 addon 必装检查 | `clusteraddon/query.go`，调用方 `deploy/deploy_precheck.go`、`deploy/appmodel/service/service.go` | 用 `app.Type` 匹配 `requiredForAppTypes` | 新增 `AddonAppTypeOf(app)`：AppModel 类应用用 framework 匹配，Helm / Agones 仍用 type；内置 YAML、库内记录、前端不变 |

`AppModel.Workload.Type` 创建时照常写入，渲染不再读取，M9 删除。

---

## 3. 存量影响

插件选择的依据从 `Workload.Type` 换成 Application 的技术栈，两处不一致的数据渲染结果会变。上线前用扫描脚本逐个应用对比 M4 前后的插件（脚本为本地 `build/m4_render_plugin_scan.js`，未入库）：

| 环境 | AppModel 类应用 | 不变 | 插件变化 | M4 后渲染报错 |
| --- | --- | --- | --- | --- |
| 测试 | 115（trpc 102，taf 13） | 115 | 0 | 0 |
| 正式 | 356（trpc 355，taf 1） | 356 | 0 | 0 |

两个环境的应用均为 `type` 与 `workload.type` 一致，没有 bkmsApp。M4 对存量渲染无影响。

两个环境各有 1 条孤儿 AppModel（应用已不存在），不参与渲染，与 M4 无关，可另行清理。

---

## 4. 测试

- trpc / taf 渲染测试原样通过，资源名（`trpc-init`、`/taf-config-template` 等）不变。
- 新增：插件按 framework 选择，blank 选 standard，忽略 `workload.type`；技术栈非法时渲染报错；bkmsApp 通过泳道预检、获得容器名变量；框架文件修改规则按 framework 判断；addon 匹配值按 framework 得出。
- 旧接口类型校验只是把比较对象换成 framework，trpc / taf 互斥由现有 Bruno 用例覆盖。bkmsApp 目前无法通过接口创建，这条路径暂无接口级用例，M5 打通 blank 后端闭环时补上。
- Helm / Agones 现有测试不回归。
- 相关 Bruno 用例（`tests/apis/appmodel-deploys`）通过。

---

## 5. 留意事项

| 事项 | 说明 | 处理时机 |
| --- | --- | --- |
| addon 按管理方式表达 | YAML 里 `requiredForAppTypes: [trpc, taf]` 表达的其实是"AppModel 渲染出 GameDeployment 所需的组件"，与框架无关。blank 同样渲染 GameDeployment，但 M4 后仍匹配不到。需改为按管理方式表达，查询兼容旧值，并同步前端集群组件页的分组 | M5 |
| BSCP 元信息中的工作负载类型（L-9） | 初始化时写死 GameDeployment，联邦环境实际渲染 Deployment。AppModel 应用注入 BSCP 时直接合并 PodSpec，不读取该字段，目前只影响展示 | 有实际使用方时处理，不晚于 M9 |

---

## 6. 验收

- 运行链路中不再有对 trpc / taf 字面值的判断（§1 列出的范围外项除外）。
- trpc / taf 新旧渲染结果一致；Helm / Agones 无回归。
- `go build ./...`、`make lint`、`make test`、相关 Bruno 用例通过。
