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

// Package apmservicename 声明 APM 服务名提取能力：各框架读取服务名的解析方式。
package apmservicename

import (
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime/runtimemod"
)

// Parser 读取 APM 服务名的解析方式。
type Parser string

const (
	// ParserTrpcYAML 解析 tRPC 的 YAML 配置。
	ParserTrpcYAML Parser = "trpcYAML"
	// ParserTafXML 解析 TAF 的 XML 配置。
	ParserTafXML Parser = "tafXML"
)

// Props APM 服务名提取的执行参数。
type Props struct {
	// Parser 解析方式。
	Parser Parser
}

// Public 无对外内容。
func (p Props) Public() any { return nil }

// Spec APM 服务名提取：从框架配置文件中读取 APM 服务名。tRPC、TAF 支持。
var Spec = runtimemod.Spec[Props]{
	ID: runtimemod.ModAPMServiceName,
	Resolve: runtimemod.ByFramework(map[string]Props{
		appruntime.FrameworkTRPC: {Parser: ParserTrpcYAML},

		appruntime.FrameworkTAF: {Parser: ParserTafXML},
	}),
}
