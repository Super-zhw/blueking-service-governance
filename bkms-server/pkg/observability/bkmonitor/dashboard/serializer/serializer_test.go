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

package serializer_test

import (
	"encoding/json"

	"github.com/go-playground/validator/v10"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	bkmapi "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/infras/cloudapi/bkmonitor"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/observability/bkmonitor/dashboard/serializer"
)

var _ = Describe("AppDashboard Serializer", func() {
	var validate *validator.Validate

	BeforeEach(func() {
		validate = validator.New()
		validate.SetTagName("binding")
	})

	Describe("AppDashboardCreateInput", func() {
		It("should require uid and title", func() {
			Expect(validate.Struct(serializer.AppDashboardCreateInput{})).To(HaveOccurred())
			Expect(validate.Struct(serializer.AppDashboardCreateInput{UID: "u", Title: "t"})).
				NotTo(HaveOccurred())
		})
	})

	Describe("AppDashboardUpdateInput", func() {
		It("should allow empty update (both fields optional)", func() {
			Expect(validate.Struct(serializer.AppDashboardUpdateInput{})).NotTo(HaveOccurred())
		})

		It("should accept title only", func() {
			title := "新标题"
			Expect(validate.Struct(serializer.AppDashboardUpdateInput{Title: &title})).NotTo(HaveOccurred())
		})
	})

	Describe("NewDashboardDirectoryOutputs", func() {
		It("should convert tree into minimal output", func() {
			tree := []*bkmapi.DashboardDirectoryNode{
				{
					ID:    100268,
					UID:   "afwxwj5d28x6of",
					Title: "各服务Prometheus指标",
					Dashboards: []bkmapi.DashboardItem{
						{
							ID: 100272, UID: "bfwy0guc537y8a", Title: "yxscampaignserver",
							URI: "db/yxscampaignserver", URL: "/grafana/d/bfwy0guc537y8a/yxscampaignserver",
						},
					},
				},
			}

			outputs := serializer.NewDashboardDirectoryOutputs(tree)

			Expect(outputs).To(HaveLen(1))
			Expect(outputs[0].ID).To(Equal(int64(100268)))
			Expect(outputs[0].UID).To(Equal("afwxwj5d28x6of"))
			Expect(outputs[0].Dashboards).To(HaveLen(1))
			Expect(outputs[0].Dashboards[0].UID).To(Equal("bfwy0guc537y8a"))
		})
	})

	Describe("ListDashboardsResp", func() {
		It("should marshal to the minimal dashboard tree shape", func() {
			resp := serializer.ListDashboardsResp{
				Data: []*serializer.DashboardDirectoryOutput{
					{
						ID:    100268,
						UID:   "afwxwj5d28x6of",
						Title: "各服务Prometheus指标",
						Dashboards: []*serializer.DashboardOutput{
							{
								ID:    100272,
								UID:   "bfwy0guc537y8a",
								Title: "yxscampaignserver",
								URI:   "db/yxscampaignserver",
								URL:   "/grafana/d/bfwy0guc537y8a/yxscampaignserver",
							},
						},
					},
				},
			}

			data, err := json.Marshal(resp)
			Expect(err).NotTo(HaveOccurred())

			var decoded serializer.ListDashboardsResp
			Expect(json.Unmarshal(data, &decoded)).NotTo(HaveOccurred())
			Expect(decoded.Data).To(HaveLen(1))
			Expect(decoded.Data[0].ID).To(Equal(int64(100268)))
			Expect(decoded.Data[0].UID).To(Equal("afwxwj5d28x6of"))
			Expect(decoded.Data[0].Dashboards).To(HaveLen(1))
			Expect(decoded.Data[0].Dashboards[0].UID).To(Equal("bfwy0guc537y8a"))
		})
	})
})
