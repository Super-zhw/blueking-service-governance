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

// DevModeScriptSet 选择注入容器的管理脚本集。
type DevModeScriptSet string

const (
	// DevModeScriptSetTRPC 注入 tRPC 脚本集。
	DevModeScriptSetTRPC DevModeScriptSet = "trpc"
	// DevModeScriptSetTAF 注入 TAF 脚本集。
	DevModeScriptSetTAF DevModeScriptSet = "taf"
)

// DevModePathScriptsSubDir 存放管理脚本的子目录。
const DevModePathScriptsSubDir = "configmap-scripts"

// 开发模式工作路径。这些是平台策略而非用户配置：AppSpec 中的 workPath/mountPath
// 仅为兼容现有 API 而保留。
const (
	// DevModeTrpcWorkPath tRPC 开发模式根目录。
	DevModeTrpcWorkPath = "/data/bkms/dev-mode/trpc"
	// DevModeTrpcMountPath tRPC 脚本挂载路径。
	DevModeTrpcMountPath = DevModeTrpcWorkPath + "/" + DevModePathScriptsSubDir
	// DevModeTrpcBinaryPath tRPC 脚本操作的二进制目录。
	DevModeTrpcBinaryPath = "/usr/local/trpc/bin"
	// DevModeTafWorkPath TAF 开发模式根目录。
	DevModeTafWorkPath = "/data/bkms/dev-mode/taf"
	// DevModeTafMountPath TAF 脚本挂载路径。
	DevModeTafMountPath = DevModeTafWorkPath + "/" + DevModePathScriptsSubDir
)

// DevModeProps 开发模式脚本的平台策略。
type DevModeProps struct {
	// ScriptSet 选择脚本模板。
	ScriptSet DevModeScriptSet
	// WorkPath 容器内开发模式根目录。
	WorkPath string
	// MountPath 脚本 ConfigMap 的挂载路径。
	MountPath string
	// BinaryPath 脚本替换的二进制目录。脚本集不需要时为空，目前 TAF 脚本集即如此。
	BinaryPath string
}

// Public 对外暴露路径，因为 AppSpec API 已经在返回这两个字段。
func (p DevModeProps) Public() any {
	return struct {
		WorkPath  string `json:"workPath"`
		MountPath string `json:"mountPath"`
	}{p.WorkPath, p.MountPath}
}

// DevModeDef 开发模式能力声明。是否启用、所在环境是否允许，属于运行时检查，仍由 devmode 模块负责。
var DevModeDef = Def[DevModeProps]{
	ID:        "devMode",
	Dimension: ByFramework,
	Framework: FrameworkTable[DevModeProps]{
		appruntime.FrameworkTRPC: {
			ScriptSet:  DevModeScriptSetTRPC,
			WorkPath:   DevModeTrpcWorkPath,
			MountPath:  DevModeTrpcMountPath,
			BinaryPath: DevModeTrpcBinaryPath,
		},
		appruntime.FrameworkTAF: {
			ScriptSet: DevModeScriptSetTAF,
			WorkPath:  DevModeTafWorkPath,
			MountPath: DevModeTafMountPath,
		},
	},
}
