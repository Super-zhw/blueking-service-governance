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

// Package runtimemod 声明各技术栈支持哪些运行时能力（Mod）。
//
// 技术栈以框架为主维度：多数能力只取决于 framework，少数只取决于 language，同时取决于两者的
// 极少。因此声明采用维度表，而不是枚举 (framework, language) 规则；解析一个技术栈的开销与能力
// 数量无关。
//
// 本包是纯函数包：无 I/O、不持久化、不缓存。不得引入领域依赖，保证各能力 owner 包都能 import
// 它而不产生循环引用。
package runtimemod

import (
	"slices"

	"github.com/pkg/errors"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime"
)

// ID 能力的稳定标识，作为 JSON key 对外暴露。
type ID string

// ReasonCode 说明能力不支持的原因，供机器识别，不得携带执行器细节。
type ReasonCode string

const (
	// ReasonUnsupportedAppType 应用不由 AppModel 管理（helm / agones）。
	ReasonUnsupportedAppType ReasonCode = "unsupported_app_type"
	// ReasonUnsupportedStack 该框架（或框架+语言）没有对应声明。
	ReasonUnsupportedStack ReasonCode = "unsupported_stack"
	// ReasonIncompleteLanguage 声明要求具体语言，但记录的技术栈没有 language。
	// BRIDGE-ONLY：存量记录可能缺少 language。
	ReasonIncompleteLanguage ReasonCode = "incomplete_language"
	// ReasonDisabledByPolicy 预留给运行时开关，开关只能收紧结果，不能放宽。
	ReasonDisabledByPolicy ReasonCode = "disabled_by_policy"
)

// ErrModUnsupported 由 Mod.Require 返回，handler 可据此映射为请求错误。
var ErrModUnsupported = errors.New("runtime mod unsupported")

// Match 约束声明在某一维度上的取值。零值匹配任意值（包括空值）；Only 只匹配列出的值，
// 且永远不匹配空值。
type Match struct {
	values []string
}

// Any 匹配任意值，包括缺失的 language。
var Any = Match{}

// Only 只匹配列出的值。传入空列表会 panic：这一定是声明写错了，而不是运行时情况。
func Only(values ...string) Match {
	if len(values) == 0 {
		panic("runtimemod: Only requires at least one value, use Any instead")
	}
	return Match{values: slices.Clone(values)}
}

func (m Match) matches(value string) bool {
	if len(m.values) == 0 {
		return true
	}
	return slices.Contains(m.values, value)
}

// FrameworkTable 框架 → props，缺少 key 表示不支持。
type FrameworkTable[P Props] map[string]P

// LanguageTable 语言 → props，缺少 key 表示不支持。
type LanguageTable[P Props] map[string]P

// Pair (framework, language) 组合，仅用于确实因组合而不同的行为。
type Pair struct {
	Framework string
	Language  string
}

// PairTable 框架+语言组合 → props。
type PairTable[P Props] map[Pair]P

// Dimension 标明声明预期填写哪张表。解析时会读取所有表，因此该字段只用于测试中校验声明。
type Dimension int

const (
	// ByFramework 默认维度：能力只取决于框架。
	ByFramework Dimension = iota
	// ByLanguage 能力只取决于语言。
	ByLanguage
	// ByPair 能力同时取决于框架和语言，应尽量少用。
	ByPair
	// StackIndependent 所有 bkmsApp 框架都支持该能力，包括后续新增的框架。
	// 仅用于在框架代码之外实现的功能。
	StackIndependent
)

// Props 每个能力的 props 类型都要实现。Public 决定哪些字段对外输出：执行器 key 保留在内部，
// 没有可公开字段的类型返回 nil。
type Props interface {
	Public() any
}

// Def 声明一个能力。字段全部导出，因为后续阶段声明会搬到各能力自己的包中。
type Def[P Props] struct {
	ID        ID
	Dimension Dimension
	// Framework 仅按框架区分的条目。优先级高于 Language，这样框架专属能力在后续新增语言时
	// 依然正确。
	Framework FrameworkTable[P]
	// Language 仅按语言区分的条目，在 Framework 没有命中时才查。
	Language LanguageTable[P]
	// Pair 优先级最高，覆盖确实因 (framework, language) 组合而不同的情况。
	Pair PairTable[P]
	// Universal 上面各表都没有命中时，对所有框架生效。优先级最低，便于用表条目单独细化某个框架。
	Universal *P
	// Languages 无论命中哪张表，都在语言维度上额外限制结果。零值表示不限制。
	Languages Match
}

