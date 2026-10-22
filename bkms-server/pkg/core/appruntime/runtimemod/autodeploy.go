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

// AutoDeployProps 无参数。
type AutoDeployProps struct{}

// Public 无对外内容。
func (p AutoDeployProps) Public() any { return nil }

// AutoDeployDef 一键构建部署：构建成功后自动部署到指定环境，与技术栈无关，所有 bkmsApp 应用都支持。
// 现有的 trpc、taf 两个接口走同一套实现，只在路由层校验应用类型，M6 合并为一个接口。
var AutoDeployDef = Def[AutoDeployProps]{
	ID:        "autoDeploy",
	Dimension: StackIndependent,
	Universal: &AutoDeployProps{},
}
