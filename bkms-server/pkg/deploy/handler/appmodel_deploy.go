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

package handler

import (
	"github.com/gin-gonic/gin"

	bkmsapp "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/app"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime/runtimemod/mods"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/deploy/serializer"
)

// 占位声明，用于让 swag 能在本文件解析 swagger 注释中引用的 serializer 类型
var _ = serializer.EmptyOutput{}

// anyFramework 统一接口不限制应用框架，适用于所有 AppModel 应用（含 bkmsApp 与旧 trpc / taf）。
const anyFramework = ""

// appMatchesFramework 判断应用能否使用某组 AppModel 部署接口：须为 AppModel 应用，旧专属接口还要求框架一致。
func appMatchesFramework(app *bkmsapp.Application, expectedFramework string) bool {
	if !appruntime.IsAppModelType(app.Type) {
		return false
	}
	return expectedFramework == anyFramework || mods.Of(app).Stack.Framework == expectedFramework
}

// ListAppModelDeployRecords 获取 AppModel 应用部署记录列表
//
//	@ID			ListAppModelDeployRecords
//	@Summary	获取 AppModel 应用部署记录列表
//	@Tags		deploy
//	@Produce	json
//	@Security	BkUserInfo
//	@Security	BkUserCredential
//	@Param		appID			path		string	true	"应用 ID"
//	@Param		envName			path		string	true	"部署环境名称"
//	@Param		trafficLaneName	query		string	false	"部署的泳道名称（空字符串表示不使用泳道）"
//	@Param		keyword			query		string	false	"搜索关键字"
//	@Param		page			query		int		true	"分页页码（从 1 开始）"
//	@Param		pageSize		query		int		true	"分页大小"
//	@Success	200				{object}	serializer.ListAppModelDeployRecordsOutput
//	@Failure	400				{object}	bkerrs.GinErrorOutput
//	@Router		/apps/{appID}/envs/{envName}/appmodel-deploys [get]
func (h *Handler) ListAppModelDeployRecords(c *gin.Context) {
	h.listAppModelDeployRecords(c)
}

// PreCheckAppModelDeploy checks an AppModel deployment before create.
//
//	@ID			PreCheckAppModelDeploy
//	@Summary	AppModel 部署前检查
//	@Tags		deploy
//	@Produce	json
//	@Security	BkUserInfo
//	@Security	BkUserCredential
//	@Param		appID	path		string	true	"应用 ID"
//	@Param		envName	path		string	true	"部署环境名称"
//	@Success	200		{object}	serializer.DeployPreCheckOutput
//	@Failure	400		{object}	bkerrs.GinErrorOutput
//	@Failure	404		{object}	bkerrs.GinErrorOutput
//	@Router		/apps/{appID}/envs/{envName}/appmodel-deploys/precheck [get]
func (h *Handler) PreCheckAppModelDeploy(c *gin.Context) {
	h.preCheckDeploy(c, anyFramework)
}

// CreateAppModelDeploy 创建 AppModel 应用部署
//
//	@ID			CreateAppModelDeploy
//	@Summary	创建 AppModel 应用部署
//	@Tags		deploy
//	@Accept		json
//	@Produce	json
//	@Security	BkUserInfo
//	@Security	BkUserCredential
//	@Param		appID	path		string								true	"应用 ID"
//	@Param		envName	path		string								true	"部署环境名称"
//	@Param		body	body		serializer.CreateAppModelDeployInput	true	"创建 AppModel 部署请求"
//	@Success	200		{object}	serializer.EmptyOutput
//	@Failure	400		{object}	bkerrs.GinErrorOutput
//	@Router		/apps/{appID}/envs/{envName}/appmodel-deploys [post]
func (h *Handler) CreateAppModelDeploy(c *gin.Context) {
	h.createAppModelDeploy(c)
}

