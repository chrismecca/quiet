// @(#)quiet / go / internal/port/zed/zed.go / 2026-10-25

// Package zed maps quiet-core semantics onto a Zed theme family.
//
// A Zed theme family is one JSON file holding several themes. quiet's
// family holds quiet-day and quiet-night. Each theme sets interface colors
// by key, such as "editor.background", syntax colors by highlight name,
// such as "keyword", and the local user's cursor and selection.
//
// The port controls color. It sets no font style or weight.
package zed

import (
	"bufio"
	"fmt"
	"io"
	"strconv"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
	"github.com/chrismecca/quiet/internal/port/terminal"
)

const schema = "https://zed.dev/schema/themes/v0.2.0.json"

// transparent is the one non-quiet value the theme uses: Zed draws ghost
// elements and some borders with no fill.
const transparent = "transparent"

// entry assigns one style key a role.
type entry struct {
	key  string
	role string
}

// section groups entries in the generated file.
type section struct {
	comment string
	entries []entry
}

// ansiNames are Zed's names for the eight ANSI hues, in slot order.
var ansiNames = [8]string{
	"black",
	"red",
	"green",
	"yellow",
	"blue",
	"magenta",
	"cyan",
	"white",
}

// styleSections returns every interface key the theme sets.
//
// Backgrounds drawn under source set roles that keep every syntax color
// readable: Field, Structure, and the Wash line tiers. Zed draws word-level
// diff changes over the line wash, so those take the emphasis tier, which
// guarantees only Ink.
func styleSections() []section {
	terminal := []entry{
		{"terminal.background", "field"},
		{"terminal.ansi.background", "field"},
		{"terminal.foreground", "ink"},
		{"terminal.bright_foreground", "ink"},
		{"terminal.dim_foreground", "muted"},
	}

	for i, name := range ansiNames {
		terminal = append(
			terminal,
			entry{"terminal.ansi." + name, fmt.Sprintf("ansi-%d", i)},
			entry{"terminal.ansi.bright_" + name, fmt.Sprintf("ansi-%d", i+8)},
			entry{"terminal.ansi.dim_" + name, fmt.Sprintf("dim-%d", i)},
		)
	}

	return []section{
		{"Window and surfaces", []entry{
			{"background.appearance", "opaque"},
			{"background", "structure"},
			{"surface.background", "surface"},
			{"elevated_surface.background", "surface"},
			{"panel.background", "surface"},
			{"panel.overlay_background", "surface"},
			{"panel.overlay_hover", "callout"},
			{"status_bar.background", "structure"},
			{"title_bar.background", "structure"},
			{"title_bar.inactive_background", "structure"},
			{"toolbar.background", "field"},
			{"tab_bar.background", "surface"},
			{"tab.inactive_background", "surface"},
			{"tab.active_background", "field"},
		}},
		{"Borders", []entry{
			{"border", "line"},
			{"border.variant", "line"},
			{"border.focused", "mineral"},
			{"border.selected", "mineral"},
			{"border.transparent", transparent},
			{"border.disabled", "line"},
			{"pane.focused_border", "mineral"},
			{"pane_group.border", "line"},
			{"panel.focused_border", "mineral"},
			{"panel.indent_guide", "line"},
			{"panel.indent_guide_active", "muted"},
			{"panel.indent_guide_hover", "muted"},
		}},
		{"Text and icons", []entry{
			{"text", "ink"},
			{"text.muted", "muted"},
			{"text.placeholder", "muted"},
			{"text.disabled", "muted"},
			{"text.accent", "mineral"},
			{"icon", "ink"},
			{"icon.muted", "muted"},
			{"icon.disabled", "muted"},
			{"icon.placeholder", "muted"},
			{"icon.accent", "mineral"},
			{"link_text.hover", "mineral"},
		}},
		{"Elements", []entry{
			{"element.background", "structure"},
			{"element.hover", "callout"},
			{"element.active", "selection"},
			{"element.selected", "selection"},
			{"element.disabled", "surface"},
			{"element.selection_background", "selection"},
			{"ghost_element.background", transparent},
			{"ghost_element.hover", "callout"},
			{"ghost_element.active", "selection"},
			{"ghost_element.selected", "selection"},
			{"ghost_element.disabled", transparent},
			{"drop_target.background", "mineral-wash"},
			{"drop_target.border", "mineral"},
		}},
		{"Scrollbars and minimap", []entry{
			{"scrollbar.thumb.background", "line"},
			{"scrollbar.thumb.hover_background", "muted"},
			{"scrollbar.thumb.active_background", "mineral"},
			{"scrollbar.thumb.border", "line"},
			{"scrollbar.track.background", "field"},
			{"scrollbar.track.border", "line"},
			{"minimap.thumb.background", "line"},
			{"minimap.thumb.hover_background", "muted"},
			{"minimap.thumb.active_background", "mineral"},
			{"minimap.thumb.border", "line"},
		}},
		{"Editor", []entry{
			{"editor.foreground", "ink"},
			{"editor.background", "field"},
			{"editor.gutter.background", "field"},
			{"editor.subheader.background", "surface"},
			{"editor.active_line.background", "callout"},
			{"editor.highlighted_line.background", "mineral-wash"},
			{"editor.debugger_active_line.background", "ochre-wash"},
			{"editor.line_number", "muted"},
			{"editor.active_line_number", "mineral"},
			{"editor.hover_line_number", "ink"},
			{"editor.invisible", "line"},
			{"editor.wrap_guide", "line"},
			{"editor.active_wrap_guide", "muted"},
			{"editor.indent_guide", "line"},
			{"editor.indent_guide_active", "muted"},
			{"editor.code_lens.foreground", "muted"},
			{"editor.document_highlight.read_background", "mineral-wash"},
			{"editor.document_highlight.write_background", "ochre-wash"},
			{"editor.document_highlight.bracket_background", "selection"},
			{"search.match_background", "mineral-wash"},
			{"search.active_match_background", "ochre-wash"},
		}},
		{"Diffs and version control", []entry{
			{"editor.diff_hunk.added.background", "moss-wash"},
			{"editor.diff_hunk.added.hollow_background", "field"},
			{"editor.diff_hunk.added.hollow_border", "moss"},
			{"editor.diff_hunk.deleted.background", "rust-wash"},
			{"editor.diff_hunk.deleted.hollow_background", "field"},
			{"editor.diff_hunk.deleted.hollow_border", "rust"},
			{"version_control.added", "moss"},
			{"version_control.modified", "ochre"},
			{"version_control.deleted", "rust"},
			{"version_control.renamed", "moss"},
			{"version_control.conflict", "rust"},
			{"version_control.ignored", "muted"},
			{"version_control.word_added", "moss-wash-emphasis"},
			{"version_control.word_deleted", "rust-wash-emphasis"},
			{"version_control.conflict_marker.ours", "mineral-wash"},
			{"version_control.conflict_marker.theirs", "moss-wash"},
		}},
		{"Status", []entry{
			{"error", "rust"},
			{"error.background", "rust-wash"},
			{"error.border", "rust"},
			{"warning", "ochre"},
			{"warning.background", "ochre-wash"},
			{"warning.border", "ochre"},
			{"info", "mineral"},
			{"info.background", "mineral-wash"},
			{"info.border", "mineral"},
			{"hint", "mineral"},
			{"hint.background", "mineral-wash"},
			{"hint.border", "mineral"},
			{"success", "moss"},
			{"success.background", "moss-wash"},
			{"success.border", "moss"},
			{"conflict", "rust"},
			{"conflict.background", "rust-wash"},
			{"conflict.border", "rust"},
			{"created", "moss"},
			{"created.background", "moss-wash"},
			{"created.border", "moss"},
			{"deleted", "rust"},
			{"deleted.background", "rust-wash"},
			{"deleted.border", "rust"},
			{"modified", "ochre"},
			{"modified.background", "ochre-wash"},
			{"modified.border", "ochre"},
			{"renamed", "moss"},
			{"renamed.background", "moss-wash"},
			{"renamed.border", "moss"},
			{"hidden", "muted"},
			{"hidden.background", "structure"},
			{"hidden.border", "line"},
			{"ignored", "muted"},
			{"ignored.background", "structure"},
			{"ignored.border", "line"},
			{"predictive", "muted"},
			{"predictive.background", "structure"},
			{"predictive.border", "line"},
			{"unreachable", "muted"},
			{"unreachable.background", "structure"},
			{"unreachable.border", "line"},
			{"debugger.accent", "rust"},
		}},
		{"Modal editing: normal Mineral, insert Moss, visual Ochre, replace Rust", []entry{
			{"vim.normal.background", "mineral"},
			{"vim.normal.foreground", "field"},
			{"vim.insert.background", "moss"},
			{"vim.insert.foreground", "field"},
			{"vim.replace.background", "rust"},
			{"vim.replace.foreground", "field"},
			{"vim.visual.background", "ochre"},
			{"vim.visual.foreground", "field"},
			{"vim.visual_line.background", "ochre"},
			{"vim.visual_line.foreground", "field"},
			{"vim.visual_block.background", "ochre"},
			{"vim.visual_block.foreground", "field"},
			{"vim.helix_normal.background", "mineral"},
			{"vim.helix_normal.foreground", "field"},
			{"vim.helix_select.background", "ochre"},
			{"vim.helix_select.foreground", "field"},
			{"vim.helix_jump_label.foreground", "mineral"},
			{"vim.yank.background", "mineral-wash"},
		}},
		{"Terminal, from quiet's generic terminal palette", terminal},
	}
}

