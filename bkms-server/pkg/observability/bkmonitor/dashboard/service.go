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
	"bytes"
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

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

// List 获取应用绑定的仪表盘目录树（已缝合应用自定义数据）。
func (s *Service) List(
	ctx context.Context,
	ws *workspace.Workspace,
	appID, operator string,
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
	records, err := s.store.ListByApp(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "list app dashboard bindings, appID=%s", appID)
	}
	return MergeDashboardTree(tree, records), nil
}

// Create 创建应用仪表盘绑定。
func (s *Service) Create(ctx context.Context, appID, uid, title, creator string) error {
	_, err := s.store.Create(ctx, &AppDashboard{
		AppID:   appID,
		UID:     uid,
		Title:   title,
		Creator: creator,
	})
	if err != nil {
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

// MergeDashboardTree 将应用已绑定的仪表盘记录缝合进目录树：
// 对树中每个仪表盘，若其 uid 命中应用的自定义绑定，则用绑定里的 title 覆盖，
// 并重建 url/uri（grafana 的 url 携带 title 对应的 slug，title 变化会导致 url 变化）。
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

	for _, node := range tree {
		if node == nil {
			continue
		}
		for i := range node.Dashboards {
			item := &node.Dashboards[i]
			record, ok := recordMap[item.UID]
			if !ok {
				continue
			}
			item.Title = record.Title
			item.Slug = slugify(record.Title)
			item.URI = "db/" + item.Slug
			item.URL = "/grafana/d/" + item.UID + "/" + item.Slug
		}
	}

	return tree
}

// slugify 生成 grafana 风格的 slug，参考 grafana/pkg/infra/slugify 的核心算法：
// ASCII 字母数字转小写保留；空白与常见标点（空格、'-'、'_'、'/'、'.' 等）被忽略；
// 其它字符（含 CJK）按 UTF-8 字节的十六进制（小写）编码；段与段之间以 '-' 连接。
func slugify(value string) string {
	value = strings.ToLower(value)
	var buffer bytes.Buffer
	lastInvalid := false

	for _, c := range value {
		if isValidSlugChar(c) {
			if lastInvalid {
				buffer.WriteByte('-')
			}
			buffer.WriteRune(c)
			lastInvalid = false
			continue
		}

		if isOmittedSlugChar(c) {
			lastInvalid = true
			continue
		}

		p := make([]byte, utf8.UTFMax)
		n := utf8.EncodeRune(p, c)
		if lastInvalid {
			buffer.WriteByte('-')
		}
		for i := 0; i < n; i++ {
			fmt.Fprintf(&buffer, "%x", p[i])
		}
		lastInvalid = true
	}

	return strings.Trim(buffer.String(), "-")
}

// isValidSlugChar reports whether c is a valid slug character (lowercase ASCII letter or digit).
func isValidSlugChar(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')
}

// isOmittedSlugChar reports whether c should be omitted from the slug.
func isOmittedSlugChar(c rune) bool {
	return strings.ContainsRune(" ,\"'\n\r\x00?().-_[]/\\!{}%", c)
}
