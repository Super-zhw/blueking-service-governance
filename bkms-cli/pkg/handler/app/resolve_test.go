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

	It("should prefer ID match over Name match", func() {
		apps := []client.AppMinimal{
			{ID: "demo-abcde", Name: "demo"},
			{ID: "demo-abcde-fghij", Name: "demo-abcde"},
		}
		result, err := resolveAppID(apps, "demo-abcde")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("demo-abcde"))
	})

	It("should return ID when Name matches but ID does not", func() {
		apps := []client.AppMinimal{
			{ID: "demo-abcde", Name: "demo"},
			{ID: "demo-svc-fghij", Name: "demo-svc"},
		}
		result, err := resolveAppID(apps, "demo")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("demo-abcde"))
	})

	It("should handle app where ID equals Name", func() {
		apps := []client.AppMinimal{
			{ID: "demo-abcde", Name: "demo-abcde"},
		}
		result, err := resolveAppID(apps, "demo-abcde")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("demo-abcde"))
	})

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

	It("should handle single-char input that matches a name", func() {
		apps := []client.AppMinimal{
			{ID: "a-abcde", Name: "a"},
		}
		result, err := resolveAppID(apps, "a")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("a-abcde"))
	})

	It("should pass through partial name that does not exactly match", func() {
		apps := []client.AppMinimal{
			{ID: "abc-abcde", Name: "abc"},
		}
		result, err := resolveAppID(apps, "a")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("a"))
	})

	It("should handle app name with hyphens", func() {
		apps := []client.AppMinimal{
			{ID: "my-cool-svc-abcde", Name: "my-cool-svc"},
		}
		result, err := resolveAppID(apps, "my-cool-svc")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("my-cool-svc-abcde"))
	})

	It("should handle full ID match when workspace has many apps", func() {
		apps := []client.AppMinimal{
			{ID: "aaa-abcde", Name: "aaa"},
			{ID: "aaa-fghij", Name: "aaa-svc"},
			{ID: "bbb-klmno", Name: "bbb"},
		}
		result, err := resolveAppID(apps, "aaa-fghij")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("aaa-fghij"))
	})

	It("should pass through input that is longer than any ID", func() {
		apps := []client.AppMinimal{
			{ID: "demo-abcde", Name: "demo"},
		}
		result, err := resolveAppID(apps, "demo-abcdefghij")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("demo-abcdefghij"))
	})

	It("should pass through empty input", func() {
		result, err := resolveAppID(twoApps, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(""))
	})
})