// Mod 单个能力的解析结果。
type Mod[P Props] struct {
	ID         ID
	Supported  bool
	ReasonCode ReasonCode
	Props      P
}

// DefMeta 声明的非泛型元数据，供测试与后续 catalog API 遍历异构声明，无需对 props 做类型断言。
type DefMeta struct {
	ID         ID
	Dimension  Dimension
	Frameworks []string
	Languages  []string
	// Universal 表示该声明是否对所有框架生效。
	Universal bool
}

// Meta 返回声明中提到的框架和语言。
func (d Def[P]) Meta() DefMeta {
	frameworks := make([]string, 0, len(d.Framework)+len(d.Pair))
	languages := make([]string, 0, len(d.Language)+len(d.Pair)+len(d.Languages.values))
	for framework := range d.Framework {
		frameworks = append(frameworks, framework)
	}
	for pair := range d.Pair {
		frameworks = append(frameworks, pair.Framework)
		languages = append(languages, pair.Language)
	}
	for language := range d.Language {
		languages = append(languages, language)
	}
	languages = append(languages, d.Languages.values...)
	return DefMeta{
		ID:         d.ID,
		Dimension:  d.Dimension,
		Frameworks: slices.Compact(slices.Sorted(slices.Values(frameworks))),
		Languages:  slices.Compact(slices.Sorted(slices.Values(languages))),
		Universal:  d.Universal != nil,
	}
}

// Resolve 用一条声明解析已校验的技术栈，不做任何 I/O。
func (d Def[P]) Resolve(stack appruntime.Stack) Mod[P] {
	mod := Mod[P]{ID: d.ID}
	if stack.Type != appruntime.AppTypeBKMSApp {
		mod.ReasonCode = ReasonUnsupportedAppType
		return mod
	}

	if props, ok := d.Pair[Pair{Framework: stack.Framework, Language: stack.Language}]; ok &&
		d.Languages.matches(stack.Language) {
		mod.Supported, mod.Props = true, props
		return mod
	}

	if props, ok := d.Framework[stack.Framework]; ok {
		if d.Languages.matches(stack.Language) {
			mod.Supported, mod.Props = true, props
			return mod
		}
		// 框架已声明，是语言限制拦住了。language 缺失时限制本可能通过，因此报 incomplete。
		if stack.Language == "" {
			mod.ReasonCode = ReasonIncompleteLanguage
			return mod
		}
		mod.ReasonCode = ReasonUnsupportedStack
		return mod
	}

	if props, ok := d.Language[stack.Language]; ok && d.Languages.matches(stack.Language) {
		mod.Supported, mod.Props = true, props
		return mod
	}

	if d.Universal != nil && d.Languages.matches(stack.Language) {
		mod.Supported, mod.Props = true, *d.Universal
		return mod
	}

	// 只有在"如果记录了 language，声明就会生效"时才报 incomplete_language，
	// 否则就是普通的不支持。
	mod.ReasonCode = d.missReason(stack, d.hasPairFor(stack.Framework))
	return mod
}

// missReason 区分"没有声明"和"声明因缺少 language 而无法生效"。
func (d Def[P]) missReason(stack appruntime.Stack, pairWouldApply bool) ReasonCode {
	if stack.Language == "" {
		if pairWouldApply || len(d.Language) > 0 {
			return ReasonIncompleteLanguage
		}
	}
	return ReasonUnsupportedStack
}

func (d Def[P]) hasPairFor(framework string) bool {
	for pair := range d.Pair {
		if pair.Framework == framework {
			return true
		}
	}
	return false
}

// Require 支持时返回 props；不支持时返回带能力 ID 与原因的 ErrModUnsupported。
// 执行器统一以它作为入口。
func (m Mod[P]) Require() (P, error) {
	if !m.Supported {
		var zero P
		return zero, errors.Wrapf(ErrModUnsupported, "%s: %s", m.ID, m.ReasonCode)
	}
	return m.Props, nil
}

// PublicModView 单个能力的对外投影。
type PublicModView struct {
	Supported  bool       `json:"supported"`
	ReasonCode ReasonCode `json:"reasonCode,omitempty"`
	Props      any        `json:"props,omitempty"`
}

// view 非泛型接口，用于遍历由不同 props 类型组成的 ModSet。
type view interface {
	view() PublicModView
}

func (m Mod[P]) view() PublicModView {
	if !m.Supported {
		return PublicModView{Supported: false, ReasonCode: m.ReasonCode}
	}
	return PublicModView{Supported: true, Props: m.Props.Public()}
}
