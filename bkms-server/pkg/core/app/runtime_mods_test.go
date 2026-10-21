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

package app_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	bkmsapp "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/app"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime/runtimemod"
)

var _ = Describe("RuntimeMods", func() {
	It("projects a legacy trpc app whose language only lives on TrpcSpec", func() {
		application := &bkmsapp.Application{
			Type:     bkmsapp.AppTypeTRPC,
			TrpcSpec: &bkmsapp.TrpcSpec{Language: appruntime.LanguageGo},
		}

		mods, err := application.RuntimeMods()
		Expect(err).NotTo(HaveOccurred())
		Expect(mods.Stack.Framework).To(Equal(appruntime.FrameworkTRPC))
		Expect(mods.Stack.Language).To(Equal(appruntime.LanguageGo))
		Expect(mods.DevMode.Supported).To(BeTrue())
		Expect(mods.PlatformBuild.Supported).To(BeTrue())
	})

	It("keeps devmode available for a legacy taf app that has no language at all", func() {
		application := &bkmsapp.Application{Type: bkmsapp.AppTypeTAF}

		mods, err := application.RuntimeMods()
		Expect(err).NotTo(HaveOccurred())
		Expect(mods.Stack.Language).To(BeEmpty())
		Expect(mods.DevMode.Supported).To(BeTrue())
		Expect(mods.DevMode.Props.ScriptSet).To(Equal(runtimemod.DevModeScriptSetTAF))
		Expect(mods.PlatformBuild.ReasonCode).To(Equal(runtimemod.ReasonUnsupportedStack))
	})

	It("returns an error for a stack that contradicts itself", func() {
		application := &bkmsapp.Application{
			Type:     bkmsapp.AppTypeTRPC,
			Language: appruntime.LanguageCpp,
			TrpcSpec: &bkmsapp.TrpcSpec{Language: appruntime.LanguageGo},
		}

		_, err := application.RuntimeMods()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(appruntime.IssueLanguageConflict))
	})
})
