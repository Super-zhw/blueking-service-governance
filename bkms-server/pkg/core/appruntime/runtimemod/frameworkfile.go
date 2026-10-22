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

// FrameworkFileFormat 框架配置文件的格式，取值与配置文件存储的 appcfg.FileFormat 一致。
type FrameworkFileFormat string

const (
	// FrameworkFileYAML YAML 格式，tRPC 使用。
	FrameworkFileYAML FrameworkFileFormat = "yaml"
	// FrameworkFileTAF TAF conf 格式（标签加 key=value，不是标准 XML），TAF 使用。
	FrameworkFileTAF FrameworkFileFormat = "taf"
)

// FrameworkFileProps 框架配置文件的参数。
type FrameworkFileProps struct {
	Format FrameworkFileFormat
}

// Public 对外展示文件格式，前端据此选择编辑器的语法（taf 按 xml 高亮）。
func (p FrameworkFileProps) Public() any {
	return struct {
		Format FrameworkFileFormat `json:"format"`
	}{p.Format}
}

// FrameworkFileDef 框架配置文件：tRPC、TAF 支持；blank 没有框架配置文件。
var FrameworkFileDef = Def[FrameworkFileProps]{
	ID:        "frameworkFile",
	Dimension: ByFramework,
	Framework: FrameworkTable[FrameworkFileProps]{
		appruntime.FrameworkTRPC: {Format: FrameworkFileYAML},
		appruntime.FrameworkTAF:  {Format: FrameworkFileTAF},
	},
}
