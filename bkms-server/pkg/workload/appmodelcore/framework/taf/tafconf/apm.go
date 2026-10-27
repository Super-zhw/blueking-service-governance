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

package tafconf

import (
	"github.com/pkg/errors"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime/capability/apmservicename"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/render"
)

// APMServiceName 从 TAF 配置中获取 APM 服务名称
// taf 配置使用 XML 格式，变量渲染使用 Renderer（支持 ${{env.KEY}}）
func APMServiceName(content string, appEnvVars map[string]string) (string, error) {
	rendered, err := render.New(render.SetEnvContext(appEnvVars)).Render(content)
	if err != nil {
		return "", errors.Wrapf(apmservicename.ErrConfigParse, "render taf config: %s", err)
	}

	tafConfig, err := Parse([]byte(rendered))
	if err != nil {
		return "", errors.Wrapf(apmservicename.ErrConfigParse, "parse taf config: %s", err)
	}
	return tafConfig.GetServiceName()
}
