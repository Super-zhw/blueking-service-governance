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

package serializer_test

import (
	"encoding/json"

	"github.com/gin-gonic/gin/binding"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/app/serializer"
)

var _ = Describe("bkmsApp serializers", func() {
	Describe("UpdateAppStartupConfigInput.ToUpdateStartupConfigParams", func() {
		decode := func(body string) *serializer.UpdateAppStartupConfigInput {
			var input serializer.UpdateAppStartupConfigInput
			Expect(json.Unmarshal([]byte(body), &input)).To(Succeed())
			return &input
		}

		It("keeps frameworkConfig when it is omitted", func() {
			params, err := decode(`{"command":["/app/server"]}`).ToUpdateStartupConfigParams()
			Expect(err).NotTo(HaveOccurred())
			Expect(params.Command).To(Equal([]string{"/app/server"}))
			Expect(params.FrameworkConfig).To(BeNil())
		})

		It("replaces frameworkConfig with an explicit empty object", func() {
			params, err := decode(`{"frameworkConfig":{}}`).ToUpdateStartupConfigParams()
			Expect(err).NotTo(HaveOccurred())
			Expect(params.FrameworkConfig).To(Equal(map[string]any{}))
		})

		It("rejects a null frameworkConfig", func() {
			_, err := decode(`{"frameworkConfig":null}`).ToUpdateStartupConfigParams()
			Expect(err).To(MatchError(ContainSubstring("must not be null")))
		})

		It("rejects a non-object frameworkConfig", func() {
			_, err := decode(`{"frameworkConfig":[1]}`).ToUpdateStartupConfigParams()
			Expect(err).To(MatchError(ContainSubstring("must be an object")))
		})
	})

	Describe("CreateBKMSAppInput validation", func() {
		newInput := func() *serializer.CreateBKMSAppInput {
			return &serializer.CreateBKMSAppInput{
				Name:         "demo",
				ID:           "demo-abc",
				Framework:    "blank",
				Language:     "go",
				BuildConfig:  &serializer.BuildConfigInput{SourceType: "imageRegistry"},
				AppModelSpec: &serializer.BKMSAppModelSpecInput{Command: []string{"/app/server"}},
			}
		}

		It("accepts a valid input", func() {
			Expect(binding.Validator.ValidateStruct(newInput())).To(Succeed())
		})

		It("rejects shCommand together with command", func() {
			input := newInput()
			input.AppModelSpec.ShCommand = "./start.sh"
			Expect(binding.Validator.ValidateStruct(input)).
				To(MatchError(ContainSubstring("exec_command_or_sh_command_exclusive")))
		})

		It("rejects duplicated env var keys", func() {
			input := newInput()
			input.AppModelSpec.EnvVars = []serializer.VariableInput{{Key: "FOO"}, {Key: "FOO"}}
			Expect(binding.Validator.ValidateStruct(input)).To(MatchError(ContainSubstring("env_var_key_unique")))
		})

		It("rejects an invalid framework config file mount dir", func() {
			input := newInput()
			input.FrameworkConfigFile = &serializer.FrameworkConfigFileInput{Name: "a.yaml", MountDir: "relative"}
			Expect(binding.Validator.ValidateStruct(input)).To(MatchError(ContainSubstring("mount_dir")))
		})
	})
})
