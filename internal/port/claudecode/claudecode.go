// @(#)quiet / go / internal/port/claudecode/claudecode.go / 2026-10-27

// Package claudecode maps quiet-core semantics onto a Claude Code custom
// theme.
//
// A Claude Code custom theme is one JSON file in ~/.claude/themes. It names
// a built-in base theme and overrides colors by palette key. Claude Code
// ignores keys its base palette does not define, and any key the theme does
// not set keeps the base theme's color.
//
// Claude Code draws on the terminal's own background, so the mapping
// assumes quiet's Field there: use the theme with a quiet terminal port.
//
// The port controls color. It leaves Claude Code's own identity colors, the
// categorical colors users assign to subagents, and decorative effects to
// the base theme.
package claudecode

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
)

// role names a quiet color used by the Claude Code mapping.
type role int

const (
	field role = iota
	surface
	structure
	ink
	muted
	line
	subtle
	mineral
	moss
	ochre
	rust
	selection
	callout
	mossWash
	rustWash
	mossWashEmphasis
	rustWashEmphasis
)

type palette [rustWashEmphasis + 1]color.SRGB

// resolve returns every role the mapping uses.
//
// Subtle is contextual. Claude Code draws subtle text on the terminal
// background, which the mapping assumes is Field, so it resolves to one
// color here.
func resolve(p core.Palette) (palette, error) {
	derived, err := core.Derive(p)
	if err != nil {
		return palette{}, fmt.Errorf(
			"derive quiet roles: %w",
			err,
		)
	}

	subtleOverField, err := core.SubtleOver(
		p,
		p.Anchors.Field,
	)
	if err != nil {
		return palette{}, err
	}

	a := p.Anchors
	wash := derived.Wash

	return palette{
		field:            a.Field,
		surface:          a.Surface,
		structure:        a.Structure,
		ink:              a.Ink,
		muted:            a.Muted,
		line:             a.Line,
		subtle:           subtleOverField,
		mineral:          a.Mineral,
		moss:             a.Moss,
		ochre:            a.Ochre,
		rust:             a.Rust,
		selection:        derived.Selection,
		callout:          derived.Callout,
		mossWash:         wash.Moss.Line,
		rustWash:         wash.Rust.Line,
		mossWashEmphasis: wash.Moss.Emphasis,
		rustWashEmphasis: wash.Rust.Emphasis,
	}, nil
}

// entry assigns one Claude Code palette key a role.
type entry struct {
	key  string
	role role
}

// mapping is the complete Claude Code policy, in the order of Claude Code's
// own palette.
//
// Shimmer keys are the highlight Claude Code sweeps across an animated
// label. Each takes the next role up from its label's: Ink over a chromatic
// or Muted label, Muted over a Line border. The animation stays visible
// without a new color.
//
// Permission modes rise with consequence. Claude Code colors Manual
// inactive, Plan planMode, Accept edits autoAccept, Auto warning, and Bypass
// error, and names each mode beside its symbol. Plan is Mineral, an active
// state. Accept edits is Ochre, like Auto: changes apply without review.
//
// Pull request labels carry review state by color alone. approved,
// changes_requested, and pending already resolve to success, error, and
// warning; merged takes Mineral so all four states stay distinct.
//
// diffAddedWord and diffRemovedWord color the changed words within a diff
// line, and also the +N and -N counters in file and pull request lists. They
// take the Wash emphasis tier, as changed words do in every quiet port; the
// counters drawn in them are not readable. No single color can serve both:
// it would have to be dark enough to keep Ink readable over it and light
// enough to read on Field.
var mapping = []entry{
	{"autoAccept", ochre},
	{"autoAcceptShimmer", ink},
	{"skill", mineral},
	{"bashBorder", mineral},
	{"claudeBlue_FOR_SYSTEM_SPINNER", mineral},
	{"claudeBlueShimmer_FOR_SYSTEM_SPINNER", ink},
	{"permission", mineral},
	{"permissionShimmer", ink},
	{"planMode", mineral},
	{"ide", mineral},
	{"promptBorder", line},
	{"promptBorderShimmer", muted},
	{"text", ink},
	// inverseText is drawn on chromatic badge backgrounds.
	{"inverseText", field},
	{"inactive", muted},
	{"inactiveShimmer", ink},
	// subtle marks nonessential hints, as Subtle does.
	{"subtle", subtle},
	{"suggestion", mineral},
	{"remember", mineral},
	// background marks background tasks, not a surface.
	{"background", mineral},
	{"success", moss},
	{"error", rust},
	{"warning", ochre},
	{"merged", mineral},
	{"warningShimmer", ink},
	// Diff line backgrounds. The dimmed variants, for diffs Claude Code
	// shows dimmed, collapse onto the line tier: the host dims their text.
	{"diffAdded", mossWash},
	{"diffRemoved", rustWash},
	{"diffAddedDimmed", mossWash},
	{"diffRemovedDimmed", rustWash},
	{"diffAddedWord", mossWashEmphasis},
	{"diffRemovedWord", rustWashEmphasis},
	// professionalBlue titles account and policy notices.
	{"professionalBlue", mineral},
	{"userMessageBackground", surface},
	{"userMessageBackgroundHover", callout},
	{"composerSidebarBackground", structure},
	{"selectionBg", selection},
	{"bashMessageBackgroundColor", surface},
	{"memoryBackgroundColor", surface},
	{"rate_limit_fill", mineral},
	{"rate_limit_empty", line},
	{"fastMode", mineral},
	{"fastModeShimmer", ink},
	{"effortUltra", mineral},
	// briefLabelYou names the user's messages in brief view: secondary
	// text, since the label itself says who spoke.
	{"briefLabelYou", muted},
}

// Encode writes a Claude Code custom theme for p.
//
// The base theme matches the adaptation, so every key quiet leaves unset
// keeps a color designed for the same ground. The theme's name is p.Name,
// and Claude Code lists the theme under its file name, so the file is named
// after the adaptation as well.
//
// The output follows Claude Code's own: two-space JSON in the order name,
// base, overrides.
func Encode(
	w io.Writer,
	p core.Palette,
) error {
	resolved, err := resolve(p)
	if err != nil {
		return err
	}

	var base string

	switch p.Mode {
	case core.ModeDay:
		base = "light"

	case core.ModeNight:
		base = "dark"

	default:
		return fmt.Errorf(
			"unsupported Claude Code palette mode %q",
			p.Mode,
		)
	}

	var b strings.Builder

	fmt.Fprintf(
		&b,
		"{\n  \"name\": %s,\n  \"base\": %s,\n  \"overrides\": {\n",
		quote(p.Name),
		quote(base),
	)

	for i, e := range mapping {
		hex, err := resolved[e.role].Hex()
		if err != nil {
			return fmt.Errorf(
				"claude-code %s: %w",
				e.key,
				err,
			)
		}

		separator := ","
		if i == len(mapping)-1 {
			separator = ""
		}

		fmt.Fprintf(
			&b,
			"    %s: %s%s\n",
			quote(e.key),
			quote(hex),
			separator,
		)
	}

	b.WriteString("  }\n}\n")

	if _, err := io.WriteString(
		w,
		b.String(),
	); err != nil {
		return fmt.Errorf(
			"write Claude Code theme: %w",
			err,
		)
	}

	return nil
}

// quote returns s as a JSON string. Marshaling a string cannot fail.
func quote(s string) string {
	b, _ := json.Marshal(s)

	return string(b)
}
