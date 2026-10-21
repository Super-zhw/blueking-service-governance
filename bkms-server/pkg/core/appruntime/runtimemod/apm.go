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

// APMConfigParser 选择从框架配置文件中读取 APM 服务名的解析器。
type APMConfigParser string

const (
	// APMConfigParserTrpcYAML 读取 tRPC YAML 的 telemetry 配置。
	APMConfigParserTrpcYAML APMConfigParser = "trpcYAML"
	// APMConfigParserTafXML 读取 TAF XML 配置。
	APMConfigParserTafXML APMConfigParser = "tafXML"
)

// APMServiceNameProps 选择配置解析器。
type APMServiceNameProps struct {
	Parser APMConfigParser
}

// Public 不对外暴露任何字段：解析器是执行器 key。
func (p APMServiceNameProps) Public() any { return nil }

// APMServiceNameDef APM 服务名提取能力声明。这里只是解析器能力，监控功能整体仍对所有应用类型开放。
var APMServiceNameDef = Def[APMServiceNameProps]{
	ID:        "apmServiceName",
	Dimension: ByFramework,
	Framework: FrameworkTable[APMServiceNameProps]{
		appruntime.FrameworkTRPC: {Parser: APMConfigParserTrpcYAML},
		appruntime.FrameworkTAF:  {Parser: APMConfigParserTafXML},
	},
}
