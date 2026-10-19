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

package serializer

import (
	"github.com/samber/lo"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/common/bkerrs"
	devmode "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/extension/component/devmode"
	_ "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/server/ginutils/validators" // register global validators
)

// RestartURIInput DevMode 进程重启 URI 路径参数
type RestartURIInput struct {
	// AppID 应用 ID
	AppID string `uri:"appID" binding:"required"`
}

// RestartInput DevMode 进程重启请求体，envs / envTypes / instances 必须且只能指定其中一个
type RestartInput struct {
	// Envs 按环境名称重启（各环境下所有 Running 实例）
	Envs *RestartEnvsTarget `json:"envs"`
	// EnvTypes 按环境类型重启（匹配环境下所有 Running 实例）
	EnvTypes *RestartEnvTypesTarget `json:"envTypes"`
	// Instances 重启指定环境下的指定实例
	Instances *RestartInstancesTarget `json:"instances"`
}

// RestartEnvsTarget 按环境名称重启的目标
type RestartEnvsTarget struct {
	// Names 环境名称列表
	Names []string `json:"names" binding:"required,min=1,unique,dive,uri_slug"`
}

// RestartEnvTypesTarget 按环境类型重启的目标
type RestartEnvTypesTarget struct {
	// Types 环境类型列表，不允许 production
	Types []string `json:"types" binding:"required,min=1,unique,dive,oneof=development test staging"`
}

// RestartInstancesTarget 指定实例的重启目标
type RestartInstancesTarget struct {
	// EnvName 环境名称
	EnvName string `json:"envName" binding:"required,uri_slug"`
	// InstanceIDs 实例 ID 列表，须属于 EnvName 环境
	InstanceIDs []string `json:"instanceIDs" binding:"required,min=1,unique,dive,required"`
}

// Validate 校验 envs / envTypes / instances 必须且只能指定其中一个，必须在 Bind 之后单独调用
func (i *RestartInput) Validate() error {
	specified := lo.CountBy([]bool{i.Envs != nil, i.EnvTypes != nil, i.Instances != nil}, func(ok bool) bool {
		return ok
	})
	if specified != 1 {
		return bkerrs.New(bkerrs.ErrCodeInvalidArgument, "exactly one of envs, envTypes or instances is required")
	}
	return nil
}

// RestartOutput DevMode 进程重启成功响应
type RestartOutput struct {
	Data *RestartData `json:"data"`
}

// RestartData DevMode 进程重启结果
type RestartData struct {
	// Envs 按环境分组的实例重启结果
	Envs []*RestartEnvResult `json:"envs"`
	// Skipped 被跳过的环境（仅 targetType=envTypes 时可能出现）
	Skipped []*RestartSkippedEnv `json:"skipped"`
}

// RestartEnvResult 单个环境的重启结果
type RestartEnvResult struct {
	// EnvName 环境名称
	EnvName string `json:"envName"`
	// EnvType 环境类型
	EnvType string `json:"envType"`
	// Instances 实例重启结果
	Instances []*RestartInstanceResult `json:"instances"`
}

// RestartInstanceResult 单个实例的重启结果
type RestartInstanceResult struct {
	// Instance 实例（pod）名称
	Instance string `json:"instance"`
	// Status 重启结果：success / failed
	Status string `json:"status"`
	// Message 结果说明
	Message string `json:"message"`
	// OldPID 重启前的业务进程 PID
	OldPID string `json:"oldPID"`
	// NewPID 重启后的业务进程 PID
	NewPID string `json:"newPID"`
}

// FromModel 将领域模型转换为输出对象
func (o *RestartInstanceResult) FromModel(result devmode.RestartInstanceResult) *RestartInstanceResult {
	o.Instance = result.Instance
	o.Status = string(result.Status)
	o.Message = result.Message
	o.OldPID = result.OldPID
	o.NewPID = result.NewPID
	return o
}

// RestartSkippedEnv 被跳过的环境
type RestartSkippedEnv struct {
	// EnvName 环境名称
	EnvName string `json:"envName"`
	// EnvType 环境类型
	EnvType string `json:"envType"`
	// Reason 跳过原因
	Reason string `json:"reason"`
}
