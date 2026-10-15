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

package publish

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-cli/pkg/client"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-cli/pkg/client/mocks"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-cli/pkg/utils/clierr"
)

var _ = Describe("BuildRestartOptions", func() {
	It("builds envs target with deduplicated env names", func() {
		opts, err := BuildRestartOptions("stage, test,stage", "", "")
		Expect(err).NotTo(HaveOccurred())
		Expect(opts).To(Equal(client.DevModeRestartOptions{
			TargetType: client.DevModeRestartTargetEnvs,
			EnvNames:   []string{"stage", "test"},
		}))
	})

	It("builds envTypes target", func() {
		opts, err := BuildRestartOptions("", "development,test", "")
		Expect(err).NotTo(HaveOccurred())
		Expect(opts).To(Equal(client.DevModeRestartOptions{
			TargetType: client.DevModeRestartTargetEnvTypes,
			EnvTypes:   []string{"development", "test"},
		}))
	})

	It("builds instances target with a single env", func() {
		opts, err := BuildRestartOptions("stage", "", "pod-1,pod-2")
		Expect(err).NotTo(HaveOccurred())
		Expect(opts).To(Equal(client.DevModeRestartOptions{
			TargetType:  client.DevModeRestartTargetInstances,
			EnvName:     "stage",
			InstanceIDs: []string{"pod-1", "pod-2"},
		}))
	})

	DescribeTable("rejects invalid flag combinations",
		func(envs, envTypes, instances, msg string) {
			_, err := BuildRestartOptions(envs, envTypes, instances)
			Expect(err).To(HaveOccurred())
			var usageErr *clierr.UsageError
			Expect(errors.As(err, &usageErr)).To(BeTrue())
			Expect(err.Error()).To(ContainSubstring(msg))
		},
		Entry("both env and env-type", "stage", "test", "", "cannot be used together"),
		Entry("neither env nor env-type", "", "", "", "one of --env or --env-type is required"),
		Entry("blank env", " , ", "", "", "one of --env or --env-type is required"),
		Entry("instances with multiple envs", "stage,test", "", "pod-1", "single --env"),
		Entry("instances with env-type", "", "test", "pod-1", "single --env"),
	)
})

var _ = Describe("Restart", func() {
	const appID = "demo-app"

	var (
		ctx  context.Context
		cli  *mocks.MockClient
		opts client.DevModeRestartOptions
	)

	BeforeEach(func() {
		ctx = context.Background()
		cli = mocks.NewMockClient(GinkgoT())
		opts = client.DevModeRestartOptions{TargetType: client.DevModeRestartTargetEnvs, EnvNames: []string{"stage"}}
	})

	It("succeeds when all instances are restarted", func() {
		cli.EXPECT().RestartDevModeProcesses(mock.Anything, appID, opts).Return(&client.DevModeRestartData{
			Envs: []client.DevModeRestartEnvResult{{
				EnvName: "stage", EnvType: "staging",
				Instances: []client.DevModeRestartInstanceResult{
					{Instance: "pod-1", Status: "success", OldPID: "1", NewPID: "2"},
				},
			}},
		}, nil)

		Expect(Restart(ctx, cli, appID, opts)).To(Succeed())
	})

	It("returns an error when any instance fails", func() {
		cli.EXPECT().RestartDevModeProcesses(mock.Anything, appID, opts).Return(&client.DevModeRestartData{
			Envs: []client.DevModeRestartEnvResult{{
				EnvName: "stage", EnvType: "staging",
				Instances: []client.DevModeRestartInstanceResult{
					{Instance: "pod-1", Status: "success", OldPID: "1", NewPID: "2"},
					{Instance: "pod-2", Status: "failed", Message: "boom"},
				},
			}},
			Skipped: []client.DevModeRestartSkippedEnv{
				{EnvName: "dev", EnvType: "development", Reason: "devmode disabled"},
			},
		}, nil)

		err := Restart(ctx, cli, appID, opts)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("restart failed on 1 instance(s)"))
	})

	It("succeeds when all environments are skipped", func() {
		cli.EXPECT().RestartDevModeProcesses(mock.Anything, appID, opts).Return(&client.DevModeRestartData{
			Skipped: []client.DevModeRestartSkippedEnv{
				{EnvName: "dev", EnvType: "development", Reason: "devmode disabled"},
			},
		}, nil)

		Expect(Restart(ctx, cli, appID, opts)).To(Succeed())
	})

	It("wraps the client error", func() {
		cli.EXPECT().RestartDevModeProcesses(mock.Anything, appID, opts).Return(nil, errors.New("server down"))

		err := Restart(ctx, cli, appID, opts)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("server down"))
	})
})
