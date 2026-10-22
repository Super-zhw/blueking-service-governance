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
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime/runtimemod"
)

var outOfCatalogLanguages = []string{
	appruntime.LanguageJava, appruntime.LanguageNode, appruntime.LanguagePython,
}

var _ = Describe("Declarations", func() {
	It("declares every capability exactly once", func() {
		declarations := runtimemod.Declarations()
		ids := lo.Map(declarations, func(d runtimemod.DefMeta, _ int) runtimemod.ID { return d.ID })
		Expect(lo.Uniq(ids)).To(HaveLen(len(ids)))

		// 新增能力若漏加到 all()，对外输出中会缺少该能力
		mods := mustResolve(bkmsStack(appruntime.FrameworkTRPC, appruntime.LanguageGo)).Public()
		Expect(lo.Keys(mods.Mods)).To(ConsistOf(ids))
	})

	It("only references frameworks and languages from the catalog", func() {
		frameworks := lo.SliceToMap(
			appruntime.Catalog(),
			func(def appruntime.FrameworkDef) (string, appruntime.FrameworkDef) {
				return def.Name, def
			},
		)

		for _, declaration := range runtimemod.Declarations() {
			for _, framework := range declaration.Frameworks {
				Expect(frameworks).To(HaveKey(framework), "unknown framework %q in %s", framework, declaration.ID)
			}
			for _, language := range declaration.Languages {
				// 目录外语言只允许出现在 adminCommand 中
				if lo.Contains(outOfCatalogLanguages, language) {
					Expect(declaration.ID).To(Equal(runtimemod.AdminCommandDef.ID),
						"out-of-catalog language %q in %s", language, declaration.ID)
					continue
				}
				Expect(appruntime.KnownLanguages.Contains(language)).
					To(BeTrue(), "unknown language %q in %s", language, declaration.ID)
			}
		}
	})

	It("documents each declaration with a dimension", func() {
		for _, declaration := range runtimemod.Declarations() {
			switch declaration.Dimension {
			case runtimemod.ByFramework:
				Expect(declaration.Frameworks).NotTo(BeEmpty(), "%s has no framework entry", declaration.ID)
			case runtimemod.ByLanguage:
				Expect(declaration.Languages).NotTo(BeEmpty(), "%s has no language entry", declaration.ID)
			case runtimemod.StackIndependent:
				Expect(declaration.Universal).To(BeTrue(), "%s must declare a universal entry", declaration.ID)
				Expect(declaration.Frameworks).To(BeEmpty(), "%s must not list frameworks", declaration.ID)
			}
		}
	})

	It("keeps every ModSet field in declarations and the public projection", func() {
		// 用反射取出 ModSet 中每个能力的 ID，与 Declarations()、对外输出逐一比对
		mods := mustResolve(bkmsStack(appruntime.FrameworkTRPC, appruntime.LanguageGo))
		value := reflect.ValueOf(mods)
		fieldIDs := make([]runtimemod.ID, 0, value.NumField())
		for i := range value.NumField() {
			idField := value.Field(i).FieldByName("ID")
			if !idField.IsValid() {
				continue
			}
			fieldIDs = append(fieldIDs, idField.Interface().(runtimemod.ID))
		}

		declared := lo.Map(runtimemod.Declarations(), func(d runtimemod.DefMeta, _ int) runtimemod.ID { return d.ID })
		Expect(fieldIDs).To(ConsistOf(declared))
		Expect(lo.Keys(mods.Public().Mods)).To(ConsistOf(declared))
	})

	It("does not import domain packages", func() {
		// 只允许依赖标准库、pkg/errors、lo 与 appruntime
		allowed := []string{
			"github.com/pkg/errors",
			"github.com/samber/lo",
			"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime",
		}
		entries, err := os.ReadDir(".")
		Expect(err).NotTo(HaveOccurred())
		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			file, parseErr := parser.ParseFile(token.NewFileSet(), name, nil, parser.ImportsOnly)
			Expect(parseErr).NotTo(HaveOccurred())
			for _, spec := range file.Imports {
				path, _ := strconv.Unquote(spec.Path.Value)
				isStdlib := !strings.Contains(strings.SplitN(path, "/", 2)[0], ".")
				Expect(isStdlib || lo.Contains(allowed, path)).
					To(BeTrue(), "%s imports %s", name, path)
			}
		}
	})
})
