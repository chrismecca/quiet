// @(#)quiet / go / internal/color/srgb.go / 2026-10-24

package color

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// gamutEpsilon absorbs floating-point noise introduced by reversible color
// transforms. It is deliberately too small to act as gamut mapping.
const gamutEpsilon = 1e-12

// SRGB stores gamma-encoded sRGB components as normalized floating-point
// values. Canonical sRGB colors are in [0, 1], but intermediate conversions
// may produce values outside that range; callers can detect that with InGamut.
type SRGB struct {
	R float64
	G float64
	B float64
}

// ParseHex parses a six-digit sRGB color with an optional leading '#'. It does
// not accept shorthand or alpha-bearing forms because quiet's canonical color
// data uses explicit 24-bit sRGB values.
func ParseHex(s string) (SRGB, error) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")

	if len(s) != 6 {
		return SRGB{}, fmt.Errorf(
			"expected 6 hex digits, got %q",
			s,
		)
	}

	v, err := strconv.ParseUint(s, 16, 24)
	if err != nil {
		return SRGB{}, fmt.Errorf(
			"parse hex color %q: %w",
			s,
			err,
		)
	}

	return SRGB{
		R: float64((v>>16)&0xff) / 255.0,
		G: float64((v>>8)&0xff) / 255.0,
		B: float64(v&0xff) / 255.0,
	}, nil
}

// MustParseHex is intended for trusted constants and test fixtures. User input
// should use ParseHex so malformed input can be reported rather than panicked.
func MustParseHex(s string) SRGB {
	c, err := ParseHex(s)
	if err != nil {
		panic(err)
	}

	return c
}

// InGamut reports whether c is representable in sRGB. The small tolerance is
// only for transform round-off; it must not be treated as gamut mapping.
func (c SRGB) InGamut() bool {
	return inUnitInterval(c.R) &&
		inUnitInterval(c.G) &&
		inUnitInterval(c.B)
}

// Hex quantizes an in-gamut color to canonical uppercase #RRGGBB. It refuses
// materially out-of-gamut values instead of silently clipping them: gamut
// mapping is a separate design decision and must remain explicit.
func (c SRGB) Hex() (string, error) {
	if !c.InGamut() {
		return "", fmt.Errorf(
			"sRGB color is out of gamut: r=%.12f g=%.12f b=%.12f",
			c.R,
			c.G,
			c.B,
		)
	}

	// InGamut permits only epsilon-sized excursions. Clamp those numerical
	// artifacts before 8-bit quantization; larger excursions have returned.
	r := uint8(math.Round(clamp01(c.R) * 255))
	g := uint8(math.Round(clamp01(c.G) * 255))
	b := uint8(math.Round(clamp01(c.B) * 255))

	return fmt.Sprintf("#%02X%02X%02X", r, g, b), nil
}

// MustHex is the assertion form of Hex for values whose gamut membership is an
// established invariant, such as frozen test fixtures and validated derivations.
func (c SRGB) MustHex() string {
	s, err := c.Hex()
	if err != nil {
		panic(err)
	}

	return s
}

func inUnitInterval(x float64) bool {
	return x >= -gamutEpsilon && x <= 1+gamutEpsilon
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}

	if x > 1 {
		return 1
	}

	return x
}

// srgbToLinear applies the IEC sRGB electro-optical transfer function. Values
// outside [0, 1] are allowed so intermediate out-of-gamut colors remain
// representable until an explicit gamut decision is made.
func srgbToLinear(x float64) float64 {
	if x <= 0.04045 {
		return x / 12.92
	}

	return math.Pow(
		(x+0.055)/1.055,
		2.4,
	)
}

// linearToSRGB applies the inverse sRGB transfer function. As above, the
// function intentionally preserves out-of-gamut intermediate values.
func linearToSRGB(x float64) float64 {
	if x <= 0.0031308 {
		return 12.92 * x
	}

	return 1.055*math.Pow(
		x,
		1.0/2.4,
	) - 0.055
}
