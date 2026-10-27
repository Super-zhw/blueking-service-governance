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

// Package bkmonitor 提供蓝鲸监控相关功能
package bkmonitor

import (
	"github.com/pkg/errors"

	bkmsapp "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/app"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime/capability/apmservicename"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime/runtimemod/mods"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/workload/appmodelcore/framework/taf/tafconf"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/workload/appmodelcore/framework/trpc/trpcconf"
)

const (
	// ApmTypeCustom 自定义 APM
	// 其他 APM 类型，在 env 中(pkg/core/env/constants.go) ，与环境类型相同，因此，不再重复定义。
	ApmTypeCustom = "custom"
)

// GetApmServiceName 根据应用技术栈获取 APM 服务名称（统一入口）
func GetApmServiceName(app *bkmsapp.Application, content string, appEnvVars map[string]string) (string, error) {
	runtimeMods := mods.Of(app)
	props, err := runtimeMods.APMServiceName.Require()
	if err != nil {
		return "", errors.Wrap(err, "apm service name")
	}
	switch props.Parser {
	case apmservicename.ParserTrpcYAML:
		return trpcconf.APMServiceName(content, appEnvVars)
	case apmservicename.ParserTafXML:
		return tafconf.APMServiceName(content, appEnvVars)
	default:
		return "", errors.Errorf("unsupported apm config parser %q", props.Parser)
	}
}
