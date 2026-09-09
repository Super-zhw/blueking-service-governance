# AppModel 应用类型体系

## 概述

本文描述 AppModel 大类的应用类型体系：trpc / taf 是框架特化类型，standard 是新增的「语言无关的通用
应用类型」，支持 go / python / nodejs 等语言。standard 复用 tRPC 应用的 AppModel + GameDeployment
部署链路，但**不绑定任何特定框架**。

与现有类型定位对照：

| 类型 | 大类 | 定义方式 |
|------|------|---------|
| helm / agones | Helm 大类 | Helm Chart + values |
| trpc / taf | AppModel 大类（框架特化） | 框架 + 框架配置文件 + appmodel |
| **standard** | AppModel 大类（通用基座） | 纯 appmodel + plain 配置文件 |

命名定为 `standard`，挂进 AppModel 大类，语言做子字段（`standardSpec.language`）。

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

## 构建（三种构建方式）

standard 需支持现有三种构建方式，`buildConfig.sourceType` 取值与 trpc/helm 一致：

| 构建方式 | sourceType | 说明 | 当前限制 | standard |
|---------|-----------|------|---------|---------|
| 源码仓库构建 | `codeRepository` | 从 Git 仓库构建镜像 | 平台通用构建仅 trpc-go | ✅ |
| 镜像仓库 | `imageRegistry` | 直接使用已构建镜像 | 现仅 helm | ✅（需放开） |
| 流水线构建 | `pipeline` | 蓝盾流水线构建 | 现仅 trpc | ✅（需放开） |

其中 `codeRepository` 内部又分两种镜像构建方式（`imageBuildMode`）：

- `repositoryDockerfile`：仓库内 Dockerfile 构建（默认），与类型无关。
- `platform`：平台通用构建（builder/runner 基础镜像 + 命令），当前后端硬校验仅 trpc-go，
  standard 需放开到 go/python/node。

构建配置统一存 `build_configs`（`sourceType` + 对应的 `imageBuildConfig` / `repoBuildConfig` /
`pipelineBuildConfig`），standard 与 trpc 同构、无需新增表；需改动的点是放开 imageRegistry /
pipeline / platform 三处现存的类型限制。

## 模型

### DB 落库（7 张表）

- `applications`：`{ ..., type:"standard", standardSpec:{language} }`
- `app_models`：`{ workload:{ type:"standard", name, command, args, envVars[], standardConfig:{language} } }`
- `app_specs`：与 trpc/taf 完全一致（默认 + 每环境一份）
- `build_configs`：与 trpc 一致
- `app_config_file_metas`：`{ _id, appID, name:"app.properties", configKind:"plain", mountPath:"/data/app/conf", isUnifiedConfig:true }`
- `app_config_files`：内容记录（`metaID` 关联 Meta，默认 + 各环境实例）
- `app_config_file_versions`：不可变版本快照

## 一期实现范围与依赖

- **不依赖 PR #142 的部分**（可先做）：新增类型并归入 AppModel 大类、语言子字段
  （`standardSpec.language`）、工作负载的 standard 配置、standard 的创建逻辑、`/standard-deploys`
  部署路由、envVars/AppSpec/组件复用、spec 更新接口、放开三种构建方式（imageRegistry / pipeline / platform）。
- **依赖 PR #142 的部分**（待合入后再接）：plain 配置文件创建（三表模型 + `configKind=plain`）。
- **一期不做**：框架特有能力（admin 命令、APM 提取）；polaris/devmode 的通用化下沉（后续单独排期）。
