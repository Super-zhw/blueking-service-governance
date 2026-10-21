# BKMSApp 运行时能力集合（RuntimeModSet）：阶段 1 方案

状态：M1 已实现（纯新增，未接入执行路径）。执行端切换、pkg 划分分别为 [里程碑](./bkms_app_runtime_model_milestones.md) M2、M3。

对应父文档 [bkms_app_runtime_model.md](./bkms_app_runtime_model.md) 的阶段 2（能力层），依赖已完成并冻结的 [阶段 0](./bkms_app_runtime_model_phase0.md) 与阶段 1（`pkg/core/appruntime` 的 Stack、catalog、Normalize、frameworkConfig）。

本文件只解决两件事：**能力（Mod）如何组织与声明**、**能力代码如何划分**。API、前端、数据迁移不在本阶段范围。

---

## 1. 决策

| 问题 | 结论 |
| --- | --- |
| 能力维度 | 以 **framework** 为主维度；language 只影响 catalog 合法组合、`platformBuild` 可用性、`adminCommand` 的 props 取值 |
| 声明形式 | 维度表（`FrameworkTable` / `LanguageTable`），不写 (framework, language) 规则枚举 |
| n×m | 不需要手写：每个能力只写自己支持的框架条目，解析分支与能力数量无关 |
| 声明位置 | 编译进代码。不落 DB / YAML，不持久化解析结果，不做进程内缓存 |
| 组装方式 | 集中式 `Resolve(stack)`，`ModSet` 为强类型 struct；不使用 init / FX 动态注册 |
| 运行时开关 | 预留 gate，只允许"收紧"，不允许放宽 |
| 解析失败语义 | 技术栈非法 → error；能力不支持 → `Supported=false` + `reasonCode` |
| 语言为空 | 允许（桥接期），`Any` 匹配空值、`Only(...)` 不匹配 |

---

## 2. 声明规范

包：`pkg/core/appruntime/runtimemod`。只依赖 `appruntime` 与标准库（`lo` / `copier` 允许），不 import `core/app`、appmodel、component、具体框架或 handler。

```go
type Dimension int

const (
    ByFramework Dimension = iota // 默认
    ByLanguage
    ByPair
    StackIndependent // 所有 bkmsApp 框架均支持，含后续新增框架
)

// 字段与方法全部导出：M3 起声明要搬到各 owner 包的 mod/ 子包，
// 未导出成员无法跨包构造。
type Def[P Props] struct {
    ID        ID
    Dimension Dimension
    Framework FrameworkTable[P] // map[string]P，缺键 = 不支持
    Language  LanguageTable[P]  // map[string]P
    Pair      PairTable[P]      // 极少使用
    Universal *P                // 与技术栈无关的能力，优先级最低
    Languages Match             // 第二轴限制，零值 = 不限制
}

type Props interface {
    // Public 返回可对外投影的 props；没有可公开字段时返回 nil。
    Public() any
}

type Mod[P Props] struct {
    ID         ID
    Supported  bool
    ReasonCode ReasonCode
    Props      P
}

func (d Def[P]) Resolve(stack appruntime.Stack) Mod[P]
func (m Mod[P]) Require() (P, error) // 不支持时返回 ErrModUnsupported + ID + reason
```

`ModSet` 与 `Resolve` 在 M1 内联于 `runtimemod`。M3 把声明移入 owner 包后，`ModSet` 必须独立成薄组装包（`runtimemod` 只留规范定义），否则 `runtimemod → capability/* → runtimemod` 形成循环。

`reasonCode` 取值：

| 值 | 含义 |
| --- | --- |
| `unsupported_app_type` | helm / agones |
| `unsupported_stack` | 该框架（或框架+语言）无对应条目 |
| `incomplete_language` | 规则要求具体语言但 language 为空（仅桥接期） |
| `disabled_by_policy` | 被运行时 gate 收紧（预留） |

解析规则：

1. `Type != bkmsApp` → 全能力 `unsupported_app_type`。
2. 按主维度取条目，`Languages` 不匹配或 language 为空 → `unsupported_stack` / `incomplete_language`。
3. 技术栈本身非法（未知框架、非法组合）由 `Resolve` 返回 error，不进入 ModSet。

