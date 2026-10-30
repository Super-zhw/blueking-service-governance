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

package lifecycle

import (
	"context"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/common/utils/datatype"
	bkmsapp "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/app"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/store"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/workload/appmodelcore/appmodel"
)

// UpdateStartupConfigParams 修改 bkmsApp 应用启动配置的参数。
type UpdateStartupConfigParams struct {
	// Command 容器启动命令，整体替换。
	Command []string
	// Args 容器启动参数，整体替换。
	Args []string
	// ShCommand 容器启动脚本，整体替换。
	ShCommand string
	// FrameworkConfig 框架扩展配置。为 nil 时不修改，否则整体替换并重新校验。
	FrameworkConfig map[string]any
}

// UpdateStartupConfig 修改 bkmsApp 应用的启动配置：启动命令、参数、脚本与 frameworkConfig，不修改框架与语言。
// 返回修改前后的 AppModel，供调用方审计。
func (s *Service) UpdateStartupConfig(
	ctx context.Context,
	app *bkmsapp.Application,
	params *UpdateStartupConfigParams,
) (*appmodel.AppModel, *appmodel.AppModel, error) {
	if err := ValidateStack(app); err != nil {
		return nil, nil, err
	}
	if params.FrameworkConfig != nil {
		if err := validateFrameworkConfig(app.Framework, params.FrameworkConfig); err != nil {
			return nil, nil, err
		}
	}

	appModel, err := store.Of(s.db, appmodel.AppModelMeta).Get(ctx, appmodel.ByApp(app.ID))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "get model of app %s", app.ID)
	}
	var oldAppModel appmodel.AppModel
	if err = copier.CopyWithOption(&oldAppModel, appModel, copier.Option{DeepCopy: true}); err != nil {
		return nil, nil, errors.Wrapf(err, "copy model of app %s", app.ID)
	}

	appModel.Workload.Command = params.Command
	appModel.Workload.Args = params.Args
	appModel.Workload.ShCommand = params.ShCommand
	if params.FrameworkConfig != nil {
		appModel.Workload.FrameworkConfig = datatype.CloneMap(params.FrameworkConfig)
		appModel.Workload.FrameworkConfigVersion = appruntime.FrameworkConfigVersionV1
	}
	if err = appmodel.UpdateAppModel(ctx, s.db, appModel); err != nil {
		return nil, nil, errors.Wrapf(err, "update model of app %s", app.ID)
	}
	return &oldAppModel, appModel, nil
}
