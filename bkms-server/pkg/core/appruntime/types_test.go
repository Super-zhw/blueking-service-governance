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

package appruntime_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime"
)

var _ = Describe("Management type", func() {
	DescribeTable("IsAppModelType",
		func(appType string, expected bool) {
			Expect(appruntime.IsAppModelType(appType)).To(Equal(expected))
		},
		Entry("bkmsApp", appruntime.AppTypeBKMSApp, true),
		Entry("legacy trpc", appruntime.AppTypeTRPC, true),
		Entry("legacy taf", appruntime.AppTypeTAF, true),
		Entry("helm", appruntime.AppTypeHelm, false),
		Entry("agones", appruntime.AppTypeAgones, false),
		Entry("empty", "", false),
	)

	DescribeTable("IsHelmBasedType",
		func(appType string, expected bool) {
			Expect(appruntime.IsHelmBasedType(appType)).To(Equal(expected))
		},
		Entry("helm", appruntime.AppTypeHelm, true),
		Entry("agones", appruntime.AppTypeAgones, true),
		Entry("bkmsApp", appruntime.AppTypeBKMSApp, false),
		Entry("legacy trpc", appruntime.AppTypeTRPC, false),
		Entry("empty", "", false),
	)
})
