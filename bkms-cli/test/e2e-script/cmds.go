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

package e2escript_test

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rogpeppe/go-internal/testscript"
)

// customCmds 返回脚本中可用的自定义命令映射。
// 所有命令均通过 ts.ReadFile("stdout") 读取上一条 exec 命令的标准输出。
func customCmds() map[string]func(ts *testscript.TestScript, neg bool, args []string) {
	return map[string]func(ts *testscript.TestScript, neg bool, args []string){
		"jsonhas":     cmdJSONHas,
		"jsonfield":   cmdJSONField,
		"outcontains": cmdOutContains,
	}
}

// cmdJSONHas 断言上一条命令的 stdout 是合法 JSON，且包含指定的顶层键。
//
// 用法：
//
//	jsonhas <key> [<key>...]
//	! jsonhas <key>        — 断言键不存在
func cmdJSONHas(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) == 0 {
		ts.Fatalf("jsonhas: at least one key required")
	}

	raw := ts.ReadFile("stdout")
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		ts.Fatalf("jsonhas: stdout is not a JSON object: %v\nstdout: %s", err, raw)
	}

	for _, key := range args {
		_, has := m[key]
		if neg && has {
			ts.Fatalf("jsonhas: key %q unexpectedly present in JSON output", key)
		}
		if !neg && !has {
			ts.Fatalf("jsonhas: key %q not found in JSON output\nstdout: %s", key, raw)
		}
	}
}

// cmdJSONField 断言上一条命令的 stdout 是合法 JSON，且指定字段的字符串值与期望相符。
//
// 用法：
//
//	jsonfield <key>=<expected>
//	! jsonfield <key>=<expected>   — 断言字段值不等于 expected
func cmdJSONField(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) != 1 {
		ts.Fatalf("jsonfield: exactly one key=value argument required")
	}

	parts := strings.SplitN(args[0], "=", 2)
	if len(parts) != 2 {
		ts.Fatalf("jsonfield: argument must be key=value, got %q", args[0])
	}
	key, expected := parts[0], parts[1]

	raw := ts.ReadFile("stdout")
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		ts.Fatalf("jsonfield: stdout is not a JSON object: %v\nstdout: %s", err, raw)
	}

	val, has := m[key]
	if !has {
		ts.Fatalf("jsonfield: key %q not found in JSON output\nstdout: %s", key, raw)
	}

	actual := fmt.Sprintf("%v", val)
	match := actual == expected
	if neg && match {
		ts.Fatalf("jsonfield: field %q unexpectedly equals %q", key, expected)
	}
	if !neg && !match {
		ts.Fatalf("jsonfield: field %q: got %q, want %q\nstdout: %s", key, actual, expected, raw)
	}
}

// cmdOutContains 断言上一条命令的 stdout+stderr 合并输出包含指定子串。
// 用于需要同时检查两个流的场景，等价于旧框架的 ExpectOutputContains。
//
// 用法：
//
//	outcontains <substring>
//	! outcontains <substring>   — 断言输出不包含该子串
func cmdOutContains(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) != 1 {
		ts.Fatalf("outcontains: exactly one argument required")
	}
	sub := args[0]

	combined := ts.ReadFile("stdout") + ts.ReadFile("stderr")
	has := strings.Contains(combined, sub)
	if neg && has {
		ts.Fatalf("outcontains: output unexpectedly contains %q\ncombined: %s", sub, combined)
	}
	if !neg && !has {
		ts.Fatalf("outcontains: output does not contain %q\ncombined: %s", sub, combined)
	}
}
