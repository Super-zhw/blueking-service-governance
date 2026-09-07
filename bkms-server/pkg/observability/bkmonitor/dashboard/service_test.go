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

package dashboard_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	bkmapi "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/infras/cloudapi/bkmonitor"
	. "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/observability/bkmonitor/dashboard"
)

var _ = Describe("MergeDashboardTree", func() {
	It("should override title and rebuild url/uri for bound dashboards", func() {
		tree := []*bkmapi.DashboardDirectoryNode{
			{
				ID:    0,
				UID:   "",
				Title: "General",
				Dashboards: []bkmapi.DashboardItem{
					{ID: 1, UID: "uid-1", Title: "老标题", URI: "db/old", URL: "/grafana/d/uid-1/old"},
					{ID: 2, UID: "uid-2", Title: "另一个", URI: "db/another", URL: "/grafana/d/uid-2/another"},
				},
			},
		}
		records := []AppDashboard{
			{AppID: "app-a", UID: "uid-1", Title: "各服务Prometheus指标"},
		}

		merged := MergeDashboardTree(tree, records)

		Expect(merged[0].Dashboards[0].Title).To(Equal("各服务Prometheus指标"))
		Expect(merged[0].Dashboards[0].URL).To(Equal(
			"/grafana/d/uid-1/e59084-e69c8d-e58aa1-prometheuse68c87-e6a087"))
		Expect(merged[0].Dashboards[0].URI).To(Equal(
			"db/e59084-e69c8d-e58aa1-prometheuse68c87-e6a087"))

		// 未绑定的仪表盘保持不变
		Expect(merged[0].Dashboards[1].Title).To(Equal("另一个"))
		Expect(merged[0].Dashboards[1].URL).To(Equal("/grafana/d/uid-2/another"))
	})

	It("should handle pure ASCII title slug", func() {
		tree := []*bkmapi.DashboardDirectoryNode{
			{
				ID:    0,
				UID:   "",
				Title: "General",
				Dashboards: []bkmapi.DashboardItem{
					{ID: 1, UID: "bfwy0guc537y8a", Title: "Old", URI: "db/old", URL: "/grafana/d/bfwy0guc537y8a/old"},
				},
			},
		}
		records := []AppDashboard{
			{AppID: "app-a", UID: "bfwy0guc537y8a", Title: "yxscampaignserver"},
		}

		merged := MergeDashboardTree(tree, records)

		Expect(merged[0].Dashboards[0].Title).To(Equal("yxscampaignserver"))
		Expect(merged[0].Dashboards[0].URL).To(Equal("/grafana/d/bfwy0guc537y8a/yxscampaignserver"))
		Expect(merged[0].Dashboards[0].URI).To(Equal("db/yxscampaignserver"))
	})

	It("should leave tree unchanged when there are no records", func() {
		tree := []*bkmapi.DashboardDirectoryNode{
			{
				ID:    0,
				UID:   "",
				Title: "General",
				Dashboards: []bkmapi.DashboardItem{
					{ID: 1, UID: "uid-1", Title: "t1", URI: "db/t1", URL: "/grafana/d/uid-1/t1"},
				},
			},
		}

		merged := MergeDashboardTree(tree, nil)

		Expect(merged).To(Equal(tree))
		Expect(merged[0].Dashboards[0].Title).To(Equal("t1"))
	})
})
