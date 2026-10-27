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

package trpcconf

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime/capability/apmservicename"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/render"
)

// APMServiceName 从 tRPC 配置中获取 APM 服务名称
// trpc 配置使用 YAML 格式，变量渲染使用 ExpandShellVars（Shell 变量格式 ${VAR}）
// 1、解析 telemetry 配置
// 2、如果是 OpenTelemetry 类型, 直接获取服务名称，并返回给调用者
// 3、如果是 伽利略 类型, 解析 trpc 服务配置, 获取服务名称，并返回给调用者
func APMServiceName(content string, appEnvVars map[string]string) (string, error) {
	contentByte := []byte(render.RenderShellVars(content, appEnvVars))

	telemetryConfig, err := ParseTelemetryConfig(contentByte)
	if err != nil {
		return "", errors.Wrapf(apmservicename.ErrConfigParse, "parse telemetry config: %s", err)
	}
	if telemetryConfig.GetTelemetryType() == TelemetryTypeUnknown {
		return "", errors.Wrap(apmservicename.ErrConfigMissing, "telemetry config type unknown")
	}
	if telemetryConfig.IsOpenTelemetry() {
		return telemetryConfig.OpenTelemetry.GetServiceName()
	}

	trpcServerConfig, err := ParseServerConfig(contentByte)
	if err != nil {
		return "", errors.Wrapf(apmservicename.ErrConfigParse, "parse trpc server config: %s", err)
	}
	if trpcServerConfig.Server.App == "" {
		return "", errors.Wrap(apmservicename.ErrConfigMissing, "trpc server config app empty")
	}
	if trpcServerConfig.Server.Server == "" {
		return "", errors.Wrap(apmservicename.ErrConfigMissing, "trpc server config server empty")
	}

	return fmt.Sprintf("%s.%s", trpcServerConfig.Server.App, trpcServerConfig.Server.Server), nil
}
