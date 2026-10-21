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

// PolarisFrameworkPatchProps 无参数：补丁内容是 plugins.registry.polaris.service，
// 取自当前环境的北极星配置，与技术栈无关。
type PolarisFrameworkPatchProps struct{}

// Public 不对外暴露任何字段。
func (p PolarisFrameworkPatchProps) Public() any { return nil }

// PolarisFrameworkPatchDef 北极星注册信息注入能力声明，即在渲染框架配置时把
// plugins.registry.polaris.service 写入配置。
//
// 补丁器挂在 tRPC 插件的 configPatchers 链上，按 YAML 路径注入，其余框架没有对应实现。
// 该能力不等于北极星注册本身：注册本身对所有应用类型可用。
var PolarisFrameworkPatchDef = Def[PolarisFrameworkPatchProps]{
	ID:        "polarisFrameworkPatch",
	Dimension: ByFramework,
	Framework: FrameworkTable[PolarisFrameworkPatchProps]{
		appruntime.FrameworkTRPC: {},
	},
}