错误语义必须区分：**不支持不是错误**，只有执行器调 `Require()` 才变成 `ErrModUnsupported`。

### 2.1 Stack 来源

`Application.RuntimeMods()`：

```go
stack, err := appruntime.CapabilityStack(a.RuntimeSnapshot())
return runtimemod.Resolve(stack)
```

`CapabilityStack` 是桥接期函数（标记 BRIDGE-ONLY，阶段 6 删除）：调用 `Inspect`，忽略 `incomplete_language`，其余 blocking issue 返回错误。

语言为空必须放行：TAF 存量无 language；`PUT /trpc-spec` 只写 `Workload.TrpcConfig.Language`，部分 tRPC 应用 Application 侧也为空。若套用 `Normalize` 的 blocking 判定，现网可用的 devmode 会被拒。

---

## 3. 能力目录

首期 10 个，均已声明（`runtimemod` 内各文件）。执行端在 M2 统一切换，autoDeploy 随 M5 接入。

| ID | 维度 | trpc+go | trpc+cpp | taf | blank | props（执行器用） |
| --- | --- | --- | --- | --- | --- | --- |
| `frameworkFile` | ByFramework | yaml | yaml | xml | — | format |
| `plainConfigFiles` | StackIndependent | ✓ | ✓ | ✓ | ✓ | — |
| `devMode` | ByFramework | trpc 脚本集 | trpc 脚本集 | taf 脚本集 | — | scriptSet、workPath、mountPath、binaryPath |
| `platformBuild` | ByFramework + `Only(go)` | ✓ | — | — | — | dockerfileLanguage |
| `adminCommand` | ByFramework + 语言对 | nested 键 | flat 键 | 私有协议 | — | protocol、configLayout |
| `apmServiceName` | ByFramework | trpcYAML | trpcYAML | tafXML | — | parser |
| `polarisServiceCheck` | ByFramework | ✓ | ✓ | — | — | — |
| `polarisFrameworkPatch` | ByFramework | ✓ | ✓ | — | — | — |
| `bscpWorkloadMeta` | ByFramework | ✓ | ✓ | — | — | workloadKind |
| `autoDeploy` | ByFramework | ✓ | ✓ | ✓ | — | —（M5 合并两套实现后接入） |

解析时 `Pair` 优先于 `Framework`，`Framework` 优先于 `Language`。`adminCommand` 的 tRPC 两种配置键布局用 `Pair` 表达，因此不写 `Framework[trpc]`，否则会掩盖"该语言未声明"的情况。

`plainConfigFiles` 由 workload 构建的公共步骤（`plainfiles.BuildPlainConfigFiles`）实现，与框架插件无关，因此用 `Universal` 声明：新增框架自动获得，不需要逐个登记。它是 blank 的主要配置途径。plain 已在现网使用，是 bkmsApp 的通用配置路径。现状：plain 只在 AppModel 渲染链路挂载，但 def 创建接口不按应用类型拦截（UI 对 helm 不暴露）；M2 起创建入口经 `PlainConfigFiles.Require()` 判断。

`Def.Meta()` / `Declarations()` 暴露声明维度与条目名，供测试校验覆盖度，后续 catalog API 复用。

### 3.0 全量能力核对

按 router 注册的功能域逐项核对（41 个域），差异只有四类归属：

| 归属 | 判据 | 处理 |
| --- | --- | --- |
| 技术栈能力 | 同一功能在不同 framework / language 下是否可用或参数不同 | 建 Mod（上表 10 项） |
| 管理方式 | AppModel 与 Helm/Agones 的差异 | 继续用 `IsAppModelType`，不进 ModSet |
| 协议实现 | 同一能力的不同实现细节 | 留在框架适配器内部，由 props 选择 |
| 历史兼容 | 旧字段、旧类型、旧路由 | 桥接期代码，M9 删除 |

