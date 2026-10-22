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

// Package runtimemod 声明各技术栈（框架 + 语言）支持哪些能力，以及执行能力所需的参数。
//
// 每个能力用一个 Def 声明；按应用的技术栈查表，得到该能力的 Mod：是否支持、执行参数、不支持的原因。
// 本包只做查表，不访问数据库，也不依赖任何业务包。查表结果不落库，每次按当前代码重新计算。
package runtimemod

import (
	"slices"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime"
)

// ID 能力标识，也是对外 JSON 的 key，发布后不能修改。
type ID string

// ReasonCode 能力不支持的原因。
type ReasonCode string

const (
	// ReasonUnsupportedAppType 非 bkmsApp 应用（helm、agones），所有能力都不支持。
	ReasonUnsupportedAppType ReasonCode = "unsupported_app_type"
	// ReasonUnsupportedStack 该框架（或框架与语言的组合）未声明此能力。
	ReasonUnsupportedStack ReasonCode = "unsupported_stack"
	// ReasonIncompleteLanguage 能否支持取决于语言，但应用未记录语言。
	// BRIDGE-ONLY：只有存量应用会缺少语言，回填后删除。
	ReasonIncompleteLanguage ReasonCode = "incomplete_language"
	// ReasonDisabledByPolicy 预留：被运行时开关关闭。开关只能关闭能力，不能开启。
	ReasonDisabledByPolicy ReasonCode = "disabled_by_policy"
)

// ErrModUnsupported 能力不支持时，Mod.Require 返回此错误。
var ErrModUnsupported = errors.New("runtime mod unsupported")

// Match 语言限制。零值不限制，语言为空也放行；Only 只放行列出的语言，语言为空不放行。
type Match struct {
	values []string
}

// Any 不限制语言。
var Any = Match{}

// Only 只放行列出的语言。不传参数会 panic，说明声明写错了。
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

// FrameworkTable 按框架查参数，未列出的框架不支持。
type FrameworkTable[P Props] map[string]P

// LanguageTable 按语言查参数，未列出的语言不支持。
type LanguageTable[P Props] map[string]P

// Pair 框架与语言的组合。
type Pair struct {
	Framework string
	Language  string
}

// PairTable 按框架与语言的组合查参数，用于同一框架下参数随语言变化的能力。
type PairTable[P Props] map[Pair]P

// Dimension 能力随哪个维度变化。只用于测试中检查声明是否完整，不参与查表。
type Dimension int

const (
	// ByFramework 随框架变化。
	ByFramework Dimension = iota
	// ByLanguage 随语言变化。
	ByLanguage
	// ByPair 随框架与语言的组合变化。
	ByPair
	// StackIndependent 与技术栈无关，所有 bkmsApp 应用都支持，包括以后新增的框架。
	StackIndependent
)

// Props 能力参数需实现的接口。
//
// Public 返回可对外展示的部分，内部执行参数不得包含在内；没有可展示内容时返回 nil。
type Props interface {
	Public() any
}

// Def 能力声明。
//
// 查表顺序为 Pair、Framework、Language、Universal，先命中的生效；命中后还须满足 Languages 限制。
// 字段全部导出，是因为 M3 起各能力的声明会移到各自的业务包中定义。
type Def[P Props] struct {
	// ID 能力标识。
	ID ID
	// Dimension 能力的主要变化维度，只用于测试校验。
	Dimension Dimension
	// Framework 按框架声明的参数。
	Framework FrameworkTable[P]
	// Language 按语言声明的参数。
	Language LanguageTable[P]
	// Pair 按框架与语言组合声明的参数。
	Pair PairTable[P]
	// Universal 对所有框架生效的参数，用于与技术栈无关的能力。
	Universal *P
	// Languages 额外的语言限制，零值不限制。
	// 例如 platformBuild 设为 Only(go)：即使是 tRPC，也只有 go 支持。
	Languages Match
}

// Mod 能力的查表结果。
type Mod[P Props] struct {
	// ID 能力标识。
	ID ID
	// Supported 是否支持。
	Supported bool
	// ReasonCode 不支持的原因，支持时为空。
	ReasonCode ReasonCode
	// Props 执行参数，不支持时为零值。
	Props P
}

// DefMeta 声明摘要，不含参数，用于统一遍历各能力的声明。
type DefMeta struct {
	// ID 能力标识。
	ID        ID
	Dimension Dimension
	// Frameworks 声明涉及的框架，已去重排序。
	Frameworks []string
	// Languages 声明涉及的语言（含 Languages 限制），已去重排序。
	Languages []string
	// Universal 是否对所有框架生效。
	Universal bool
}

// Meta 生成声明摘要。
func (d Def[P]) Meta() DefMeta {
	pairs := lo.Keys(d.Pair)
	frameworks := append(
		lo.Keys(d.Framework),
		lo.Map(pairs, func(p Pair, _ int) string { return p.Framework })...,
	)
	languages := slices.Concat(
		lo.Keys(d.Language),
		lo.Map(pairs, func(p Pair, _ int) string { return p.Language }),
		d.Languages.values,
	)
	return DefMeta{
		ID:         d.ID,
		Dimension:  d.Dimension,
		Frameworks: sortedUniq(frameworks),
		Languages:  sortedUniq(languages),
		Universal:  d.Universal != nil,
	}
}

func sortedUniq(values []string) []string {
	out := lo.Uniq(values)
	slices.Sort(out)
	return out
}

// Resolve 按技术栈查表，得出该能力的支持情况。调用前技术栈应已校验通过。
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
		// 框架已声明，但未通过语言限制。语言为空时无法判断，报 incomplete_language。
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

	mod.ReasonCode = d.missReason(stack)
	return mod
}

// missReason 各表都未命中时的原因：语言为空、且补上语言后可能命中时报 incomplete_language，
// 其余情况报 unsupported_stack。
func (d Def[P]) missReason(stack appruntime.Stack) ReasonCode {
	if stack.Language != "" {
		return ReasonUnsupportedStack
	}
	hasPairForFramework := lo.ContainsBy(lo.Keys(d.Pair), func(p Pair) bool {
		return p.Framework == stack.Framework
	})
	if hasPairForFramework || len(d.Language) > 0 {
		return ReasonIncompleteLanguage
	}
	return ReasonUnsupportedStack
}

// Require 支持时返回参数，不支持时返回 ErrModUnsupported。执行器应通过它取参数。
func (m Mod[P]) Require() (P, error) {
	if !m.Supported {
		var zero P
		return zero, errors.Wrapf(ErrModUnsupported, "%s: %s", m.ID, m.ReasonCode)
	}
	return m.Props, nil
}

// PublicModView 单个能力的对外结构。
type PublicModView struct {
	Supported bool `json:"supported"`
	// ReasonCode 不支持的原因，支持时省略。
	ReasonCode ReasonCode `json:"reasonCode,omitempty"`
	// Props 对外参数，没有时省略。
	Props any `json:"props,omitempty"`
}

// view 把不同参数类型的 Mod 统一转成对外结构。
type view interface {
	view() PublicModView
}

func (m Mod[P]) view() PublicModView {
	if !m.Supported {
		return PublicModView{Supported: false, ReasonCode: m.ReasonCode}
	}
	return PublicModView{Supported: true, Props: m.Props.Public()}
}
