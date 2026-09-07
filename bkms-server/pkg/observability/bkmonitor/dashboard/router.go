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

package dashboard

import "github.com/gin-gonic/gin"

// Handler contains views required by app dashboard Gin routes.
type Handler interface {
	ListAppDashboards(c *gin.Context)
	CreateAppDashboard(c *gin.Context)
	UpdateAppDashboard(c *gin.Context)
	DeleteAppDashboard(c *gin.Context)
}

// Register 注册应用仪表盘绑定相关路由。
func Register(rg *gin.RouterGroup, h Handler) {
	// 查询应用绑定的仪表盘列表
	rg.GET("/apps/:appID/bkmonitor/dashboards", h.ListAppDashboards)
	// 创建应用仪表盘绑定
	rg.POST("/apps/:appID/bkmonitor/dashboards", h.CreateAppDashboard)
	// 更新应用仪表盘绑定
	rg.PUT("/apps/:appID/bkmonitor/dashboards/:uid", h.UpdateAppDashboard)
	// 删除应用仪表盘绑定
	rg.DELETE("/apps/:appID/bkmonitor/dashboards/:uid", h.DeleteAppDashboard)
}
