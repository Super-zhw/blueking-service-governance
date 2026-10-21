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

// BSCPWorkloadMetaProps 描述 workload 在 BSCP 元信息中的登记方式。
type BSCPWorkloadMetaProps struct {
	// WorkloadKind 登记到 BSCP 的 k8s kind。
	WorkloadKind string
}

// Public 不对外暴露任何字段：kind 属于集成细节。
func (p BSCPWorkloadMetaProps) Public() any { return nil }

// BSCPWorkloadMetaDef BSCP workload 元信息能力声明。TAF 目前没有接入 BSCP，
// 若声明为支持，会出现没有实现的集成。
var BSCPWorkloadMetaDef = Def[BSCPWorkloadMetaProps]{
	ID:        "bscpWorkloadMeta",
	Dimension: ByFramework,
	Framework: FrameworkTable[BSCPWorkloadMetaProps]{
		appruntime.FrameworkTRPC: {WorkloadKind: "GameDeployment"},
	},
}
