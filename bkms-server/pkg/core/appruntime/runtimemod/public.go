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

package runtimemod

import (
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime"
)

// PublicView 应用能力的对外结构。
//
// 只反映技术栈决定的支持情况，不包含环境、集群等运行时条件。
type PublicView struct {
	SchemaVersion int    `json:"schemaVersion"`
	Type          string `json:"type"`
	// Framework 框架，helm、agones 应用为空。
	Framework string `json:"framework,omitempty"`
	// Language 语言，未记录时为空。
	Language string `json:"language,omitempty"`
	// Mods 能力 ID 到能力的映射，包含全部能力；新增能力只会增加 key。
	Mods map[ID]PublicModView `json:"mods"`
}

// Public 转为对外结构。
func (s ModSet) Public() PublicView {
	return PublicView{
		SchemaVersion: SchemaVersion,
		Type:          s.Stack.Type,
		Framework:     s.Stack.Framework,
		Language:      s.Stack.Language,
		Mods:          lo.MapValues(s.all(), func(mod view, _ ID) PublicModView { return mod.view() }),
	}
}

// Matrix 返回框架目录中每个合法组合的能力，供前端创建应用时展示。
func Matrix() ([]PublicView, error) {
	var out []PublicView
	for _, def := range appruntime.Catalog() {
		for _, language := range def.SupportedLanguages {
			set, err := Resolve(appruntime.Stack{
				Type:      appruntime.AppTypeBKMSApp,
				Framework: def.Name,
				Language:  language,
			})
			if err != nil {
				return nil, errors.Wrapf(err, "resolve %s+%s", def.Name, language)
			}
			out = append(out, set.Public())
		}
	}
	return out, nil
}
