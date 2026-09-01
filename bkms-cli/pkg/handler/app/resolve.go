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
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-cli/pkg/client"
)

// ResolveAppID 将用户输入的 app 标识（ID、name 或其前缀）解析为确定的 app ID。
//
// 匹配策略（精确优先 → 前缀最长匹配）：
//  1. 精确匹配：input 完全等于某 app 的 ID 或 Name。
//  2. 前缀匹配：input 是某 app ID 的前缀（app ID = name + "-" + 5位后缀），
//     按 ID 长度升序排序后取最短（与 input 最接近的即为最长匹配）。
//  3. 都未命中 → 当作 ID 透传（可能是其他空间的 app，由服务端最终校验）。
func ResolveAppID(ctx context.Context, workspaceID, input string) (string, error) {
	if workspaceID == "" {
		return input, nil
	}

	apps, err := client.New().ListApps(ctx, workspaceID)
	if err != nil {
		return "", errors.Wrap(err, "list apps for resolution")
	}

	return resolveAppID(apps, input)
}

// resolveAppID 精确匹配
func resolveAppID(apps []client.AppMinimal, input string) (string, error) {
	exactID, idFound := lo.Find(apps, func(a client.AppMinimal) bool { return a.ID == input })
	exactName, nameFound := lo.Find(apps, func(a client.AppMinimal) bool { return a.Name == input })

	switch {
	case idFound && nameFound && exactID.ID == exactName.ID:
		return exactID.ID, nil
	case idFound && nameFound:
		return "", errors.Errorf(
			"ambiguous: '%s' matches app '%s' by ID and app '%s' (id=%s) by name, "+
				"please use the full app ID to disambiguate",
			input, exactID.Name, exactName.Name, exactName.ID,
		)
	case idFound:
		return exactID.ID, nil
	case nameFound:
		return exactName.ID, nil
	}

	return resolveByPrefix(apps, input)
}

// resolveByPrefix 前缀匹配
// 在精确匹配均未命中时，按 app ID 前缀最长匹配规则查找。
// app ID 的结构为 name + "-" + 5位随机后缀，因此用户输入 name 本身就是 ID 的前缀。
func resolveByPrefix(apps []client.AppMinimal, input string) (string, error) {
	hits := lo.Filter(apps, func(a client.AppMinimal, _ int) bool {
		return strings.HasPrefix(a.ID, input)
	})
	if len(hits) == 0 {
		return input, nil
	}

	sort.Slice(hits, func(i, j int) bool { return len(hits[i].ID) < len(hits[j].ID) })

	if len(hits) > 1 && len(hits[0].ID) == len(hits[1].ID) {
		return "", errors.Errorf(
			"ambiguous: '%s' matches multiple apps by prefix, "+
				"please provide a longer identifier to disambiguate", input,
		)
	}

	return hits[0].ID, nil
}
