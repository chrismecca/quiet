// @(#)quiet / go / internal/port/terminal/palette.go / 2026-10-17

// Package terminal maps Quiet Core semantics onto the conventional sixteen
// ANSI terminal color slots.
//
// The package is a reference-port adapter, not part of Quiet Core. Terminal-
// specific colors derived here must not be interpreted as additional canonical
// Quiet anchors.
package terminal

import (
	"fmt"
	"math"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
)

const (
	// brightDeltaL increases perceptual lightness for ANSI bright chromatic
	// slots. "Bright" is treated as a terminal-local intensity relationship,
	// not as increased contrast against Quiet Field.
	brightDeltaL = 0.010

	// brightChromaScale modestly increases chroma while preserving the base hue
	// family. The scale is intentionally conservative: ANSI bright colors must
	// remain recognizably Quiet rather than becoming saturated terminal colors.
	brightChromaScale = 1.20

	// minimumChromaticContrast applies to ANSI chromatic foreground slots.
	//
	// Neutral ANSI slots are deliberately excluded. ANSI neutral slots are
	// overloaded as both foreground and background colors, and the terminal
	// model cannot guarantee foreground readability for both perceptual
	// extremes in both light and dark adaptations.
	minimumChromaticContrast = 4.5
)

// Palette is a fully resolved generic ANSI terminal palette.
//
// It contains host-level roles rather than Quiet semantic roles. Platform
// adapters such as Terminal.app consume this representation and should not
// need to know how Quiet anchors produced it.
type Palette struct {
	Name       string
	Background color.SRGB
	Text       color.SRGB
	Bold       color.SRGB
	Selection  color.SRGB
	Cursor     color.SRGB
	ANSI       [16]color.SRGB
}

// Resolve maps one canonical Quiet palette onto the generic ANSI terminal
// model.
func Resolve(p core.Palette) (Palette, error) {
	derived, err := core.Derive(p)
	if err != nil {
		return Palette{}, fmt.Errorf(
			"derive Quiet roles: %w",
			err,
		)
	}

	cyan, err := deriveCyan(p)
	if err != nil {
		return Palette{}, err
	}

	magenta, err := deriveMagenta(p)
	if err != nil {
		return Palette{}, err
	}

	black, brightBlack, white, brightWhite, err := neutralRamp(p)
	if err != nil {
		return Palette{}, err
	}

	baseChromatic := [6]color.SRGB{
		p.Anchors.Rust,
		p.Anchors.Moss,
		p.Anchors.Ochre,
		p.Anchors.Mineral,
		magenta,
		cyan,
	}

	var brightChromatic [6]color.SRGB

	for i, base := range baseChromatic {
		if err := requireChromaticContrast(
			fmt.Sprintf("ANSI %d", i+1),
			base,
			p.Anchors.Field,
		); err != nil {
			return Palette{}, err
		}

		bright, err := deriveBrightChromatic(base)
		if err != nil {
			return Palette{}, fmt.Errorf(
				"derive ANSI %d: %w",
				i+9,
				err,
			)
		}

		if err := requireChromaticContrast(
			fmt.Sprintf("ANSI %d", i+9),
			bright,
			p.Anchors.Field,
		); err != nil {
			return Palette{}, err
		}

		brightChromatic[i] = bright
	}

	return Palette{
		Name:       p.Name,
		Background: p.Anchors.Field,
		Text:       p.Anchors.Ink,
		Bold:       p.Anchors.Ink,
		Selection:  derived.Selection,
		Cursor:     p.Anchors.Mineral,

		ANSI: [16]color.SRGB{
			black,
			baseChromatic[0],
			baseChromatic[1],
			baseChromatic[2],
			baseChromatic[3],
			baseChromatic[4],
			baseChromatic[5],
			white,

			brightBlack,
			brightChromatic[0],
			brightChromatic[1],
			brightChromatic[2],
			brightChromatic[3],
			brightChromatic[4],
			brightChromatic[5],
			brightWhite,
		},
	}, nil
}

