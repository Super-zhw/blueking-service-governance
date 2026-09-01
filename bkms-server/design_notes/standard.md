# Standard 应用类型（通用/标准应用）

## 概述

Standard 是新增的一种「语言无关的通用应用类型」，支持 go / python / nodejs 等语言。它复用 tRPC 应用
的 AppModel + GameDeployment 部署链路，但**不绑定任何特定框架**。

与现有类型定位对照：

| 类型 | 大类 | 定义方式 |
|------|------|---------|
| helm / agones | Helm 大类 | Helm Chart + values |
| trpc / taf | AppModel 大类（框架特化） | 框架 + 框架配置文件 + appmodel |
| **standard** | AppModel 大类（通用基座） | 纯 appmodel + plain 配置文件 |

命名定为 `standard`，挂进 appmodel 家族（`IsAppModelType`），语言做子字段（`standardSpec.language`）。

## 与 AppModel 大类的关系

平台应用分两大族：**AppModel 族**（trpc/taf/standard）与 **Helm 族**（helm/agones）。

trpc / taf 本质上可视为 **standard 的「框架特化」版本**——它们在通用的 AppModel 基座（workload /
envVars / AppSpec / 组件 / GameDeployment 部署）之上，额外挂载了框架专属能力（框架配置文件、admin
命令、APM 提取、语言相关解析等）。standard 是去掉框架特化后的通用基座。

## 配置文件：使用 plain 配置（遵循 PR #142 设计）

standard 应用**有配置文件**，使用 PR #142 引入的 **plain 配置文件**（`configKind=plain`）。

plain 与 framework 配置文件的核心差异（决定 standard 的配置能力）：

1. 同一应用下可创建**多个** plain 文件（framework 仍是 1:1）
2. plain 文件可**选择挂载环境**（文件与环境 1:n），framework 仍挂载全部环境
3. plain 文件**存储完整内容**，非 patch 模式
4. plain 文件开启「按环境配置」后：环境未独立修改时不产生真实文件，内容随默认文件同步；一旦独立修改
   才产生一份独立文件，与默认文件解耦；可一键恢复默认（再次与默认同步）
5. plain 文件可**移除挂载**，移除后对应配置文件及版本信息全部删除

**三表模型与字段定义以 PR #142 为准，本文不重复给出。** standard 的配置能力直接复用该 PR 的
`appcfg` Meta 三表模型（`app_config_file_metas` / `app_config_files` / `app_config_file_versions`）
与 `plainfiles` 渲染，包含 `mountPath` 挂载、按环境配置、环境变量模板渲染。

standard 创建时：写入 `configKind=plain` 的配置文件（用户指定 name + mountPath + 内容）。后续
trpc/taf 的 framework 配置也迁移到这套模型，与 plain 共用能力。

## JSON 模型

### 创建请求 `POST /workspaces/:workspaceID/apps`

```jsonc
{
  "id": "demo-standard-a1b2c3",
  "name": "demo-standard",
  "type": "standard",                      // 新增枚举值
  "buildConfig": { "...同 trpc（codeRepository 源码构建）..." },
  "appModelSpec": {
    "command": ["./server"], "args": ["--config", "conf/app.properties"],
    "envVars": [{ "key": "POD_IP", "value": "", "description": "pod ip", "isSensitive": false }],
    "standardSpec": { "language": "go" }   // go / python / node
  }
}
```

### 详情响应 `GET /apps/:appID`

```jsonc
{ "data": { "id": "demo-standard-a1b2c3", "workspaceID": "bkms-workspace",
    "name": "demo-standard", "type": "standard", "displayName": "demo-standard", "creator": "admin",
    "buildConfig": { "...同 input..." },
    "appModelSpec": {
      "command": ["./server"], "args": ["--config", "conf/app.properties"],
      "envVars": [{ "key": "POD_IP", "value": "", "description": "pod ip", "isSensitive": false }],
      "components": [], "standardSpec": { "language": "go" } } } }
```

### DB 落库（7 张表）

- `applications`：`{ ..., type:"standard", standardSpec:{language} }`
- `app_models`：`{ workload:{ type:"standard", name, command, args, envVars[], standardConfig:{language} } }`
- `app_specs`：与 trpc/taf 完全一致（默认 + 每环境一份）
- `build_configs`：与 trpc 一致
- `app_config_file_metas`：`{ _id, appID, name:"app.properties", configKind:"plain", mountPath:"/data/app/conf", isUnifiedConfig:true }`
- `app_config_files`：内容记录（`metaID` 关联 Meta，默认 + 各环境实例）
- `app_config_file_versions`：不可变版本快照

## 能力分类（修正）

能力分三类，避免把「开放能力」误判为「框架特有」：

### 1. 通用基座能力（AppModel 族内 type 无关，standard 天然拥有）

