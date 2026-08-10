# Metrics（Prometheus）

基于 `prometheus/client_golang` 的指标采集与暴露能力。

## 架构概览

```
┌──────────────┐  :8081/debug/metrics  ┌────────────────────┐
│  Prometheus  │ ◄───── scrape ─────── │  Metrics HTTP Server│
│  (Operator)  │                       │  (独立端口)          │
└──────────────┘                       └────────────────────┘
       ▲                                        │
       │ ServiceMonitor                         │ promhttp.Handler()
       │                                        ▼
       │                               ┌────────────────────┐
       │                               │  promauto 全局指标   │
       └───────────────────────────────│  Counter            │
                                       └────────────────────┘
```

## 包结构

| 文件 | 职责 |
|------|------|
| `metrics.go` | 使用 `promauto` 定义全局 Counter 指标（自动注册）、`recordFailure` 通用记录函数及各场景便捷函数 |
| `server.go` | Metrics HTTP Server 启停逻辑（`StartServer(ctx)` / `StopServer(ctx)`） |

## 配置

`config.yaml` 中新增：

```yaml
metrics:
  port: 8081   # 默认端口，范围 1-65535，无效值自动回退为 8081
```

对应结构体 `MetricsConfig`（`pkg/common/config/types.go`），加载时自动校验回退（`pkg/common/config/config.go`）。
