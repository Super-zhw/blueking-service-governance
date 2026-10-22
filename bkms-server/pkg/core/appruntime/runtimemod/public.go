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

// PublicView 一个应用的能力集合对外输出的 JSON 结构。
//
// 只描述技术栈决定的静态能力。某能力在具体环境下是否可用（如生产环境禁用 devMode、
// 集群未安装依赖组件）不在这里体现，需要结合环境另行判断。
type PublicView struct {
	// SchemaVersion 结构版本号，见 SchemaVersion 常量。
	SchemaVersion int `json:"schemaVersion"`
	// Type 应用类型，如 bkmsApp、helm。
	Type string `json:"type"`
	// Framework 应用框架，helm / agones 应用为空并省略。
	Framework string `json:"framework,omitempty"`
	// Language 应用语言，未记录时为空并省略。
	Language string `json:"language,omitempty"`
	// Mods 能力 ID 到能力视图的映射。使用 map 是因为 encoding/json 会按 key 排序输出，结果稳定。
	Mods map[ID]PublicModView `json:"mods"`
}

// Public 把能力集合转换为对外 JSON 结构。
func (s ModSet) Public() PublicView {
	return PublicView{
		SchemaVersion: SchemaVersion,
		Type:          s.Stack.Type,
		Framework:     s.Stack.Framework,
		Language:      s.Stack.Language,
		Mods:          lo.MapValues(s.all(), func(mod view, _ ID) PublicModView { return mod.view() }),
	}
}

// Matrix 组合约定 (framework, language) 的能力集合，按顺序返回。
//
// 用于前端创建应用时展示各技术栈支持哪些能力，前端无需自行维护一份支持规则。
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
