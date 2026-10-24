// @(#)quiet / go / internal/port/tmux/tmux.go / 2026-10-24

// Package tmux maps quiet-core semantics onto tmux interface chrome.
//
// The port is deliberately a color overlay. It does not configure tmux
// content, layout, key bindings, formats, plugins, or terminal behavior.
package tmux

import (
	"fmt"
	"io"
	"strings"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
)

// Theme is the resolved quiet color vocabulary used by tmux.
//
// Values are serialized sRGB colors because tmux accepts #RRGGBB directly.
// This is a host-level representation; it does not introduce new quiet-core
// roles.
type Theme struct {
	Name         string
	ArtifactPath string

	Field     string
	Surface   string
	Structure string
	Ink       string
	Muted     string
	Line      string

	Mineral string
	Ochre   string
	Rust    string

	Selection string
	Callout   string
}

// Resolve maps one canonical quiet palette onto tmux interface semantics.
func Resolve(p core.Palette) (Theme, error) {
	derived, err := core.Derive(p)
	if err != nil {
		return Theme{}, fmt.Errorf(
			"derive quiet roles: %w",
			err,
		)
	}

	artifactPath, err := artifactPath(p.Name)
	if err != nil {
		return Theme{}, err
	}

	theme := Theme{
		Name:         p.Name,
		ArtifactPath: artifactPath,
	}

	values := []struct {
		name  string
		value color.SRGB
		dst   *string
	}{
		{
			name:  "Field",
			value: p.Anchors.Field,
			dst:   &theme.Field,
		},
		{
			name:  "Surface",
			value: p.Anchors.Surface,
			dst:   &theme.Surface,
		},
		{
			name:  "Structure",
			value: p.Anchors.Structure,
			dst:   &theme.Structure,
		},
		{
			name:  "Ink",
			value: p.Anchors.Ink,
			dst:   &theme.Ink,
		},
		{
			name:  "Muted",
			value: p.Anchors.Muted,
			dst:   &theme.Muted,
		},
		{
			name:  "Line",
			value: p.Anchors.Line,
			dst:   &theme.Line,
		},
		{
			name:  "Mineral",
			value: p.Anchors.Mineral,
			dst:   &theme.Mineral,
		},
		{
			name:  "Ochre",
			value: p.Anchors.Ochre,
			dst:   &theme.Ochre,
		},
		{
			name:  "Rust",
			value: p.Anchors.Rust,
			dst:   &theme.Rust,
		},
		{
			name:  "Selection",
			value: derived.Selection,
			dst:   &theme.Selection,
		},
		{
			name:  "Callout",
			value: derived.Callout,
			dst:   &theme.Callout,
		},
	}

	for _, item := range values {
		value, err := item.value.Hex()
		if err != nil {
			return Theme{}, fmt.Errorf(
				"%s: %w",
				item.name,
				err,
			)
		}

		*item.dst = value
	}

	return theme, nil
}

// role names a resolved Theme color. The overlay table refers to roles
// rather than hexadecimal values, so the tmux mapping reads as quiet
// semantics and cannot drift out of step with a positional argument list.
type role int

const (
	field role = iota
	surface
	structure
	ink
	muted
	line
	mineral
	ochre
	rust
	selection
	callout
)

// paint is a tmux color expression: a role, or a format conditional that
// tmux evaluates at draw time.
type paint interface {
	expression(t Theme) string
}

func (r role) expression(t Theme) string {
	return [...]string{
		field:     t.Field,
		surface:   t.Surface,
		structure: t.Structure,
		ink:       t.Ink,
		muted:     t.Muted,
		line:      t.Line,
		mineral:   t.Mineral,
		ochre:     t.Ochre,
		rust:      t.Rust,
		selection: t.Selection,
		callout:   t.Callout,
	}[r]
}

// conditional selects between paints with a tmux format conditional, so
// pane and window state change color without hooks or scripts.
type conditional struct {
	condition string
	then      paint
	otherwise paint
}

func (c conditional) expression(t Theme) string {
	return "#{?" + c.condition +
		"," + c.then.expression(t) +
		"," + c.otherwise.expression(t) + "}"
}

// style is one tmux style value. Unset components are omitted; set ones are
// written in bg, fg, fill order.
type style struct {
	bg   paint
	fg   paint
	fill paint
}

func (s style) expression(t Theme) string {
	var parts []string

	for _, component := range []struct {
		name  string
		value paint
	}{
		{"bg", s.bg},
		{"fg", s.fg},
		{"fill", s.fill},
	} {
		if component.value != nil {
			parts = append(
				parts,
				component.name+"="+component.value.expression(t),
			)
		}
	}

	return strings.Join(parts, ",")
}

// scope is the tmux command that sets an option.
type scope string

const (
	sessionOption scope = "set-option"
	windowOption  scope = "set-window-option"
)

type assignment struct {
	scope  scope
	option string
	style  style
}

type section struct {
	comment     string
	assignments []assignment
}

// activePaneBorder carries tmux's pane states. synchronize-panes is the
// consequential state, because input reaches more than one pane, so it alone
// takes Rust. tmux's own behavior remains the primary cue; color supplements
// it.
var activePaneBorder = conditional{
	condition: "pane_modal_flag",
	then:      mineral,
	otherwise: conditional{
		condition: "pane_marked",
		then:      ochre,
		otherwise: conditional{
			condition: "synchronize-panes",
			then:      rust,
			otherwise: conditional{
				condition: "pane_in_mode",
				then:      ochre,
				otherwise: mineral,
			},
		},
	},
}

