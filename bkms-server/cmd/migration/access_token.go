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

package migration

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/term"
)

// readAccessToken 交互式读取 access token（密文输入，不回显）。
//
// 相比通过 --access-token 命令行参数明文传递，密文输入不会残留到 shell history、
// 进程参数列表（ps 可见）等，避免令牌泄露。
func readAccessToken() (string, error) {
	fmt.Fprint(os.Stderr, "Enter access token: ")
	token, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", errors.Wrap(err, "read access token")
	}

	tokenStr := strings.TrimSpace(string(token))
	if tokenStr == "" {
		return "", errors.New("access token must not be empty")
	}
	return tokenStr, nil
}
