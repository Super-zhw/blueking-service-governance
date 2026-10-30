# 阶段 3B / 3F / 4：新 API 与 blank 闭环（M5 + M6）

状态：已提 PR #320。

对应 [里程碑](./bkms_app_runtime_model_milestones.md) M5、M6 合并实施，总体方案阶段 3B、3F、4（服务端）。前置为 M4.5（PR #311）合入。

---

## 1. 目标与边界

**目标：** 提供 bkmsApp 的新 API，新应用只写新字段；用新 API 跑通 tRPC、TAF、blank 三种应用的完整链路。旧 API 保持原样，旧数据由 M8 统一迁移。

**范围外：**

| 事项 | 原因 / 去向 |
| --- | --- |
| 旧 API 改为薄适配层、新旧字段双写 | 旧 API 按旧方式运行，旧数据由 M8 回填 |
| 旧 API 拒绝切换技术栈 | 属于旧 API 行为，不改；冲突数据由 M8 扫描处理 |
| 创建失败的跨集合一致性 | 改造前已存在，与本需求无关；新 API 沿用"Application 最后写"的顺序 |
| 旧路由埋点 | 切换时统一下线旧路由，不需要观察数据 |
| 合并 `admin-cmds` / `taf-admin-cmds` | 两者已按能力判断协议（M2），bkmsApp 应用可直接使用 |
| 环境级能力接口 `envs/:envName/runtime-mods` | 总体方案标注"若产品需要"，目前无需求 |
| `startup-config` 并发写冲突检查 | AppModel 没有版本号，旧更新接口也没有；单独评估 |
| 前端改用新 API、输出 `type=bkmsApp` 给旧客户端 | M7；CLI 需同步识别 `bkmsApp` 并改用新接口 |
| 修改 AppModel 的审计记录含敏感环境变量明文 | 旧 trpc-spec / taf-spec 与新 startup-config 都存在，环境变量接口同样存在；单独 PR 修复并清理历史记录 |

---

## 2. 决策

| 问题 | 结论 |
| --- | --- |
| 新 API 是否加开关 | 不加。正式环境不主动调用，只在测试环境使用 |
| 新建应用入口 | 新开路由 `POST /workspaces/:workspaceID/bkmsapps`，旧入口不变 |
| 路由命名 | AppModel 是内部概念，对外按应用类型 `bkmsApp` 命名：`bkmsapps`、`bkmsapp-deploys`（`bkms` 是整个项目的名字，单独使用有歧义）；构建部署只有平台托管应用才有，用 `build-deploys`；修改接口覆盖启动命令、参数、脚本与 frameworkConfig，用 `startup-config` |
| 旧 spec 更新接口遇到 bkmsApp 应用 | `PUT trpc-spec`、`PUT taf-spec` 拒绝 `type=bkmsApp`，避免把旧字段写回新应用 |
| 框架差异如何提供 | 不引入框架适配器接口。是否创建框架文件及其格式读能力 `frameworkFile`；frameworkConfig 按框架查表调用对应的 `DecodeConfig` |
| frameworkConfig | 新建时写 `{}` 与版本 1。trpc / taf 的 `Config` 去掉已收敛到 def 的 fileName / filePath；blank 只接受 `{}` |
| 新建应用的 language | 必填，按框架目录校验；TAF 只接受 `cpp`。语言目录新增 nodejs、python，只对 blank 开放（blank 支持 go、cpp、nodejs、python），blank 没有任何框架特性，与语言无关 |
| 存量 TAF 的 language | 已核对：正式环境 1 个应用，确认是 C++；测试环境 13 个为测试数据。由 M8 回填为 `cpp`，不手工修复 |
| 集群组件登记值 | YAML 在 trpc / taf 之后追加 `bkmsApp`，旧值标为待废弃，M9 删除；AppModel 应用按 `bkmsApp` 与框架匹配，命中任一即可；前端暂时隐藏 `bkmsApp` 分组，界面不变 |

---

## 3. 接口

| 接口 | 说明 |
| --- | --- |
| `POST /workspaces/:workspaceID/bkmsapps` | 创建 bkmsApp 应用 |
| `PUT /apps/:appID/startup-config` | 修改启动命令、参数、脚本与 frameworkConfig，不能修改 framework / language |
| `GET /app-runtime-definitions` | 框架目录（框架、支持的语言）与能力矩阵 |
| `GET /apps/:appID/runtime-mods` | 应用的能力集合及公开参数 |
| `/apps/:appID/envs/:envName/bkmsapp-deploys` | 部署记录、预检、部署、下架、资源快照列表与详情、最新状态；复用现有通用 handler，只要求是 AppModel 应用（trpc、taf、bkmsApp），不校验框架 |
| `POST /apps/:appID/envs/:envName/build-deploys` | 一键构建部署；同上 |
| 详情 `GET /apps/:appID`、列表 `GET /workspaces/:workspaceID/apps` | `framework`、`language` 已由 M0 的兼容投影返回；bkmsApp 应用的详情另外返回 `appModelSpec.frameworkConfig` |

创建请求：

| 字段 | 说明 |
| --- | --- |
| `name`、`id`、`buildConfig` | 与旧接口一致 |
| `framework`、`language` | 必填，按框架目录校验 |
| `appModelSpec` | 启动命令、参数、启动脚本、环境变量、`frameworkConfig`（缺省为 `{}`） |
| `frameworkConfigFile` | 框架配置文件的名称、挂载目录、内容。框架支持框架配置文件时必填，否则（blank）传入即报错 |

