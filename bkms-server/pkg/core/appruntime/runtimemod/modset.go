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

package runtimemod

import (
	"slices"

	"github.com/pkg/errors"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime"
)

// SchemaVersion 对外结构的版本号，只在结构发生不兼容变更时递增。
const SchemaVersion = 1

// ModSet 应用的全部能力。
//
// 新增能力时，需同步修改本结构体、Resolve、Declarations、all，测试会检查四处是否一致。
type ModSet struct {
	// Stack 查表所用的技术栈。
	Stack appruntime.Stack

	FrameworkFile         Mod[FrameworkFileProps]
	PlainConfigFiles      Mod[PlainConfigFilesProps]
	DevMode               Mod[DevModeProps]
	PlatformBuild         Mod[PlatformBuildProps]
	AdminCommand          Mod[AdminCommandProps]
	APMServiceName        Mod[APMServiceNameProps]
	PolarisServiceCheck   Mod[PolarisServiceCheckProps]
	PolarisFrameworkPatch Mod[PolarisFrameworkPatchProps]
	BSCPWorkloadMeta      Mod[BSCPWorkloadMetaProps]
	AutoDeploy            Mod[AutoDeployProps]
}

// Resolve 按技术栈得出全部能力。
//
// 只有技术栈本身非法时才返回 error；能力不支持不算错误，体现在各 Mod 的 Supported 上。
func Resolve(stack appruntime.Stack) (ModSet, error) {
	if err := validate(stack); err != nil {
		return ModSet{}, errors.Wrap(err, "resolve runtime mods")
	}
	return ModSet{
		Stack:                 stack,
		FrameworkFile:         FrameworkFileDef.Resolve(stack),
		PlainConfigFiles:      PlainConfigFilesDef.Resolve(stack),
		DevMode:               DevModeDef.Resolve(stack),
		PlatformBuild:         PlatformBuildDef.Resolve(stack),
		AdminCommand:          AdminCommandDef.Resolve(stack),
		APMServiceName:        APMServiceNameDef.Resolve(stack),
		PolarisServiceCheck:   PolarisServiceCheckDef.Resolve(stack),
		PolarisFrameworkPatch: PolarisFrameworkPatchDef.Resolve(stack),
		BSCPWorkloadMeta:      BSCPWorkloadMetaDef.Resolve(stack),
		AutoDeploy:            AutoDeployDef.Resolve(stack),
	}, nil
}

// validate 校验技术栈。与 appruntime.ValidateStack 的区别是允许 bkmsApp 不填语言，
// 原因同 appruntime.CapabilityStack。
// BRIDGE-ONLY：存量语言回填后，改用 appruntime.ValidateStack。
func validate(stack appruntime.Stack) error {
	switch stack.Type {
	case appruntime.AppTypeHelm, appruntime.AppTypeAgones:
		if stack.Framework != "" || stack.Language != "" {
			return errors.Wrapf(appruntime.ErrUnexpectedStackFields, "app type %q", stack.Type)
		}
		return nil
	case appruntime.AppTypeBKMSApp:
		def, ok := appruntime.LookupFramework(stack.Framework)
		if !ok {
			return errors.Wrapf(appruntime.ErrUnknownFramework, "framework %q", stack.Framework)
		}
		if stack.Language == "" {
			return nil
		}
		if !appruntime.KnownLanguages.Contains(stack.Language) {
			return errors.Wrapf(appruntime.ErrUnknownLanguage, "language %q", stack.Language)
		}
		if !slices.Contains(def.SupportedLanguages, stack.Language) {
			return errors.Wrapf(
				appruntime.ErrInvalidCombination,
				"framework %q does not support language %q", stack.Framework, stack.Language,
			)
		}
		return nil
	default:
		return errors.Wrapf(appruntime.ErrUnknownAppType, "app type %q", stack.Type)
	}
}

// Declarations 返回全部能力的声明摘要。
func Declarations() []DefMeta {
	return []DefMeta{
		FrameworkFileDef.Meta(),
		PlainConfigFilesDef.Meta(),
		DevModeDef.Meta(),
		PlatformBuildDef.Meta(),
		AdminCommandDef.Meta(),
		APMServiceNameDef.Meta(),
		PolarisServiceCheckDef.Meta(),
		PolarisFrameworkPatchDef.Meta(),
		BSCPWorkloadMetaDef.Meta(),
		AutoDeployDef.Meta(),
	}
}

// all 以能力 ID 为 key 返回全部能力，供 Public 使用。
func (s ModSet) all() map[ID]view {
	return map[ID]view{
		s.FrameworkFile.ID:         s.FrameworkFile,
		s.PlainConfigFiles.ID:      s.PlainConfigFiles,
		s.DevMode.ID:               s.DevMode,
		s.PlatformBuild.ID:         s.PlatformBuild,
		s.AdminCommand.ID:          s.AdminCommand,
		s.APMServiceName.ID:        s.APMServiceName,
		s.PolarisServiceCheck.ID:   s.PolarisServiceCheck,
		s.PolarisFrameworkPatch.ID: s.PolarisFrameworkPatch,
		s.BSCPWorkloadMeta.ID:      s.BSCPWorkloadMeta,
		s.AutoDeploy.ID:            s.AutoDeploy,
	}
}
