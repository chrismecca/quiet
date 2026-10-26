// @(#)quiet / go / internal/port/helix/helix.go / 2026-10-25

// Package helix maps quiet-core semantics onto Helix editor themes.
//
// A Helix theme styles scopes: syntax scopes from tree-sitter captures, such
// as "keyword", and interface scopes, such as "ui.statusline". A scope
// without its own style falls back to its parent, so "keyword.control"
// takes the style of "keyword".
//
// The port controls color. It keeps the cues that mark state in Helix's own
// default theme, such as curly underlines on diagnostics and struck-out
// deprecated names, and drops typographic emphasis such as bold and italic.
package helix

import (
	"fmt"
	"io"
	"strings"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
)

// style is one scope's style. Colors are palette names.
type style struct {
	scope string
	fg    string
	bg    string

	// curl colors a curly underline.
	curl string

	// modifier is a non-color cue Helix's default theme gives the scope.
	modifier string
}

// styles returns every scope the theme sets.
//
// Backgrounds drawn under source set no foreground when they keep every
// syntax color readable: Wash line tiers and Structure.
func styles() []style {
	return []style{
		// Syntax. Ordinary source is Ink, comments Muted, and language
		// structure Mineral.
		{scope: "attribute", fg: "mineral"},
		{scope: "type", fg: "mineral"},
		{scope: "type.enum.variant", fg: "ink"},
		{scope: "constructor", fg: "ink"},
		{scope: "constant", fg: "ink"},
		{scope: "string", fg: "ink"},
		{scope: "comment", fg: "muted"},
		{scope: "variable", fg: "ink"},
		{scope: "label", fg: "mineral"},
		{scope: "punctuation", fg: "ink"},
		{scope: "keyword", fg: "mineral"},
		{scope: "operator", fg: "mineral"},
		{scope: "function", fg: "ink"},
		{scope: "function.macro", fg: "mineral"},
		{scope: "tag", fg: "mineral"},
		{scope: "namespace", fg: "mineral"},
		{scope: "special", fg: "mineral"},

		{scope: "markup.heading", fg: "mineral"},
		{scope: "markup.list", fg: "mineral"},
		{scope: "markup.link", fg: "mineral"},
		{scope: "markup.link.url", fg: "mineral", modifier: "underlined"},
		{scope: "markup.quote", fg: "muted"},
		{scope: "markup.raw", fg: "ink"},

		{scope: "diff.plus", fg: "moss"},
		{scope: "diff.minus", fg: "rust"},
		{scope: "diff.delta", fg: "ochre"},

		// Editing area.
		{scope: "ui.background", bg: "field"},
		{scope: "ui.background.separator", fg: "line"},
		{scope: "ui.text", fg: "ink"},
		{scope: "ui.text.focus", fg: "ink", bg: "selection"},
		{scope: "ui.text.inactive", fg: "muted"},
		{scope: "ui.text.info", fg: "ink"},
		{scope: "ui.text.directory", fg: "mineral"},

		{scope: "ui.cursor", fg: "field", bg: "mineral"},
		{scope: "ui.cursor.primary", fg: "field", bg: "mineral"},
		{scope: "ui.cursor.match", fg: "ink", bg: "selection"},
		{scope: "ui.selection", fg: "ink", bg: "selection"},
		{scope: "ui.selection.primary", fg: "ink", bg: "selection"},
		{scope: "ui.cursorline.primary", bg: "callout"},
		{scope: "ui.cursorline.secondary", bg: "callout"},
		{scope: "ui.cursorcolumn.primary", bg: "callout"},
		{scope: "ui.cursorcolumn.secondary", bg: "callout"},
		{scope: "ui.highlight", bg: "mineral-wash"},
		{scope: "ui.highlight.frameline", bg: "ochre-wash"},
		{scope: "tabstop", bg: "structure"},

		{scope: "ui.gutter", bg: "field"},
		{scope: "ui.gutter.selected", bg: "callout"},
		{scope: "ui.linenr", fg: "muted"},
		{scope: "ui.linenr.selected", fg: "mineral"},
		{scope: "ui.debug.breakpoint", fg: "rust"},
		{scope: "ui.debug.active", fg: "ochre"},

		{scope: "ui.virtual", fg: "line"},
		{scope: "ui.virtual.ruler", bg: "surface"},
		{scope: "ui.virtual.whitespace", fg: "line"},
		{scope: "ui.virtual.indent-guide", fg: "line"},
		{scope: "ui.virtual.wrap", fg: "line"},
		{scope: "ui.virtual.inlay-hint", fg: "muted", bg: "surface"},
		{scope: "ui.virtual.jump-label", fg: "field", bg: "mineral"},

		// Status and buffer lines. The mode colors apply only when
		// editor.color-modes is enabled.
		{scope: "ui.statusline", fg: "ink", bg: "structure"},
		{scope: "ui.statusline.inactive", fg: "muted", bg: "surface"},
		{scope: "ui.statusline.normal", fg: "field", bg: "mineral"},
		{scope: "ui.statusline.insert", fg: "field", bg: "moss"},
		{scope: "ui.statusline.select", fg: "field", bg: "ochre"},
		{scope: "ui.statusline.separator", fg: "line"},
		{scope: "ui.bufferline", fg: "muted", bg: "surface"},
		{scope: "ui.bufferline.active", fg: "ink", bg: "field"},
		{scope: "ui.bufferline.background", bg: "structure"},

		// Popups, menus, and pickers.
		{scope: "ui.window", fg: "line"},
		{scope: "ui.popup", fg: "ink", bg: "surface"},
		{scope: "ui.popup.info", fg: "ink", bg: "surface"},
		{scope: "ui.help", fg: "ink", bg: "surface"},
		{scope: "ui.menu", fg: "ink", bg: "surface"},
		{scope: "ui.menu.selected", fg: "ink", bg: "selection"},
		{scope: "ui.menu.scroll", fg: "line", bg: "structure"},
		{scope: "ui.picker.header", fg: "ink", bg: "structure"},
		{scope: "ui.picker.header.column", fg: "muted"},
		{scope: "ui.picker.header.column.active", fg: "ink"},

		// Diagnostics: gutter markers are colored, and the editing area
		// keeps Helix's curly underline in the diagnostic's color.
		{scope: "error", fg: "rust"},
		{scope: "warning", fg: "ochre"},
		{scope: "info", fg: "mineral"},
		{scope: "hint", fg: "mineral"},
		{scope: "diagnostic.error", curl: "rust"},
		{scope: "diagnostic.warning", curl: "ochre"},
		{scope: "diagnostic.info", curl: "mineral"},
		{scope: "diagnostic.hint", curl: "mineral"},
		{scope: "diagnostic.unnecessary", fg: "muted"},
		{scope: "diagnostic.deprecated", modifier: "crossed_out"},
	}
}

