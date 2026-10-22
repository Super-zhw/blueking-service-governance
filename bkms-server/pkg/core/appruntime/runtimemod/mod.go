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

// Package runtimemod 声明每种技术栈（framework + language）支持哪些运行时能力，以及支持时的参数。
//
// 一个能力对应一个 Def，解析某个技术栈得到一个 Mod：Supported 表示是否支持，Props 是支持时
// 执行器需要的参数，ReasonCode 是不支持的原因。
//
// 大多数能力只随 framework 变化，少数只随 language 变化，同时依赖两者的极少。因此 Def 用
// 按维度划分的查找表来声明，而不是穷举所有 (framework, language) 组合。
//
// 本包只做声明与查表：不读写数据库、不缓存结果、不依赖任何业务包，各能力所在的业务包都可以
// 直接引用它而不会产生循环依赖。
package runtimemod

import (
	"slices"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/TencentBlueKing/blueking-service-governance/bkms-server/pkg/core/appruntime"
)

// ID 能力标识，如 "devMode"。对外作为 JSON 的 key，发布后不可更改。
type ID string

// ReasonCode 能力不支持的原因码，供前端判断展示方式；不得包含脚本集、解析器等内部实现信息。
type ReasonCode string

const (
	// ReasonUnsupportedAppType 应用类型不是 bkmsApp（如 helm / agones），所有能力都不支持。
	ReasonUnsupportedAppType ReasonCode = "unsupported_app_type"
	// ReasonUnsupportedStack 当前 framework（或 framework + language）没有声明该能力。
	ReasonUnsupportedStack ReasonCode = "unsupported_stack"
	// ReasonIncompleteLanguage 该能力的支持与否取决于 language，但应用没有记录 language，无法判断。
	// BRIDGE-ONLY：只有存量记录会缺少 language，数据回填后删除。
	ReasonIncompleteLanguage ReasonCode = "incomplete_language"
	// ReasonDisabledByPolicy 预留：运行时开关关闭了该能力。开关只能把支持改为不支持，不能反向。
	ReasonDisabledByPolicy ReasonCode = "disabled_by_policy"
)

// ErrModUnsupported 执行器调用 Mod.Require 时，能力不支持则返回该错误；handler 可用 errors.Is 识别。
var ErrModUnsupported = errors.New("runtime mod unsupported")

// Match 声明在某个维度（目前只用于 language）上的取值限制。
//
//   - 零值（即 Any）：不限制，任何值都匹配，包括空值。
//   - Only(...)：只匹配列出的值，空值一律不匹配。
type Match struct {
	values []string
}

// Any 不做限制。language 为空时也视为匹配。
var Any = Match{}

// Only 只匹配给定的值。传入空列表会 panic：这只可能是声明写错，应在开发期暴露。
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

// FrameworkTable 按 framework 查 props。没有对应 key 表示该 framework 不支持。
type FrameworkTable[P Props] map[string]P

// LanguageTable 按 language 查 props。没有对应 key 表示该 language 不支持。
type LanguageTable[P Props] map[string]P

// Pair framework 与 language 的组合，用于参数确实随组合变化的能力（如 tRPC 不同语言的 admin 配置键）。
type Pair struct {
	Framework string
	Language  string
}

// PairTable 按 (framework, language) 组合查 props。
type PairTable[P Props] map[Pair]P

// Dimension 说明能力主要随哪个维度变化，用于测试中检查声明是否填写了对应的表。
// 解析时不读取该字段，所有表都会按优先级查找。
type Dimension int

const (
	// ByFramework 能力只随 framework 变化（默认）。
	ByFramework Dimension = iota
	// ByLanguage 能力只随 language 变化。
	ByLanguage
	// ByPair 能力随 framework 与 language 的组合变化，应尽量少用。
	ByPair
	// StackIndependent 能力与技术栈无关，所有 bkmsApp 应用都支持，包括以后新增的 framework。
	// 只适用于不经过框架代码实现的功能，如 plain 配置文件。
	StackIndependent
)

