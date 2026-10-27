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

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/workload/appmodelcore/framework/trpc/trpcconf"
)

var _ = Describe("TelemetryConfig", func() {
	Describe("ParseTelemetryConfig", func() {
		Context("when parsing valid YAML data", func() {
			It("should parse galileo config successfully", func() {
				yamlData := `
plugins:
  telemetry:
    galileo:
      ocp_addr: "http://ocp.example.com"
      config:
        access_point: 3
        metrics_config:
          enable: true
          exporter:
            collector:
              addr: "http://metrics.example.com"
        traces_config:
          enable: true
          exporter:
            collector:
              addr: "http://traces.example.com"
        logs_config:
          enable: false
        profiles_config:
          enable: true
          exporter:
            collector:
              addr: "http://profiles.example.com"
      resource:
        platform: BCS
        tenant_id: test-tenant
`
				telemetry, err := trpcconf.ParseTelemetryConfig([]byte(yamlData))

				Expect(err).NotTo(HaveOccurred())
				Expect(telemetry).NotTo(BeNil())
				Expect(telemetry.Galileo).NotTo(BeNil())
				Expect(telemetry.Galileo.OCPAddr).To(Equal("http://ocp.example.com"))
				Expect(telemetry.Galileo.Config.AccessPoint).To(Equal(3))
				Expect(telemetry.Galileo.Config.MetricsConfig.Enable).To(BeTrue())
				Expect(
					telemetry.Galileo.Config.MetricsConfig.Exporter.Collector.Addr,
				).To(Equal("http://metrics.example.com"))
				Expect(telemetry.Galileo.Config.TracesConfig.Enable).To(BeTrue())
				Expect(telemetry.Galileo.Config.LogsConfig.Enable).To(BeFalse())
				Expect(telemetry.Galileo.Config.ProfilesConfig.Enable).To(BeTrue())
				Expect(telemetry.Galileo.Resource.Platform).To(Equal("BCS"))
				Expect(telemetry.Galileo.Resource.TenantID).To(Equal("test-tenant"))
			})

			It("should parse opentelemetry config successfully", func() {
				yamlData := `
plugins:
  telemetry:
    opentelemetry:
      addr: "http://otel.example.com"
      tenant_id: "tenant-123"
      attributes:
        - key: "service_name"
          value: "my-service"
        - key: "env"
          value: "production"
`
				telemetry, err := trpcconf.ParseTelemetryConfig([]byte(yamlData))

				Expect(err).NotTo(HaveOccurred())
				Expect(telemetry).NotTo(BeNil())
				Expect(telemetry.OpenTelemetry).NotTo(BeNil())
				Expect(telemetry.OpenTelemetry.Addr).To(Equal("http://otel.example.com"))
				Expect(telemetry.OpenTelemetry.TenantID).To(Equal("tenant-123"))
				Expect(telemetry.OpenTelemetry.Attributes).To(HaveLen(2))
				Expect(telemetry.OpenTelemetry.Attributes[0].Key).To(Equal("service_name"))
				Expect(telemetry.OpenTelemetry.Attributes[0].Value).To(Equal("my-service"))
			})
		})

		Context("when parsing empty or minimal config", func() {
			It("should handle empty telemetry config", func() {
				yamlData := `
plugins:
  telemetry:
`
				telemetry, err := trpcconf.ParseTelemetryConfig([]byte(yamlData))

				Expect(err).NotTo(HaveOccurred())
				Expect(telemetry).NotTo(BeNil())
				Expect(telemetry.OpenTelemetry).To(BeNil())
				Expect(telemetry.Galileo).To(BeNil())
			})
		})

		Context("when parsing invalid YAML data", func() {
			It("should return error for invalid yaml syntax", func() {
				yamlData := `
plugins:
  telemetry:
    galileo:
      ocp_addr: [invalid yaml
`
				telemetry, err := trpcconf.ParseTelemetryConfig([]byte(yamlData))

				Expect(err).To(HaveOccurred())
				Expect(telemetry).To(BeNil())
			})
		})
	})

	Describe("GetTelemetryType", func() {
		Context("when determining telemetry type", func() {
			It("should return Galileo type when galileo has ocp_addr", func() {
				telemetry := &trpcconf.Telemetry{
					Galileo: &trpcconf.GalileoConfig{
						OCPAddr: "http://ocp.example.com",
					},
				}

				Expect(telemetry.GetTelemetryType()).To(Equal(trpcconf.TelemetryTypeGalileo))
				Expect(telemetry.IsGalileo()).To(BeTrue())
				Expect(telemetry.IsOpenTelemetry()).To(BeFalse())
			})

			It("should return OpenTelemetry type when opentelemetry is fully configured", func() {
				telemetry := &trpcconf.Telemetry{
					OpenTelemetry: &trpcconf.OpenTelemetryConfig{
						Addr:     "http://otel.example.com",
						TenantID: "tenant-123",
						Attributes: []trpcconf.Attribute{
							{Key: "service_name", Value: "test"},
						},
					},
				}

				Expect(telemetry.GetTelemetryType()).To(Equal(trpcconf.TelemetryTypeOpenTelemetry))
				Expect(telemetry.IsOpenTelemetry()).To(BeTrue())
				Expect(telemetry.IsGalileo()).To(BeFalse())
			})

			It("should return Unknown type when both are nil", func() {
				telemetry := &trpcconf.Telemetry{}

				Expect(telemetry.GetTelemetryType()).To(Equal(trpcconf.TelemetryTypeUnknown))
				Expect(telemetry.IsOpenTelemetry()).To(BeFalse())
				Expect(telemetry.IsGalileo()).To(BeFalse())
			})
		})
	})

	Describe("OpenTelemetryConfig.GetServiceName", func() {
		Context("when getting service name from attributes", func() {
			It("should return service name when present", func() {
				config := &trpcconf.OpenTelemetryConfig{
					Attributes: []trpcconf.Attribute{
						{Key: "env", Value: "production"},
						{Key: "service_name", Value: "my-service.my-server"},
						{Key: "version", Value: "1.0.0"},
					},
				}

				serviceName, err := config.GetServiceName()
				Expect(err).NotTo(HaveOccurred())
				Expect(serviceName).To(Equal("my-service.my-server"))
			})

			It("should return error when service_name not found", func() {
				config := &trpcconf.OpenTelemetryConfig{
					Attributes: []trpcconf.Attribute{
						{Key: "env", Value: "production"},
						{Key: "version", Value: "1.0.0"},
					},
				}

				serviceName, err := config.GetServiceName()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("service name not found"))
				Expect(serviceName).To(BeEmpty())
			})
		})
	})

	Describe("ParseServerConfig", func() {
		Context("when parsing valid YAML data", func() {
			It("should parse trpc server config successfully", func() {
				yamlData := `
server:
  app: myapp
  server: myserver
`
				config, err := trpcconf.ParseServerConfig([]byte(yamlData))

				Expect(err).NotTo(HaveOccurred())
				Expect(config).NotTo(BeNil())
				Expect(config.Server.App).To(Equal("myapp"))
				Expect(config.Server.Server).To(Equal("myserver"))
			})

			It("should handle config embedded in full trpc yaml", func() {
				yamlData := `
global:
  namespace: Production
  env_name: formal

server:
  app: helloworld
  server: Greeter
  local_ip: 127.0.0.1
`
				config, err := trpcconf.ParseServerConfig([]byte(yamlData))

				Expect(err).NotTo(HaveOccurred())
				Expect(config).NotTo(BeNil())
				Expect(config.Server.App).To(Equal("helloworld"))
				Expect(config.Server.Server).To(Equal("Greeter"))
			})
		})

		Context("when parsing empty config", func() {
			It("should handle yaml without server section", func() {
				yamlData := `
global:
  namespace: Production
`
				config, err := trpcconf.ParseServerConfig([]byte(yamlData))

				Expect(err).NotTo(HaveOccurred())
				Expect(config).NotTo(BeNil())
				Expect(config.Server.App).To(BeEmpty())
				Expect(config.Server.Server).To(BeEmpty())
			})

			It("should handle empty server section", func() {
				yamlData := `
server:
`
				config, err := trpcconf.ParseServerConfig([]byte(yamlData))

				Expect(err).NotTo(HaveOccurred())
				Expect(config).NotTo(BeNil())
				Expect(config.Server.App).To(BeEmpty())
				Expect(config.Server.Server).To(BeEmpty())
			})
		})
	})

	Describe("Galileo config details", func() {
		Context("when parsing complete galileo config", func() {
			It("should parse all nested config correctly", func() {
				yamlData := `
plugins:
  telemetry:
    galileo:
      ocp_addr: "http://ocp.example.com:8080"
      config:
        access_point: 5
        metrics_config:
          enable: true
          exporter:
            collector:
              addr: "http://metrics.example.com/api/v1/metric/write"
        traces_config:
          enable: true
          exporter:
            collector:
              addr: "http://traces.example.com:4317"
        logs_config:
          enable: false
        profiles_config:
          enable: true
          exporter:
            collector:
              addr: "http://profiles.example.com/api/v1/profile/write"
      resource:
        platform: BCS
        tenant_id: my-tenant-token
`
				telemetry, err := trpcconf.ParseTelemetryConfig([]byte(yamlData))

				Expect(err).NotTo(HaveOccurred())
				Expect(telemetry.Galileo).NotTo(BeNil())

				galileo := telemetry.Galileo
				Expect(galileo.OCPAddr).To(Equal("http://ocp.example.com:8080"))
				Expect(galileo.Config.AccessPoint).To(Equal(5))

				// Metrics config
				Expect(galileo.Config.MetricsConfig.Enable).To(BeTrue())
				Expect(galileo.Config.MetricsConfig.Exporter.Collector.Addr).To(
					Equal("http://metrics.example.com/api/v1/metric/write"))

				// Traces config
				Expect(galileo.Config.TracesConfig.Enable).To(BeTrue())
				Expect(galileo.Config.TracesConfig.Exporter.Collector.Addr).To(
					Equal("http://traces.example.com:4317"))

				// Logs config
				Expect(galileo.Config.LogsConfig.Enable).To(BeFalse())

				// Profiles config
				Expect(galileo.Config.ProfilesConfig.Enable).To(BeTrue())
				Expect(galileo.Config.ProfilesConfig.Exporter.Collector.Addr).To(
					Equal("http://profiles.example.com/api/v1/profile/write"))

				// Resource
				Expect(galileo.Resource.Platform).To(Equal("BCS"))
				Expect(galileo.Resource.TenantID).To(Equal("my-tenant-token"))
			})
		})
	})
})