// palette returns the theme's palette entries, in order.
func palette(p core.Palette) ([][2]string, error) {
	derived, err := core.Derive(p)
	if err != nil {
		return nil, fmt.Errorf(
			"derive quiet roles: %w",
			err,
		)
	}

	a := p.Anchors

	values := []struct {
		name  string
		value color.SRGB
	}{
		{"field", a.Field},
		{"surface", a.Surface},
		{"structure", a.Structure},
		{"ink", a.Ink},
		{"muted", a.Muted},
		{"line", a.Line},
		{"mineral", a.Mineral},
		{"moss", a.Moss},
		{"ochre", a.Ochre},
		{"rust", a.Rust},
		{"selection", derived.Selection},
		{"callout", derived.Callout},
		{"mineral-wash", derived.Wash.Mineral.Line},
		{"ochre-wash", derived.Wash.Ochre.Line},
	}

	entries := make([][2]string, len(values))

	for i, v := range values {
		hex, err := v.value.Hex()
		if err != nil {
			return nil, fmt.Errorf(
				"%s: %w",
				v.name,
				err,
			)
		}

		entries[i] = [2]string{v.name, hex}
	}

	return entries, nil
}

// Encode writes one quiet palette as a Helix theme.
func Encode(
	w io.Writer,
	p core.Palette,
) error {
	entries, err := palette(p)
	if err != nil {
		return err
	}

	var b strings.Builder

	fmt.Fprintf(
		&b,
		"# Generated by quiet port helix. Do not edit.\n"+
			"#\n"+
			"# %s\n\n",
		p.Name,
	)

	for _, s := range styles() {
		var attrs []string

		if s.fg != "" {
			attrs = append(attrs, fmt.Sprintf("fg = %q", s.fg))
		}

		if s.bg != "" {
			attrs = append(attrs, fmt.Sprintf("bg = %q", s.bg))
		}

		if s.curl != "" {
			attrs = append(
				attrs,
				fmt.Sprintf("underline = { color = %q, style = \"curl\" }", s.curl),
			)
		}

		if s.modifier != "" {
			attrs = append(attrs, fmt.Sprintf("modifiers = [%q]", s.modifier))
		}

		fmt.Fprintf(
			&b,
			"%q = { %s }\n",
			s.scope,
			strings.Join(attrs, ", "),
		)
	}

	b.WriteString("\n[palette]\n")

	for _, e := range entries {
		fmt.Fprintf(&b, "%s = %q\n", e[0], e[1])
	}

	if _, err := io.WriteString(
		w,
		b.String(),
	); err != nil {
		return fmt.Errorf(
			"write Helix theme: %w",
			err,
		)
	}

	return nil
}
