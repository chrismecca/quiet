// @(#)quiet / go / internal/port/textmate/textmate.go / 2026-10-24

// Package textmate defines quiet's TextMate scope policy.
//
// Every port whose host highlights source with TextMate scopes uses these
// rules, so VS Code, tmTheme hosts, and anything built on them share one
// syntax hierarchy: Ink for ordinary source, Muted for comments, Mineral for
// language structure. Moss, Ochre, and Rust are reserved for state such as
// invalid code and diffs.
package textmate

import (
	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
)

// Rule assigns one quiet color to a group of TextMate scopes.
type Rule struct {
	Name   string
	Scopes []string
	Color  color.SRGB
}

// SourceRules returns the syntax rules for source code and markup.
func SourceRules(p core.Palette) []Rule {
	a := p.Anchors

	return []Rule{
		{
			Name: "Comments",
			Scopes: []string{
				"comment",
				"punctuation.definition.comment",
			},
			Color: a.Muted,
		},
		{
			Name: "Ordinary source",
			Scopes: []string{
				"constant",
				"string",
				"variable",
				"variable.parameter",
				"entity.name.function",
				"entity.name.function.member",
				"entity.name.method",
				"entity.other.attribute-name",
				"punctuation",
			},
			Color: a.Ink,
		},
		{
			Name: "Language structure",
			Scopes: []string{
				"keyword",
				"keyword.operator",
				"storage.type",
				"storage.modifier",
				"entity.name.type",
				"entity.other.inherited-class",
				"support.type",
				"support.class",
				"support.function",
				"meta.preprocessor",
				"keyword.control.directive",
				"punctuation.definition.directive",
			},
			Color: a.Mineral,
		},
		{
			Name: "Markup structure",
			Scopes: []string{
				"markup.heading",
				"markup.link",
				"entity.name.tag",
			},
			Color: a.Mineral,
		},
		{
			Name: "Markup quotation",
			Scopes: []string{
				"markup.quote",
			},
			Color: a.Muted,
		},
		{
			Name: "Invalid",
			Scopes: []string{
				"invalid.illegal",
			},
			Color: a.Rust,
		},
		{
			Name: "Deprecated",
			Scopes: []string{
				"invalid.deprecated",
			},
			Color: a.Ochre,
		},
	}
}

// DiffRules returns the rules for unified-diff scopes. They match the Vim
// and Neovim diff mapping: Moss for insertions, Ochre for changes, Rust for
// deletions, and Mineral for diff structure such as headers and ranges.
func DiffRules(p core.Palette) []Rule {
	a := p.Anchors

	return []Rule{
		{
			Name: "Diff structure",
			Scopes: []string{
				"meta.diff.header",
				"meta.diff.range",
				"meta.diff.index",
			},
			Color: a.Mineral,
		},
		{
			Name: "Diff insertion",
			Scopes: []string{
				"markup.inserted",
			},
			Color: a.Moss,
		},
		{
			Name: "Diff change",
			Scopes: []string{
				"markup.changed",
			},
			Color: a.Ochre,
		},
		{
			Name: "Diff deletion",
			Scopes: []string{
				"markup.deleted",
			},
			Color: a.Rust,
		},
	}
}
