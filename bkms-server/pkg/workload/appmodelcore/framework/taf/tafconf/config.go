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

// Package tafconf 解析 TAF 配置文件（标签加 key=value）：APM 服务名等。
//
// 只依赖能力声明与通用渲染工具，供能力包和执行端引用。
package tafconf

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime/capability/apmservicename"
)

// Config TAF 配置文件根结构（XML 格式）
// TAF 配置使用非标准 XML 格式，在 XML 标签内使用 key=value 格式存储配置项
type Config struct {
	XMLName     xml.Name    `xml:"taf"`
	Application Application `xml:"application"`
}

// Application TAF application 节点
type Application struct {
	// CharData 包含 setdivision 等 key=value 格式的配置项
	CharData string `xml:",chardata"`
	Server   Server `xml:"server"`
}

// Server TAF server 节点
type Server struct {
	// CharData 包含 app、server 等 key=value 格式的配置项
	CharData string `xml:",chardata"`
}

// Parse 从 XML 字节解析 TAF 配置
func Parse(data []byte) (*Config, error) {
	var config Config
	if err := xml.Unmarshal(data, &config); err != nil {
		return nil, errors.Wrap(err, "unmarshal taf config")
	}
	return &config, nil
}

// GetServiceName 从 TAF 配置中获取 APM 服务名称
// 拼接规则：{app}.{server}.{setdivision去除点号保留星号}
// setdivision 为可选字段，为空时拼接结果为 {app}.{server}
func (c *Config) GetServiceName() (string, error) {
	serverKV := parseKeyValues(c.Application.Server.CharData)
	app := serverKV["app"]
	server := serverKV["server"]

	if app == "" || server == "" {
		return "", errors.Wrap(apmservicename.ErrConfigMissing, "taf config missing app or server field")
	}

	appKV := parseKeyValues(c.Application.CharData)
	setDivision := appKV["setdivision"]

	if setDivision == "" {
		return fmt.Sprintf("%s.%s", app, server), nil
	}

	// 去除点号，保留星号
	setDivision = strings.ReplaceAll(setDivision, ".", "")
	return fmt.Sprintf("%s.%s.%s", app, server, setDivision), nil
}

// parseKeyValues 解析 TAF 配置中 key=value 格式的文本内容
func parseKeyValues(text string) map[string]string {
	result := make(map[string]string)
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return result
}
