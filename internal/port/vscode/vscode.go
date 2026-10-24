// @(#)quiet / go / internal/port/vscode/vscode.go / 2026-10-24

// Package vscode maps quiet-core semantics onto Visual Studio Code's native
// workbench, syntax, semantic-token, and integrated-terminal color systems.
package vscode

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
	"github.com/chrismecca/quiet/internal/port/terminal"
	"github.com/chrismecca/quiet/internal/port/textmate"
)

type document struct {
	Type                 string            `json:"type"`
	SemanticHighlighting bool              `json:"semanticHighlighting"`
	Colors               map[string]string `json:"colors"`
	TokenColors          []tokenRule       `json:"tokenColors"`
	SemanticTokenColors  map[string]string `json:"semanticTokenColors"`
}

type tokenRule struct {
	Name     string        `json:"name"`
	Scope    []string      `json:"scope"`
	Settings tokenSettings `json:"settings"`
}

type tokenSettings struct {
	Foreground string `json:"foreground"`
}

type resolved struct {
	field     string
	surface   string
	structure string
	ink       string
	muted     string
	line      string
	mineral   string
	moss      string
	ochre     string
	rust      string
	selection string
	callout   string

	terminalBackground string
	terminalForeground string
	terminalCursor     string
	terminalSelection  string
	ansi               [16]string
}

