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

// Env 环境
type Env struct {
	ID          string          `json:"id" yaml:"id"`
	Name        string          `json:"name" yaml:"name"`
	DisplayName string          `json:"displayName" yaml:"displayName"`
	Type        string          `json:"type" yaml:"type"`
	Description string          `json:"description" yaml:"description"`
	UpdatedAt   string          `json:"updatedAt" yaml:"updatedAt"`
	Cluster     *EnvClusterInfo `json:"cluster" yaml:"cluster"`
}

// EnvClusterInfo 环境运行时配置（集群信息）
type EnvClusterInfo struct {
	ClusterID   string `json:"clusterID" yaml:"clusterID"`
	ClusterType string `json:"clusterType" yaml:"clusterType"`
	Namespace   string `json:"namespace" yaml:"namespace"`
	ProjectCode string `json:"projectCode" yaml:"projectCode"`
}

// ListEnvsRespData 获取环境列表返回数据
type ListEnvsRespData struct {
	Data []Env `json:"data"`
}

// GetEnvRespData 获取环境详情返回数据
type GetEnvRespData struct {
	Data Env `json:"data"`
}

// CreateEnvBody 创建环境请求体
type CreateEnvBody struct {
	Name        string      `json:"name" yaml:"name"`
	DisplayName string      `json:"displayName" yaml:"displayName"`
	Type        string      `json:"type" yaml:"type"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	Cluster     *EnvCluster `json:"cluster" yaml:"cluster"`
}

// EnvCluster 创建环境时的集群配置
type EnvCluster struct {
	ClusterID   string `json:"clusterID" yaml:"clusterID"`
	ClusterType string `json:"clusterType" yaml:"clusterType"`
	Namespace   string `json:"namespace" yaml:"namespace"`
}

// CreateEnvRespData 创建环境返回数据
// server 只返回 {"data": {"id": "..."}}
type CreateEnvRespData struct {
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

// UpdateEnvBasicInfoBody 更新环境基本信息请求体
type UpdateEnvBasicInfoBody struct {
	DisplayName string `json:"displayName,omitempty" yaml:"displayName,omitempty"`
	// Type 环境类型，可选值：development | test | staging | production
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}
