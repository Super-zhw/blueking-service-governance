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

// Package app 提供应用创建相关的处理逻辑
package app

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-cli/pkg/client"
)

var _ = Describe("resolveAppID", func() {
	twoApps := []client.AppMinimal{
		{ID: "myapp-abcde", Name: "myapp"},
		{ID: "other-fghij", Name: "other"},
	}

	// ==================== exact match ====================

	It("should match by ID exactly", func() {
		result, err := resolveAppID(twoApps, "myapp-abcde")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("myapp-abcde"))
	})

	It("should match by Name exactly and return app ID", func() {
		result, err := resolveAppID(twoApps, "myapp")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("myapp-abcde"))
	})

	It("should report ambiguity when ID of one app equals Name of another", func() {
		apps := []client.AppMinimal{
			{ID: "demo-abcde", Name: "demo"},
			{ID: "demo-abcde-fghij", Name: "demo-abcde"},
		}
		_, err := resolveAppID(apps, "demo-abcde")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("ambiguous"))
	})

	It("should prefer exact Name match over prefix ID match", func() {
		apps := []client.AppMinimal{
			{ID: "demo-abcde", Name: "demo"},
			{ID: "demo-svc-fghij", Name: "demo-svc"},
		}
		result, err := resolveAppID(apps, "demo")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("demo-abcde"))
	})

	It("should prefer exact ID match over exact Name match when they point to the same app", func() {
		apps := []client.AppMinimal{
			{ID: "demo-abcde", Name: "demo-abcde"},
		}
		result, err := resolveAppID(apps, "demo-abcde")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("demo-abcde"))
	})

	// ==================== no match → pass through ====================

	It("should pass through input when nothing matches", func() {
		result, err := resolveAppID(twoApps, "unknown-app")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("unknown-app"))
	})

	It("should pass through input when app list is empty", func() {
		result, err := resolveAppID(nil, "myapp")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("myapp"))
	})

	// ==================== edge / boundary cases ====================

	It("should handle single-char input that matches a name", func() {
		apps := []client.AppMinimal{
			{ID: "a-abcde", Name: "a"},
		}
		result, err := resolveAppID(apps, "a")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("a-abcde"))
	})

	It("should fall through to prefix when input is a single char not matching any name", func() {
		apps := []client.AppMinimal{
			{ID: "abc-abcde", Name: "abc"},
		}
		result, err := resolveAppID(apps, "a")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("abc-abcde"))
	})

	It("should handle app name with hyphens (name itself contains '-')", func() {
		apps := []client.AppMinimal{
			{ID: "my-cool-svc-abcde", Name: "my-cool-svc"},
		}
		result, err := resolveAppID(apps, "my-cool-svc")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("my-cool-svc-abcde"))
	})

	It("should handle input equal to full ID when workspace has many apps", func() {
		apps := []client.AppMinimal{
			{ID: "aaa-abcde", Name: "aaa"},
			{ID: "aaa-fghij", Name: "aaa-svc"},
			{ID: "bbb-klmno", Name: "bbb"},
		}
		result, err := resolveAppID(apps, "aaa-fghij")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("aaa-fghij"))
	})

	It("should not match input that is longer than any ID", func() {
		apps := []client.AppMinimal{
			{ID: "demo-abcde", Name: "demo"},
		}
		result, err := resolveAppID(apps, "demo-abcdefghij")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("demo-abcdefghij"))
	})

	It("should report ambiguity when empty input matches all apps", func() {
		_, err := resolveAppID(twoApps, "")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("ambiguous"))
	})

	It("should match exact ID even when multiple apps share the same name prefix", func() {
		apps := []client.AppMinimal{
			{ID: "svc-abcde", Name: "svc"},
			{ID: "svc-ab-fghij", Name: "svc-ab"},
			{ID: "svc-abc-klmno", Name: "svc-abc"},
		}
		result, err := resolveAppID(apps, "svc-abcde")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("svc-abcde"))
	})
})

