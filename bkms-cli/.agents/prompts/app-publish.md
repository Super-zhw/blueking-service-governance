# bkms-cli app publish Reference

`bkms-cli app publish` 用于将本地二进制文件发布到应用开发模式（dev mode）容器中，并串行对目标实例执行上传与重启流程。

该命令仅适用于已开启 dev mode 的非正式环境。发布前会先校验应用、环境、dev mode 配置和目标实例；发布时会把本地文件上传到容器工作目录下的 `bin` 路径，并执行 `restart.sh` 完成切换。

BCS Auth Info 由服务端自动通过当前登录凭据获取，无需手动传入 token。

实例选择有两种方式：

- 使用 `--instance-ids` 显式指定要发布的实例，多个实例用逗号分隔。
- 使用 `--all` 自动获取当前环境下所有 `Running` 状态实例。

## 常用场景

发布到指定实例，适合小范围验证。

```bash
# 发布到单个实例
bkms-cli app publish --app myapp --env stage -f ./bin/server --instance-ids pod-1

# 发布到多个实例
bkms-cli app publish --app myapp --env stage -f ./bin/server --instance-ids pod-1,pod-2
```

自动发布到当前环境全部 Running 实例，适合开发模式下做整体验证。

```bash
# 自动选择所有 Running 实例
bkms-cli app publish --app myapp --env stage -f ./bin/server --all
```

## publish

`publish` 命令负责把本地二进制文件上传到开发模式容器，并逐个实例执行重启脚本。命令执行前会打印本次将操作的实例总数和实例名列表，随后按实例串行处理。

```bash
# 发布到指定实例
bkms-cli app publish --app myapp --env test -f ./bin/server --instance-ids pod-1

# 发布到多个实例
bkms-cli app publish --app myapp --env test -f ./bin/server --instance-ids pod-1,pod-2

# 自动发布到所有 Running 实例
bkms-cli app publish --app myapp --env test -f ./bin/server --all
```

### 参数说明

| 参数 | 必填 | 说明 |
|------|------|------|
| `--app` | 是 | 应用 ID |
| `--env` | 是 | 环境名称 |
| `-f, --file` | 是 | 本地二进制文件路径 |
| `--instance-ids` | 否 | 实例 ID（pod 名），逗号分隔；与 `--all` 二选一 |
| `--all` | 否 | 发布到所有 Running 实例；与 `--instance-ids` 二选一 |

### 使用约束

- `--instance-ids` 和 `--all` 互斥，二选一。
- 未使用 `--all` 时，必须显式提供 `--instance-ids`。
- 目标环境必须是已开启 dev mode 的非正式环境。
