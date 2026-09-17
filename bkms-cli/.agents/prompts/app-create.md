# bkms-cli app create Reference

`bkms-cli app create` 用于从 YAML spec 文件创建应用。支持的应用类型：trpc、taf、helm、agones。

YAML spec 文件结构与后端 API 请求体一致。顶层 `id` 字段可选（不填时由 `name` + 随机后缀自动生成）。

## 常用场景

```bash
# 从 YAML spec 文件创建应用（使用默认工作空间）
bkms-cli app create -f app.yaml

# 显式指定工作空间
bkms-cli app create -f app.yaml --workspace ws-demo
```

创建成功后输出应用 ID、名称和类型：

```
✓ App created successfully
  ID:   my-app-a1b2
  Name: my-app
  Type: trpc
```

## 完整 YAML 示例

### trpc + imageRegistry

```yaml
name: my-trpc-app
type: trpc

buildConfig:
  sourceType: imageRegistry
  imageBuildConfig:
    name: example.com/my-team/my-trpc-app:latest
    # 可选，私有仓库用户名；密码写入后加密存储
    username: myuser
    # 可选，私有仓库密码（加密存储）
    # password: secret

appModelSpec:
  trpcSpec:
    language: go
    fileName: trpc_go.yaml
    filePath: /app/conf
```

### trpc + codeRepository

```yaml
# tRPC 应用，使用代码仓库构建
name: my-trpc-app
type: trpc

buildConfig:
  sourceType: codeRepository
  repoBuildConfig:
    type: TGit
    repoAlias: my-repo
    repoURL: https://github.com/my-team/my-trpc-app.git
    defaultBranch: main
    dockerfile: Dockerfile

appModelSpec:
  command: ["/app/bin/server"]
  args: ["--config", "/app/conf/trpc_go.yaml"]
  envVars:
    - key: ENV
      value: prod
      # description 为可选字段
      description: "Runtime environment"
  trpcSpec:
    language: go
    fileName: trpc_go.yaml
    filePath: /app/conf
    fileContent: |
      server:
        service:
          - name: trpc.app.server.service
            ip: 0.0.0.0
            port: 8080
```

### trpc + pipeline（蓝盾流水线构建）

```yaml
# tRPC 应用，由蓝盾流水线负责构建
name: my-trpc-app
type: trpc

buildConfig:
  sourceType: pipeline
  pipelineBuildConfig:
    pipelineID: p-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    params:
      BKMS_IMAGE_NAME: my-trpc-app
      BKMS_IMAGE_TAG: latest

appModelSpec:
  trpcSpec:
    language: go
    fileName: trpc_go.yaml
    filePath: /app/conf
```

### helm + GitRepo

```yaml
# Helm 应用，使用 Git 仓库中的 Chart
name: my-helm-app
type: helm

buildConfig:
  sourceType: imageRegistry
  imageBuildConfig:
    name: example.com/my-team/my-helm-app:v1.0.0

helmSpec:
  helmSource:
    repoType: GitRepo
    valueFiles:
      - values.yaml
      - values-prod.yaml
    gitRepoConfig:
      type: TGit
      repoAlias: my-chart-repo
      repoURL: https://github.com/my-team/helm-charts.git
      revision: main
      sourceDir: charts/my-app
```

### helm + HelmRepo

```yaml
# Helm 应用，使用 Helm 仓库中的 Chart
name: my-helm-app
type: helm

buildConfig:
  sourceType: imageRegistry
  imageBuildConfig:
    name: example.com/my-team/my-helm-app:v1.0.0

helmSpec:
  helmSource:
    repoType: HelmRepo
    helmRepoConfig:
      repoURL: https://charts.example.com/stable
      chartName: my-chart
```

### helm + BCSRepo

```yaml
# Helm 应用，使用 BCS 仓库中的 Chart
name: my-helm-app
type: helm

buildConfig:
  sourceType: imageRegistry
  imageBuildConfig:
    name: example.com/my-team/my-helm-app:v1.0.0

helmSpec:
  helmSource:
    repoType: BCSRepo
    bcsRepoConfig:
      projectCode: my-project
      repoName: my-bcs-repo
      chartName: my-chart
```