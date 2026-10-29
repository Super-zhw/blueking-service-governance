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

package appmodel_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/workload/appmodelcore/appmodel"
)

var _ = Describe("Workload.GetFrameworkConfig", func() {
	DescribeTable("returns the framework config only when its version is set",
		func(w appmodel.Workload, expected map[string]any, expectedOK bool) {
			cfg, ok := w.GetFrameworkConfig()
			Expect(ok).To(Equal(expectedOK))
			Expect(cfg).To(Equal(expected))
		},
		Entry("version not set", appmodel.Workload{FrameworkConfig: map[string]any{"a": 1}}, nil, false),
		Entry("version set with nil config", appmodel.Workload{FrameworkConfigVersion: 1}, map[string]any{}, true),
		Entry("version set with config",
			appmodel.Workload{FrameworkConfigVersion: 1, FrameworkConfig: map[string]any{"a": 1}},
			map[string]any{"a": 1}, true),
	)
})
