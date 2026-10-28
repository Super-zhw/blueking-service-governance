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

package trpc

import (
	"github.com/pkg/errors"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime"
)

// Config tRPC 的 frameworkConfig。文件名与挂载目录已收敛到框架配置文件 def，目前没有扩展字段。
type Config struct{}

// DecodeConfig 严格解码 tRPC 的 frameworkConfig，拒绝未知 key 与不匹配的版本。
func DecodeConfig(cfg map[string]any, version int) (Config, error) {
	dest, err := appruntime.DecodeStrict[Config](cfg, version, appruntime.FrameworkConfigVersionV1)
	if err != nil {
		return Config{}, errors.Wrap(err, "tRPC")
	}
	return dest, nil
}
