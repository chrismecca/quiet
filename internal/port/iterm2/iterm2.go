// @(#)quiet / go / internal/port/iterm2/iterm2.go / 2026-10-24

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
	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(
		`<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" ` +
			`"http://www.apple.com/DTDs/PropertyList-1.0.dtd">` + "\n",
	)
	b.WriteString(`<plist version="1.0">` + "\n")
	b.WriteString("<dict>\n")

	for i := 0; i < len(night.ANSI); i++ {
		if err := writeAdaptiveColor(
			&b,
			fmt.Sprintf("Ansi %d Color", i),
			night.ANSI[i],
			day.ANSI[i],
			1.0,
		); err != nil {
			return err
		}
	}

	type role struct {
		name  string
		night color.SRGB
		day   color.SRGB
		alpha float64
	}

	roles := []role{
		{
			name:  "Background Color",
			night: night.Background,
			day:   day.Background,
			alpha: 1.0,
		},
		{
			name:  "Foreground Color",
			night: night.Text,
			day:   day.Text,
			alpha: 1.0,
		},
		{
			name:  "Bold Color",
			night: night.Bold,
			day:   day.Bold,
			alpha: 1.0,
		},
		{
			name:  "Cursor Color",
			night: night.Cursor,
			day:   day.Cursor,
			alpha: 1.0,
		},
		{
			name:  "Cursor Text Color",
			night: night.Background,
			day:   day.Background,
			alpha: 1.0,
		},
		{
			name:  "Selection Color",
			night: night.Selection,
			day:   day.Selection,
			alpha: 1.0,
		},
		{
			name:  "Selected Text Color",
			night: night.Text,
			day:   day.Text,
			alpha: 1.0,
		},
		{
			name:  "Link Color",
			night: night.Cursor,
			day:   day.Cursor,
			alpha: 1.0,
		},
		{
			name:  "Active Pane Border Color",
			night: night.Cursor,
			day:   day.Cursor,
			alpha: 1.0,
		},
		{
			name:  "IME Cursor Color",
			night: night.Cursor,
			day:   day.Cursor,
			alpha: 1.0,
		},
		{
			name:  "Cursor Guide Color",
			night: night.Cursor,
			day:   day.Cursor,
			alpha: 0.25,
		},
		{
			name:  "Badge Color",
			night: night.Cursor,
			day:   day.Cursor,
			alpha: 0.50,
		},
		{
			name:  "Match Background Color",
			night: night.Selection,
			day:   day.Selection,
			alpha: 1.0,
		},
	}

	for _, r := range roles {
		if err := writeAdaptiveColor(
			&b,
			r.name,
			r.night,
			r.day,
			r.alpha,
		); err != nil {
			return err
		}
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