修改请求中 `frameworkConfig` 的语义按总体方案 §5.1：缺省不修改，`null` 拒绝，`{}` 为显式替换并重新校验。

---

## 4. 创建流程

编排服务位于 `workload/appmodelcore/lifecycle`，只服务新 API：

1. 校验技术栈、frameworkConfig、框架配置文件输入，全部通过后再写库。
2. 写构建配置（与旧接口相同的校验）。
3. 能力 `frameworkFile` 支持时，按其格式创建框架配置文件 def 与内容。
4. 写 AppModel：启动配置、环境变量、`frameworkConfig`、`frameworkConfigVersion=1`；不写 `Workload.Type`、`TrpcConfig`、`TafConfig`。
5. 按工作空间默认规则写 AppSpec（与旧流程相同）。
6. 最后写 Application：`type=bkmsApp`、`framework`、`language`；不写 `TrpcSpec`。
7. 异步任务与审计与旧接口相同：自定义镜像记录、默认告警策略、操作记录。

旧的 `framework/{trpc,taf}/{trpc,taf}.go` 继续服务旧 API，M9 删除。

---

## 5. 其他改动

| 事项 | 改动 |
| --- | --- |
| APM 服务名同步 | 部署轮询先判断应用是否支持 `apmServiceName`，再读框架文件；否则 blank 每次部署都会记一条 Error |
| 集群组件 | 内置 YAML 追加 `bkmsApp`、保留 trpc / taf，启动时同步到库。新旧版本都能匹配，滚动发布与回滚期间不会漏查必装组件；前端暂时隐藏 `bkmsApp` 分组，M7 开放 bkmsApp 应用时切换为只按 `bkmsApp` 分组 |
| frameworkConfig 类型 | trpc / taf 的 `Config` 改为空结构；新增 blank 的解码；`Workload.GetFrameworkConfig` 统一读取（未写版本视为没有） |
| 空间组件引用检查 | 原来只统计 trpc 应用，改为统计全部 AppModel 应用（含 taf、bkmsApp），被引用的组件不能删除 |
| 配置文件定义 | `POST /apps/:appID/app-config-file-defs` 拒绝为 `frameworkFile` 不支持的 AppModel 应用（blank）新建 framework 文件，返回 400；Helm、trpc、taf 不受影响。旧 `app-config-files` 即将下架，不再改动 |
| 旧 CreateApp | 抽出创建前校验、写构建配置、创建后异步任务三段公共逻辑供新接口复用，行为不变 |
| 文档 | 里程碑合并 M5、M6，删除"MongoDB 事务保护"，M4.5 标为 ✅ #311；阶段 0 文档记录 TAF 存量核对结果 |

---

## 6. 与总体方案、里程碑的差异

| 原规划 | 本方案 | 原因 |
| --- | --- | --- |
| §5.1 复用 `POST /workspaces/:workspaceID/apps` | 新开 `bkmsapps` | 旧 API 不接收新参数，避免回归验证 |
| §5.3 旧 create / spec 接口转换为统一领域命令；M6 旧接口改薄适配层 | 旧接口不改 | 同上 |
| M5 新旧字段双写 | 不双写 | 新接口只写新字段，旧数据由 M8 统一 |
| M5 MongoDB 事务保护 | 不做 | 与本需求无关，见 §1 |
| M5 旧路由埋点 | 不做 | 切换时统一下线 |
| M6 合并管理命令接口 | 不合并 | 现有接口已按能力分派 |
| §5.1 路由 `appmodel-spec`、`appmodel-deploys[...]`、`appmodel-build-deploys` | `startup-config`、`bkmsapp-deploys[...]`、`build-deploys` | 见 §2 路由命名 |
| §2.2 框架适配器提供差异 | 能力声明加各框架的 `DecodeConfig` | 新接口不写旧投影后，剩余差异不足以构成接口 |

---

## 7. 测试

- 单测：
  - 三种框架经编排服务创建后的字段正确，没有旧字段；
  - 非法技术栈、blank 传入框架配置文件、frameworkConfig 含未知 key 时被拒绝；
  - 修改接口的 frameworkConfig 语义；
  - 集群组件按新值与旧值匹配；
  - 空间组件引用检查覆盖 trpc、taf、blank；
  - 配置文件定义按应用类型与 `frameworkFile` 能力校验 configKind。
- 能力查询接口、旧 spec 接口拒绝 bkmsApp 由 Bruno 用例覆盖。
- Bruno：新增 `tests/apis/bkmsapps`。`api` 目录覆盖创建、详情、能力查询、修改及各类拒绝场景（含 blank 新建 framework 文件被拒）；`{blank,trpc,taf}-lifecycle` 走 `bkmsapp-deploys` 完成预检 → 部署 → 资源快照 → 最新状态 → 下架，blank 另外先创建 plain 配置文件。部署类用例由 `RUN_BKMSAPP_DEPLOY_API_TESTS` 控制，默认跳过。
- 回归：旧 API 的 Bruno 用例、`make lint`、`make test`。
- Swagger：已执行 `make apidocs`，新接口均已生成；旧接口不变。

---

## 8. 验收

- 新 API 创建的三种应用只含新字段，能完整走完部署链路。
- 旧 API 行为不变，旧 Bruno 用例通过。
- blank 应用不产生框架资源，不触发框架能力的报错日志。
