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

package appcfgfiledef

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	bkmsapp "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/app"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/app/appcfg"
)

var _ = Describe("validateConfigKindForApp", func() {
	DescribeTable("allows or rejects config kinds by app",
		func(app *bkmsapp.Application, kind appcfg.ConfigKind, errSubstring string) {
			err := validateConfigKindForApp(app, kind)
			if errSubstring == "" {
				Expect(err).NotTo(HaveOccurred())
				return
			}
			Expect(err).To(MatchError(ContainSubstring(errSubstring)))
		},
		Entry("plain for trpc", &bkmsapp.Application{Type: bkmsapp.AppTypeTRPC}, appcfg.ConfigKindPlain, ""),
		Entry("plain for blank",
			&bkmsapp.Application{Type: bkmsapp.AppTypeBKMSApp, Framework: "blank", Language: "python"},
			appcfg.ConfigKindPlain, ""),
		Entry("plain for helm", &bkmsapp.Application{Type: bkmsapp.AppTypeHelm}, appcfg.ConfigKindPlain,
			"only supported for appmodel applications"),
		Entry("framework for helm", &bkmsapp.Application{Type: bkmsapp.AppTypeHelm}, appcfg.ConfigKindFramework, ""),
		Entry("framework for legacy trpc", &bkmsapp.Application{Type: bkmsapp.AppTypeTRPC},
			appcfg.ConfigKindFramework, ""),
		Entry("framework for legacy taf", &bkmsapp.Application{Type: bkmsapp.AppTypeTAF},
			appcfg.ConfigKindFramework, ""),
		Entry("framework for bkmsApp trpc",
			&bkmsapp.Application{Type: bkmsapp.AppTypeBKMSApp, Framework: "trpc", Language: "go"},
			appcfg.ConfigKindFramework, ""),
		Entry("framework for blank",
			&bkmsapp.Application{Type: bkmsapp.AppTypeBKMSApp, Framework: "blank", Language: "go"},
			appcfg.ConfigKindFramework, `framework "blank" does not support framework config file`),
	)
})