// activeTreePreview marks the preview of the active pane or window.
var activeTreePreview = conditional{
	condition: "#{||:#{&&:#{pane_format},#{pane_active}},#{&&:#{window_format},#{window_active}}}",
	then:      mineral,
	otherwise: muted,
}

// overlay is the complete tmux mapping policy. Each row maps one tmux style
// option onto quiet roles; Encode only serializes it.
var overlay = []section{
	{
		comment: "Status line.",
		assignments: []assignment{
			{sessionOption, "status-style", style{bg: structure, fg: ink}},
			{sessionOption, "status-left-style", style{fg: ink}},
			{sessionOption, "status-right-style", style{fg: muted}},
		},
	},
	{
		comment: "Messages and command prompts.",
		assignments: []assignment{
			{sessionOption, "message-style", style{bg: surface, fg: ink, fill: surface}},
			{sessionOption, "message-command-style", style{bg: surface, fg: mineral, fill: surface}},
		},
	},
	{
		comment: "Window status.",
		assignments: []assignment{
			{windowOption, "window-status-style", style{bg: structure, fg: muted}},
			{windowOption, "window-status-current-style", style{bg: selection, fg: ink}},
			{windowOption, "window-status-last-style", style{bg: structure, fg: muted}},
			{windowOption, "window-status-activity-style", style{bg: structure, fg: mineral}},
			{windowOption, "window-status-bell-style", style{bg: structure, fg: ochre}},
		},
	},
	{
		comment: "Pane borders.",
		assignments: []assignment{
			{windowOption, "pane-border-style", style{fg: line}},
			{windowOption, "pane-active-border-style", style{fg: activePaneBorder}},
		},
	},
	{
		comment: "Pane and session status, when enabled by the user's configuration.",
		assignments: []assignment{
			{windowOption, "pane-status-style", style{fg: muted}},
			{windowOption, "pane-status-current-style", style{bg: selection, fg: ink}},
			{windowOption, "session-status-style", style{fg: muted}},
			{windowOption, "session-status-current-style", style{bg: selection, fg: ink}},
		},
	},
	{
		comment: "Copy and mode UI.",
		assignments: []assignment{
			{windowOption, "mode-style", style{bg: selection, fg: ink}},
			{windowOption, "copy-mode-match-style", style{bg: callout, fg: ink}},
			{windowOption, "copy-mode-current-match-style", style{bg: selection, fg: ink}},
			{windowOption, "copy-mode-mark-style", style{bg: callout, fg: ochre}},
			{windowOption, "copy-mode-line-number-style", style{fg: muted}},
			{windowOption, "copy-mode-current-line-number-style", style{fg: mineral}},
			{windowOption, "copy-mode-position-style", style{fg: muted}},
			{windowOption, "switch-mode-match-style", style{bg: callout, fg: ink}},
		},
	},
	{
		comment: "Menus.",
		assignments: []assignment{
			{windowOption, "menu-style", style{bg: surface, fg: ink}},
			{windowOption, "menu-selected-style", style{bg: selection, fg: ink}},
			{windowOption, "menu-border-style", style{bg: surface, fg: line}},
		},
	},
	{
		comment: "Popups.",
		assignments: []assignment{
			{windowOption, "popup-style", style{bg: surface, fg: ink}},
			{windowOption, "popup-border-style", style{bg: surface, fg: line}},
		},
	},
	{
		comment: "Tree mode.",
		assignments: []assignment{
			{windowOption, "tree-mode-border-style", style{bg: surface, fg: line}},
			{windowOption, "tree-mode-preview-style", style{fg: activeTreePreview}},
		},
	},
	{
		comment: "Display-panes overlay and scrollbars.",
		assignments: []assignment{
			{windowOption, "display-panes-border-style", style{bg: structure, fg: line}},
			{windowOption, "pane-scrollbars-style", style{bg: field, fg: muted}},
		},
	},
}

// Encode writes a tmux color overlay.
//
// Every style assignment uses -a so existing non-color style attributes
// remain intact. quiet is intended to be sourced after user configuration
// and any plugin that establishes tmux colors.
func Encode(
	w io.Writer,
	t Theme,
) error {
	var b strings.Builder

	fmt.Fprintf(
		&b,
		"# @(#)quiet / tmux / %s / 2026-10-19\n",
		t.ArtifactPath,
	)

	fmt.Fprintf(
		&b,
		`
# %s
#
# quiet color overlay for tmux.
#
# This file changes tmux interface color only.
# It does not configure content, layout, key bindings, formats,
# plugins, or terminal behavior.
`,
		t.Name,
	)

	for _, s := range overlay {
		fmt.Fprintf(
			&b,
			"\n# %s\n",
			s.comment,
		)

		for _, a := range s.assignments {
			fmt.Fprintf(
				&b,
				"%s -ag %s '%s'\n",
				a.scope,
				a.option,
				a.style.expression(t),
			)
		}
	}

	if _, err := io.WriteString(
		w,
		b.String(),
	); err != nil {
		return fmt.Errorf(
			"write tmux configuration: %w",
			err,
		)
	}

	return nil
}

func artifactPath(name string) (string, error) {
	switch name {
	case "quiet-day":
		return "ports/tmux/quiet-day.conf", nil

	case "quiet-night":
		return "ports/tmux/quiet-night.conf", nil

	default:
		return "", fmt.Errorf(
			"unsupported tmux palette %q",
			name,
		)
	}
}
