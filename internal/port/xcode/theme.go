// @(#)quiet / go / internal/port/xcode/theme.go / 2026-10-24

// Package xcode maps quiet-core semantics onto Xcode's native workspace
// color-theme model.
package xcode

import (
	"fmt"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
)

const (
	overrideColor            = "color"
	overrideColorWithOpacity = "colorWithOpacity"
)

// Override maps one native Xcode color role to a resolved quiet color.
type Override struct {
	Role  string
	Kind  string
	Color color.SRGB
}

// Theme is the resolved host-level representation consumed by the Xcode
// native theme serializer.
//
// These roles belong to Xcode. They are not additional quiet-core semantics.
type Theme struct {
	Name        string
	ColorScheme string

	Background color.SRGB
	Primary    color.SRGB
	Secondary  color.SRGB

	Overrides []Override
}

// Resolve maps one canonical quiet palette onto Xcode's native workspace
// theme vocabulary.
func Resolve(
	p core.Palette,
) (Theme, error) {
	var colorScheme string

	switch fmt.Sprint(p.Mode) {
	case "day":
		colorScheme = "light"

	case "night":
		colorScheme = "dark"

	default:
		return Theme{}, fmt.Errorf(
			"unsupported quiet mode %q",
			fmt.Sprint(p.Mode),
		)
	}

	derived, err := core.Derive(p)
	if err != nil {
		return Theme{}, fmt.Errorf(
			"derive quiet roles: %w",
			err,
		)
	}

	a := p.Anchors

	return Theme{
		Name:        p.Name,
		ColorScheme: colorScheme,

		Background: a.Field,

		// Xcode derives additional workspace colors from its primary and
		// secondary palette seeds. Mineral represents information, focus,
		// and active state; Field remains the environmental base.
		Primary:   a.Mineral,
		Secondary: a.Field,

		Overrides: []Override{
			{
				Role:  "plainText",
				Kind:  overrideColor,
				Color: a.Ink,
			},
			{
				Role:  "comment",
				Kind:  overrideColor,
				Color: a.Muted,
			},
			{
				Role:  "commentEmphasized",
				Kind:  overrideColor,
				Color: a.Ink,
			},
			{
				Role:  "markdown",
				Kind:  overrideColor,
				Color: a.Muted,
			},
			{
				Role:  "markdownEmphasized",
				Kind:  overrideColor,
				Color: a.Ink,
			},
			{
				Role:  "keyword",
				Kind:  overrideColor,
				Color: a.Mineral,
			},
			{
				Role:  "keyword.attribute",
				Kind:  overrideColor,
				Color: a.Mineral,
			},
			{
				Role:  "string",
				Kind:  overrideColor,
				Color: a.Ink,
			},
			{
				Role:  "string.regex",
				Kind:  overrideColor,
				Color: a.Ink,
			},
			{
				Role:  "number",
				Kind:  overrideColor,
				Color: a.Ink,
			},
			{
				Role:  "number.character",
				Kind:  overrideColor,
				Color: a.Ink,
			},
			{
				Role:  "link",
				Kind:  overrideColor,
				Color: a.Mineral,
			},
			{
				Role:  "preprocessor",
				Kind:  overrideColor,
				Color: a.Mineral,
			},
			{
				Role:  "preprocessor.macro",
				Kind:  overrideColor,
				Color: a.Mineral,
			},
			{
				Role:  "preprocessor.macroSystem",
				Kind:  overrideColor,
				Color: a.Muted,
			},
			{
				Role:  "typeDeclaration",
				Kind:  overrideColor,
				Color: a.Mineral,
			},
			{
				Role:  "memberDeclaration",
				Kind:  overrideColor,
				Color: a.Ink,
			},
			{
				Role:  "projectType",
				Kind:  overrideColor,
				Color: a.Mineral,
			},
			{
				Role:  "projectType.class",
				Kind:  overrideColor,
				Color: a.Mineral,
			},
			{
				Role:  "projectType.type",
				Kind:  overrideColor,
				Color: a.Mineral,
			},
			{
				Role:  "projectMember",
				Kind:  overrideColor,
				Color: a.Ink,
			},
			{
				Role:  "projectMember.function",
				Kind:  overrideColor,
				Color: a.Ink,
			},
			{
				Role:  "projectMember.constant",
				Kind:  overrideColor,
				Color: a.Ink,
			},
			{
				Role:  "projectMember.variable",
				Kind:  overrideColor,
				Color: a.Ink,
			},
			{
				Role:  "otherType",
				Kind:  overrideColor,
				Color: a.Muted,
			},
			{
				Role:  "otherType.class",
				Kind:  overrideColor,
				Color: a.Muted,
			},
			{
				Role:  "otherType.type",
				Kind:  overrideColor,
				Color: a.Muted,
			},
			{
				Role:  "otherMember",
				Kind:  overrideColor,
				Color: a.Muted,
			},
			{
				Role:  "otherMember.function",
				Kind:  overrideColor,
				Color: a.Muted,
			},
			{
				Role:  "otherMember.constant",
				Kind:  overrideColor,
				Color: a.Muted,
			},
			{
				Role:  "otherMember.variable",
				Kind:  overrideColor,
				Color: a.Muted,
			},
			{
				Role:  "selectedTextBackgroundColor",
				Kind:  overrideColorWithOpacity,
				Color: derived.Selection,
			},
			{
				Role:  "insertionPointColor",
				Kind:  overrideColor,
				Color: a.Mineral,
			},
			{
				Role:  "currentLineHighlight",
				Kind:  overrideColorWithOpacity,
				Color: derived.Callout,
			},
			{
				Role:  "invisibles",
				Kind:  overrideColor,
				Color: a.Line,
			},
			{
				Role:  "diffAddition",
				Kind:  overrideColor,
				Color: a.Moss,
			},
			{
				Role:  "diffDeletion",
				Kind:  overrideColor,
				Color: a.Rust,
			},
			{
				Role:  "debugger",
				Kind:  overrideColor,
				Color: a.Mineral,
			},
		},
	}, nil
}
