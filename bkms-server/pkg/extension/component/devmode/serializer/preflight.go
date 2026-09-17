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

package serializer

// PreflightURIInput DevMode Publish Preflight URI 路径参数
type PreflightURIInput struct {
	// AppID 应用 ID
	AppID string `uri:"appID" binding:"required"`
	// EnvName 环境名称
	EnvName string `uri:"envName" binding:"required"`
}

// PreflightBodyInput DevMode Publish Preflight 请求体参数
type PreflightBodyInput struct {
	// InstanceIDs 指定的实例 ID 列表（与 PublishAll 二选一）
	InstanceIDs []string `json:"instanceIDs"`
	// PublishAll 是否发布到所有 Running 状态的实例（与 InstanceIDs 二选一）
	PublishAll bool `json:"publishAll"`
}

// PreflightOutput DevMode Publish Preflight 成功响应
type PreflightOutput struct {
	Data *PreflightData `json:"data"`
}

// PreflightData Preflight 响应数据
type PreflightData struct {
	// Token 用户 Token，用于访问集群 API
	Token string `json:"token"`
	// Address 已组装好的集群完整地址；如 {baseUrl}/clusters/{clusterID}/
	Address string `json:"address"`
	// Namespace 目标命名空间
	Namespace string `json:"namespace"`
	// InstanceIDs 校验通过的实例 ID 列表
	InstanceIDs []string `json:"instanceIDs"`
	// DevMode 开发模式相关路径配置
	DevMode *PreflightDevMode `json:"devMode"`
}

// PreflightDevMode 开发模式路径信息
type PreflightDevMode struct {
	// WorkPath 开发模式根目录
	WorkPath string `json:"workPath"`
	// MountPath 脚本挂载路径
	MountPath string `json:"mountPath"`
}
