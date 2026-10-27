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

package tafconf_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime/capability/apmservicename"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/workload/appmodelcore/framework/taf/tafconf"
)

var _ = Describe("APMServiceName", func() {
	Context("when taf config is valid", func() {
		It("should return app.server without setdivision", func() {
			content := `
<taf>
<application>
	<server>
		app=MyApp
		server=MyServer
	</server>
</application>
</taf>`
			serviceName, err := tafconf.APMServiceName(content, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(serviceName).To(Equal("MyApp.MyServer"))
		})

		It("should return app.server.setdivision with setdivision", func() {
			content := `
<taf>
<application>
	setdivision=prod.sh.*
	<server>
		app=XXXXGAME
		server=GameAAAAServer
	</server>
</application>
</taf>`
			serviceName, err := tafconf.APMServiceName(content, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(serviceName).To(Equal("XXXXGAME.GameAAAAServer.prodsh*"))
		})
	})

	Context("when taf config has variable references", func() {
		It("should expand ${{KEY}} variables before parsing", func() {
			content := `
<taf>
<application>
	<server>
		app=${{env.APP_NAME}}
		server=${{env.SERVER_NAME}}
	</server>
</application>
</taf>`
			envVars := map[string]string{
				"APP_NAME":    "DynamicApp",
				"SERVER_NAME": "DynamicServer",
			}
			serviceName, err := tafconf.APMServiceName(content, envVars)
			Expect(err).NotTo(HaveOccurred())
			Expect(serviceName).To(Equal("DynamicApp.DynamicServer"))
		})
	})

	Context("when taf config is missing required fields", func() {
		It("should return error when app is missing", func() {
			content := `
<taf>
<application>
	<server>
		server=MyServer
	</server>
</application>
</taf>`
			_, err := tafconf.APMServiceName(content, nil)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, apmservicename.ErrConfigMissing)).To(BeTrue())
		})

		It("should return error when server is missing", func() {
			content := `
<taf>
<application>
	<server>
		app=MyApp
	</server>
</application>
</taf>`
			_, err := tafconf.APMServiceName(content, nil)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, apmservicename.ErrConfigMissing)).To(BeTrue())
		})
	})

	Context("when taf config is invalid XML", func() {
		It("should return ErrAPMConfigParse for invalid xml", func() {
			content := `<taf><application><server>app=Test`
			_, err := tafconf.APMServiceName(content, nil)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, apmservicename.ErrConfigParse)).To(BeTrue())
		})
	})
})
