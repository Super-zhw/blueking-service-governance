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

package polaris

import (
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime/runtimemod"
)

// FrameworkPatchProps 无参数。
type FrameworkPatchProps struct{}

// Public 无对外内容。
func (p FrameworkPatchProps) Public() any { return nil }

// FrameworkPatch 北极星注册信息注入：渲染框架配置文件时，把北极星注册信息写入
// plugins.registry.polaris.service，已有该配置时不覆盖。仅 tRPC 支持。
var FrameworkPatch = runtimemod.Spec[FrameworkPatchProps]{
	ID: runtimemod.ModPolarisFrameworkPatch,
	Resolve: runtimemod.ByFramework(map[string]FrameworkPatchProps{
		appruntime.FrameworkTRPC: {},
	}),
}