// deriveBrightChromatic produces the ANSI bright member of one chromatic
// family.
//
// ANSI bright slots are used as both foregrounds and backgrounds. The previous
// reference-port experiment moved each base color toward Ink, which increased
// foreground contrast but made Quiet Day bright backgrounds darker than their
// corresponding normal backgrounds.
//
// Candidate 0.2 instead gives "bright" an adaptation-independent meaning:
// slightly greater perceptual lightness and chroma with hue preserved.
func deriveBrightChromatic(
	base color.SRGB,
) (color.SRGB, error) {
	lch := base.OKLab().OKLCH()

	if math.IsNaN(lch.H) {
		return color.SRGB{}, fmt.Errorf(
			"cannot brighten color with undefined hue",
		)
	}

	bright := color.OKLCH{
		L: lch.L + brightDeltaL,
		C: lch.C * brightChromaScale,
		H: lch.H,
	}.OKLab().SRGB()

	if !bright.InGamut() {
		return color.SRGB{}, fmt.Errorf(
			"bright chromatic is outside sRGB gamut",
		)
	}

	return bright, nil
}

// neutralRamp maps existing Quiet neutrals onto the four ANSI grayscale slots.
//
// The mapping is based on perceptual lightness instead of Day/Night names:
//
//   - 0  = darker of Field and Ink
//   - 8  = darker of Muted and Line
//   - 7  = lighter of Muted and Line
//   - 15 = lighter of Field and Ink
//
// ANSI neutral slots are inherently overloaded: the same indexed color may be
// requested as either foreground or background. The extreme slot matching
// Field may consequently disappear when misused as a foreground. Candidate
// reference-port testing treats that as a host-model limitation rather than
// inventing a mode-specific reversal of ANSI neutral semantics.
func neutralRamp(
	p core.Palette,
) (
	black color.SRGB,
	brightBlack color.SRGB,
	white color.SRGB,
	brightWhite color.SRGB,
	err error,
) {
	black, brightWhite = orderByLightness(
		p.Anchors.Field,
		p.Anchors.Ink,
	)

	brightBlack, white = orderByLightness(
		p.Anchors.Muted,
		p.Anchors.Line,
	)

	values := []color.SRGB{
		black,
		brightBlack,
		white,
		brightWhite,
	}

	for i := 1; i < len(values); i++ {
		if lightness(values[i]) <= lightness(values[i-1]) {
			return color.SRGB{},
				color.SRGB{},
				color.SRGB{},
				color.SRGB{},
				fmt.Errorf(
					"ANSI neutral ramp is not monotonically increasing",
				)
		}
	}

	return black, brightBlack, white, brightWhite, nil
}

func orderByLightness(
	a color.SRGB,
	b color.SRGB,
) (color.SRGB, color.SRGB) {
	if lightness(a) <= lightness(b) {
		return a, b
	}

	return b, a
}

func lightness(c color.SRGB) float64 {
	return c.OKLab().L
}

// deriveCyan fills the perceptual interval between Moss and Mineral.
//
// ANSI requires a cyan family that Quiet Core does not define. The terminal
// reference port therefore creates one by averaging L and C and taking the
// shortest hue midpoint between those two neighboring Quiet families.
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
			"derive terminal cyan: %w",
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

// deriveMagenta fills the otherwise unoccupied outer hue interval between
// Mineral and Rust.
//
// The midpoint intentionally follows increasing hue from Mineral through 360°
// to Rust. The opposite arc passes through Moss and Ochre and would therefore
// duplicate existing Quiet families instead of filling ANSI's missing magenta
// region.
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
			"derive terminal magenta: %w",
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
			"derived terminal %s is outside sRGB gamut",
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

func requireChromaticContrast(
	name string,
	foreground color.SRGB,
	background color.SRGB,
) error {
	actual := color.Contrast(
		foreground,
		background,
	)

	if actual >= minimumChromaticContrast {
		return nil
	}

	return fmt.Errorf(
		"%s contrast %.6f below %.6f",
		name,
		actual,
		minimumChromaticContrast,
	)
}
