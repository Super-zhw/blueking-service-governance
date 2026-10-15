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

package devmode

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"

	log "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/common/logging"
	k8sclient "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/infras/kubernetes/client"
)

const (
	// restartConcurrency 单次重启请求内并行执行的实例数上限
	restartConcurrency = 10
	// restartExecTimeout 单次读取 pid.conf 的 exec 超时时间
	restartExecTimeout = 10 * time.Second
	// restartStopTimeout 执行 stop.sh 并清除停止标志的超时时间（taf 优雅退出最长约 15s）
	restartStopTimeout = 60 * time.Second
	// defaultRestartWaitStartTimeout 等待 monitor 拉起进程（pid.conf 重新生成）的默认超时时间
	defaultRestartWaitStartTimeout = 30 * time.Second
	// defaultRestartPollInterval 轮询 pid.conf 的默认间隔
	defaultRestartPollInterval = time.Second
)

// 以下脚本通过 `bash -c <script> bash <mountPath>` 执行，$1 为脚本挂载路径。
// source utils.sh 时屏蔽其输出（会 source /etc/profile），避免污染 stdout。
const (
	// readPIDScript 读取 pid.conf，文件不存在时以非 0 退出码结束
	readPIDScript = `source "$1/utils.sh" >/dev/null 2>&1 && cat "${PID_CONF}" 2>/dev/null`

	// stopProcessScript 停止业务进程，随后删除 pid.conf 并清除停止标志，交由 monitor 重新拉起。
	// - source 失败时不执行 stop，避免出现「进程已被杀、停止标志无法清除」
	// - stop.sh 失败也会删除 pid.conf 并清除停止标志，让 monitor 兜底，退出码透传 stop.sh 的结果
	// - 删除 pid.conf：start.sh 确认进程启动后才会重新写入，以此判断进程已被拉起
	stopProcessScript = `source "$1/utils.sh" >/dev/null 2>&1 || { echo "failed to source $1/utils.sh" >&2; exit 1; }
bash "$1/stop.sh"
rc=$?
rm -f "${PID_CONF}"
clear_stop_flag
exit ${rc}`
)

// PodExecutor 在 pod 内执行命令，*k8sclient.PodClient 实现了该接口
type PodExecutor interface {
	Exec(ctx context.Context, namespace, podName, container string, cmd []string) (*k8sclient.ExecResult, error)
}

// RestartTarget 单个环境下的进程重启目标
type RestartTarget struct {
	// EnvName 环境名称
	EnvName string
	// Namespace 实例所在命名空间
	Namespace string
	// Container 业务容器名称
	Container string
	// MountPath devmode 脚本挂载路径
	MountPath string
	// InstanceIDs 目标实例（pod）名称列表
	InstanceIDs []string
	// Executor 访问该环境所在集群的 pod 命令执行器
	Executor PodExecutor
}

// RestartInstanceResult 单个实例的重启结果
type RestartInstanceResult struct {
	// EnvName 环境名称
	EnvName string
	// Instance 实例（pod）名称
	Instance string
	// Status 重启结果：success / failed
	Status PublishStatus
	// Message 结果说明
	Message string
	// OldPID 重启前的业务进程 PID（pid.conf 不存在时为空）
	OldPID string
	// NewPID 重启后的业务进程 PID
	NewPID string
}

// Restarter 开发模式业务进程重启器：stop 业务进程后由 monitor 重新拉起，pod 不重建
type Restarter struct {
	store            PublishRecordStore
	pollInterval     time.Duration
	waitStartTimeout time.Duration
}

// RestarterOption Restarter 可选配置
type RestarterOption func(*Restarter)

// WithRestartPollInterval 设置轮询 pid.conf 的间隔
func WithRestartPollInterval(d time.Duration) RestarterOption {
	return func(r *Restarter) { r.pollInterval = d }
}

// WithRestartWaitStartTimeout 设置等待进程被拉起的超时时间
func WithRestartWaitStartTimeout(d time.Duration) RestarterOption {
	return func(r *Restarter) { r.waitStartTimeout = d }
}

