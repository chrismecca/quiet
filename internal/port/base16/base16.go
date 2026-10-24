// @(#)quiet / go / internal/port/base16/base16.go / 2026-10-24

// Package base16 projects quiet-core onto the Base16 compatibility model.
package base16

import (
	"fmt"
	"math"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
)

const minimumForegroundContrast = 4.5

// Scheme is a resolved Base16 compatibility scheme.
//
// Base16 is not a representation of quiet-core. It has its own fixed set of
// sixteen roles. Values in this structure are a compatibility projection and
// must not be interpreted as additional quiet canonical anchors.
type Scheme struct {
	Name        string
	Slug        string
	Author      string
	Description string
	Variant     string
	Palette     [16]color.SRGB
}

// Resolve projects one quiet palette onto Base16.
func Resolve(p core.Palette) (Scheme, error) {
	derived, err := core.Derive(p)
	if err != nil {
		return Scheme{}, fmt.Errorf(
			"derive quiet roles: %w",
			err,
		)
	}

	var (
		variant string
		slug    string
	)

	switch fmt.Sprint(p.Mode) {
	case "day":
		variant = "light"
		slug = "quiet-day"

	case "night":
		variant = "dark"
		slug = "quiet-night"

	default:
		return Scheme{}, fmt.Errorf(
			"unsupported quiet mode %q",
			fmt.Sprint(p.Mode),
		)
	}

	cyan, err := deriveCyan(p)
	if err != nil {
		return Scheme{}, err
	}

	magenta, err := deriveMagenta(p)
	if err != nil {
		return Scheme{}, err
	}

	// Base16 base04 is a foreground between its secondary/comment foreground
	// and default foreground. quiet has no canonical role corresponding
	// exactly to that slot, so the compatibility port derives the midpoint
	// between Muted and Ink in OKLab.
	base04 := color.MixOKLab(
		p.Anchors.Muted,
		p.Anchors.Ink,
		0.50,
	)

	if !base04.InGamut() {
		return Scheme{}, fmt.Errorf(
			"Base16 base04 is outside sRGB gamut",
		)
	}

	result := Scheme{
		Name:        p.Name,
		Slug:        slug,
		Author:      "quiet",
		Description: "A Base16 compatibility projection of " + p.Name + ".",
		Variant:     variant,

		Palette: [16]color.SRGB{
			// Neutral and structural roles.
			p.Anchors.Field,     // base00
			p.Anchors.Structure, // base01
			derived.Selection,   // base02
			p.Anchors.Muted,     // base03
			base04,              // base04
			p.Anchors.Ink,       // base05
			p.Anchors.Ink,       // base06
			p.Anchors.Ink,       // base07

			// Chromatic compatibility roles.
			p.Anchors.Rust,    // base08
			p.Anchors.Ochre,   // base09
			p.Anchors.Ochre,   // base0A
			p.Anchors.Moss,    // base0B
			cyan,              // base0C
			p.Anchors.Mineral, // base0D
			magenta,           // base0E
			p.Anchors.Rust,    // base0F
		},
	}

	// Base16 uses base03 through base0F extensively as foreground syntax
	// colors. quiet's compatibility projection keeps every such color at the
	// same readable-foreground floor used by quiet-core where possible.
	for i := 3; i < len(result.Palette); i++ {
		actual := color.Contrast(
			result.Palette[i],
			p.Anchors.Field,
		)

		if actual < minimumForegroundContrast {
			return Scheme{}, fmt.Errorf(
				"Base16 base%02X contrast %.6f below %.6f",
				i,
				actual,
				minimumForegroundContrast,
			)
		}
	}

	return result, nil
}

func deriveCyan(
	p core.Palette,
) (color.SRGB, error) {
	moss := p.Anchors.Moss.OKLab().OKLCH()
	mineral := p.Anchors.Mineral.OKLab().OKLCH()

	hue, err := shortestHueMidpoint(
		moss.H,
		mineral.H,
	)
	if err != nil {
		return color.SRGB{}, fmt.Errorf(
			"derive Base16 cyan: %w",
			err,
		)
	}

	return deriveExtension(
		"cyan",
		moss,
		mineral,
		hue,
	)
}

func deriveMagenta(
	p core.Palette,
) (color.SRGB, error) {
	mineral := p.Anchors.Mineral.OKLab().OKLCH()
	rust := p.Anchors.Rust.OKLab().OKLCH()

	hue, err := forwardHueMidpoint(
		mineral.H,
		rust.H,
	)
	if err != nil {
		return color.SRGB{}, fmt.Errorf(
			"derive Base16 magenta: %w",
			err,
		)
	}

	return deriveExtension(
		"magenta",
		mineral,
		rust,
		hue,
	)
}

func deriveExtension(
	name string,
	a color.OKLCH,
	b color.OKLCH,
	hue float64,
) (color.SRGB, error) {
	result := color.OKLCH{
		L: (a.L + b.L) / 2,
		C: (a.C + b.C) / 2,
		H: hue,
	}.OKLab().SRGB()

	if !result.InGamut() {
		return color.SRGB{}, fmt.Errorf(
			"derived Base16 %s is outside sRGB gamut",
			name,
		)
	}

	return result, nil
}

func shortestHueMidpoint(
	a float64,
	b float64,
) (float64, error) {
	if math.IsNaN(a) || math.IsNaN(b) {
		return 0, fmt.Errorf(
			"cannot interpolate undefined hue",
		)
	}

	delta := math.Mod(
		b-a+540,
		360,
	) - 180

	return normalizeHue(
		a + delta/2,
	), nil
}

func forwardHueMidpoint(
	a float64,
	b float64,
) (float64, error) {
	if math.IsNaN(a) || math.IsNaN(b) {
		return 0, fmt.Errorf(
			"cannot interpolate undefined hue",
		)
	}

	delta := math.Mod(
		b-a+360,
		360,
	)

	return normalizeHue(
		a + delta/2,
	), nil
}

func normalizeHue(h float64) float64 {
	h = math.Mod(h, 360)

	if h < 0 {
		h += 360
	}

	return h
}
