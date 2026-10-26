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

package bcs

import (
	"net/url"
	"strings"
)

// WeTermParams 生成 WeTerm 登录容器链接所需的参数
type WeTermParams struct {
	// EnvID BCS 侧分配的环境 ID（bkBcsEnvID）
	EnvID string
	// ProjectID BCS 项目 ID
	ProjectID string
	// ProjectCode BCS 项目 Code
	ProjectCode string
	// ClusterID 集群 ID
	ClusterID string
	// Namespace 命名空间
	Namespace string
	// PodName Pod 名称
	PodName string
	// ContainerName 容器名称
	ContainerName string
}

// BuildWeTermURL 生成 WeTerm 登录容器的链接，格式由 BCS 侧约定：
//
//	weterm://session/open/bcs?ns={namespace}&pod={podName}&container={containerName}&type=login
//	&clusterId={clusterId}&envId={bkBcsEnvID}&projectID={projectID}&projectCode={projectCode}
//
// 不使用 url.Values.Encode，因为它会按 key 排序，打乱 BCS 约定的参数顺序。
func BuildWeTermURL(p WeTermParams) string {
	query := []struct{ key, value string }{
		{"ns", p.Namespace},
		{"pod", p.PodName},
		{"container", p.ContainerName},
		{"type", "login"},
		{"clusterId", p.ClusterID},
		{"envId", p.EnvID},
		{"projectID", p.ProjectID},
		{"projectCode", p.ProjectCode},
	}
	pairs := make([]string, 0, len(query))
	for _, q := range query {
		pairs = append(pairs, q.key+"="+url.QueryEscape(q.value))
	}
	return "weterm://session/open/bcs?" + strings.Join(pairs, "&")
}