// DeleteAppModelDeploy 删除 AppModel 应用部署（下架当前环境最新版本）
//
//	@ID			DeleteAppModelDeploy
//	@Summary	删除 AppModel 应用部署（下架当前环境最新版本）
//	@Tags		deploy
//	@Produce	json
//	@Security	BkUserInfo
//	@Security	BkUserCredential
//	@Param		appID			path		string	true	"应用 ID"
//	@Param		envName			path		string	true	"部署环境名称"
//	@Param		trafficLaneName	query		string	false	"部署的泳道名称（空字符串表示不使用泳道）"
//	@Success	200				{object}	serializer.EmptyOutput
//	@Failure	400				{object}	bkerrs.GinErrorOutput
//	@Router		/apps/{appID}/envs/{envName}/appmodel-deploys [delete]
func (h *Handler) DeleteAppModelDeploy(c *gin.Context) {
	h.deleteAppModelDeploy(c)
}

// ListAppModelResourceSnapshots 获取 AppModel 应用资源快照列表
//
//	@ID			ListAppModelResourceSnapshots
//	@Summary	获取 AppModel 应用资源快照列表
//	@Tags		deploy
//	@Produce	json
//	@Security	BkUserInfo
//	@Security	BkUserCredential
//	@Param		appID		path		string	true	"应用 ID"
//	@Param		envName		path		string	true	"部署环境名称"
//	@Param		deployID	path		string	true	"部署记录 ID"
//	@Param		page		query		int		true	"分页页码（从 1 开始）"
//	@Param		pageSize	query		int		true	"分页大小"
//	@Success	200			{object}	serializer.ListAppModelResourceSnapshotsOutput
//	@Failure	400			{object}	bkerrs.GinErrorOutput
//	@Router		/apps/{appID}/envs/{envName}/appmodel-deploys/{deployID}/resource-snapshots [get]
func (h *Handler) ListAppModelResourceSnapshots(c *gin.Context) {
	h.listAppModelResourceSnapshots(c)
}

// GetAppModelResourceSnapshot 获取 AppModel 应用部署记录下某个资源的快照详情（具体 k8s 资源的 Manifest 等）
//
//	@ID			GetAppModelResourceSnapshot
//	@Summary	获取 AppModel 应用部署记录下某个资源的快照详情（具体 k8s 资源的 Manifest 等）
//	@Tags		deploy
//	@Produce	json
//	@Security	BkUserInfo
//	@Security	BkUserCredential
//	@Param		appID		path		string	true	"应用 ID"
//	@Param		envName		path		string	true	"部署环境名称"
//	@Param		deployID	path		string	true	"部署记录 ID"
//	@Param		snapshotID	path		string	true	"资源清单快照 ID"
//	@Success	200			{object}	serializer.GetAppModelResourceSnapshotOutput
//	@Failure	400			{object}	bkerrs.GinErrorOutput
//	@Router		/apps/{appID}/envs/{envName}/appmodel-deploys/{deployID}/resource-snapshots/{snapshotID} [get]
func (h *Handler) GetAppModelResourceSnapshot(c *gin.Context) {
	h.getAppModelResourceSnapshot(c)
}

// GetLatestAppModelDeployStatus 获取 AppModel 应用最新一次部署的状态
//
//	@ID			GetLatestAppModelDeployStatus
//	@Summary	获取 AppModel 应用最新一次部署的状态
//	@Tags		deploy
//	@Produce	json
//	@Security	BkUserInfo
//	@Security	BkUserCredential
//	@Param		appID			path		string	true	"应用 ID"
//	@Param		envName			path		string	true	"部署环境名称"
//	@Param		trafficLaneName	query		string	false	"部署的泳道名称（空字符串表示不使用泳道）"
//	@Success	200				{object}	serializer.GetLatestAppModelDeployStatusOutput
//	@Failure	400				{object}	bkerrs.GinErrorOutput
//	@Router		/apps/{appID}/envs/{envName}/appmodel-deploys/latest-status [get]
func (h *Handler) GetLatestAppModelDeployStatus(c *gin.Context) {
	h.getLatestAppModelDeployStatus(c, anyFramework)
}