创建/查询/列表/删除应用、构建配置、AppModel workload（command/args/envVars/image/resources）、
AppSpec（resources/updateStrategy/probe/lifecycle/labels/annotations 等 8 个 section）、环境变量
（app 级 + scoped）、组件、GameDeployment 部署（记录/状态/资源快照）、实例（list/watch/webconsole）、
拓扑、实例日志、镜像 tag 记录/晋级。

### 2. 开放类能力（跨类型、不应与框架强耦合，standard 后续也应支持）

- **polaris（服务注册/发现）**：能力本身与框架无关，任何 go/python/node 服务都可能注册北极星。
  当前实现耦合在 trpc plugin 里（对 trpc YAML 做 patch），后续应下沉为通用能力（如走组件或 AppSpec）。
- **devmode（开发模式热更新）**：能力本身与框架无关。当前实现按 trpc/taf 硬编码了 work 路径与脚本集，
  后续应抽象出通用 devmode（standard 用通用路径/脚本）。

### 3. 框架特有能力（真正与框架绑定，standard 一期不做）

- 框架配置文件（`configKind=framework`，trpc YAML / taf XML）
- admin 命令（trpc 走 HTTP + YAML 解析发现 admin IP/port；taf 走 Tars SDK + XML）
- APM 服务名提取（trpc 解析 telemetry/server 配置；taf 解析 XML）
- 语言相关解析（trpc 的 go/cpp 影响配置文件名后缀与 admin 解析路径）
- BSCP 元数据、平台 Dockerfile 构建（现仅 trpc）

## 各类型特有能力清单

| 能力 | helm | agones | trpc | taf | standard |
|------|:---:|:---:|:---:|:---:|:---:|
| Helm chart 来源（Helm/BCS/Git） | ✅ | ✅ | - | - | - |
| values 文件（默认 default） | ✅ | ✅ | - | - | - |
| Helm 部署（install/rollback） | ✅ | ✅ | - | - | - |
| GameDeployment 部署 | - | - | ✅ | ✅ | ✅ |
| 框架配置文件（framework） | - | - | ✅ | ✅ | - |
| plain 配置文件 | - | - | -（迁移中） | -（迁移中） | ✅ |
| admin 命令 | - | - | ✅ | ✅ | - |
| APM 服务名提取 | - | - | ✅ | ✅ | - |
| polaris（开放类，待下沉） | - | - | ✅（YAML patch） | - | 待支持 |
| devmode（开放类，待抽象） | - | - | ✅ | ✅ | 待支持 |
| 语言子字段 | - | - | go/cpp | - | go/python/node |
| AppSpec / envVars / 组件 | - | - | ✅ | ✅ | ✅ |

## AppModel 大类内差别梳理（问题 1、2）

**问题 1：如何维护 trpc/taf/standard 三者在功能特性上的区分。**

三者的差别本质是「框架特化度」不同，可抽象为三层，避免散落在几十处 `switch app.Type`：

1. **通用基座**（workload/envVars/AppSpec/组件/部署）→ 共享，type 无关。
2. **框架特化点**（框架配置文件、admin 命令、APM 提取、语言解析）→ 走 `workload/plugin` 注册表
   （`GetWorkloadPlugin(WorkloadType)`），每种框架一个 plugin，框架差异收敛到 plugin 内。
3. **开放能力**（polaris/devmode）→ 下沉为通用能力，通过 AppSpec section 或组件按需启用，与框架解耦。

这样新增一种框架（或 standard 增加某能力）只需：注册一个 framework plugin（或新增一个通用 section/组件），
而不是在每处 `switch app.Type` 补分支。

**问题 2：AppModel 大类内差别 + 未来框架扩展。**

AppModel 大类内，trpc 与 taf 的差别仅在三处：`configKind`+配置格式（yaml vs xml）、admin 命令协议
（HTTP vs Tars）、APM 提取（yaml vs xml）。其余（workload/AppSpec/envVars/部署）完全一致。

未来若要支持开源 Go/Python 框架，模式是：以 standard 为基座 + 一个 framework plugin + 一份
`XxxSpec`（类比 TrpcSpec）+ `configKind=framework` 的框架配置解析。平台只需新增「框架特化层」，
通用基座与开放能力全部复用——这正是把 trpc/taf 视为「specialized generic app」的收益。

## 一期实现范围与依赖

- **不依赖 PR #142 的部分**（可先做）：type 常量 + `IsAppModelType`、`standardSpec.language`、
  AppModel `standardConfig`、standard 创建 service、`/standard-deploys` 部署路由、envVars/AppSpec/组件
  复用、`UpdateAppStandardSpec`。
- **依赖 PR #142 的部分**（待合入后再接）：plain 配置文件创建（三表模型 + `configKind=plain`）。
- **一期不做**：框架特有能力（admin 命令、APM 提取）；polaris/devmode 的通用化下沉（后续单独排期）。
