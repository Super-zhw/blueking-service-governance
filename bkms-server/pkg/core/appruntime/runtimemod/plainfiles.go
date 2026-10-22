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

// PlainConfigFilesProps 无参数。
type PlainConfigFilesProps struct{}

// Public 无对外内容。
func (p PlainConfigFilesProps) Public() any { return nil }

// PlainConfigFilesDef plain 配置文件：用户自定义、挂载到容器的配置文件。
// 不经过框架插件，所有 bkmsApp 应用都支持。
var PlainConfigFilesDef = Def[PlainConfigFilesProps]{
	ID:        "plainConfigFiles",
	Dimension: StackIndependent,
	Universal: &PlainConfigFilesProps{},
}
