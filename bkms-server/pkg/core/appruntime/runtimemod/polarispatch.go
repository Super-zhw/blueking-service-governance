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

package runtimemod

import "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime"

// PolarisFrameworkPatchProps 无参数。注入的内容来自应用在当前环境的北极星配置，
// 不随技术栈变化。
type PolarisFrameworkPatchProps struct{}

// Public 无对外字段。
func (p PolarisFrameworkPatchProps) Public() any { return nil }

// PolarisFrameworkPatchDef 北极星注册信息注入能力。
//
// 渲染框架配置文件时，若应用在该环境启用了北极星健康检查，就把注册信息写入配置的
// plugins.registry.polaris.service 路径（原配置已有该路径时不覆盖）。
//
// 目前只有 tRPC 插件挂载了对应的补丁器（trpc/patcher.PolarisRegistryPatcher），
// 其他 framework 不会注入。本能力只指"写入框架配置"这一步，北极星注册本身对所有应用可用。
var PolarisFrameworkPatchDef = Def[PolarisFrameworkPatchProps]{
	ID:        "polarisFrameworkPatch",
	Dimension: ByFramework,
	Framework: FrameworkTable[PolarisFrameworkPatchProps]{
		appruntime.FrameworkTRPC: {},
	},
}