// syntax assigns Zed's highlight names their quiet roles: ordinary source is
// Ink, comments Muted, and language structure Mineral.
var syntax = []entry{
	{"attribute", "mineral"},
	{"boolean", "ink"},
	{"comment", "muted"},
	{"comment.doc", "muted"},
	{"constant", "ink"},
	{"constructor", "ink"},
	{"diff.minus", "rust"},
	{"diff.plus", "moss"},
	{"embedded", "ink"},
	{"emphasis", "ink"},
	{"emphasis.strong", "ink"},
	{"enum", "mineral"},
	{"function", "ink"},
	{"function.builtin", "ink"},
	{"hint", "muted"},
	{"keyword", "mineral"},
	{"label", "mineral"},
	{"link_text", "mineral"},
	{"link_uri", "mineral"},
	{"namespace", "mineral"},
	{"number", "ink"},
	{"operator", "mineral"},
	{"predictive", "muted"},
	{"preproc", "mineral"},
	{"primary", "ink"},
	{"property", "ink"},
	{"punctuation", "ink"},
	{"punctuation.bracket", "ink"},
	{"punctuation.delimiter", "ink"},
	{"punctuation.list_marker", "mineral"},
	{"punctuation.markup", "mineral"},
	{"punctuation.special", "ink"},
	{"selector", "mineral"},
	{"selector.pseudo", "mineral"},
	{"string", "ink"},
	{"string.escape", "ink"},
	{"string.regex", "ink"},
	{"string.special", "ink"},
	{"string.special.symbol", "ink"},
	{"tag", "mineral"},
	{"text.literal", "ink"},
	{"title", "mineral"},
	{"type", "mineral"},
	{"variable", "ink"},
	{"variable.parameter", "ink"},
	{"variable.special", "ink"},
	{"variant", "ink"},
}