// Encode writes one quiet palette as a native VS Code color-theme document.
//
// Workbench and editor semantics map directly from quiet-core. The integrated
// terminal deliberately consumes the generic terminal reference mapping so
// ANSI behavior remains identical to quiet's other terminal ports.
//
// The port controls color only. It does not impose typography, iconography,
// layout, fonts, font weight, italics, or other application styling.
func Encode(
	w io.Writer,
	p core.Palette,
) error {
	themeType, err := themeType(p.Name)
	if err != nil {
		return err
	}

	c, err := resolve(p)
	if err != nil {
		return err
	}

	tokenColors, err := textMateColors(p)
	if err != nil {
		return err
	}

	doc := document{
		Type:                 themeType,
		SemanticHighlighting: true,
		Colors:               workbenchColors(c),
		TokenColors:          tokenColors,
		SemanticTokenColors:  semanticColors(c),
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(doc); err != nil {
		return fmt.Errorf(
			"encode VS Code theme: %w",
			err,
		)
	}

	return nil
}

func workbenchColors(
	c resolved,
) map[string]string {
	result := map[string]string{
		// Global workbench.
		"foreground":            c.ink,
		"descriptionForeground": c.muted,
		"disabledForeground":    c.muted,
		"focusBorder":           c.mineral,
		"errorForeground":       c.rust,

		// Editor.
		"editor.background":                   c.field,
		"editor.foreground":                   c.ink,
		"editorCursor.foreground":             c.mineral,
		"editorCursor.background":             c.field,
		"editor.selectionBackground":          c.selection,
		"editor.inactiveSelectionBackground":  c.callout,
		"editor.lineHighlightBackground":      c.callout,
		"editorLineNumber.foreground":         c.muted,
		"editorLineNumber.activeForeground":   c.mineral,
		"editorWhitespace.foreground":         c.line,
		"editorIndentGuide.background1":       c.line,
		"editorIndentGuide.activeBackground1": c.mineral,

		"editorBracketMatch.background": c.selection,
		"editorBracketMatch.foreground": c.ink,
		"editorBracketMatch.border":     c.mineral,

		"editorBracketHighlight.foreground1":                  c.ink,
		"editorBracketHighlight.foreground2":                  c.ink,
		"editorBracketHighlight.foreground3":                  c.ink,
		"editorBracketHighlight.foreground4":                  c.ink,
		"editorBracketHighlight.foreground5":                  c.ink,
		"editorBracketHighlight.foreground6":                  c.ink,
		"editorBracketHighlight.unexpectedBracket.foreground": c.rust,

		"editorBracketPairGuide.background1": c.line,
		"editorBracketPairGuide.background2": c.line,
		"editorBracketPairGuide.background3": c.line,
		"editorBracketPairGuide.background4": c.line,
		"editorBracketPairGuide.background5": c.line,
		"editorBracketPairGuide.background6": c.line,

		"editorBracketPairGuide.activeBackground1": c.mineral,
		"editorBracketPairGuide.activeBackground2": c.mineral,
		"editorBracketPairGuide.activeBackground3": c.mineral,
		"editorBracketPairGuide.activeBackground4": c.mineral,
		"editorBracketPairGuide.activeBackground5": c.mineral,
		"editorBracketPairGuide.activeBackground6": c.mineral,

		"editor.findMatchBackground":           c.selection,
		"editor.findMatchHighlightBackground":  c.callout,
		"editor.wordHighlightBackground":       c.callout,
		"editor.wordHighlightStrongBackground": c.selection,

		// Diagnostics.
		"editorError.foreground":   c.rust,
		"editorWarning.foreground": c.ochre,
		"editorInfo.foreground":    c.mineral,
		"editorHint.foreground":    c.mineral,

		"problemsErrorIcon.foreground":   c.rust,
		"problemsWarningIcon.foreground": c.ochre,
		"problemsInfoIcon.foreground":    c.mineral,

		// Editor widgets and completion.
		"editorWidget.background": c.surface,
		"editorWidget.foreground": c.ink,
		"editorWidget.border":     c.line,

		"editorHoverWidget.background": c.surface,
		"editorHoverWidget.foreground": c.ink,
		"editorHoverWidget.border":     c.line,

		"editorSuggestWidget.background":          c.surface,
		"editorSuggestWidget.foreground":          c.ink,
		"editorSuggestWidget.border":              c.line,
		"editorSuggestWidget.selectedBackground":  c.selection,
		"editorSuggestWidget.highlightForeground": c.mineral,

		// Lists and trees.
		"list.activeSelectionBackground":   c.selection,
		"list.activeSelectionForeground":   c.ink,
		"list.inactiveSelectionBackground": c.callout,
		"list.inactiveSelectionForeground": c.ink,
		"list.hoverBackground":             c.callout,
		"list.hoverForeground":             c.ink,
		"list.focusBackground":             c.selection,
		"list.focusForeground":             c.ink,
		"list.highlightForeground":         c.mineral,
		"tree.indentGuidesStroke":          c.line,

		// Activity bar.
		"activityBar.background":         c.structure,
		"activityBar.foreground":         c.ink,
		"activityBar.inactiveForeground": c.muted,
		"activityBar.border":             c.line,
		"activityBarBadge.background":    c.mineral,
		"activityBarBadge.foreground":    c.field,

		// Side bar.
		"sideBar.background":              c.surface,
		"sideBar.foreground":              c.ink,
		"sideBar.border":                  c.line,
		"sideBarTitle.foreground":         c.ink,
		"sideBarSectionHeader.background": c.structure,
		"sideBarSectionHeader.foreground": c.ink,
		"sideBarSectionHeader.border":     c.line,

		// Tabs/editor groups.
		"editorGroupHeader.tabsBackground": c.surface,
		"editorGroup.border":               c.line,

		"tab.activeBackground":   c.field,
		"tab.activeForeground":   c.ink,
		"tab.inactiveBackground": c.surface,
		"tab.inactiveForeground": c.muted,
		"tab.border":             c.line,
		"tab.activeBorderTop":    c.mineral,

		// Panel.
		"panel.background":              c.surface,
		"panel.border":                  c.line,
		"panelTitle.activeForeground":   c.ink,
		"panelTitle.inactiveForeground": c.muted,
		"panelTitle.activeBorder":       c.mineral,

		// Status bar.
		"statusBar.background":         c.structure,
		"statusBar.foreground":         c.ink,
		"statusBar.border":             c.line,
		"statusBar.noFolderBackground": c.structure,
		"statusBar.noFolderForeground": c.ink,

		// Title bar.
		"titleBar.activeBackground":   c.surface,
		"titleBar.activeForeground":   c.ink,
		"titleBar.inactiveBackground": c.surface,
		"titleBar.inactiveForeground": c.muted,
		"titleBar.border":             c.line,

		// Inputs.
		"input.background":            c.field,
		"input.foreground":            c.ink,
		"input.border":                c.line,
		"input.placeholderForeground": c.muted,
		"inputOption.activeBorder":    c.mineral,

		"dropdown.background": c.surface,
		"dropdown.foreground": c.ink,
		"dropdown.border":     c.line,

		// Buttons and information indicators.
		"button.background":          c.structure,
		"button.foreground":          c.ink,
		"button.hoverBackground":     c.selection,
		"button.secondaryBackground": c.surface,
		"button.secondaryForeground": c.ink,

		"badge.background": c.mineral,
		"badge.foreground": c.field,

		"progressBar.background": c.mineral,

		// Source-control state.
		"gitDecoration.addedResourceForeground":       c.moss,
		"gitDecoration.untrackedResourceForeground":   c.moss,
		"gitDecoration.modifiedResourceForeground":    c.ochre,
		"gitDecoration.deletedResourceForeground":     c.rust,
		"gitDecoration.conflictingResourceForeground": c.rust,
		"gitDecoration.ignoredResourceForeground":     c.muted,

		// Integrated terminal.
		"terminal.background":          c.terminalBackground,
		"terminal.foreground":          c.terminalForeground,
		"terminalCursor.foreground":    c.terminalCursor,
		"terminalCursor.background":    c.terminalBackground,
		"terminal.selectionBackground": c.terminalSelection,
	}

	terminalIDs := [...]string{
		"terminal.ansiBlack",
		"terminal.ansiRed",
		"terminal.ansiGreen",
		"terminal.ansiYellow",
		"terminal.ansiBlue",
		"terminal.ansiMagenta",
		"terminal.ansiCyan",
		"terminal.ansiWhite",

		"terminal.ansiBrightBlack",
		"terminal.ansiBrightRed",
		"terminal.ansiBrightGreen",
		"terminal.ansiBrightYellow",
		"terminal.ansiBrightBlue",
		"terminal.ansiBrightMagenta",
		"terminal.ansiBrightCyan",
		"terminal.ansiBrightWhite",
	}

	for i, id := range terminalIDs {
		result[id] = c.ansi[i]
	}

	return result
}

// textMateColors serializes quiet's shared TextMate scope policy, so VS Code
// and tmTheme hosts cannot drift apart.
func textMateColors(
	p core.Palette,
) ([]tokenRule, error) {
	var rules []tokenRule

	for _, rule := range textmate.SourceRules(p) {
		hex, err := rule.Color.Hex()
		if err != nil {
			return nil, fmt.Errorf(
				"%s: %w",
				rule.Name,
				err,
			)
		}

		rules = append(rules, tokenRule{
			Name:  rule.Name,
			Scope: rule.Scopes,
			Settings: tokenSettings{
				Foreground: hex,
			},
		})
	}

	return rules, nil
}

func semanticColors(
	c resolved,
) map[string]string {
	return map[string]string{
		// Secondary information.
		"comment": c.muted,

		// Ordinary source/data.
		"parameter":  c.ink,
		"variable":   c.ink,
		"property":   c.ink,
		"enumMember": c.ink,
		"event":      c.ink,
		"function":   c.ink,
		"method":     c.ink,
		"string":     c.ink,
		"number":     c.ink,
		"regexp":     c.ink,

		// Language structure.
		"namespace":     c.mineral,
		"type":          c.mineral,
		"class":         c.mineral,
		"enum":          c.mineral,
		"interface":     c.mineral,
		"struct":        c.mineral,
		"typeParameter": c.mineral,
		"macro":         c.mineral,
		"label":         c.mineral,
		"keyword":       c.mineral,
		"operator":      c.mineral,
		"decorator":     c.mineral,

		// Semantic state.
		"*.deprecated": c.ochre,
	}
}

func resolve(
	p core.Palette,
) (resolved, error) {
	derived, err := core.Derive(p)
	if err != nil {
		return resolved{}, fmt.Errorf(
			"derive quiet roles: %w",
			err,
		)
	}

	term, err := terminal.Resolve(p)
	if err != nil {
		return resolved{}, fmt.Errorf(
			"resolve terminal palette: %w",
			err,
		)
	}

	var c resolved

	values := []struct {
		name  string
		value color.SRGB
		dst   *string
	}{
		{"Field", p.Anchors.Field, &c.field},
		{"Surface", p.Anchors.Surface, &c.surface},
		{"Structure", p.Anchors.Structure, &c.structure},
		{"Ink", p.Anchors.Ink, &c.ink},
		{"Muted", p.Anchors.Muted, &c.muted},
		{"Line", p.Anchors.Line, &c.line},
		{"Mineral", p.Anchors.Mineral, &c.mineral},
		{"Moss", p.Anchors.Moss, &c.moss},
		{"Ochre", p.Anchors.Ochre, &c.ochre},
		{"Rust", p.Anchors.Rust, &c.rust},
		{"Selection", derived.Selection, &c.selection},
		{"Callout", derived.Callout, &c.callout},

		{"Terminal Background", term.Background, &c.terminalBackground},
		{"Terminal Foreground", term.Text, &c.terminalForeground},
		{"Terminal Cursor", term.Cursor, &c.terminalCursor},
		{"Terminal Selection", term.Selection, &c.terminalSelection},
	}

	for _, item := range values {
		hex, err := item.value.Hex()
		if err != nil {
			return resolved{}, fmt.Errorf(
				"%s: %w",
				item.name,
				err,
			)
		}

		*item.dst = hex
	}

	for i, value := range term.ANSI {
		hex, err := value.Hex()
		if err != nil {
			return resolved{}, fmt.Errorf(
				"ANSI %d: %w",
				i,
				err,
			)
		}

		c.ansi[i] = hex
	}

	return c, nil
}

func themeType(
	name string,
) (string, error) {
	switch name {
	case "quiet-day":
		return "light", nil
	case "quiet-night":
		return "dark", nil
	default:
		return "", fmt.Errorf(
			"unsupported VS Code palette %q",
			name,
		)
	}
}
