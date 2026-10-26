// @(#)quiet / go / internal/port/iterm2/iterm2.go / 2026-10-26

// Package iterm2 serializes quiet's generic terminal palettes into iTerm2's
// native .itermcolors property-list format.
package iterm2

import (
	"fmt"
	"io"
	"strings"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/port/terminal"
)

// entry is one iTerm2 color key with its quiet-night and quiet-day values.
type entry struct {
	name  string
	night color.SRGB
	day   color.SRGB
	alpha float64
}

func entries(
	day terminal.Palette,
	night terminal.Palette,
) []entry {
	var result []entry

	for i := range night.ANSI {
		result = append(result, entry{
			name:  fmt.Sprintf("Ansi %d Color", i),
			night: night.ANSI[i],
			day:   day.ANSI[i],
			alpha: 1.0,
		})
	}

	return append(
		result,
		entry{"Background Color", night.Background, day.Background, 1.0},
		entry{"Foreground Color", night.Text, day.Text, 1.0},
		entry{"Bold Color", night.Bold, day.Bold, 1.0},
		entry{"Cursor Color", night.Cursor, day.Cursor, 1.0},
		entry{"Cursor Text Color", night.Background, day.Background, 1.0},
		entry{"Selection Color", night.Selection, day.Selection, 1.0},
		entry{"Selected Text Color", night.Text, day.Text, 1.0},
		entry{"Link Color", night.Cursor, day.Cursor, 1.0},
		entry{"Active Pane Border Color", night.Cursor, day.Cursor, 1.0},
		entry{"IME Cursor Color", night.Cursor, day.Cursor, 1.0},
		entry{"Cursor Guide Color", night.Cursor, day.Cursor, 0.25},
		entry{"Badge Color", night.Cursor, day.Cursor, 0.50},
		entry{"Match Background Color", night.Selection, day.Selection, 1.0},
	)
}

// Encode writes one adaptive iTerm2 color preset.
//
// Unsuffixed values use quiet-night as the compatibility fallback.
// Explicit (Dark) values use quiet-night.
// Explicit (Light) values use quiet-day.
//
// No quiet color derivation occurs here. Both inputs are already-resolved
// generic terminal palettes.
func Encode(
	w io.Writer,
	day terminal.Palette,
	night terminal.Palette,
) error {
	return writePreset(w, func(b *strings.Builder) error {
		for _, e := range entries(day, night) {
			if err := writeAdaptiveColor(b, e.name, e.night, e.day, e.alpha); err != nil {
				return err
			}
		}

		return nil
	})
}

// EncodeVariant writes a preset for one palette, without (Light) and (Dark)
// variants, as theme collections that hold one file per variant expect.
func EncodeVariant(
	w io.Writer,
	p terminal.Palette,
) error {
	return writePreset(w, func(b *strings.Builder) error {
		for _, e := range entries(p, p) {
			if err := writeColor(b, e.name, e.night, e.alpha); err != nil {
				return err
			}
		}

		return nil
	})
}

func writePreset(
	w io.Writer,
	body func(*strings.Builder) error,
) error {
	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(
		`<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" ` +
			`"http://www.apple.com/DTDs/PropertyList-1.0.dtd">` + "\n",
	)
	b.WriteString(`<plist version="1.0">` + "\n")
	b.WriteString("<dict>\n")

	if err := body(&b); err != nil {
		return err
	}

	b.WriteString("</dict>\n")
	b.WriteString("</plist>\n")

	if _, err := io.WriteString(
		w,
		b.String(),
	); err != nil {
		return fmt.Errorf(
			"write iTerm2 preset: %w",
			err,
		)
	}

	return nil
}

func writeAdaptiveColor(
	b *strings.Builder,
	name string,
	night color.SRGB,
	day color.SRGB,
	alpha float64,
) error {
	if err := writeColor(
		b,
		name,
		night,
		alpha,
	); err != nil {
		return err
	}

	if err := writeColor(
		b,
		name+" (Dark)",
		night,
		alpha,
	); err != nil {
		return err
	}

	if err := writeColor(
		b,
		name+" (Light)",
		day,
		alpha,
	); err != nil {
		return err
	}

	return nil
}

func writeColor(
	b *strings.Builder,
	name string,
	value color.SRGB,
	alpha float64,
) error {
	canonical, err := canonicalSRGB(value)
	if err != nil {
		return fmt.Errorf(
			"%s: %w",
			name,
			err,
		)
	}

	fmt.Fprintf(
		b,
		"\t<key>%s</key>\n"+
			"\t<dict>\n"+
			"\t\t<key>Alpha Component</key>\n"+
			"\t\t<real>%.6f</real>\n"+
			"\t\t<key>Blue Component</key>\n"+
			"\t\t<real>%.9f</real>\n"+
			"\t\t<key>Color Space</key>\n"+
			"\t\t<string>sRGB</string>\n"+
			"\t\t<key>Green Component</key>\n"+
			"\t\t<real>%.9f</real>\n"+
			"\t\t<key>Red Component</key>\n"+
			"\t\t<real>%.9f</real>\n"+
			"\t</dict>\n",
		name,
		alpha,
		canonical.B,
		canonical.G,
		canonical.R,
	)

	return nil
}

// canonicalSRGB applies the same 8-bit export boundary used by quiet's other
// host artifacts before recovering normalized components for plist encoding.
func canonicalSRGB(
	value color.SRGB,
) (color.SRGB, error) {
	hex, err := value.Hex()
	if err != nil {
		return color.SRGB{}, err
	}

	canonical, err := color.ParseHex(hex)
	if err != nil {
		return color.SRGB{}, fmt.Errorf(
			"reparse canonical color %s: %w",
			hex,
			err,
		)
	}

	return canonical, nil
}