// roles resolves every role the theme uses to its JSON value.
func roles(p core.Palette) (map[string]string, error) {
	derived, err := core.Derive(p)
	if err != nil {
		return nil, fmt.Errorf(
			"derive quiet roles: %w",
			err,
		)
	}

	term, err := terminal.Resolve(p)
	if err != nil {
		return nil, fmt.Errorf(
			"resolve terminal palette: %w",
			err,
		)
	}

	a := p.Anchors
	wash := derived.Wash

	colors := map[string]color.SRGB{
		"field":                 a.Field,
		"surface":               a.Surface,
		"structure":             a.Structure,
		"ink":                   a.Ink,
		"muted":                 a.Muted,
		"line":                  a.Line,
		"mineral":               a.Mineral,
		"moss":                  a.Moss,
		"ochre":                 a.Ochre,
		"rust":                  a.Rust,
		"selection":             derived.Selection,
		"callout":               derived.Callout,
		"mineral-wash":          wash.Mineral.Line,
		"moss-wash":             wash.Moss.Line,
		"ochre-wash":            wash.Ochre.Line,
		"rust-wash":             wash.Rust.Line,
		"mineral-wash-emphasis": wash.Mineral.Emphasis,
		"moss-wash-emphasis":    wash.Moss.Emphasis,
		"ochre-wash-emphasis":   wash.Ochre.Emphasis,
		"rust-wash-emphasis":    wash.Rust.Emphasis,
	}

	// Dim terminal colors use quiet's Subtle rule: each normal color
	// composited over the terminal background.
	for i := range 8 {
		colors[fmt.Sprintf("ansi-%d", i)] = term.ANSI[i]
		colors[fmt.Sprintf("ansi-%d", i+8)] = term.ANSI[i+8]

		dim, err := color.CompositeLegacySRGB(
			term.ANSI[i],
			term.Background,
			core.SubtleAlpha,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"dim ANSI %d: %w",
				i,
				err,
			)
		}

		colors[fmt.Sprintf("dim-%d", i)] = dim
	}

	result := map[string]string{
		transparent: "#00000000",
		"opaque":    "opaque",
	}

	for name, c := range colors {
		hex, err := c.Hex()
		if err != nil {
			return nil, fmt.Errorf(
				"%s: %w",
				name,
				err,
			)
		}

		result[name] = hex
	}

	return result, nil
}

