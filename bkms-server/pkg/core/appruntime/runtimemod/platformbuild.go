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

// PlatformBuildProps 平台生成 Dockerfile 构建所需的参数。
type PlatformBuildProps struct {
	// DockerfileLanguage 作为 Dockerfile 生成器语言传给流水线。
	DockerfileLanguage string
}

// Public 不对外暴露任何字段：客户端只需要知道该构建方式是否可用，语言是流水线内部参数。
func (p PlatformBuildProps) Public() any { return nil }

// PlatformBuildDef 平台生成 Dockerfile 构建能力声明。目前仅限 tRPC 且仅限 Go，
// 两个限制都在这里表达，不再写在构建服务里。
var PlatformBuildDef = Def[PlatformBuildProps]{
	ID:        "platformBuild",
	Dimension: ByFramework,
	Framework: FrameworkTable[PlatformBuildProps]{
		appruntime.FrameworkTRPC: {DockerfileLanguage: appruntime.LanguageGo},
	},
	Languages: Only(appruntime.LanguageGo),
}
