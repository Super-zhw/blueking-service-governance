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

package appmodel

import (
	"context"

	"github.com/TencentBlueKing/gopkg/stringx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	bkmsapp "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/app"
	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/infras/account/auth"
	k8skind "github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/infras/kubernetes/kind"
)

var _ = Describe("createInstancesUpdateRecord", func() {
	var diApp *fxtest.App

	var ctx context.Context
	var recordStore RecordStore
	var appID string

	BeforeEach(func() {
		ctx = auth.WithUser(context.Background(), auth.User{ID: "tester"})
		appID = "inst-update-app-" + stringx.Random(6)

		diApp = fxtest.New(
			GinkgoT(),
			FxModule,
			fx.Populate(&recordStore),
		)
		diApp.RequireStart()
	})

	AfterEach(func() {
		diApp.RequireStop()
	})

	Context("createInstancesUpdateRecord", func() {
		It("should create a deploying record inheriting the base record", func() {
			base := &Record{
				WorkspaceID:     "ws-1",
				AppID:           appID,
				EnvName:         "prod",
				TrafficLaneName: "lane-1",
				ClusterID:       "BCS-K8S-00001",
				Namespace:       "ns-1",
				ImageTag:        "v1",
				UpdateStrategy:  "RollingUpdate",
				Replicas:        3,
				DeletionCost:    -2,
				LabelSelector:   map[string]string{"app": "gd1"},
				WorkloadKind:    k8skind.GameDeploy,
				ResourceKeys: ResourceKeys{
					{Kind: k8skind.GameDeploy, Name: "gd1"},
					{Kind: "ConfigMap", Name: "cm1"},
				},
				Status:  StatusDeployed,
				Extras:  map[string]string{ExtraKeyDeploySource: DeploySourceBuildAutoDeploy},
				Creator: "creator",
			}
			baseID, err := recordStore.Create(ctx, base)
			Expect(err).NotTo(HaveOccurred())
			base.ID, _ = bson.ObjectIDFromHex(baseID)

			deployer := &Deployer{recordStore: recordStore, app: &bkmsapp.Application{ID: appID}}
			deployID, err := deployer.createInstancesUpdateRecord(ctx, base, "v2", "InplaceUpdate", 4)
			Expect(err).NotTo(HaveOccurred())
			Expect(deployID).NotTo(Equal(baseID))

			latest, err := recordStore.GetLatest(ctx, appID, "prod", "lane-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(latest.ID.Hex()).To(Equal(deployID))
			Expect(latest.Status).To(Equal(StatusDeploying))
			Expect(latest.ImageTag).To(Equal("v2"))
			Expect(latest.UpdateStrategy).To(Equal("InplaceUpdate"))
			Expect(latest.Replicas).To(Equal(int32(4)))
			Expect(latest.WorkspaceID).To(Equal("ws-1"))
			Expect(latest.ClusterID).To(Equal("BCS-K8S-00001"))
			Expect(latest.Namespace).To(Equal("ns-1"))
			Expect(latest.DeletionCost).To(Equal(int64(-2)))
			Expect(latest.LabelSelector).To(Equal(base.LabelSelector))
			Expect(latest.WorkloadKind).To(Equal(k8skind.GameDeploy))
			Expect(latest.ResourceKeys).To(Equal(base.ResourceKeys))
			// 不继承构建信息，按直接部署展示
			Expect(latest.Extras).To(BeEmpty())
			Expect(latest.Creator).To(Equal("tester"))
			Expect(latest.Updater).To(Equal("tester"))
			Expect(latest.StartedAt.IsZero()).To(BeFalse())

			// 源记录保持不变
			old, err := recordStore.Get(ctx, appID, baseID)
			Expect(err).NotTo(HaveOccurred())
			Expect(old.Status).To(Equal(StatusDeployed))
			Expect(old.ImageTag).To(Equal("v1"))
		})
	})
})

// newPodForTest 创建用于测试的 Pod
func newPodForTest(name, phase string) unstructured.Unstructured {
	return unstructured.Unstructured{Object: map[string]any{
		"metadata": map[string]any{"name": name},
		"status":   map[string]any{"phase": phase},
	}}
}

var _ = Describe("coversAllRunningPods", func() {
	m := &podManager{}
	pods := []unstructured.Unstructured{
		newPodForTest("pod-1", "Running"),
		newPodForTest("pod-2", "Pending"),
		newPodForTest("pod-3", "Failed"),
	}

	It("should return true when all running pods are specified", func() {
		Expect(m.coversAllRunningPods(pods, []string{"pod-1", "pod-2"})).To(BeTrue())
		// 包含终止态 Pod 也视为全量
		Expect(m.coversAllRunningPods(pods, []string{"pod-1", "pod-2", "pod-3"})).To(BeTrue())
	})

	It("should return false when some running pods are not specified", func() {
		Expect(m.coversAllRunningPods(pods, []string{"pod-1"})).To(BeFalse())
		Expect(m.coversAllRunningPods(pods, []string{"pod-1", "pod-3"})).To(BeFalse())
	})

	It("should return false when there is no running pod", func() {
		Expect(m.coversAllRunningPods(nil, []string{"pod-1"})).To(BeFalse())
		Expect(m.coversAllRunningPods(
			[]unstructured.Unstructured{newPodForTest("pod-3", "Succeeded")}, []string{"pod-3"},
		)).To(BeFalse())
	})
})
