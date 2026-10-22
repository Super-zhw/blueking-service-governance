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

// AdminProtocol admin 命令的协议。
type AdminProtocol string

const (
	// AdminProtocolTRPC tRPC admin，基于 HTTP。
	AdminProtocolTRPC AdminProtocol = "trpc"
	// AdminProtocolTAF TAF 私有协议。
	AdminProtocolTAF AdminProtocol = "taf"
)

// TrpcAdminConfigLayout tRPC 配置文件中 admin 地址的写法，不同语言写法不同。
type TrpcAdminConfigLayout string

const (
	// TrpcAdminLayoutNested server.admin.ip / server.admin.port，go、python 使用。
	TrpcAdminLayoutNested TrpcAdminConfigLayout = "nested"
	// TrpcAdminLayoutFlat server.admin_ip / server.admin_port，cpp、node 使用。
	TrpcAdminLayoutFlat TrpcAdminConfigLayout = "flat"
	// TrpcAdminLayoutNestedPrefixed server.admin.admin_ip / server.admin.admin_port，java 使用。
	TrpcAdminLayoutNestedPrefixed TrpcAdminConfigLayout = "nestedPrefixed"
)

// AdminCommandProps admin 命令的执行参数。
type AdminCommandProps struct {
	Protocol AdminProtocol
	// ConfigLayout admin 地址的写法，仅 tRPC 使用。
	ConfigLayout TrpcAdminConfigLayout
}

// Public 无对外内容。
func (p AdminCommandProps) Public() any { return nil }

// AdminCommandDef admin 命令：tRPC、TAF 支持。
//
// tRPC 按语言区分 admin 地址的写法。java、node、python 不在框架目录中，对应条目暂不生效，
// 纳入目录后自动生效。
var AdminCommandDef = Def[AdminCommandProps]{
	ID:        "adminCommand",
	Dimension: ByFramework,
	Framework: FrameworkTable[AdminCommandProps]{
		appruntime.FrameworkTAF: {Protocol: AdminProtocolTAF},
	},
	Pair: PairTable[AdminCommandProps]{
		{Framework: appruntime.FrameworkTRPC, Language: appruntime.LanguageGo}: {
			Protocol:     AdminProtocolTRPC,
			ConfigLayout: TrpcAdminLayoutNested,
		},
		{Framework: appruntime.FrameworkTRPC, Language: appruntime.LanguageCpp}: {
			Protocol:     AdminProtocolTRPC,
			ConfigLayout: TrpcAdminLayoutFlat,
		},
		{Framework: appruntime.FrameworkTRPC, Language: appruntime.LanguagePython}: {
			Protocol:     AdminProtocolTRPC,
			ConfigLayout: TrpcAdminLayoutNested,
		},
		{Framework: appruntime.FrameworkTRPC, Language: appruntime.LanguageNode}: {
			Protocol:     AdminProtocolTRPC,
			ConfigLayout: TrpcAdminLayoutFlat,
		},
		{Framework: appruntime.FrameworkTRPC, Language: appruntime.LanguageJava}: {
			Protocol:     AdminProtocolTRPC,
			ConfigLayout: TrpcAdminLayoutNestedPrefixed,
		},
	},
}
