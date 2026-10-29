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

// ListBKMSDeployRecords 获取 bkmsApp 应用部署记录列表
//
//	@ID			ListBKMSDeployRecords
//	@Summary	获取 bkmsApp 应用部署记录列表
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
//	@Router		/apps/{appID}/envs/{envName}/bkms-deploys [get]
func (h *Handler) ListBKMSDeployRecords(c *gin.Context) {
	h.listAppModelDeployRecords(c)
}

// PreCheckBKMSDeploy bkmsApp 应用部署前检查
//
//	@ID			PreCheckBKMSDeploy
//	@Summary	bkmsApp 应用部署前检查
//	@Tags		deploy
//	@Produce	json
//	@Security	BkUserInfo
//	@Security	BkUserCredential
//	@Param		appID	path		string	true	"应用 ID"
//	@Param		envName	path		string	true	"部署环境名称"
//	@Success	200		{object}	serializer.DeployPreCheckOutput
//	@Failure	400		{object}	bkerrs.GinErrorOutput
//	@Failure	404		{object}	bkerrs.GinErrorOutput
//	@Router		/apps/{appID}/envs/{envName}/bkms-deploys/precheck [get]
func (h *Handler) PreCheckBKMSDeploy(c *gin.Context) {
	h.preCheckDeploy(c, "")
}

// CreateBKMSDeploy 创建 bkmsApp 应用部署
//
//	@ID			CreateBKMSDeploy
//	@Summary	创建 bkmsApp 应用部署
//	@Tags		deploy
//	@Accept		json
//	@Produce	json
//	@Security	BkUserInfo
//	@Security	BkUserCredential
//	@Param		appID	path		string								true	"应用 ID"
//	@Param		envName	path		string								true	"部署环境名称"
//	@Param		body	body		serializer.CreateAppModelDeployInput	true	"创建部署请求"
//	@Success	200		{object}	serializer.EmptyOutput
//	@Failure	400		{object}	bkerrs.GinErrorOutput
//	@Router		/apps/{appID}/envs/{envName}/bkms-deploys [post]
func (h *Handler) CreateBKMSDeploy(c *gin.Context) {
	h.createAppModelDeploy(c)
}

// DeleteBKMSDeploy 删除 bkmsApp 应用部署（下架当前环境最新版本）
//
//	@ID			DeleteBKMSDeploy
//	@Summary	删除 bkmsApp 应用部署（下架当前环境最新版本）
//	@Tags		deploy
//	@Produce	json
//	@Security	BkUserInfo
//	@Security	BkUserCredential
//	@Param		appID			path		string	true	"应用 ID"
//	@Param		envName			path		string	true	"部署环境名称"
//	@Param		trafficLaneName	query		string	false	"部署的泳道名称（空字符串表示不使用泳道）"
//	@Success	200				{object}	serializer.EmptyOutput
//	@Failure	400				{object}	bkerrs.GinErrorOutput
//	@Router		/apps/{appID}/envs/{envName}/bkms-deploys [delete]
func (h *Handler) DeleteBKMSDeploy(c *gin.Context) {
	h.deleteAppModelDeploy(c)
}

// ListBKMSResourceSnapshots 获取 bkmsApp 应用资源快照列表
//
//	@ID			ListBKMSResourceSnapshots
//	@Summary	获取 bkmsApp 应用资源快照列表
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
//	@Router		/apps/{appID}/envs/{envName}/bkms-deploys/{deployID}/resource-snapshots [get]
func (h *Handler) ListBKMSResourceSnapshots(c *gin.Context) {
	h.listAppModelResourceSnapshots(c)
}

// GetBKMSResourceSnapshot 获取 bkmsApp 应用部署记录下某个资源的快照详情（具体 k8s 资源的 Manifest 等）
//
//	@ID			GetBKMSResourceSnapshot
//	@Summary	获取 bkmsApp 应用部署记录下某个资源的快照详情（具体 k8s 资源的 Manifest 等）
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
//	@Router		/apps/{appID}/envs/{envName}/bkms-deploys/{deployID}/resource-snapshots/{snapshotID} [get]
func (h *Handler) GetBKMSResourceSnapshot(c *gin.Context) {
	h.getAppModelResourceSnapshot(c)
}

// GetLatestBKMSDeployStatus 获取 bkmsApp 应用最新一次部署的状态
//
//	@ID			GetLatestBKMSDeployStatus
//	@Summary	获取 bkmsApp 应用最新一次部署的状态
//	@Tags		deploy
//	@Produce	json
//	@Security	BkUserInfo
//	@Security	BkUserCredential
//	@Param		appID			path		string	true	"应用 ID"
//	@Param		envName			path		string	true	"部署环境名称"
//	@Param		trafficLaneName	query		string	false	"部署的泳道名称（空字符串表示不使用泳道）"
//	@Success	200				{object}	serializer.GetLatestAppModelDeployStatusOutput
//	@Failure	400				{object}	bkerrs.GinErrorOutput
//	@Router		/apps/{appID}/envs/{envName}/bkms-deploys/latest-status [get]
func (h *Handler) GetLatestBKMSDeployStatus(c *gin.Context) {
	h.getLatestAppModelDeployStatus(c, "")
}

// appMatchesFramework 判断应用能否使用某组部署接口：须为 AppModel 应用（trpc、taf、bkmsApp）；
// expectedFramework 为旧 trpc / taf 专属接口要求的框架，bkms-deploys 接口传空字符串，不校验框架。
func appMatchesFramework(app *bkmsapp.Application, expectedFramework string) bool {
	if !appruntime.IsAppModelType(app.Type) {
		return false
	}
	return expectedFramework == "" || mods.Of(app).Stack.Framework == expectedFramework
}