func appearance(p core.Palette) (string, error) {
	switch p.Mode {
	case core.ModeDay:
		return "light", nil

	case core.ModeNight:
		return "dark", nil

	default:
		return "", fmt.Errorf(
			"unsupported Zed palette mode %q",
			p.Mode,
		)
	}
}

// Encode writes the quiet theme family, holding one theme per palette.
func Encode(
	w io.Writer,
	palettes ...core.Palette,
) error {
	out := bufio.NewWriter(w)

	fmt.Fprintf(out, "{\n")
	fmt.Fprintf(out, "  \"$schema\": %s,\n", strconv.Quote(schema))
	fmt.Fprintf(out, "  \"name\": \"quiet\",\n")
	fmt.Fprintf(out, "  \"author\": \"Chris Mecca\",\n")
	fmt.Fprintf(out, "  \"themes\": [\n")

	for i, p := range palettes {
		if err := writeTheme(out, p); err != nil {
			return fmt.Errorf(
				"%s: %w",
				p.Name,
				err,
			)
		}

		if i < len(palettes)-1 {
			fmt.Fprintf(out, ",")
		}

		fmt.Fprintf(out, "\n")
	}

	fmt.Fprintf(out, "  ]\n}\n")

	return out.Flush()
}

func writeTheme(out *bufio.Writer, p core.Palette) error {
	mode, err := appearance(p)
	if err != nil {
		return err
	}

	values, err := roles(p)
	if err != nil {
		return err
	}

	value := func(role string) (string, error) {
		v, ok := values[role]
		if !ok {
			return "", fmt.Errorf("unknown role %q", role)
		}

		return strconv.Quote(v), nil
	}

	fmt.Fprintf(out, "    {\n")
	fmt.Fprintf(out, "      \"name\": %s,\n", strconv.Quote(p.Name))
	fmt.Fprintf(out, "      \"appearance\": %s,\n", strconv.Quote(mode))
	fmt.Fprintf(out, "      \"style\": {\n")

	for _, s := range styleSections() {
		for _, e := range s.entries {
			v, err := value(e.role)
			if err != nil {
				return fmt.Errorf("%s: %w", e.key, err)
			}

			fmt.Fprintf(out, "        %s: %s,\n", strconv.Quote(e.key), v)
		}
	}

	ink, err := value("ink")
	if err != nil {
		return err
	}

	mineral, err := value("mineral")
	if err != nil {
		return err
	}

	selection, err := value("selection")
	if err != nil {
		return err
	}

	// Bracket colorization takes one accent, so brackets stay Ink. The
	// local user is Mineral; collaborators keep Zed's own colors, which
	// identify people rather than carry meaning.
	fmt.Fprintf(out, "        \"accents\": [%s],\n", ink)
	fmt.Fprintf(out, "        \"players\": [\n")
	fmt.Fprintf(
		out,
		"          { \"cursor\": %s, \"background\": %s, \"selection\": %s }\n",
		mineral,
		mineral,
		selection,
	)
	fmt.Fprintf(out, "        ],\n")
	fmt.Fprintf(out, "        \"syntax\": {\n")

	for i, e := range syntax {
		v, err := value(e.role)
		if err != nil {
			return fmt.Errorf("syntax %s: %w", e.key, err)
		}

		comma := ","
		if i == len(syntax)-1 {
			comma = ""
		}

		fmt.Fprintf(
			out,
			"          %s: { \"color\": %s }%s\n",
			strconv.Quote(e.key),
			v,
			comma,
		)
	}

	fmt.Fprintf(out, "        }\n")
	fmt.Fprintf(out, "      }\n")
	fmt.Fprintf(out, "    }")

	return nil
}
