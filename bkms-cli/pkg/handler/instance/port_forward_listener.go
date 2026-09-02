/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 服务治理 (BlueKing Service Governance) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 *  http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 * to the current version of the project delivered to anyone in the future.
 */

package instance

import (
	"context"
	"log/slog"
	"net"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-cli/pkg/client"
)

// RunPortForwardListener 启动本地 TCP listener 并管理端口转发生命周期。
func RunPortForwardListener(
	ctx context.Context,
	cli portForwardTunnelClient,
	opts PortForwardOptions,
) error {
	ln, err := net.Listen("tcp", opts.ListenAddress())
	if err != nil {
		return errors.Wrapf(err, "listen on %s failed, local port may already be in use", opts.ListenAddress())
	}
	defer ln.Close()
	slog.Debug("tcp listener started", "address", opts.ListenAddress())

	if !isLoopbackAddress(opts.LocalAddress) {
		slog.Warn("localAddress may be reachable from other machines", "local_address", opts.LocalAddress)
	}
	slog.InfoContext(ctx, "forwarding established",
		"listen_address", opts.ListenAddress(),
		"instance_id", opts.InstanceID,
		"remote_port", opts.RemotePort,
	)

	// 使用 semaphore 控制最大并发连接数。
	sem := semaphore.NewWeighted(maxConcurrentConnections)
	g, gCtx := errgroup.WithContext(ctx)

	// accept 循环在独立 goroutine 中运行，ctx 取消时通过关闭 listener 退出。
	g.Go(func() error {
		<-gCtx.Done()
		slog.Debug("context canceled, shutting down listener")
		return ln.Close()
	})

	g.Go(func() error {
		return acceptAndForward(gCtx, ln, sem, g, cli, opts)
	})

	if err = g.Wait(); err != nil {
		slog.Error("port-forward listener error", "error", err)
	}
	slog.Info("port-forward stopped")
	return nil
}

func acceptAndForward(
	ctx context.Context,
	ln net.Listener,
	sem *semaphore.Weighted,
	g *errgroup.Group,
	cli portForwardTunnelClient,
	opts PortForwardOptions,
) error {
	for {
		conn, acceptErr := ln.Accept()
		if acceptErr != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				return errors.Wrap(acceptErr, "accept connection")
			}
		}

		slog.Debug("accepted new connection",
			"local_addr", conn.LocalAddr().String(),
			"remote_addr", conn.RemoteAddr().String(),
		)

		if acquireErr := sem.Acquire(ctx, 1); acquireErr != nil {
			_ = conn.Close()
			return acquireErr //nolint:wrapcheck // context 已取消，直接返回
		}

		g.Go(func() error {
			defer sem.Release(1)

			slog.Debug("opening port-forward tunnel",
				"instance_id", opts.InstanceID,
				"remote_port", opts.RemotePort,
			)

			if connErr := handlePortForwardConnection(ctx, cli, opts, conn); connErr != nil {
				slog.Error("port-forward connection failed",
					"instance_id", opts.InstanceID,
					"remote_port", opts.RemotePort,
					"error", connErr.Error(),
				)
			} else {
				slog.Debug("port-forward connection closed normally",
					"instance_id", opts.InstanceID,
					"remote_port", opts.RemotePort,
				)
			}
			return nil
		})
	}
}

func isLoopbackAddress(address string) bool {
	ip := net.ParseIP(address)
	return ip != nil && ip.IsLoopback()
}

func handlePortForwardConnection(
	ctx context.Context,
	cli portForwardTunnelClient,
	opts PortForwardOptions,
	localConn net.Conn,
) error {
	defer localConn.Close()
	if cli == nil {
		return errors.New("port-forward tunnel client is nil")
	}

	tunnel, err := cli.OpenPortForwardTunnel(ctx, opts.AppID, opts.EnvName, client.PortForwardTunnelOptions{
		InstanceID: opts.InstanceID,
		RemotePort: opts.RemotePort,
		LocalPort:  opts.LocalPort,
	})
	if err != nil {
		return errors.Wrap(err, "open port-forward tunnel failed")
	}
	defer tunnel.Close()

	slog.InfoContext(ctx, "handling connection",
		"instance_id", opts.InstanceID,
		"local_port", opts.LocalPort,
		"remote_port", opts.RemotePort,
	)

	slog.Debug("starting bidirectional copy",
		"local_port", opts.LocalPort,
		"remote_port", opts.RemotePort,
	)

	return copyBidirectional(ctx, localConn, tunnel)
}
