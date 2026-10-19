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

package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"k8s.io/client-go/rest"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/common/bkerrs"
	bkmsapp "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/app"
	envmodel "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/env/model"
	devmode "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/extension/component/devmode"
	slz "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/extension/component/devmode/serializer"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/infras/account/auth"
	k8sclient "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/infras/kubernetes/client"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/infras/kubernetes/cluster"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/misc/audit"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/server/ginutils"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/server/ginutils/perm"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/store"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/workload/appmodelcore/appspec"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/workload/appmodelcore/workload/defaults"
)

// restartEnvTarget 单个环境解析后的重启目标
type restartEnvTarget struct {
	env         *envmodel.Environment
	devModeSpec *appspec.DevModeSpec
	instanceIDs []string
}

// RestartDevModeProcesses 开发模式进程重启接口。
// 同步完成：stop 业务进程并清除停止标志，由 monitor 重新拉起，pod 不重建。
//
//	@ID			RestartDevModeProcesses
//	@Summary	开发模式进程重启
//	@Tags		devmode
//	@Accept		json
//	@Produce	json
//	@Security	BkUserInfo
//	@Security	BkUserCredential
//	@Param		appID	path		string				true	"应用 ID"
//	@Param		body	body		slz.RestartInput	true	"进程重启请求体"
//	@Success	200		{object}	slz.RestartOutput
//	@Failure	400		{object}	bkerrs.GinErrorOutput
//	@Router		/devmode/{appID}/restart [post]
func (h *Handler) RestartDevModeProcesses(c *gin.Context) {
	var uriInput slz.RestartURIInput
	var input slz.RestartInput
	if err := ginutils.BindURI(c, &uriInput); err != nil {
		bkerrs.AbortWithErr(c, err)
		return
	}
	if err := ginutils.BindJSON(c, &input); err != nil {
		bkerrs.AbortWithErr(c, err)
		return
	}
	if err := input.Validate(); err != nil {
		bkerrs.AbortWithErr(c, err)
		return
	}
	ctx := c.Request.Context()

	if err := checkClusterAPIConfigured(); err != nil {
		bkerrs.AbortWithErr(c, err)
		return
	}
	app, err := perm.ValidateAppByID(ctx, h.registry, uriInput.AppID, perm.TypeDeploy)
	if err != nil {
		bkerrs.AbortWithErr(c, err)
		return
	}

	// 解析重启目标
	envTargets, skipped, err := h.resolveRestartTargets(ctx, app, input)
	if err != nil {
		bkerrs.AbortWithErr(c, err)
		return
	}
	if len(envTargets) == 0 {
		ginutils.OK(c, buildRestartOutput(nil, nil, skipped))
		return
	}

	// 获取用户 Token，用于访问各环境所在集群
	bcsProjectCode, err := h.getBCSProjectCode(ctx, app.WorkspaceID)
	if err != nil {
		bkerrs.AbortWithErr(c, err)
		return
	}
	token, err := h.fetchActiveToken(ctx, bcsProjectCode)
	if err != nil {
		bkerrs.AbortWithErr(c, err)
		return
	}

	results, err := devmode.NewRestarter(store.Of(h.registry.DB, devmode.PublishRecordMeta)).
		Restart(ctx, app.ID, auth.MustGetUser(ctx).ID, buildRestartTargets(envTargets, token))
	if err != nil {
		bkerrs.AbortWithErr(c, bkerrs.Wrap(err, bkerrs.ErrCodeInternalServerError, "restart devmode processes"))
		return
	}

	addRestartAuditRecords(ctx, app, results)
	ginutils.OK(c, buildRestartOutput(envTargets, results, skipped))
}

// resolveRestartTargets 按目标类型解析重启目标。
// envs / instances 为严格模式，任一环境不满足条件即返回错误；
// envTypes 模式下不满足条件的环境被跳过，并返回跳过原因。
func (h *Handler) resolveRestartTargets(
	ctx context.Context, app *bkmsapp.Application, input slz.RestartInput,
) ([]*restartEnvTarget, []*slz.RestartSkippedEnv, error) {
	switch {
	case input.Envs != nil:
		targets, err := lo.MapErr(input.Envs.Names, func(envName string, _ int) (*restartEnvTarget, error) {
			return h.resolveRestartTargetByEnvName(ctx, app, envName, nil)
		})
		return targets, nil, err

	case input.Instances != nil:
		target, err := h.resolveRestartTargetByEnvName(
			ctx, app, input.Instances.EnvName, input.Instances.InstanceIDs,
		)
		if err != nil {
			return nil, nil, err
		}
		return []*restartEnvTarget{target}, nil, nil

	case input.EnvTypes != nil:
		return h.resolveRestartTargetsByEnvTypes(ctx, app, input.EnvTypes.Types)

	default:
		return nil, nil, bkerrs.New(bkerrs.ErrCodeInvalidArgument, "no restart target specified")
	}
}

// resolveRestartTargetsByEnvTypes 解析匹配环境类型的所有环境，不满足条件的环境被跳过并返回跳过原因
func (h *Handler) resolveRestartTargetsByEnvTypes(
	ctx context.Context, app *bkmsapp.Application, envTypes []string,
) ([]*restartEnvTarget, []*slz.RestartSkippedEnv, error) {
	envs, err := store.Of(h.registry.DB, envmodel.EnvMeta).
		List(ctx, envmodel.InWorkspace(app.WorkspaceID), envmodel.VisibleToApp(app.ID), envmodel.NewestFirst())
	if err != nil {
		return nil, nil, bkerrs.Wrapf(err, bkerrs.ErrCodeInternalServerError, "list envs of app %s", app.ID)
	}
	matched := lo.Filter(envs, func(env envmodel.Environment, _ int) bool {
		return lo.Contains(envTypes, env.Type)
	})

	var targets []*restartEnvTarget
	var skipped []*slz.RestartSkippedEnv
	for i := range matched {
		env := &matched[i]
		target, err := h.resolveRestartTarget(ctx, app, env, nil)
		if err != nil {
			skipped = append(
				skipped,
				&slz.RestartSkippedEnv{EnvName: env.Name, EnvType: env.Type, Reason: err.Error()},
			)
			continue
		}
		targets = append(targets, target)
	}
	return targets, skipped, nil
}

