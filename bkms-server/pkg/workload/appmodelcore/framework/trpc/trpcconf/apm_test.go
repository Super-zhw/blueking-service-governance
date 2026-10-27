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

package trpcconf_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime/capability/apmservicename"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/workload/appmodelcore/framework/trpc/trpcconf"
)

var _ = Describe("APMServiceName", func() {
	Context("when config has OpenTelemetry telemetry", func() {
		It("should return service_name from opentelemetry attributes", func() {
			content := `
plugins:
  telemetry:
    opentelemetry:
      addr: "http://otel.example.com"
      tenant_id: "tenant-123"
      attributes:
        - key: "service_name"
          value: "otel-service"
        - key: "env"
          value: "production"
server:
  app: myapp
  server: myserver
`
			serviceName, err := trpcconf.APMServiceName(content, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(serviceName).To(Equal("otel-service"))
		})

		It("should return error when opentelemetry attributes missing service_name", func() {
			content := `
plugins:
  telemetry:
    opentelemetry:
      addr: "http://otel.example.com"
      tenant_id: "tenant-123"
      attributes:
        - key: "env"
          value: "production"
`
			_, err := trpcconf.APMServiceName(content, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("service name not found"))
		})
	})

	Context("when config has Galileo telemetry", func() {
		It("should return app.server from trpc server config", func() {
			content := `
server:
  app: helloworld
  server: Greeter
plugins:
  telemetry:
    galileo:
      ocp_addr: "http://ocp.example.com"
      config:
        access_point: 3
      resource:
        platform: BCS
        tenant_id: test-tenant
`
			serviceName, err := trpcconf.APMServiceName(content, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(serviceName).To(Equal("helloworld.Greeter"))
		})

		It("should return ErrAPMConfigMissing when server.app is empty", func() {
			content := `
server:
  server: Greeter
plugins:
  telemetry:
    galileo:
      ocp_addr: "http://ocp.example.com"
`
			_, err := trpcconf.APMServiceName(content, nil)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, apmservicename.ErrConfigMissing)).To(BeTrue())
		})

		It("should return ErrAPMConfigMissing when server.server is empty", func() {
			content := `
server:
  app: helloworld
plugins:
  telemetry:
    galileo:
      ocp_addr: "http://ocp.example.com"
`
			_, err := trpcconf.APMServiceName(content, nil)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, apmservicename.ErrConfigMissing)).To(BeTrue())
		})

		It("should return ErrAPMConfigMissing when server section is missing", func() {
			content := `
plugins:
  telemetry:
    galileo:
      ocp_addr: "http://ocp.example.com"
`
			_, err := trpcconf.APMServiceName(content, nil)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, apmservicename.ErrConfigMissing)).To(BeTrue())
		})
	})

	Context("when telemetry config is empty or unknown", func() {
		It("should return ErrAPMConfigMissing when telemetry is empty", func() {
			content := `
plugins:
  telemetry:
server:
  app: myapp
  server: myserver
`
			_, err := trpcconf.APMServiceName(content, nil)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, apmservicename.ErrConfigMissing)).To(BeTrue())
		})

		It("should return ErrAPMConfigMissing when plugins section is missing", func() {
			content := `
server:
  app: myapp
  server: myserver
`
			_, err := trpcconf.APMServiceName(content, nil)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, apmservicename.ErrConfigMissing)).To(BeTrue())
		})
	})

	Context("when content has shell variable references", func() {
		It("should expand ${VAR} before parsing", func() {
			content := `
server:
  app: ${APP_NAME}
  server: ${SERVER_NAME}
plugins:
  telemetry:
    galileo:
      ocp_addr: "http://ocp.example.com"
`
			envVars := map[string]string{
				"APP_NAME":    "expanded-app",
				"SERVER_NAME": "expanded-server",
			}
			serviceName, err := trpcconf.APMServiceName(content, envVars)
			Expect(err).NotTo(HaveOccurred())
			Expect(serviceName).To(Equal("expanded-app.expanded-server"))
		})

		It("should expand ${VAR} in opentelemetry service_name", func() {
			content := `
plugins:
  telemetry:
    opentelemetry:
      addr: "http://otel.example.com"
      tenant_id: "tenant-123"
      attributes:
        - key: "service_name"
          value: "${MY_SERVICE}"
`
			envVars := map[string]string{
				"MY_SERVICE": "dynamic-service-name",
			}
			serviceName, err := trpcconf.APMServiceName(content, envVars)
			Expect(err).NotTo(HaveOccurred())
			Expect(serviceName).To(Equal("dynamic-service-name"))
		})
	})

	Context("when content is invalid YAML", func() {
		It("should return ErrAPMConfigParse for invalid yaml", func() {
			content := `
plugins:
  telemetry:
    opentelemetry:
      addr: [invalid yaml
`
			_, err := trpcconf.APMServiceName(content, nil)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, apmservicename.ErrConfigParse)).To(BeTrue())
		})
	})
})
