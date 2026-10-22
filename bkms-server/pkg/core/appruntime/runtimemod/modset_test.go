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

package runtimemod_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime/runtimemod"
)

func bkmsStack(framework, language string) appruntime.Stack {
	return appruntime.Stack{
		Type: appruntime.AppTypeBKMSApp, Framework: framework, Language: language,
	}
}

func mustResolve(stack appruntime.Stack) runtimemod.ModSet {
	mods, err := runtimemod.Resolve(stack)
	Expect(err).NotTo(HaveOccurred())
	return mods
}

var _ = Describe("Resolve", func() {
	DescribeTable("rejects an invalid stack",
		func(stack appruntime.Stack, substring string) {
			_, err := runtimemod.Resolve(stack)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(substring))
		},
		Entry("unknown framework", bkmsStack("quic", appruntime.LanguageGo), "unknown framework"),
		Entry("taf with go", bkmsStack(appruntime.FrameworkTAF, appruntime.LanguageGo), "does not support language"),
		Entry("unknown language", bkmsStack(appruntime.FrameworkTRPC, "java"), "unknown language"),
		Entry("legacy type", appruntime.Stack{Type: appruntime.AppTypeTRPC}, "unknown app type"),
		Entry("helm with framework", appruntime.Stack{
			Type: appruntime.AppTypeHelm, Framework: appruntime.FrameworkTRPC,
		}, "must not carry"),
	)

	It("tolerates a missing language so legacy rows keep their capabilities", func() {
		mods := mustResolve(bkmsStack(appruntime.FrameworkTAF, ""))
		Expect(mods.DevMode.Supported).To(BeTrue())
		Expect(mods.DevMode.Props.ScriptSet).To(Equal(runtimemod.DevModeScriptSetTAF))
		Expect(mods.DevMode.Props.WorkPath).To(Equal(runtimemod.DevModeTafWorkPath))
	})

	It("reports every capability as unsupported for helm", func() {
		mods := mustResolve(appruntime.Stack{Type: appruntime.AppTypeHelm})
		Expect(mods.FrameworkFile.ReasonCode).To(Equal(runtimemod.ReasonUnsupportedAppType))
		Expect(mods.DevMode.ReasonCode).To(Equal(runtimemod.ReasonUnsupportedAppType))
		Expect(mods.AutoDeploy.ReasonCode).To(Equal(runtimemod.ReasonUnsupportedAppType))
	})

	It("supports devmode but not platformBuild for trpc+cpp", func() {
		mods := mustResolve(bkmsStack(appruntime.FrameworkTRPC, appruntime.LanguageCpp))
		Expect(mods.DevMode.Supported).To(BeTrue())
		Expect(mods.DevMode.Props.ScriptSet).To(Equal(runtimemod.DevModeScriptSetTRPC))
		Expect(mods.PlatformBuild.Supported).To(BeFalse())
		Expect(mods.PlatformBuild.ReasonCode).To(Equal(runtimemod.ReasonUnsupportedStack))
	})

	It("supports plain config files for every bkmsApp framework, even without a language", func() {
		for _, framework := range []string{
			appruntime.FrameworkTRPC, appruntime.FrameworkTAF, appruntime.FrameworkBlank,
		} {
			Expect(mustResolve(bkmsStack(framework, "")).PlainConfigFiles.Supported).To(BeTrue(), framework)
		}
		Expect(mustResolve(appruntime.Stack{Type: appruntime.AppTypeHelm}).PlainConfigFiles.ReasonCode).
			To(Equal(runtimemod.ReasonUnsupportedAppType))
	})

	It("treats blank as a framework with no framework capabilities", func() {
		mods := mustResolve(bkmsStack(appruntime.FrameworkBlank, appruntime.LanguageGo))
		Expect(mods.FrameworkFile.Supported).To(BeFalse())
		Expect(mods.DevMode.Supported).To(BeFalse())
		Expect(mods.AutoDeploy.Supported).To(BeFalse())
		Expect(mods.FrameworkFile.ReasonCode).To(Equal(runtimemod.ReasonUnsupportedStack))
	})

	It("distinguishes a missing language from an unsupported stack", func() {
		trpcNoLanguage := mustResolve(bkmsStack(appruntime.FrameworkTRPC, ""))
		Expect(trpcNoLanguage.PlatformBuild.Supported).To(BeFalse())
		Expect(trpcNoLanguage.PlatformBuild.ReasonCode).To(Equal(runtimemod.ReasonIncompleteLanguage))
		Expect(trpcNoLanguage.AdminCommand.ReasonCode).To(Equal(runtimemod.ReasonIncompleteLanguage))

		taf := mustResolve(bkmsStack(appruntime.FrameworkTAF, appruntime.LanguageCpp))
		Expect(taf.PlatformBuild.ReasonCode).To(Equal(runtimemod.ReasonUnsupportedStack))
	})

	It("selects the admin config layout per language", func() {
		trpcGo := mustResolve(bkmsStack(appruntime.FrameworkTRPC, appruntime.LanguageGo))
		Expect(trpcGo.AdminCommand.Props.ConfigLayout).To(Equal(runtimemod.TrpcAdminLayoutNested))

		trpcCpp := mustResolve(bkmsStack(appruntime.FrameworkTRPC, appruntime.LanguageCpp))
		Expect(trpcCpp.AdminCommand.Props.ConfigLayout).To(Equal(runtimemod.TrpcAdminLayoutFlat))

		taf := mustResolve(bkmsStack(appruntime.FrameworkTAF, appruntime.LanguageCpp))
		Expect(taf.AdminCommand.Props.Protocol).To(Equal(runtimemod.AdminProtocolTAF))
	})

	It("pre-registers admin layouts for out-of-catalog languages without opening them", func() {
		pairs := runtimemod.AdminCommandDef.Pair
		Expect(pairs[runtimemod.Pair{Framework: appruntime.FrameworkTRPC, Language: appruntime.LanguagePython}].
			ConfigLayout).To(Equal(runtimemod.TrpcAdminLayoutNested))
		Expect(pairs[runtimemod.Pair{Framework: appruntime.FrameworkTRPC, Language: appruntime.LanguageNode}].
			ConfigLayout).To(Equal(runtimemod.TrpcAdminLayoutFlat))
		Expect(pairs[runtimemod.Pair{Framework: appruntime.FrameworkTRPC, Language: appruntime.LanguageJava}].
			ConfigLayout).To(Equal(runtimemod.TrpcAdminLayoutNestedPrefixed))

		// 目录外语言仍会被技术栈校验拒绝
		for _, language := range []string{
			appruntime.LanguageJava, appruntime.LanguageNode, appruntime.LanguagePython,
		} {
			_, err := runtimemod.Resolve(bkmsStack(appruntime.FrameworkTRPC, language))
			Expect(err).To(MatchError(appruntime.ErrUnknownLanguage), language)
		}
	})

	It("does not let a caller mutate later results through returned props", func() {
		first := mustResolve(bkmsStack(appruntime.FrameworkTRPC, appruntime.LanguageGo))
		first.DevMode.Props.WorkPath = "/mutated"

		second := mustResolve(bkmsStack(appruntime.FrameworkTRPC, appruntime.LanguageGo))
		Expect(second.DevMode.Props.WorkPath).To(Equal(runtimemod.DevModeTrpcWorkPath))
	})
})

var _ = Describe("Require", func() {
	It("returns props when supported", func() {
		mods := mustResolve(bkmsStack(appruntime.FrameworkTRPC, appruntime.LanguageGo))
		props, err := mods.PlatformBuild.Require()
		Expect(err).NotTo(HaveOccurred())
		Expect(props.DockerfileLanguage).To(Equal(appruntime.LanguageGo))
	})

	It("returns ErrModUnsupported naming the capability when unsupported", func() {
		mods := mustResolve(bkmsStack(appruntime.FrameworkTRPC, appruntime.LanguageCpp))
		_, err := mods.PlatformBuild.Require()
		Expect(err).To(MatchError(runtimemod.ErrModUnsupported))
		Expect(err.Error()).To(ContainSubstring("platformBuild"))
		Expect(err.Error()).To(ContainSubstring("unsupported_stack"))
	})
})