| 能力 / 差异点 | 归属 | 说明 |
| --- | --- | --- |
| 框架配置文件格式、TAF 配置合并 | 技术栈能力 | `frameworkFile` 的 props；合并算法仍属协议实现 |
| devMode 脚本与路径 | 技术栈能力 | `devMode` |
| 平台生成 Dockerfile | 技术栈能力 | `platformBuild` |
| admin 命令协议与配置键布局 | 技术栈能力 | `adminCommand` |
| APM 服务名解析器 | 技术栈能力 | `apmServiceName` |
| 北极星服务名校验 | 技术栈能力 | `polarisServiceCheck` |
| 北极星注册信息注入 | 技术栈能力 | `polarisFrameworkPatch` |
| BSCP 自动推导 workload name/kind | 技术栈能力 | `bscpWorkloadMeta` |
| plain 配置文件 | 技术栈能力（全体支持） | `plainConfigFiles` |
| 一键构建部署 | 技术栈能力 | `autoDeploy`，M5 接入 |
| 渲染插件选择、空 `Workload.Type` 兜底 | 技术栈能力 | M4 改为按 framework 分派，不单独建 Mod |
| 组件引用仅统计 trpc | 历史缺陷 | phase0 已定单独立项 |
| 容器名内置变量仅 trpc/taf | 管理方式 | M4 改为按 `IsAppModelType` |
| 集群 addon `requiredForAppTypes` | 管理方式 / workload 形态 | M4 改为按 workload kind 与 CRD 依赖，不按技术栈 |
| 部署、预检、状态、总览、实例、AppSpec、组件、环境变量 | 管理方式 | `IsAppModelType` |
| GPA、HostPort、端口池、北极星注册与权重、环境事件 | 管理方式 / 环境 | GPA 需 AppModel 且非联邦；HostPort 只看联邦；端口池与环境相关 |
| 构建配置、构建记录、监控仪表盘、用户组、集成类接口、审计 | 无技术栈差异 | 不建 Mod |
| Helm chart 来源、values、Helm 部署、chart 构建 | 独立链路 | 不属于 bkmsApp 能力集合 |
| `TrpcSpec`、`TrpcConfig`、`TafConfig`、`Workload.Type`、旧路由 | 历史兼容 | M9 删除 |

核对中发现的三处待处理项：

1. **adminCommand 的语言覆盖。** 执行端支持 go/cpp/java/node/python 五种布局，已全部登记进 `Pair`。java/node/python 为目录外语言（`appruntime.LanguageJava` 等常量，不在 `KnownLanguages`），技术栈校验会拒绝，条目当前不可达；存量扫描确认后再决定是否纳入目录。
2. **平台构建判定重复。** `build/service.go` 与 `PlatformBuildDef` 各写一份，M2 只保留后者。
3. **未被引用的桥接代码。** `trpc/config.go`、`taf/config.go` 的 `DecodeConfig` 与 `pkg/core/render/migrate` 目前没有生产调用方，M2 决定接入框架适配器或删除。

### 3.1 语言维度的真实占比

首期为纯语言维度的只有 `platformBuild`；`adminCommand` 是框架维度 + 语言子覆盖（`TrpcAdminConfigLayout`）。因此 `stack.Language` 在解析中只用于三处：catalog 合法组合校验、`platformBuild` 查表、`adminCommand` 的 props 取值。

### 3.2 props 要点

- `devMode`：`WorkPath` / `MountPath` 由 props 给出，不再由 `app.Type` / `Workload.Type` 推导。`taf-start.sh` 与 `TrpcBinaryPath` 属于 `scriptSet` 差异，保留在 props。
- `platformBuild`：`DockerfileLanguage` 替代 `Application.TrpcSpec.Language` 的读取。
- 执行器内部 key（scriptSet、parser、configLayout、workloadKind、dockerfileLanguage）**不对外暴露**。

### 3.3 命名统一

`blank` 为 framework 字面值；`standard` 仅保留为内部 workload 插件名。产品文案、catalog、ModSet 一律用 `blank`。

---

## 4. 对外 JSON 契约

`ModSetPublic()` 即后续 API 响应体，本阶段用 golden 测试锁定。

