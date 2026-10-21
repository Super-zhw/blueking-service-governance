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

// PolarisServiceCheckProps 无参数：该能力是将北极星服务名与 tRPC 配置文件比对的配置校验。
type PolarisServiceCheckProps struct{}

// Public 不对外暴露任何字段。
func (p PolarisServiceCheckProps) Public() any { return nil }

// PolarisServiceCheckDef 北极星服务名校验能力声明。不限制北极星注册本身，只限制基于 tRPC 配置的校验。
var PolarisServiceCheckDef = Def[PolarisServiceCheckProps]{
	ID:        "polarisServiceCheck",
	Dimension: ByFramework,
	Framework: FrameworkTable[PolarisServiceCheckProps]{
		appruntime.FrameworkTRPC: {},
	},
}
