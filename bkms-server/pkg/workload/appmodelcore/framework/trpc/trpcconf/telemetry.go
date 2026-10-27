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

// Package trpcconf 解析 tRPC 配置文件（YAML）：APM 服务名、服务列表等。
//
// 只依赖能力声明与通用渲染工具，供能力包和执行端引用。
package trpcconf

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime/capability/apmservicename"
)

const (
	// serviceNameKey 服务名称
	// 蓝鲸监控规范，定义的服务名称标识符
	serviceNameKey = "service_name"
)

// TelemetryType 上报类型
type TelemetryType string

const (
	// TelemetryTypeOpenTelemetry OpenTelemetry
	TelemetryTypeOpenTelemetry TelemetryType = "opentelemetry"

	// TelemetryTypeGalileo Galileo
	TelemetryTypeGalileo TelemetryType = "galileo"

	// TelemetryTypeUnknown 未知
	TelemetryTypeUnknown TelemetryType = "unknown"
)

// ServerConfig tRPC 配置中的 server 段
type ServerConfig struct {
	Server struct {
		App string `yaml:"app"`

		Server string `yaml:"server"`
	} `yaml:"server"`
}

// PluginsConfig 插件配置
type PluginsConfig struct {
	Plugins Plugins `yaml:"plugins"`
}

// Plugins 插件列表
type Plugins struct {
	Telemetry Telemetry `yaml:"telemetry"`
}

// Telemetry 遥测配置
// opentelemetry 和 galileo 二选一
type Telemetry struct {
	OpenTelemetry *OpenTelemetryConfig `yaml:"opentelemetry,omitempty"`

	Galileo *GalileoConfig `yaml:"galileo,omitempty"`
}

// GetTelemetryType 获取遥测上报类型
func (t *Telemetry) GetTelemetryType() TelemetryType {
	// 优先判断 galileo（因为 galileo 配置更完整时优先使用）
	if t.Galileo != nil && t.Galileo.OCPAddr != "" {
		return TelemetryTypeGalileo
	}
	// 判断 opentelemetry
	if t.OpenTelemetry != nil && len(t.OpenTelemetry.Attributes) > 0 && t.OpenTelemetry.Addr != "" &&
		t.OpenTelemetry.TenantID != "" {
		return TelemetryTypeOpenTelemetry
	}
	// 如果 galileo 有配置但 ocp_addr 为空，仍然认为是 galileo 类型
	if t.Galileo != nil {
		return TelemetryTypeGalileo
	}
	// 如果 opentelemetry 有配置
	if t.OpenTelemetry != nil {
		return TelemetryTypeOpenTelemetry
	}
	return TelemetryTypeUnknown
}

// IsOpenTelemetry 是否为 OpenTelemetry 类型
func (t *Telemetry) IsOpenTelemetry() bool {
	return t.GetTelemetryType() == TelemetryTypeOpenTelemetry
}

// IsGalileo 是否为 Galileo 类型
func (t *Telemetry) IsGalileo() bool {
	return t.GetTelemetryType() == TelemetryTypeGalileo
}

// ========== OpenTelemetry 配置 ==========

// OpenTelemetryConfig OpenTelemetry 配置
type OpenTelemetryConfig struct {
	Addr       string      `yaml:"addr"`
	TenantID   string      `yaml:"tenant_id"`
	Attributes []Attribute `yaml:"attributes,omitempty"`
}

// Attribute 属性键值对
type Attribute struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

// ========== Galileo 配置 ==========

// GalileoConfig Galileo 配置
type GalileoConfig struct {
	OCPAddr  string          `yaml:"ocp_addr"`
	Config   GalileoDetail   `yaml:"config"`
	Resource GalileoResource `yaml:"resource"`
}

// GalileoDetail Galileo 详细配置
type GalileoDetail struct {
	AccessPoint    int            `yaml:"access_point"`
	MetricsConfig  MetricsConfig  `yaml:"metrics_config"`
	TracesConfig   TracesConfig   `yaml:"traces_config"`
	LogsConfig     LogsConfig     `yaml:"logs_config"`
	ProfilesConfig ProfilesConfig `yaml:"profiles_config"`
}

// MetricsConfig 指标配置
type MetricsConfig struct {
	Enable   bool            `yaml:"enable"`
	Exporter MetricsExporter `yaml:"exporter,omitempty"`
}

// MetricsExporter 指标导出器
type MetricsExporter struct {
	Collector CollectorConfig `yaml:"collector"`
}

// TracesConfig 链路追踪配置
type TracesConfig struct {
	Enable   bool           `yaml:"enable"`
	Exporter TracesExporter `yaml:"exporter,omitempty"`
}

// TracesExporter 链路导出器
type TracesExporter struct {
	Collector CollectorConfig `yaml:"collector"`
}

// LogsConfig 日志配置
type LogsConfig struct {
	Enable bool `yaml:"enable"`
}

// ProfilesConfig 性能分析配置
type ProfilesConfig struct {
	Enable   bool             `yaml:"enable"`
	Exporter ProfilesExporter `yaml:"exporter,omitempty"`
}

// ProfilesExporter 性能分析导出器
type ProfilesExporter struct {
	Collector CollectorConfig `yaml:"collector"`
}

// CollectorConfig 收集器配置
type CollectorConfig struct {
	Addr string `yaml:"addr"`
}

// GalileoResource Galileo 资源配置
type GalileoResource struct {
	Platform string `yaml:"platform"`
	TenantID string `yaml:"tenant_id"`
}

// ========== 解析方法 ==========

// ParseTelemetryConfig 从 YAML 字节解析配置
func ParseTelemetryConfig(data []byte) (*Telemetry, error) {
	var config PluginsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.Wrap(err, "unmarshal telemetry config")
	}
	return &config.Plugins.Telemetry, nil
}

// ParseServerConfig 从 YAML 字节解析配置
func ParseServerConfig(data []byte) (*ServerConfig, error) {
	var config ServerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.Wrapf(err, "unmarshal trpc server config")
	}
	return &config, nil
}

// GetServiceName 获取服务名称
func (s *OpenTelemetryConfig) GetServiceName() (string, error) {
	for _, attribute := range s.Attributes {
		if attribute.Key == serviceNameKey {
			return attribute.Value, nil
		}
	}

	return "", errors.Wrapf(apmservicename.ErrConfigMissing, "%s service name not found", TelemetryTypeOpenTelemetry)
}
