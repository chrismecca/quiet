// @(#)quiet / go / internal/port/jetbrains/jetbrains.go / 2026-10-25

// Package jetbrains maps quiet-core semantics onto JetBrains IDEs.
//
// The port has two parts: an editor color scheme, and a UI theme for the
// IDE's interface. The plugin packages both; the editor schemes are also
// written on their own, for import without the plugin.
//
// The port controls color. It keeps the non-color cues JetBrains' own
// schemes agree on, such as underlined reassigned variables and struck-out
// deprecated names, and adds no typography of its own.
package jetbrains

import (
	"fmt"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
	"github.com/chrismecca/quiet/internal/port/terminal"
)

// role names a quiet color used by the JetBrains mapping.
type role int

const (
	none role = iota
	field
	surface
	structure
	ink
	muted
	line
	mineral
	moss
	ochre
	rust
	selection
	callout
	mineralWash
	mossWash
	mossWashEmphasis
	ochreWash
	ochreWashEmphasis
	rustWash
	rustWashEmphasis

	// The generic terminal palette, for the console.
	terminalBackground
	terminalText
	ansi0
)

// ansi returns the role of ANSI slot i.
func ansi(i int) role {
	return ansi0 + role(i)
}

type palette map[role]color.SRGB

func resolve(p core.Palette) (palette, error) {
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

	result := palette{
		field:             a.Field,
		surface:           a.Surface,
		structure:         a.Structure,
		ink:               a.Ink,
		muted:             a.Muted,
		line:              a.Line,
		mineral:           a.Mineral,
		moss:              a.Moss,
		ochre:             a.Ochre,
		rust:              a.Rust,
		selection:         derived.Selection,
		callout:           derived.Callout,
		mineralWash:       derived.Wash.Mineral.Line,
		mossWash:          derived.Wash.Moss.Line,
		mossWashEmphasis:  derived.Wash.Moss.Emphasis,
		ochreWash:         derived.Wash.Ochre.Line,
		ochreWashEmphasis: derived.Wash.Ochre.Emphasis,
		rustWash:          derived.Wash.Rust.Line,
		rustWashEmphasis:  derived.Wash.Rust.Emphasis,

		terminalBackground: term.Background,
		terminalText:       term.Text,
	}

	for i, c := range term.ANSI {
		result[ansi(i)] = c
	}

	return result, nil
}

// hex returns r as lowercase RRGGBB without a leading #, the form JetBrains
// writes in its own schemes.
func (p palette) hex(r role) (string, error) {
	c, ok := p[r]
	if !ok {
		return "", fmt.Errorf("unresolved role %d", r)
	}

	r8, g8, b8, err := c.RGB8()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%02x%02x%02x", r8, g8, b8), nil
}

// parentScheme returns the bundled editor scheme quiet's scheme inherits
// from, for keys quiet does not set.
func parentScheme(p core.Palette) (string, error) {
	switch p.Mode {
	case core.ModeDay:
		return "Default", nil

	case core.ModeNight:
		return "Darcula", nil

	default:
		return "", fmt.Errorf(
			"unsupported JetBrains palette mode %q",
			p.Mode,
		)
	}
}