// NewRestarter 创建 Restarter
func NewRestarter(store PublishRecordStore, opts ...RestarterOption) *Restarter {
	r := &Restarter{
		store:            store,
		pollInterval:     defaultRestartPollInterval,
		waitStartTimeout: defaultRestartWaitStartTimeout,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// restartTask 单个实例的重启任务
type restartTask struct {
	target   *RestartTarget
	instance string
	recordID string
}

// Restart 同步重启所有目标实例，返回结果按 targets 及其 InstanceIDs 的顺序排列。
// 所有实例并行执行（上限 restartConcurrency），单个实例失败不影响其他实例；
// 仅写入发布记录失败时返回 error。
func (r *Restarter) Restart(
	ctx context.Context, appID, operator string, targets []RestartTarget,
) ([]RestartInstanceResult, error) {
	tasks, err := r.createPendingRecords(ctx, appID, operator, targets)
	if err != nil {
		return nil, err
	}

	// 记录更新不随请求取消，避免记录停留在 running
	recordCtx := context.WithoutCancel(ctx)
	results := make([]RestartInstanceResult, len(tasks))

	var g errgroup.Group
	g.SetLimit(restartConcurrency)
	for i, task := range tasks {
		g.Go(func() error {
			r.updateRecord(recordCtx, task.recordID, PublishStatusRunning, "")
			results[i] = r.restartInstance(ctx, task.target, task.instance)
			r.updateRecord(recordCtx, task.recordID, results[i].Status, results[i].Message)
			return nil
		})
	}
	_ = g.Wait()

	return results, nil
}

// createPendingRecords 为所有目标实例写入 pending 状态的 restart 记录
func (r *Restarter) createPendingRecords(
	ctx context.Context, appID, operator string, targets []RestartTarget,
) ([]restartTask, error) {
	var tasks []restartTask
	var records []PublishRecord
	for i := range targets {
		target := &targets[i]
		for _, instance := range target.InstanceIDs {
			tasks = append(tasks, restartTask{target: target, instance: instance})
			records = append(records, PublishRecord{
				AppID:    appID,
				EnvName:  target.EnvName,
				Action:   PublishActionRestart,
				Instance: instance,
				Status:   PublishStatusPending,
				Operator: operator,
			})
		}
	}
	if len(records) == 0 {
		return nil, nil
	}

	ids, err := r.store.Create(ctx, records)
	if err != nil {
		return nil, errors.Wrapf(err, "create restart records for app %s", appID)
	}
	if len(ids) != len(tasks) {
		return nil, errors.Errorf(
			"create restart records for app %s: expect %d ids, got %d",
			appID,
			len(tasks),
			len(ids),
		)
	}
	for i := range tasks {
		tasks[i].recordID = ids[i]
	}
	return tasks, nil
}

// updateRecord 更新重启记录状态（best-effort，失败仅记录日志）
func (r *Restarter) updateRecord(ctx context.Context, id string, status PublishStatus, message string) {
	if err := r.store.UpdateStatus(ctx, id, status, message); err != nil {
		log.WarnAttrs(ctx, "update devmode restart record failed",
			slog.String("record_id", id),
			slog.String("status", string(status)),
			slog.String("err", err.Error()),
		)
	}
}

// restartInstance 重启单个实例：读取旧 PID → stop 并清除停止标志 → 等待 monitor 拉起（pid.conf 重新生成）
func (r *Restarter) restartInstance(
	ctx context.Context,
	target *RestartTarget,
	instance string,
) RestartInstanceResult {
	result := RestartInstanceResult{EnvName: target.EnvName, Instance: instance, Status: PublishStatusFailed}

	// a. 读取旧 PID，pid.conf 不存在时为空，仅用于展示
	oldPID, _, err := r.readPID(ctx, target, instance)
	if err != nil {
		result.Message = fmt.Sprintf("failed to read pid.conf: %v", err)
		return result
	}
	result.OldPID = oldPID

	// b. stop 业务进程并清除停止标志；一旦开始执行必须跑完，不随请求取消
	stopCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), restartStopTimeout)
	defer cancel()
	stopRet, err := r.exec(stopCtx, target, instance, stopProcessScript)
	if err != nil {
		result.Message = fmt.Sprintf("failed to stop process: %v", err)
		return result
	}
	if stopRet.ExitCode != 0 {
		result.Message = fmt.Sprintf(
			"failed to stop process: stop.sh exited with code %d%s", stopRet.ExitCode, formatStderr(stopRet.Stderr),
		)
		return result
	}

	// c. 轮询 pid.conf，文件重新出现即视为 monitor 已拉起进程
	newPID, err := r.waitProcessStarted(ctx, target, instance)
	if err != nil {
		result.Message = fmt.Sprintf("process stopped, but %v", err)
		return result
	}

	result.Status = PublishStatusSuccess
	result.NewPID = newPID
	result.Message = fmt.Sprintf("process restarted, pid: %s -> %s", lo.Ternary(oldPID == "", "none", oldPID), newPID)
	return result
}

// waitProcessStarted 轮询 pid.conf 直到文件出现或超时，返回新 PID
func (r *Restarter) waitProcessStarted(ctx context.Context, target *RestartTarget, instance string) (string, error) {
	waitCtx, cancel := context.WithTimeout(ctx, r.waitStartTimeout)
	defer cancel()

	ticker := time.NewTicker(r.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-waitCtx.Done():
			if errors.Is(waitCtx.Err(), context.DeadlineExceeded) && ctx.Err() == nil {
				return "", errors.Errorf("timed out waiting for monitor to start it (pid.conf not created within %s)",
					r.waitStartTimeout)
			}
			return "", errors.Wrap(waitCtx.Err(), "waiting for process to start was interrupted")
		case <-ticker.C:
			// exec 异常（如网络抖动）不中断轮询，直到超时
			pid, exists, err := r.readPID(waitCtx, target, instance)
			if err == nil && exists {
				return pid, nil
			}
		}
	}
}

// readPID 读取 pid.conf，返回 PID 及文件是否存在
func (r *Restarter) readPID(ctx context.Context, target *RestartTarget, instance string) (string, bool, error) {
	execCtx, cancel := context.WithTimeout(ctx, restartExecTimeout)
	defer cancel()

	ret, err := r.exec(execCtx, target, instance, readPIDScript)
	if err != nil {
		return "", false, err
	}
	if ret.ExitCode != 0 {
		return "", false, nil
	}
	return strings.TrimSpace(ret.Stdout), true, nil
}

// exec 在目标实例内执行脚本，$1 为脚本挂载路径
func (r *Restarter) exec(
	ctx context.Context, target *RestartTarget, instance, script string,
) (*k8sclient.ExecResult, error) {
	cmd := []string{"bash", "-c", script, "bash", target.MountPath}
	ret, err := target.Executor.Exec(ctx, target.Namespace, instance, target.Container, cmd)
	if err != nil {
		return nil, errors.Wrapf(err, "exec in instance %s of env %s", instance, target.EnvName)
	}
	return ret, nil
}

// formatStderr 将 stderr 格式化为错误信息后缀
func formatStderr(stderr string) string {
	stderr = strings.TrimSpace(stderr)
	if stderr == "" {
		return ""
	}
	return ": " + stderr
}
