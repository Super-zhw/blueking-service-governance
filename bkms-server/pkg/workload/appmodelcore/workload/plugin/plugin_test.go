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

package plugin_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	bkmsapp "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/app"
	envmodel "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/env/model"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/workload/appmodelcore/appmodel"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/workload/appmodelcore/workload/plugin"
)

// fakePlugin 只用于验证插件选择，不产生任何资源。
type fakePlugin struct{ typ string }

func (p fakePlugin) Type() string { return p.typ }

func (p fakePlugin) Start(
	_ context.Context, _ *envmodel.Environment, _ *bkmsapp.Application, _ *appmodel.AppModel, _ plugin.RenderContext,
) (plugin.WorkloadPluginSession, error) {
	return nil, nil
}

var _ = BeforeSuite(func() {
	for _, typ := range []string{appmodel.WorkloadTypeTrpc, appmodel.WorkloadTypeTaf} {
		plugin.MustRegisterWorkloadPlugin(fakePlugin{typ: typ})
	}
})

var _ = Describe("GetFrameworkPlugin", func() {
	DescribeTable("should map a framework to its plugin",
		func(framework, expectedType string) {
			p, err := plugin.GetFrameworkPlugin(framework)
			Expect(err).NotTo(HaveOccurred())
			Expect(p.Type()).To(Equal(expectedType))
		},
		Entry("trpc", "trpc", appmodel.WorkloadTypeTrpc),
		Entry("taf", "taf", appmodel.WorkloadTypeTaf),
	)

	It("should return no plugin for blank", func() {
		p, err := plugin.GetFrameworkPlugin("blank")
		Expect(err).NotTo(HaveOccurred())
		Expect(p).To(BeNil())
	})

	DescribeTable("should reject an empty or unknown framework",
		func(framework string) {
			_, err := plugin.GetFrameworkPlugin(framework)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no workload plugin for framework"))
		},
		// 技术栈非法时框架为空，不能再默认按 trpc 渲染
		Entry("empty", ""),
		Entry("unknown", "spring"),
	)
})
