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

// PlainConfigFilesProps 无参数：plain 文件由 workload 构建的公共步骤挂载，不随框架变化。
type PlainConfigFilesProps struct{}

// Public 不对外暴露任何字段。
func (p PlainConfigFilesProps) Public() any { return nil }

// PlainConfigFilesDef plain 配置文件能力声明，即用户自定义、挂载到 workload 的文本文件。
// 它的渲染不经过任何框架插件，因此是 blank 的主要配置途径，tRPC 与 TAF 同样可用。
var PlainConfigFilesDef = Def[PlainConfigFilesProps]{
	ID:        "plainConfigFiles",
	Dimension: StackIndependent,
	Universal: &PlainConfigFilesProps{},
}
