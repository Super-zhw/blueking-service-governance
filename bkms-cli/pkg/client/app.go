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

package client

// AppMinimal 应用简要信息
type AppMinimal struct {
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	DisplayName string `json:"displayName" yaml:"displayName"`
	Type        string `json:"type" yaml:"type"`
	Creator     string `json:"creator" yaml:"creator"`
}

// ListAppsRespData 获取应用列表返回数据
type ListAppsRespData struct {
	Data []AppMinimal `json:"data"`
}

// CreateAppRespData 创建应用返回数据
type CreateAppRespData struct {
	Data AppMinimal `json:"data"`
}

// GetAppIDAutoSuffixRespData 获取应用 ID 自动后缀返回数据
// 后端直接返回 {"suffix": "..."} 格式，无 data 包装
type GetAppIDAutoSuffixRespData struct {
	Suffix string `json:"suffix"`
}

// DevModeConfig 开发模式配置
type DevModeConfig struct {
	// Enabled 是否启用开发模式
	Enabled bool `json:"enabled" yaml:"enabled"`

	// WorkPath 开发模式根目录
	WorkPath string `json:"workPath" yaml:"workPath"`

	// MountPath 脚本挂载路径
	MountPath string `json:"mountPath" yaml:"mountPath"`
}

// GetEnvEffectiveDevModeRespData 获取环境实际生效的开发模式配置返回数据
type GetEnvEffectiveDevModeRespData struct {
	Data *DevModeConfig `json:"data"`
}

// DevModePreflightData 开发模式 Publish 预检响应数据
type DevModePreflightData struct {
	// Token 用户 Token
	Token string `json:"token"`
	// Address 已组装好的集群完整地址
	Address string `json:"address"`
	// Namespace 目标命名空间
	Namespace string `json:"namespace"`
	// InstanceIDs 校验通过的实例 ID 列表
	InstanceIDs []string `json:"instanceIDs"`
	// DevMode 开发模式路径配置
	DevMode *DevModeConfig `json:"devMode"`
}

// DevModePreflightRespData 开发模式 Publish 预检响应包装
type DevModePreflightRespData struct {
	Data *DevModePreflightData `json:"data"`
}
