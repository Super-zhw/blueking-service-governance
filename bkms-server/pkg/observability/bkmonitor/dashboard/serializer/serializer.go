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

// Package serializer 定义应用仪表盘绑定相关 Gin 输入输出结构。
package serializer

import (
	"github.com/samber/lo"

	bkmapi "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/infras/cloudapi/bkmonitor"
	_ "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/server/ginutils/validators" // register global validators
)

// AppURIInput 路径参数（仅应用 ID）。
type AppURIInput struct {
	AppID string `uri:"appID" binding:"required,uri_slug"`
}

// AppDashboardURIInput 应用仪表盘路径参数。
type AppDashboardURIInput struct {
	AppID string `uri:"appID" binding:"required,uri_slug"`
	UID   string `uri:"uid" binding:"required,min=1"`
}

// AppDashboardCreateInput 创建应用仪表盘绑定请求。
type AppDashboardCreateInput struct {
	// UID 仪表盘 uid
	UID string `json:"uid" binding:"required,min=1"`
	// Title 仪表盘标题
	Title string `json:"title" binding:"required,min=1"`
}

// AppDashboardUpdateInput 更新应用仪表盘绑定请求。
type AppDashboardUpdateInput struct {
	// UID 仪表盘 uid（可选，变更绑定的仪表盘）
	UID *string `json:"uid" binding:"omitempty,min=1"`
	// Title 仪表盘标题（可选）
	Title *string `json:"title" binding:"omitempty,min=1"`
}

// DashboardOutput 仪表盘输出。
type DashboardOutput struct {
	// ID 仪表盘 ID
	ID int64 `json:"id"`
	// UID 仪表盘 uid
	UID string `json:"uid"`
	// Title 仪表盘标题
	Title string `json:"title"`
	// URI 仪表盘 URI
	URI string `json:"uri"`
	// URL 仪表盘访问 URL
	URL string `json:"url"`
}

// DashboardDirectoryOutput 仪表盘目录输出。
type DashboardDirectoryOutput struct {
	// Dashboards 目录下的仪表盘列表
	Dashboards []*DashboardOutput `json:"dashboards"`
	// ID 目录 ID
	ID int64 `json:"id"`
	// UID 目录 uid
	UID string `json:"uid"`
	// Title 目录标题
	Title string `json:"title"`
	// URI 目录 URI
	URI string `json:"uri"`
	// URL 目录访问 URL
	URL string `json:"url"`
}

// FromModel 从 bkmonitor 目录树节点填充输出字段。
func (o *DashboardDirectoryOutput) FromModel(node *bkmapi.DashboardDirectoryNode) *DashboardDirectoryOutput {
	if o == nil {
		return nil
	}
	*o = DashboardDirectoryOutput{
		ID:    node.ID,
		UID:   node.UID,
		Title: node.Title,
		URI:   node.URI,
		URL:   node.URL,
		Dashboards: lo.Map(node.Dashboards, func(item bkmapi.DashboardItem, _ int) *DashboardOutput {
			return &DashboardOutput{
				ID:    item.ID,
				UID:   item.UID,
				Title: item.Title,
				URI:   item.URI,
				URL:   item.URL,
			}
		}),
	}
	return o
}

// NewDashboardDirectoryOutputs 将目录树转换为输出列表。
func NewDashboardDirectoryOutputs(tree []*bkmapi.DashboardDirectoryNode) []*DashboardDirectoryOutput {
	return lo.Map(tree, func(node *bkmapi.DashboardDirectoryNode, _ int) *DashboardDirectoryOutput {
		return new(DashboardDirectoryOutput).FromModel(node)
	})
}

// ListDashboardsResp 获取仪表盘目录树列表的响应。
type ListDashboardsResp struct {
	// Data 仪表盘目录树
	Data []*DashboardDirectoryOutput `json:"data"`
}

// EmptyOutput is the JSON response for APIs that return no data.
type EmptyOutput struct{}
