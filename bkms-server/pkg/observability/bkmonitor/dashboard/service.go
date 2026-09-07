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

import (
	"context"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/workspace"
	bkmapi "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/infras/cloudapi/bkmonitor"
)

// Service 提供应用仪表盘绑定管理能力。
type Service struct {
	store     AppDashboardStore
	newClient func(operator string) (bkmapi.MonitorClient, error)
}

// NewService 创建应用仪表盘绑定服务。
func NewService(store AppDashboardStore) *Service {
	return &Service{store: store, newClient: bkmapi.NewMonitorClient}
}

// List 获取应用绑定的仪表盘目录树
func (s *Service) List(
	ctx context.Context,
	ws *workspace.Workspace,
	appID, operator string,
) ([]*bkmapi.DashboardDirectoryNode, error) {
	tree, err := s.fetchDirectoryTree(ctx, ws, operator)
	if err != nil {
		return nil, err
	}
	records, err := s.store.ListByApp(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "list app dashboard bindings, appID=%s", appID)
	}
	return MergeDashboardTree(tree, records), nil
}

// Create 创建应用仪表盘绑定，创建前会校验 uid 在 bkmonitor 侧真实存在。
func (s *Service) Create(
	ctx context.Context,
	ws *workspace.Workspace,
	appID, uid, title, operator string,
) error {
	tree, err := s.fetchDirectoryTree(ctx, ws, operator)
	if err != nil {
		return err
	}
	if !dashboardExists(tree, uid) {
		return errors.Wrapf(ErrDashboardNotExist, "uid=%s", uid)
	}
	if _, err = s.store.Create(ctx, &AppDashboard{
		AppID:   appID,
		UID:     uid,
		Title:   title,
		Creator: operator,
	}); err != nil {
		return errors.Wrapf(err, "create app dashboard binding, appID=%s, uid=%s", appID, uid)
	}
	return nil
}

// Update 更新应用仪表盘绑定。
func (s *Service) Update(ctx context.Context, appID, uid string, updateData *AppDashboardUpdateData) error {
	if err := s.store.Update(ctx, appID, uid, updateData); err != nil {
		return errors.Wrapf(err, "update app dashboard binding, appID=%s, uid=%s", appID, uid)
	}
	return nil
}

// Delete 删除应用仪表盘绑定。
func (s *Service) Delete(ctx context.Context, appID, uid string) error {
	if err := s.store.Delete(ctx, appID, uid); err != nil {
		return errors.Wrapf(err, "delete app dashboard binding, appID=%s, uid=%s", appID, uid)
	}
	return nil
}

// MergeDashboardTree 将应用绑定的仪表盘记录与监控侧数据合并
func MergeDashboardTree(
	tree []*bkmapi.DashboardDirectoryNode,
	records []AppDashboard,
) []*bkmapi.DashboardDirectoryNode {
	if len(records) == 0 {
		return tree
	}

	recordMap := lo.SliceToMap(records, func(r AppDashboard) (string, AppDashboard) {
		return r.UID, r
	})

	lo.ForEach(tree, func(node *bkmapi.DashboardDirectoryNode, _ int) {
		if node == nil {
			return
		}
		lo.ForEach(node.Dashboards, func(item bkmapi.DashboardItem, i int) {
			if record, ok := recordMap[item.UID]; ok {
				node.Dashboards[i].Title = record.Title
			}
		})
	})

	return tree
}

// fetchDirectoryTree 解析 bkmonitor 项目并拉取仪表盘目录树。
func (s *Service) fetchDirectoryTree(
	ctx context.Context,
	ws *workspace.Workspace,
	operator string,
) ([]*bkmapi.DashboardDirectoryNode, error) {
	bkMonitorProjectID, err := ws.ResolveBkMonitorProjectID()
	if err != nil {
		return nil, errors.Wrap(err, "resolve bkmonitor space id")
	}
	client, err := s.newClient(operator)
	if err != nil {
		return nil, errors.Wrap(err, "new bkmonitor client")
	}
	tree, err := client.GetDashboardDirectoryTree(ctx, bkMonitorProjectID)
	if err != nil {
		return nil, errors.Wrapf(err, "get dashboard directory tree, bk_biz_id=%d", bkMonitorProjectID)
	}
	return tree, nil
}

// dashboardExists reports whether the given uid exists in the directory tree.
func dashboardExists(tree []*bkmapi.DashboardDirectoryNode, uid string) bool {
	return lo.SomeBy(tree, func(node *bkmapi.DashboardDirectoryNode) bool {
		return node != nil && lo.ContainsBy(node.Dashboards, func(item bkmapi.DashboardItem) bool {
			return item.UID == uid
		})
	})
}
