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

// Package polaris 声明与框架相关的北极星能力：服务名校验与注册信息注入。北极星注册本身不在此声明。
package polaris

import (
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime/runtimemod"
)

// ServiceCheckProps 无参数。
type ServiceCheckProps struct{}

// Public 无对外内容。
func (p ServiceCheckProps) Public() any { return nil }

// ServiceCheck 北极星服务名校验：北极星配置的服务名不在 tRPC 配置文件中时给出警告。仅 tRPC 支持。
// 校验由 extension/addon/polaris 执行。
var ServiceCheck = runtimemod.Spec[ServiceCheckProps]{
	ID: runtimemod.ModPolarisServiceCheck,
	Resolve: runtimemod.ByFramework(map[string]ServiceCheckProps{
		appruntime.FrameworkTRPC: {},
	}),
}
