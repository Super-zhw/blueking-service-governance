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

// AdminProtocol 选择 admin 命令的实现。
type AdminProtocol string

const (
	// AdminProtocolTRPC tRPC HTTP admin 服务。
	AdminProtocolTRPC AdminProtocol = "trpc"
	// AdminProtocolTAF TAF 私有协议。
	AdminProtocolTAF AdminProtocol = "taf"
)

// TrpcAdminConfigLayout 描述 admin IP 与端口在 tRPC 配置文件中的位置。
// 不同语言的 tRPC 配置格式不同，布局按语言选择。
type TrpcAdminConfigLayout string

const (
	// TrpcAdminLayoutNested 读取 server.admin.ip / server.admin.port（go、python）。
	TrpcAdminLayoutNested TrpcAdminConfigLayout = "nested"
	// TrpcAdminLayoutFlat 读取 server.admin_ip / server.admin_port（cpp、node）。
	TrpcAdminLayoutFlat TrpcAdminConfigLayout = "flat"
	// TrpcAdminLayoutNestedPrefixed 读取 server.admin.admin_ip / server.admin.admin_port（java）。
	TrpcAdminLayoutNestedPrefixed TrpcAdminConfigLayout = "nestedPrefixed"
)

// AdminCommandProps 选择协议；tRPC 下还选择配置键布局。
type AdminCommandProps struct {
	Protocol AdminProtocol
	// ConfigLayout 仅在 tRPC 协议下有值。
	ConfigLayout TrpcAdminConfigLayout
}

// Public 不对外暴露任何字段：布局是从配置文件推导的执行器 key。
func (p AdminCommandProps) Public() any { return nil }

// AdminCommandDef admin 命令能力声明。
//
// tRPC 的配置布局随语言变化，TAF 只有一种私有协议，因此语言差异按组合表达，
// 而不是对整个能力做语言限制。Pair 只登记框架目录中的语言；目录新增语言时在此补对应布局。
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
