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

// Package devmode 声明开发模式能力：哪些框架支持开发模式，以及脚本集与容器内路径。
package devmode

import (
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime/runtimemod"
)

// ScriptSet 开发模式注入的脚本集。
type ScriptSet string

const (
	// ScriptSetTRPC tRPC 脚本集。
	ScriptSetTRPC ScriptSet = "trpc"
	// ScriptSetTAF TAF 脚本集。
	ScriptSetTAF ScriptSet = "taf"
)

// ScriptsSubDir 脚本所在的子目录。
const ScriptsSubDir = "configmap-scripts"

// 开发模式的容器内路径，由平台决定，用户不能修改。
const (
	// TrpcWorkPath tRPC 开发模式根目录。
	TrpcWorkPath = "/data/bkms/dev-mode/trpc"
	// TrpcMountPath tRPC 脚本挂载目录。
	TrpcMountPath = TrpcWorkPath + "/" + ScriptsSubDir
	// TrpcBinaryPath tRPC 二进制所在目录。
	TrpcBinaryPath = "/usr/local/trpc/bin"

	// TafWorkPath TAF 开发模式根目录。
	TafWorkPath = "/data/bkms/dev-mode/taf"
	// TafMountPath TAF 脚本挂载目录。
	TafMountPath = TafWorkPath + "/" + ScriptsSubDir
)

// Props 开发模式的执行参数。
type Props struct {
	// ScriptSet 脚本集。
	ScriptSet ScriptSet
	// WorkPath 开发模式根目录。
	WorkPath string
	// MountPath 脚本挂载目录。
	MountPath string
	// BinaryPath 二进制所在目录，TAF 不需要，为空。
	BinaryPath string
}

// Public 对外展示根目录与挂载目录，与 AppSpec 接口返回的一致。
func (p Props) Public() any {
	return struct {
		WorkPath  string `json:"workPath"`
		MountPath string `json:"mountPath"`
	}{p.WorkPath, p.MountPath}
}

// Spec 开发模式：tRPC、TAF 支持。是否开启、当前环境是否允许，由 devmode 模块在运行时判断。
var Spec = runtimemod.Spec[Props]{
	ID: runtimemod.ModDevMode,
	Resolve: runtimemod.ByFramework(map[string]Props{
		appruntime.FrameworkTRPC: {
			ScriptSet:  ScriptSetTRPC,
			WorkPath:   TrpcWorkPath,
			MountPath:  TrpcMountPath,
			BinaryPath: TrpcBinaryPath,
		},

		appruntime.FrameworkTAF: {
			ScriptSet: ScriptSetTAF,
			WorkPath:  TafWorkPath,
			MountPath: TafMountPath,
		},
	}),
}