// Props 能力参数类型需要实现的接口。
//
// Public 返回允许对外暴露的字段；脚本集、解析器这类内部实现参数不得包含在内。
// 没有可暴露字段时返回 nil，对外 JSON 中将省略 props。
type Props interface {
	Public() any
}

// Def 一个能力的声明。
//
// 解析时按以下优先级查找，命中即返回：Pair → Framework → Language → Universal。
// 无论命中哪一项，结果都还要再满足 Languages 限制。
//
// 字段全部导出，是为了以后把各能力的声明移到其所在业务包时仍能在包外构造。
type Def[P Props] struct {
	// ID 能力标识。
	ID ID
	// Dimension 能力主要随哪个维度变化，仅供测试校验。
	Dimension Dimension
	// Framework 按 framework 声明的 props。
	Framework FrameworkTable[P]
	// Language 按 language 声明的 props，仅在 Framework 未命中时查找。
	Language LanguageTable[P]
	// Pair 按 (framework, language) 组合声明的 props，优先级最高。
	Pair PairTable[P]
	// Universal 对所有 framework 生效的 props，优先级最低；与技术栈无关的能力使用它。
	Universal *P
	// Languages 对解析结果的语言限制，零值表示不限制。例如 platformBuild 用 Only(go)
	// 表示即使 framework 支持，也只有 go 可用。
	Languages Match
}

// Mod 某个技术栈对某个能力的解析结果。
type Mod[P Props] struct {
	// ID 能力标识。
	ID ID
	// Supported 是否支持。
	Supported bool
	// ReasonCode 不支持的原因；Supported 为 true 时为空。
	ReasonCode ReasonCode
	// Props 执行器使用的参数；Supported 为 false 时为零值。
	Props P
}

// DefMeta 声明的摘要信息，不带泛型参数，便于把不同 props 类型的声明放进同一个切片里遍历。
type DefMeta struct {
	// ID 能力标识。
	ID ID
	// Dimension 能力主要随哪个维度变化。
	Dimension Dimension
	// Frameworks 声明中出现过的 framework，已去重并排序。
	Frameworks []string
	// Languages 声明中出现过的 language（含 Languages 限制），已去重并排序。
	Languages []string
	// Universal 是否声明了对所有 framework 生效的 props。
	Universal bool
}

// Meta 汇总声明中出现过的 framework 和 language。
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

// Resolve 计算该能力对指定技术栈是否支持及其参数。stack 应已通过技术栈校验。
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
		// framework 已声明，只是被 Languages 限制拦下。language 为空时无法确定是否满足限制，
		// 因此报 incomplete_language，而不是 unsupported_stack。
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

// missReason 所有表都未命中时，给出不支持的原因。
//
// 如果补上 language 后有可能命中（该 framework 有 Pair 条目，或声明了 Language 表），
// 说明是 language 缺失导致无法判断，返回 incomplete_language；否则返回 unsupported_stack。
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

// Require 支持时返回 props；不支持时返回包装了 ErrModUnsupported 的错误，错误信息包含能力 ID
// 与原因码。执行器应通过它获取 props，而不是直接读 Props 字段。
func (m Mod[P]) Require() (P, error) {
	if !m.Supported {
		var zero P
		return zero, errors.Wrapf(ErrModUnsupported, "%s: %s", m.ID, m.ReasonCode)
	}
	return m.Props, nil
}

// PublicModView 单个能力对外输出的 JSON 结构。
type PublicModView struct {
	// Supported 是否支持。
	Supported bool `json:"supported"`
	// ReasonCode 不支持的原因，支持时省略。
	ReasonCode ReasonCode `json:"reasonCode,omitempty"`
	// Props 对外参数，由各 props 类型的 Public 方法决定，无可暴露字段时省略。
	Props any `json:"props,omitempty"`
}

// view 把不同 props 类型的 Mod 统一转换成 PublicModView，供 ModSet 遍历时使用。
type view interface {
	view() PublicModView
}

func (m Mod[P]) view() PublicModView {
	if !m.Supported {
		return PublicModView{Supported: false, ReasonCode: m.ReasonCode}
	}
	return PublicModView{Supported: true, Props: m.Props.Public()}
}