// resolveRestartTargetByEnvName 按环境名称解析重启目标
func (h *Handler) resolveRestartTargetByEnvName(
	ctx context.Context, app *bkmsapp.Application, envName string, instanceIDs []string,
) (*restartEnvTarget, error) {
	env, err := store.Of(h.registry.DB, envmodel.EnvMeta).
		Get(ctx, envmodel.AppEnvByName(app.WorkspaceID, app.ID, envName))
	if err != nil {
		if errors.Is(err, envmodel.ErrEnvNotFound) {
			return nil, bkerrs.Errorf(bkerrs.ErrCodeNotFound, "environment %s not found", envName)
		}
		return nil, bkerrs.Wrapf(err, bkerrs.ErrCodeInternalServerError, "get env %s by name", envName)
	}
	return h.resolveRestartTarget(ctx, app, env, instanceIDs)
}

// resolveRestartTarget 校验环境并解析目标实例；instanceIDs 为空时取所有 Running 实例
func (h *Handler) resolveRestartTarget(
	ctx context.Context, app *bkmsapp.Application, env *envmodel.Environment, instanceIDs []string,
) (*restartEnvTarget, error) {
	if err := checkEnvSupported(env); err != nil {
		return nil, err
	}
	devModeSpec, err := h.getEffectiveDevMode(ctx, app, env.Name)
	if err != nil {
		return nil, err
	}
	if lo.FromPtr(devModeSpec.MountPath) == "" {
		return nil, bkerrs.Errorf(
			bkerrs.ErrCodeInvalidRequest, "devmode mount path is empty for app %s in environment %s", app.ID, env.Name,
		)
	}
	ids, err := h.resolveInstanceIDs(ctx, app.ID, env.Name, instanceIDs, len(instanceIDs) == 0)
	if err != nil {
		return nil, err
	}
	return &restartEnvTarget{env: env, devModeSpec: devModeSpec, instanceIDs: ids}, nil
}

// buildRestartTargets 构造 Restarter 所需的重启目标，同一集群共用一个 pod 命令执行器
func buildRestartTargets(envTargets []*restartEnvTarget, token string) []devmode.RestartTarget {
	executors := make(map[string]devmode.PodExecutor)
	return lo.Map(envTargets, func(target *restartEnvTarget, _ int) devmode.RestartTarget {
		clusterID := target.env.Cluster.ClusterID
		executor, ok := executors[clusterID]
		if !ok {
			executor = k8sclient.NewPodClient(&cluster.Config{
				Rest: &rest.Config{
					BearerToken:     token,
					Host:            buildClusterAddress(clusterID),
					TLSClientConfig: rest.TLSClientConfig{Insecure: true},
				},
				ClusterID: clusterID,
			})
			executors[clusterID] = executor
		}
		return devmode.RestartTarget{
			EnvName:     target.env.Name,
			Namespace:   target.env.Cluster.Namespace,
			Container:   defaults.WorkloadMainContainerName,
			MountPath:   lo.FromPtr(target.devModeSpec.MountPath),
			InstanceIDs: target.instanceIDs,
			Executor:    executor,
		}
	})
}

// addRestartAuditRecords 为每个实例写入操作审计记录
func addRestartAuditRecords(ctx context.Context, app *bkmsapp.Application, results []devmode.RestartInstanceResult) {
	recordCtx := context.WithoutCancel(ctx)
	for _, result := range results {
		go audit.AddOperationRecordAsync(
			recordCtx,
			audit.OperationTypeRestart,
			audit.ResourceTypeInstance,
			result.Instance,
			audit.WithResult(lo.Ternary(
				result.Status == devmode.RecordStatusSuccess, audit.ResultSuccess, audit.ResultFailed,
			)),
			audit.WithDataAfter(map[string]any{
				"oldPID":  result.OldPID,
				"newPID":  result.NewPID,
				"message": result.Message,
			}),
			audit.WithWorkspaceID(app.WorkspaceID),
			audit.WithAppID(app.ID),
			audit.WithEnvName(result.EnvName),
		)
	}
}

// buildRestartOutput 按环境分组组装重启响应
func buildRestartOutput(
	envTargets []*restartEnvTarget,
	results []devmode.RestartInstanceResult,
	skipped []*slz.RestartSkippedEnv,
) *slz.RestartOutput {
	resultsByEnv := lo.GroupBy(results, func(result devmode.RestartInstanceResult) string { return result.EnvName })
	envs := lo.Map(envTargets, func(target *restartEnvTarget, _ int) *slz.RestartEnvResult {
		return &slz.RestartEnvResult{
			EnvName: target.env.Name,
			EnvType: target.env.Type,
			Instances: lo.Map(
				resultsByEnv[target.env.Name],
				func(result devmode.RestartInstanceResult, _ int) *slz.RestartInstanceResult {
					return new(slz.RestartInstanceResult).FromModel(result)
				},
			),
		}
	})
	return &slz.RestartOutput{
		Data: &slz.RestartData{Envs: envs, Skipped: lo.Ternary(skipped == nil, []*slz.RestartSkippedEnv{}, skipped)},
	}
}
