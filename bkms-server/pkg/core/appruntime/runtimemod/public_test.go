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
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime/runtimemod"
)

// toJSON 序列化为 JSON，golden 比对的是对外输出格式
func toJSON(view any) string {
	raw, err := json.Marshal(view)
	Expect(err).NotTo(HaveOccurred())
	return string(raw)
}

var _ = Describe("Public", func() {
	It("renders trpc+go with every capability supported", func() {
		Expect(toJSON(mustResolve(bkmsStack(appruntime.FrameworkTRPC, appruntime.LanguageGo)).Public())).
			To(MatchJSON(`{
				"schemaVersion": 1,
				"type": "bkmsApp",
				"framework": "trpc",
				"language": "go",
				"mods": {
					"frameworkFile":       {"supported": true, "props": {"format": "yaml"}},
					"plainConfigFiles":    {"supported": true},
					"devMode":             {"supported": true, "props": {
						"workPath": "/data/bkms/dev-mode/trpc",
						"mountPath": "/data/bkms/dev-mode/trpc/configmap-scripts"
					}},
					"platformBuild":       {"supported": true},
					"adminCommand":        {"supported": true},
					"apmServiceName":      {"supported": true},
					"polarisServiceCheck": {"supported": true},
					"polarisFrameworkPatch": {"supported": true},
					"bscpWorkloadMeta":    {"supported": true},
					"autoDeploy":          {"supported": true}
				}
			}`))
	})

	It("renders trpc+cpp with platformBuild unsupported", func() {
		Expect(toJSON(mustResolve(bkmsStack(appruntime.FrameworkTRPC, appruntime.LanguageCpp)).Public())).
			To(MatchJSON(`{
				"schemaVersion": 1,
				"type": "bkmsApp",
				"framework": "trpc",
				"language": "cpp",
				"mods": {
					"frameworkFile":  {"supported": true, "props": {"format": "yaml"}},
					"plainConfigFiles": {"supported": true},
					"devMode":        {"supported": true, "props": {
						"workPath": "/data/bkms/dev-mode/trpc",
						"mountPath": "/data/bkms/dev-mode/trpc/configmap-scripts"
					}},
					"platformBuild":  {"supported": false, "reasonCode": "unsupported_stack"},
					"adminCommand":   {"supported": true},
					"apmServiceName": {"supported": true},
					"polarisServiceCheck": {"supported": true},
					"polarisFrameworkPatch": {"supported": true},
					"bscpWorkloadMeta":    {"supported": true},
					"autoDeploy":          {"supported": true}
				}
			}`))
	})

	It("renders a taf row that never recorded a language without the language key", func() {
		Expect(toJSON(mustResolve(bkmsStack(appruntime.FrameworkTAF, "")).Public())).
			To(MatchJSON(`{
				"schemaVersion": 1,
				"type": "bkmsApp",
				"framework": "taf",
				"mods": {
					"frameworkFile":       {"supported": true, "props": {"format": "xml"}},
					"plainConfigFiles":    {"supported": true},
					"devMode":             {"supported": true, "props": {
						"workPath": "/data/bkms/dev-mode/taf",
						"mountPath": "/data/bkms/dev-mode/taf/configmap-scripts"
					}},
					"platformBuild":       {"supported": false, "reasonCode": "unsupported_stack"},
					"adminCommand":        {"supported": true},
					"apmServiceName":      {"supported": true},
					"polarisServiceCheck": {"supported": false, "reasonCode": "unsupported_stack"},
					"polarisFrameworkPatch": {"supported": false, "reasonCode": "unsupported_stack"},
					"bscpWorkloadMeta":    {"supported": false, "reasonCode": "unsupported_stack"},
					"autoDeploy":          {"supported": true}
				}
			}`))
	})

	It("renders blank with framework capabilities unsupported and plain files supported", func() {
		Expect(toJSON(mustResolve(bkmsStack(appruntime.FrameworkBlank, appruntime.LanguageGo)).Public())).
			To(MatchJSON(`{
				"schemaVersion": 1,
				"type": "bkmsApp",
				"framework": "blank",
				"language": "go",
				"mods": {
					"frameworkFile":       {"supported": false, "reasonCode": "unsupported_stack"},
					"plainConfigFiles":    {"supported": true},
					"devMode":             {"supported": false, "reasonCode": "unsupported_stack"},
					"platformBuild":       {"supported": false, "reasonCode": "unsupported_stack"},
					"adminCommand":        {"supported": false, "reasonCode": "unsupported_stack"},
					"apmServiceName":      {"supported": false, "reasonCode": "unsupported_stack"},
					"polarisServiceCheck": {"supported": false, "reasonCode": "unsupported_stack"},
					"polarisFrameworkPatch": {"supported": false, "reasonCode": "unsupported_stack"},
					"bscpWorkloadMeta":    {"supported": false, "reasonCode": "unsupported_stack"},
					"autoDeploy":          {"supported": false, "reasonCode": "unsupported_stack"}
				}
			}`))
	})

	It("renders helm with no stack fields and an app type reason", func() {
		Expect(toJSON(mustResolve(appruntime.Stack{Type: appruntime.AppTypeHelm}).Public())).
			To(MatchJSON(`{
				"schemaVersion": 1,
				"type": "helm",
				"mods": {
					"frameworkFile":       {"supported": false, "reasonCode": "unsupported_app_type"},
					"plainConfigFiles":    {"supported": false, "reasonCode": "unsupported_app_type"},
					"devMode":             {"supported": false, "reasonCode": "unsupported_app_type"},
					"platformBuild":       {"supported": false, "reasonCode": "unsupported_app_type"},
					"adminCommand":        {"supported": false, "reasonCode": "unsupported_app_type"},
					"apmServiceName":      {"supported": false, "reasonCode": "unsupported_app_type"},
					"polarisServiceCheck": {"supported": false, "reasonCode": "unsupported_app_type"},
					"polarisFrameworkPatch": {"supported": false, "reasonCode": "unsupported_app_type"},
					"bscpWorkloadMeta":    {"supported": false, "reasonCode": "unsupported_app_type"},
					"autoDeploy":          {"supported": false, "reasonCode": "unsupported_app_type"}
				}
			}`))
	})
})

var _ = Describe("Matrix", func() {
	It("covers every legal catalog combination in stable order", func() {
		matrix, err := runtimemod.Matrix()
		Expect(err).NotTo(HaveOccurred())

		combinations := make([]string, 0, len(matrix))
		for _, view := range matrix {
			combinations = append(combinations, view.Framework+"+"+view.Language)
		}
		Expect(combinations).To(Equal([]string{
			"blank+cpp", "blank+go", "taf+cpp", "trpc+cpp", "trpc+go",
		}))
	})

	It("agrees with resolving each combination individually", func() {
		matrix, err := runtimemod.Matrix()
		Expect(err).NotTo(HaveOccurred())
		for _, view := range matrix {
			single := mustResolve(bkmsStack(view.Framework, view.Language)).Public()
			Expect(view).To(Equal(single))
		}
	})
})