var _ = Describe("resolveByPrefix", func() {
	It("should match partial ID as prefix", func() {
		apps := []client.AppMinimal{
			{ID: "myapp-abcde", Name: "myapp"},
			{ID: "other-fghij", Name: "other"},
		}
		result, err := resolveByPrefix(apps, "myapp-ab")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("myapp-abcde"))
	})

	It("should prefer shorter ID (longest match) when multiple prefix hits", func() {
		apps := []client.AppMinimal{
			{ID: "demo-service-fghij", Name: "demo-service"},
			{ID: "demo-abcde", Name: "demo"},
		}
		result, err := resolveByPrefix(apps, "demo-")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("demo-abcde"))
	})

	It("should report ambiguity when multiple apps have same-length ID", func() {
		apps := []client.AppMinimal{
			{ID: "demo-svc-abcde", Name: "demo-svc"},
			{ID: "demo-app-fghij", Name: "demo-app"},
		}
		_, err := resolveByPrefix(apps, "demo-")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("ambiguous"))
		Expect(err.Error()).To(ContainSubstring("prefix"))
	})

	It("should resolve partial name via ID prefix", func() {
		apps := []client.AppMinimal{
			{ID: "myapp-abcde", Name: "myapp"},
		}
		result, err := resolveByPrefix(apps, "myap")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("myapp-abcde"))
	})

	It("should pass through when no prefix matches", func() {
		apps := []client.AppMinimal{
			{ID: "myapp-abcde", Name: "myapp"},
		}
		result, err := resolveByPrefix(apps, "unknown")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("unknown"))
	})

	// ==================== edge / boundary cases ====================

	It("should handle empty app list", func() {
		result, err := resolveByPrefix(nil, "demo")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("demo"))
	})

	It("should handle single-char prefix matching multiple apps", func() {
		apps := []client.AppMinimal{
			{ID: "a-abcde", Name: "a"},
			{ID: "ab-fghij", Name: "ab"},
		}
		result, err := resolveByPrefix(apps, "a")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("a-abcde"))
	})

	It("should resolve correctly when all IDs have same prefix but different lengths", func() {
		apps := []client.AppMinimal{
			{ID: "svc-abc-klmno", Name: "svc-abc"},
			{ID: "svc-abcde", Name: "svc"},
			{ID: "svc-abc-def-pqrst", Name: "svc-abc-def"},
		}
		result, err := resolveByPrefix(apps, "svc-a")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("svc-abcde"))
	})

	It("should report ambiguity among three apps with same-length ID", func() {
		apps := []client.AppMinimal{
			{ID: "demo-aaa-abcde", Name: "demo-aaa"},
			{ID: "demo-bbb-fghij", Name: "demo-bbb"},
			{ID: "demo-ccc-klmno", Name: "demo-ccc"},
		}
		_, err := resolveByPrefix(apps, "demo-")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("ambiguous"))
	})

	It("should not match when input is longer than all IDs", func() {
		apps := []client.AppMinimal{
			{ID: "demo-abcde", Name: "demo"},
		}
		result, err := resolveByPrefix(apps, "demo-abcde-extra")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("demo-abcde-extra"))
	})

	It("should match unique app when input equals ID exactly (prefix includes exact)", func() {
		apps := []client.AppMinimal{
			{ID: "myapp-abcde", Name: "myapp"},
		}
		result, err := resolveByPrefix(apps, "myapp-abcde")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("myapp-abcde"))
	})

	It("should pick shorter ID even if it appears later in the list", func() {
		apps := []client.AppMinimal{
			{ID: "api-gateway-fghij", Name: "api-gateway"},
			{ID: "api-abcde", Name: "api"},
			{ID: "api-server-klmno", Name: "api-server"},
		}
		result, err := resolveByPrefix(apps, "api-")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("api-abcde"))
	})

	It("should handle empty input as prefix of all IDs and pick shortest", func() {
		apps := []client.AppMinimal{
			{ID: "long-name-abcde", Name: "long-name"},
			{ID: "ab-fghij", Name: "ab"},
		}
		result, err := resolveByPrefix(apps, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("ab-fghij"))
	})
})
