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

// Package e2escript_test 提供基于 testscript 的 bkms-cli 端到端黑盒测试。
//
// 测试脚本位于：
//   - testdata/         — 已认证场景（Setup 自动执行 login + workspace set）
//   - testdata-unauth/  — 未认证场景（Setup 仅写入不含 token 的最小配置）
//
// 运行方式：
//
//	make e2e-script-test
//
// 或直接（需先设置 test/e2e/.env 或相应环境变量）：
//
//	BKMS_CLI_BIN=./build/bkms-cli-e2e go test -v -timeout 10m ./test/e2e-script/...
package e2escript_test

import (
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

// TestE2E 运行 testdata/ 下的所有已认证场景脚本。
func TestE2E(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:   "testdata",
		Setup: setup,
		Cmds:  customCmds(),
	})
}

// TestE2EUnauth 运行 testdata-unauth/ 下的所有未认证场景脚本。
func TestE2EUnauth(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:   "testdata-unauth",
		Setup: setupUnauth,
		Cmds:  customCmds(),
	})
}
