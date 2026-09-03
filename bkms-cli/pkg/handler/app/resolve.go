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
	"context"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-cli/pkg/client"
)

// ResolveAppID 将用户输入的 app 标识解析为确定的 app ID。
//
// 匹配策略：
//  1. 精确匹配 app ID。
//  2. 精确匹配 app Name，返回对应 ID。
//  3. 都未命中 → 当作 ID 透传（可能是其他空间的 app，由服务端最终校验）。
func ResolveAppID(ctx context.Context, workspaceID, input string) (string, error) {
	if workspaceID == "" {
		return input, nil
	}
	if input == "" {
		return "", errors.New("app cannot be empty")
	}

	apps, err := client.New().ListApps(ctx, workspaceID)
	if err != nil {
		return "", errors.Wrap(err, "list apps for resolution")
	}

	return resolveAppID(apps, input)
}

func resolveAppID(apps []client.AppMinimal, input string) (string, error) {
	if _, found := lo.Find(apps, func(a client.AppMinimal) bool { return a.ID == input }); found {
		return input, nil
	}

	if match, found := lo.Find(apps, func(a client.AppMinimal) bool { return a.Name == input }); found {
		return match.ID, nil
	}

	return input, nil
}
