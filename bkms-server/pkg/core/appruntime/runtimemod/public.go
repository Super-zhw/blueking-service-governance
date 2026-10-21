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

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime"
)

// PublicView 单个应用的对外能力描述。
//
// 这里只包含静态能力。依赖环境、配置、集群或权限的可用性，属于环境维度的视图。
type PublicView struct {
	SchemaVersion int                  `json:"schemaVersion"`
	Type          string               `json:"type"`
	Framework     string               `json:"framework,omitempty"`
	Language      string               `json:"language,omitempty"`
	Mods          map[ID]PublicModView `json:"mods"`
}

// Public 投影能力集合。使用 map 作为 mods，encoding/json 会按 key 排序，输出稳定。
func (s ModSet) Public() PublicView {
	mods := s.all()
	out := make(map[ID]PublicModView, len(mods))
	for id, mod := range mods {
		out[id] = mod.view()
	}
	return PublicView{
		SchemaVersion: SchemaVersion,
		Type:          s.Stack.Type,
		Framework:     s.Stack.Framework,
		Language:      s.Stack.Language,
		Mods:          out,
	}
}

// Matrix 基于框架目录生成所有合法技术栈的支持矩阵，顺序稳定。客户端可直接用它构建创建表单，
// 无需再维护一份规则副本。
func Matrix() ([]PublicView, error) {
	frameworks := appruntime.Catalog()
	out := make([]PublicView, 0, len(frameworks))
	for _, def := range frameworks {
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
