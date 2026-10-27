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

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/workload/appmodelcore/framework/taf/tafconf"
)

var _ = Describe("Parse", func() {
	Context("when parsing valid TAF XML data", func() {
		It("should parse taf config with setdivision successfully", func() {
			xmlData := `
<taf>
<application>
	setdivision=dev.sh.*
	<server>
		app=XXXXGAME
		server=GameAAAAServer
	</server>
	<client>
	</client>
</application>
</taf>`
			config, err := tafconf.Parse([]byte(xmlData))

			Expect(err).NotTo(HaveOccurred())
			Expect(config).NotTo(BeNil())
			Expect(config.Application.Server.CharData).To(ContainSubstring("app=XXXXGAME"))
			Expect(config.Application.Server.CharData).To(ContainSubstring("server=GameAAAAServer"))
			Expect(config.Application.CharData).To(ContainSubstring("setdivision=dev.sh.*"))
		})

		It("should get service name with setdivision", func() {
			xmlData := `
<taf>
<application>
	setdivision=prod.sh.*
	<server>
		app=XXXXGAME
		server=GameAAAAServer
	</server>
	<client>
	</client>
</application>
</taf>`
			config, err := tafconf.Parse([]byte(xmlData))
			Expect(err).NotTo(HaveOccurred())

			serviceName, err := config.GetServiceName()
			Expect(err).NotTo(HaveOccurred())
			Expect(serviceName).To(Equal("XXXXGAME.GameAAAAServer.prodsh*"))
		})

		It("should get service name without setdivision", func() {
			xmlData := `
<taf>
<application>
	<server>
		app=MyApp
		server=MyServer
	</server>
	<client>
	</client>
</application>
</taf>`
			config, err := tafconf.Parse([]byte(xmlData))
			Expect(err).NotTo(HaveOccurred())

			serviceName, err := config.GetServiceName()
			Expect(err).NotTo(HaveOccurred())
			Expect(serviceName).To(Equal("MyApp.MyServer"))
		})
	})

	Context("when parsing invalid TAF XML data", func() {
		It("should return error for invalid xml syntax", func() {
			xmlData := `<taf><application><server>app=Test`
			_, err := tafconf.Parse([]byte(xmlData))
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when required fields are missing", func() {
		It("should return error when app is missing", func() {
			xmlData := `
<taf>
<application>
	<server>
		server=MyServer
	</server>
</application>
</taf>`
			config, err := tafconf.Parse([]byte(xmlData))
			Expect(err).NotTo(HaveOccurred())

			_, err = config.GetServiceName()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("missing app or server"))
		})

		It("should return error when server is missing", func() {
			xmlData := `
<taf>
<application>
	<server>
		app=MyApp
	</server>
</application>
</taf>`
			config, err := tafconf.Parse([]byte(xmlData))
			Expect(err).NotTo(HaveOccurred())

			_, err = config.GetServiceName()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("missing app or server"))
		})
	})
})
