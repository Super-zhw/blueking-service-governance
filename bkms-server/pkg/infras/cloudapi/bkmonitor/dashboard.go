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

// Package bkmonitor api client，如：蓝鲸监控的 apm、蓝鲸监控的 metadata
package bkmonitor

import (
	"context"
	"net/http"

	"github.com/TencentBlueKing/bk-apigateway-sdks/core/bkapi"
	"github.com/TencentBlueKing/gopkg/mapx"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

// GetDashboardDirectoryTree 获取蓝鲸监控仪表盘数据
func (c *MonitorGatewayClient) GetDashboardDirectoryTree(
	ctx context.Context,
	bkBizID int64,
) ([]*DashboardDirectoryNode, error) {
	params := map[string]string{
		"bk_biz_id": cast.ToString(bkBizID),
	}

	resp, err := c.handleOperation(ctx, c.NewOperation(
		bkapi.OperationConfig{
			Name:   "get_dashboard_directory_tree",
			Method: http.MethodGet,
			Path:   "/app/dashboard/get_dashboard_directory_tree/",
		},
	).SetQueryParams(params))
	if err != nil {
		return nil, errors.Wrapf(err, "get dashboard directory tree failed, bk_biz_id: %d", bkBizID)
	}

	result := make([]*DashboardDirectoryNode, 0)
	if err = mapstructure.Decode(mapx.GetList(resp, "data"), &result); err != nil {
		return nil, errors.Wrapf(err, "decode dashboard directory tree failed, bk_biz_id: %d", bkBizID)
	}

	return result, nil
}

// GetDashboardDetail 获取仪表盘详情，仪表盘不存在时返回 (nil, nil)。
func (c *MonitorGatewayClient) GetDashboardDetail(
	ctx context.Context,
	bkBizID int64,
	dashboardUID string,
) (*DashboardDetail, error) {
	params := map[string]string{
		"bk_biz_id":     cast.ToString(bkBizID),
		"dashboard_uid": dashboardUID,
	}

	resp, err := c.handleOperation(ctx, c.NewOperation(
		bkapi.OperationConfig{
			Name:   "get_dashboard_detail",
			Method: http.MethodGet,
			Path:   "/app/dashboard/get_dashboard_detail/",
		},
	).SetQueryParams(params))
	if err != nil {
		return nil, errors.Wrapf(err, "get dashboard detail failed, bk_biz_id: %d, uid: %s", bkBizID, dashboardUID)
	}

	// 仪表盘不存在时，data 为 null。
	if resp["data"] == nil {
		return nil, nil
	}

	result := new(DashboardDetail)
	if err = mapstructure.Decode(resp["data"], result); err != nil {
		return nil, errors.Wrapf(err, "decode dashboard detail failed, bk_biz_id: %d, uid: %s", bkBizID, dashboardUID)
	}

	return result, nil
}
