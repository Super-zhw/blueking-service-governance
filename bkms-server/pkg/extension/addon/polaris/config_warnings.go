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

package polaris

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"

	log "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/common/logging"
	bkmsapp "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/app"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/app/appcfg"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime/runtimemod/mods"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/store"
)

// CollectConfigWarnings 为单个 PolarisConfig 收集服务名校验 warnings。
// immediate 模式不注入 tRPC registry 配置，不校验；其余由能力 polaris.ServiceCheck 判断是否校验并执行。
func CollectConfigWarnings(
	ctx context.Context,
	db *mongo.Database,
	cfgProvider appcfg.MountableFileProvider,
	config *PolarisConfig,
) []string {
	if config.IsImmediateRegister() {
		return nil
	}

	app, err := store.Of(db, bkmsapp.AppMeta).Get(ctx, bkmsapp.ByID(config.AppID))
	if err != nil {
		log.Warnf(ctx, "get app for polaris validation failed: app=%s, err=%v", config.AppID, err)
		return nil
	}

	return checkServiceName(
		ctx,
		cfgProvider,
		mods.Of(app).PolarisServiceCheck,
		serviceCheckTarget{
			AppID:       config.AppID,
			ConfigName:  config.Name,
			PolarisName: config.PolarisName,
			EnvNames:    config.ScopeEnvNames,
		},
	)
}
