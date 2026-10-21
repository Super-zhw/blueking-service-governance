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

package appruntime

// CapabilityStack 为能力解析规范化快照。
//
// 与 Normalize 只有一处不同：language 缺失不视为致命错误。存量 TAF 记录从未保存 language，
// 部分 tRPC 记录只在 workload 上保存 language，若视为阻断，会收回这些应用当前正在使用的能力。
//
// BRIDGE-ONLY：阶段 5 回填 language 之后，调用方可直接使用 ValidateStack，本函数随 Inspect 一并删除。
func CapabilityStack(s Snapshot) (Stack, error) {
	report := Inspect(s)
	for _, issue := range report.Issues {
		if issue.Blocking && issue.Code != IssueIncompleteLanguage {
			return Stack{}, report.Err()
		}
	}
	return report.Stack, nil
}
