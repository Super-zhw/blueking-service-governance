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

// SchemaVersion 对外视图的结构版本。只在响应结构发生破坏性变更时调整，支持矩阵本身的变化
// 不需要升版本。
const SchemaVersion = 1

// ModSet 单个应用解析得到的能力集合。
type ModSet struct {
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

// Resolve 将规范化后的技术栈解析为能力集合。
//
// 技术栈非法返回 error；能力不支持不是错误。调用方进入执行器时使用 Mod.Require，
// 届时才把"不支持"转换为 ErrModUnsupported。
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

// validate 校验技术栈本身。与 appruntime.Normalize 不同，它允许 language 为空：存量 TAF 记录
// 从未保存过 language，部分 tRPC 记录只在 workload 上有 language。
// BRIDGE-ONLY：阶段 5 回填 language 之后，改用 appruntime.ValidateStack。
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

// Declarations 按声明顺序返回全部能力声明。测试用它校验声明是否覆盖框架目录；
// 后续 catalog API 用它发布框架支持矩阵，避免再维护一份规则副本。
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

// all 按声明顺序返回全部能力。完整性测试会断言它与 ModSet 字段保持一致，
// 防止新增能力在对外投影中被遗漏。
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
