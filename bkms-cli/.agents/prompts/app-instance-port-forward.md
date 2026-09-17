# bkms-cli app instance port-forward Reference

`bkms-cli app instance port-forward` 用于将本地 TCP 端口转发到应用实例 Pod 的指定端口。后端通过 WebSocket 隧道与目标 Pod 建立 TCP 连接，适用于开发调试、内部服务访问等场景。

该命令的工作方式类似于 `kubectl port-forward`，但仅针对单个 Pod 实例。如需访问多个 Pod，请启动多个 port-forward 命令并使用不同的本地端口。

## 命令格式

```
bkms-cli app instance port-forward [flags] [LOCAL_PORT:]REMOTE_PORT
```

### 端口参数格式

| 格式 | 说明 | 示例 |
|------|------|------|
| `REMOTE_PORT` | 本地端口 = 远程端口 | `8080` → localhost:8080 → pod:8080 |
| `LOCAL_PORT:REMOTE_PORT` | 指定本地端口和远程端口 | `18080:8080` → localhost:18080 → pod:8080 |

### Flags

| Flag | 类型 | 必填 | 说明 |
|------|------|------|------|
| `--app` | string | 是 | 应用 ID |
| `--env` | string | 是 | 环境名称 |
| `--instance` | string | 是 | 目标 Pod 实例 ID |
| `--local-address` | string | 否 | 本地监听地址，默认 `127.0.0.1` |

## 常用场景

### 转发本地端口到 Pod 端口（相同端口号）

```bash
# 将本地 8080 端口转发到 pod-1 的 8080 端口
bkms-cli app instance port-forward --app myapp --env test --instance pod-1 8080
```

### 转发本地端口到 Pod 端口（不同端口号）

```bash
# 将本地 18080 端口转发到 pod-1 的 8080 端口
bkms-cli app instance port-forward --app myapp --env test --instance pod-1 18080:8080
```

### 使用自定义本地地址

```bash
# 监听所有网络接口（允许其他机器访问）
bkms-cli app instance port-forward --app myapp --env test --instance pod-1 18080:8080 --local-address 0.0.0.0
```

### 访问 Pod 内的管理命令接口

```bash
# 转发本地 9090 到 Pod 的 admin 端口
bkms-cli app instance port-forward --app myapp --env test --instance pod-1 9090:9090

# 然后使用 curl 访问
curl http://127.0.0.1:9090/cmds
```

### 访问 Pod 内的 HTTP 服务

对于 Pod 中暴露的 HTTP 服务端口，转发后可在本地直接使用 `curl` 或浏览器等 HTTP 客户端访问：

```bash
# 将本地 18080 转发到 Pod 的 HTTP 服务端口 80
bkms-cli app instance port-forward --app myapp --env test --instance pod-1 18080:80

# 在本地直接通过 curl 访问
curl http://127.0.0.1:18080/api/v1/health
curl -X POST http://127.0.0.1:18080/api/v1/data -H "Content-Type: application/json" -d '{"key":"value"}'
```

### 访问 Pod 内的 tRPC 服务

对于 Pod 中暴露的 tRPC 服务端口，转发后可在本地通过 tRPC 客户端程序直接对本地代理端口发起 RPC 调用：

```bash
# 将本地 18080 转发到 Pod 的 tRPC 服务端口 8080
bkms-cli app instance port-forward --app myapp --env test --instance pod-1 18080:8080
```

再编写 tRPC 客户端代码，将 target 指向本地代理端口即可调用远端服务（伪代码）：

```go
// 构建 tRPC 客户端，target 指向本地 port-forward 代理地址
proxy := pb.NewXxxServiceClientProxy(
    client.WithTarget("ip://127.0.0.1:18080"),
    client.WithProtocol("trpc"),
)

// 像调用本地服务一样发起 RPC 请求
resp, err := proxy.SayHello(ctx, &pb.HelloRequest{Msg: "hello"})
```

> **说明**：tRPC 客户端需要对应服务生成的 proto 代码（`pb` 包）。将 `target` 设置为 `ip://127.0.0.1:<LOCAL_PORT>` 即可通过 port-forward 隧道透明访问远端 Pod 上的 tRPC 服务。

## 使用约束

- `--app`、`--env`、`--instance` 为必填参数。
- 端口参数为位置参数，必须提供且仅支持一个。
- 端口号范围为 1-65535。
- 目标实例必须处于 Running 状态。
- 需要具备应用编辑权限和目标环境的部署操作权限。
- 非 loopback 地址监听（如 `0.0.0.0`）时会显示安全警告。
- 按 `Ctrl+C` 可停止端口转发。

## 输出说明

命令启动后会输出以下结构化日志：

```
level=INFO msg="forwarding established" listen_address=127.0.0.1:18080 instance_id=pod-1 remote_port=8080
```

每当有新连接建立时会输出：

```
level=INFO msg="handling connection" instance_id=pod-1 local_port=18080 remote_port=8080
```

连接失败时会输出错误信息：

```
level=ERROR msg="port-forward connection failed" instance_id=pod-1 remote_port=8080 error=<message>
```

停止时输出：

```
Port-forward stopped
```
