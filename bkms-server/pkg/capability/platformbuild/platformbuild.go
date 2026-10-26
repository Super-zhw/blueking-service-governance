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

// Package platformbuild 声明平台构建能力：哪些技术栈支持由平台生成 Dockerfile 构建镜像。
package platformbuild

import (
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime/runtimemod"
)

// Props 平台构建的执行参数。
type Props struct {
	// DockerfileLanguage 传给流水线的 Dockerfile 生成语言。
	DockerfileLanguage string
}

// Public 无对外内容。
func (p Props) Public() any { return nil }

// Spec 平台构建：由平台生成 Dockerfile 并构建镜像。仅 tRPC + go 支持。
var Spec = runtimemod.Spec[Props]{
	ID: runtimemod.ModPlatformBuild,
	Resolve: func(framework, language string) runtimemod.Mod[Props] {
		if framework != appruntime.FrameworkTRPC {
			return runtimemod.Unsupported[Props]()
		}
		switch language {
		case appruntime.LanguageGo:
			return runtimemod.Supported(Props{DockerfileLanguage: appruntime.LanguageGo})
		case "":
			return runtimemod.Incomplete[Props]()
		default:
			return runtimemod.Unsupported[Props]()
		}
	},
}
