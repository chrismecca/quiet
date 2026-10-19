// @(#)quiet / go / internal/port/tmux/tmux.go / 2026-10-19

// Package tmux maps Quiet Core semantics onto tmux interface chrome.
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

// Theme is the resolved Quiet color vocabulary used by tmux.
//
// Values are serialized sRGB colors because tmux accepts #RRGGBB directly.
// This is a host-level representation; it does not introduce new Quiet Core
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

// Resolve maps one canonical Quiet palette onto tmux interface semantics.
func Resolve(p core.Palette) (Theme, error) {
	derived, err := core.Derive(p)
	if err != nil {
		return Theme{}, fmt.Errorf(
			"derive Quiet roles: %w",
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

// Encode writes a tmux color overlay.
//
// Every style assignment uses -a so existing non-color style attributes
// remain intact. Quiet is intended to be sourced after user configuration
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
# Quiet color overlay for tmux.
#
# This file changes tmux interface color only.
# It does not configure content, layout, key bindings, formats,
# plugins, or terminal behavior.

# Status line.
set-option -ag status-style 'bg=%s,fg=%s'
set-option -ag status-left-style 'fg=%s'
set-option -ag status-right-style 'fg=%s'

# Messages and command prompts.
set-option -ag message-style 'bg=%s,fg=%s,fill=%s'
set-option -ag message-command-style 'bg=%s,fg=%s,fill=%s'

# Window status.
set-window-option -ag window-status-style 'bg=%s,fg=%s'
set-window-option -ag window-status-current-style 'bg=%s,fg=%s'
set-window-option -ag window-status-last-style 'bg=%s,fg=%s'
set-window-option -ag window-status-activity-style 'bg=%s,fg=%s'
set-window-option -ag window-status-bell-style 'bg=%s,fg=%s'

# Pane borders.
set-window-option -ag pane-border-style 'fg=%s'
set-window-option -ag pane-active-border-style 'fg=#{?pane_modal_flag,%s,#{?pane_marked,%s,#{?synchronize-panes,%s,#{?pane_in_mode,%s,%s}}}}'

# Pane and session status, when enabled by the user's configuration.
set-window-option -ag pane-status-style 'fg=%s'
set-window-option -ag pane-status-current-style 'bg=%s,fg=%s'
set-window-option -ag session-status-style 'fg=%s'
set-window-option -ag session-status-current-style 'bg=%s,fg=%s'

# Copy and mode UI.
set-window-option -ag mode-style 'bg=%s,fg=%s'
set-window-option -ag copy-mode-match-style 'bg=%s,fg=%s'
set-window-option -ag copy-mode-current-match-style 'bg=%s,fg=%s'
set-window-option -ag copy-mode-mark-style 'bg=%s,fg=%s'
set-window-option -ag copy-mode-line-number-style 'fg=%s'
set-window-option -ag copy-mode-current-line-number-style 'fg=%s'
set-window-option -ag copy-mode-position-style 'fg=%s'
set-window-option -ag switch-mode-match-style 'bg=%s,fg=%s'

# Menus.
set-window-option -ag menu-style 'bg=%s,fg=%s'
set-window-option -ag menu-selected-style 'bg=%s,fg=%s'
set-window-option -ag menu-border-style 'bg=%s,fg=%s'

# Popups.
set-window-option -ag popup-style 'bg=%s,fg=%s'
set-window-option -ag popup-border-style 'bg=%s,fg=%s'

# Tree mode.
set-window-option -ag tree-mode-border-style 'bg=%s,fg=%s'
set-window-option -ag tree-mode-preview-style 'fg=#{?#{||:#{&&:#{pane_format},#{pane_active}},#{&&:#{window_format},#{window_active}}},%s,%s}'

# Display-panes overlay and scrollbars.
set-window-option -ag display-panes-border-style 'bg=%s,fg=%s'
set-window-option -ag pane-scrollbars-style 'bg=%s,fg=%s'
`,
		t.Name,

		// Status line.
		t.Structure,
		t.Ink,
		t.Ink,
		t.Muted,

		// Messages.
		t.Surface,
		t.Ink,
		t.Surface,
		t.Surface,
		t.Mineral,
		t.Surface,

		// Window status.
		t.Structure,
		t.Muted,
		t.Selection,
		t.Ink,
		t.Structure,
		t.Muted,
		t.Structure,
		t.Mineral,
		t.Structure,
		t.Ochre,

		// Pane borders.
		t.Line,
		t.Mineral,
		t.Ochre,
		t.Rust,
		t.Ochre,
		t.Mineral,

		// Pane/session status.
		t.Muted,
		t.Selection,
		t.Ink,
		t.Muted,
		t.Selection,
		t.Ink,

		// Copy/mode UI.
		t.Selection,
		t.Ink,
		t.Callout,
		t.Ink,
		t.Selection,
		t.Ink,
		t.Callout,
		t.Ochre,
		t.Muted,
		t.Mineral,
		t.Muted,
		t.Callout,
		t.Ink,

		// Menus.
		t.Surface,
		t.Ink,
		t.Selection,
		t.Ink,
		t.Surface,
		t.Line,

		// Popups.
		t.Surface,
		t.Ink,
		t.Surface,
		t.Line,

		// Tree mode.
		t.Surface,
		t.Line,
		t.Mineral,
		t.Muted,

		// Display panes / scrollbars.
		t.Structure,
		t.Line,
		t.Field,
		t.Muted,
	)

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
	case "Quiet Day":
		return "ports/tmux/quiet-day.conf", nil

	case "Quiet Night":
		return "ports/tmux/quiet-night.conf", nil

	default:
		return "", fmt.Errorf(
			"unsupported tmux palette %q",
			name,
		)
	}
}
