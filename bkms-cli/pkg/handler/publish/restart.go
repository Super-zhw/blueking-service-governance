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

package publish

import (
	"context"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-cli/pkg/client"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-cli/pkg/utils/clierr"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-cli/pkg/utils/console"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-cli/pkg/utils/params"
)

// restartStatusFailed 实例重启失败状态
const restartStatusFailed = "failed"

// BuildRestartOptions 根据命令行参数（逗号分隔）构造进程重启请求参数：
//   - --env 与 --env-type 互斥，必须且只能指定其中一个
//   - --instance-ids 仅可与单个 --env 一起使用
func BuildRestartOptions(envs, envTypes, instanceIDs string) (client.DevModeRestartOptions, error) {
	envNames := lo.Uniq(params.NormalizeInstIDs(envs, ","))
	types := lo.Uniq(params.NormalizeInstIDs(envTypes, ","))
	instances := lo.Uniq(params.NormalizeInstIDs(instanceIDs, ","))

	switch {
	case len(envNames) > 0 && len(types) > 0:
		return client.DevModeRestartOptions{}, clierr.Usagef("--env and --env-type cannot be used together")
	case len(envNames) == 0 && len(types) == 0:
		return client.DevModeRestartOptions{}, clierr.Usagef("one of --env or --env-type is required")
	case len(instances) > 0 && len(envNames) != 1:
		return client.DevModeRestartOptions{}, clierr.Usagef("--instance-ids can only be used with a single --env")
	}

	switch {
	case len(instances) > 0:
		return client.DevModeRestartOptions{
			Instances: &client.DevModeRestartInstancesTarget{EnvName: envNames[0], InstanceIDs: instances},
		}, nil
	case len(envNames) > 0:
		return client.DevModeRestartOptions{Envs: &client.DevModeRestartEnvsTarget{Names: envNames}}, nil
	default:
		return client.DevModeRestartOptions{EnvTypes: &client.DevModeRestartEnvTypesTarget{Types: types}}, nil
	}
}

// Restart 重启开发模式容器内的业务进程（pod 不重建），逐环境输出结果；任一实例失败时返回错误。
func Restart(ctx context.Context, cli client.Client, appID string, opts client.DevModeRestartOptions) error {
	console.Info("Restarting dev mode processes, this may take a while...")
	data, err := cli.RestartDevModeProcesses(ctx, appID, opts)
	if err != nil {
		return errors.Wrapf(err, "restart dev mode processes for app %s", appID)
	}

	for _, env := range data.Envs {
		console.Info("\nEnvironment: %s (%s)", env.EnvName, env.EnvType)
		for _, inst := range env.Instances {
			if inst.Status == restartStatusFailed {
				console.Error("  [FAILED]  %s: %s", inst.Instance, inst.Message)
				continue
			}
			console.Info(
				"  [SUCCESS] %s: pid %s -> %s",
				inst.Instance,
				lo.CoalesceOrEmpty(inst.OldPID, "none"),
				inst.NewPID,
			)
		}
	}

	if len(data.Skipped) > 0 {
		console.Warn("\nSkipped environments:")
		for _, env := range data.Skipped {
			console.Warn("  - %s (%s): %s", env.EnvName, env.EnvType, env.Reason)
		}
	}

	instances := lo.FlatMap(
		data.Envs,
		func(env client.DevModeRestartEnvResult, _ int) []client.DevModeRestartInstanceResult {
			return env.Instances
		},
	)
	total := len(instances)
	failed := lo.CountBy(instances, func(inst client.DevModeRestartInstanceResult) bool {
		return inst.Status == restartStatusFailed
	})
	console.Info("\n==================================================")
	switch {
	case total == 0:
		console.Warn("No instances were restarted.")
	case failed == 0:
		console.Info("All %d instance(s) restarted successfully!", total)
	default:
		console.Error("Restart finished with %d/%d failed instance(s).", failed, total)
	}
	console.Info("==================================================")

	if failed > 0 {
		return errors.Errorf("restart failed on %d instance(s)", failed)
	}
	return nil
}