```json
{
  "schemaVersion": 1,
  "type": "bkmsApp",
  "framework": "trpc",
  "language": "cpp",
  "mods": {
    "frameworkFile":       { "supported": true,  "props": { "format": "yaml" } },
    "plainConfigFiles":    { "supported": true },
    "devMode":             { "supported": true,  "props": { "workPath": "/data/bkms/dev-mode/trpc", "mountPath": "/data/bkms/dev-mode/trpc/configmap-scripts" } },
    "platformBuild":       { "supported": false, "reasonCode": "unsupported_stack" },
    "adminCommand":        { "supported": true },
    "apmServiceName":      { "supported": true },
    "polarisServiceCheck":    { "supported": true },
    "polarisFrameworkPatch":  { "supported": true },
    "bscpWorkloadMeta":    { "supported": true },
    "autoDeploy":          { "supported": true }
  }
}
```

规则：

- `mods` 用 map，`encoding/json` 按键排序，输出稳定。
- `props` 为空时整个键省略，不输出 `null`。
- `framework` / `language` 为空时省略键，不输出空串。
- helm / agones：`framework`、`language` 省略，全部 `unsupported_app_type`。
- `Matrix()` 对 catalog 全组合生成同样结构，供前端创建表单使用（挂 API 属 M6）。

---

## 5. 校验与测试

不设启动期校验。约束全部由测试承担。

M1 交付：

| 测试 | 保证 |
| --- | --- |
| golden JSON（5 组合 + Matrix） | 能力矩阵与对外契约不随重构漂移；新增 framework 或能力时 golden 变红，强制人工表态 |
| 表键 ⊆ catalog、ID 唯一 | 拼写错误的 framework 不会静默失效 |
| `ModSet` 字段与 `all()` 反射对齐 | 新增能力漏进 `Declarations()` 或公开投影时立即失败 |
| 不可变性 | 修改返回的 props 不影响后续解析 |
| import 约束 | `runtimemod` 不引入领域依赖 |

M2 交付：执行器覆盖测试（各 owner 包），保证凡 `Supported=true` 的组合都能构造执行器，不出现"宣称支持但无实现"。

测试按仓库约定：Ginkgo、英文描述、`dbfactory`。`runtimemod` 为纯函数包，不需要 DB。

---

## 6. 能力代码的 pkg 划分

目标形态（**分阶段搬，不在本阶段执行**）：

```
pkg/core/appruntime/
  runtimemod/          # 规范定义、ModSet、Resolve、Public、Matrix、首期 10 个声明

pkg/capability/        # 有技术栈开关的模块，一个能力一个目录
  <cap>/{mod, service, handler, serializer}

pkg/framework/         # 框架协议实现，不含能力开关
  trpc/ taf/ blank/

pkg/base/              # 管理方式基座，无 Mod
  deploy/ render/ instance/ envvars/ ...
```

划分原则：

1. 有 `mod` 的是能力；无 `mod` 的是基座。判据：该功能是否会因技术栈不同而不支持。
2. 能力壳（是否支持）与协议实现分离。`pkg/framework/*` 内允许保留 trpc/taf 命名。
3. `pkg/extension/` 逐步退役；组件体系（`extension/component`、portpool）属基座，不再接收新能力。

搬迁在 M3 一次完成，纯移动，不含逻辑改动。

---

## 7. 实施

按 [里程碑](./bkms_app_runtime_model_milestones.md)：M1 本文声明与测试，M2 执行端切换，M3 pkg 划分。

M2 补充说明：

- AppSpec 的 `workPath` / `mountPath` 今日实际不生效（`devmode.New` 会再次覆盖），props 作为唯一来源与现状一致；字段暂按原样读写，M9 清理。
- adminCommand 的配置键布局按语言区分：go/python 为 `server.admin.ip`，cpp/node 为 `server.admin_ip`，java 为 `server.admin.admin_ip`。五种语言均已登记；目录外语言纳入目录后自动生效。
- frameworkFile 的 format 驱动内容校验：现有 `FrameworkPolicy.ValidateContent` 对所有框架按 YAML 解析，含 `key: value` 行的合法 TAF 配置会被误拦，M2 一并修复。

---

## 8. 验收

- 10 个能力均有声明，`Matrix()` 覆盖 catalog 全部合法组合，5 份 golden JSON 通过。
- `trpc+cpp` 支持 devMode、不支持 platformBuild；`blank+go` 全部框架能力不支持；helm 全部 `unsupported_app_type`。
- M2 完成后，能力判断不再读取 `AppType` / `Workload.Type` / `TrpcSpec.Language`。
- `make lint`、相关包测试、`make test` 通过。
