// @(#)quiet / go / internal/core/wash.go / 2026-10-24

package core

import (
	"fmt"
	"math"

	"github.com/chrismecca/quiet/internal/color"
)

// The constants below are part of the current candidate Wash derivation
// contract. Distances and lightness offsets are OKLab units.
//
// The line distance is bounded above by sRGB gamut: quiet-day Rust over Field
// leaves the gamut just above 0.025. The emphasis lightness offsets move away
// from Field's extreme so emphasis can carry more chroma while remaining in
// gamut.
const (
	dayWashLineDistance   = 0.020
	nightWashLineDistance = 0.020

	dayWashEmphasisDistance   = 0.040
	nightWashEmphasisDistance = 0.040

	dayWashEmphasisDeltaL   = -0.030
	nightWashEmphasisDeltaL = 0.030

	// minimumWashSourceChroma rejects a family anchor too close to the
	// neutral axis for its hue angle to carry meaning. Every current
	// canonical chromatic anchor exceeds 0.038 OKLCH chroma.
	minimumWashSourceChroma = 0.020
)

// Family identifies one of the four quiet chromatic anchor families.
type Family string

const (
	FamilyMineral Family = "mineral"
	FamilyMoss    Family = "moss"
	FamilyOchre   Family = "ochre"
	FamilyRust    Family = "rust"
)

// Families returns the chromatic families in stable semantic order, matching
// Anchors.All.
func Families() []Family {
	return []Family{
		FamilyMineral,
		FamilyMoss,
		FamilyOchre,
		FamilyRust,
	}
}

// WashTier selects the strength of a Wash.
type WashTier string

const (
	// WashLine marks a region such as a diff line or diagnostic line. It
	// preserves the ground's lightness so every foreground readable on the
	// ground remains readable on the wash.
	WashLine WashTier = "line"

	// WashEmphasis marks a span inside a line wash, such as a word-level diff
	// change. It trades some foreground headroom for visibility and is
	// intended for Ink content only.
	WashEmphasis WashTier = "emphasis"
)

// Wash contains both tiers of one family's wash over Field.
type Wash struct {
	Line     color.SRGB
	Emphasis color.SRGB
}

// Washes contains the canonical Field-grounded washes for every chromatic
// family.
//
// Field is the canonical ground because diff and diagnostic regions normally
// appear in the primary document field. Washes over any other ground are
// different colors and must be resolved with WashOver.
type Washes struct {
	Mineral Wash
	Moss    Wash
	Ochre   Wash
	Rust    Wash
}

// For returns the wash for family.
func (w Washes) For(family Family) (Wash, error) {
	switch family {
	case FamilyMineral:
		return w.Mineral, nil

	case FamilyMoss:
		return w.Moss, nil

	case FamilyOchre:
		return w.Ochre, nil

	case FamilyRust:
		return w.Rust, nil

	default:
		return Wash{}, fmt.Errorf(
			"unknown chromatic family %q",
			family,
		)
	}
}

// WashOver returns family's wash of tier over ground.
//
// A wash displaces the ground in OKLab toward the family anchor's hue by a
// fixed distance. It does not interpolate toward the anchor: the anchors are
// readable foregrounds and therefore far from the ground in lightness, so
// interpolation would mostly darken or lighten the ground rather than tint it.
// The fixed distance also gives every family the same perceptual strength
// regardless of how close the ground's own hue is to the family.
//
// Like Subtle, a wash depends on its local ground. Unlike Subtle, quiet
// publishes the Field-grounded results as opaque derived roles.
func WashOver(
	p Palette,
	family Family,
	tier WashTier,
	ground color.SRGB,
) (color.SRGB, error) {
	deltaL, distance, err := washParameters(
		p.Mode,
		tier,
	)
	if err != nil {
		return color.SRGB{}, err
	}

	source, err := familyAnchor(
		p,
		family,
	)
	if err != nil {
		return color.SRGB{}, err
	}

	hue := source.OKLab().OKLCH()

	if hue.C < minimumWashSourceChroma ||
		math.IsNaN(hue.H) {
		return color.SRGB{}, fmt.Errorf(
			"derive %s wash: anchor chroma %.6f is too close to neutral",
			family,
			hue.C,
		)
	}

	result := color.DisplaceOKLab(
		ground,
		deltaL,
		distance,
		hue.H,
	)

	if err := requireDisplayGamut(
		fmt.Sprintf(
			"%s %s wash",
			family,
			tier,
		),
		result,
	); err != nil {
		return color.SRGB{}, err
	}

	return result, nil
}

func deriveWashes(p Palette) (Washes, error) {
	var washes Washes

	targets := map[Family]*Wash{
		FamilyMineral: &washes.Mineral,
		FamilyMoss:    &washes.Moss,
		FamilyOchre:   &washes.Ochre,
		FamilyRust:    &washes.Rust,
	}

	for _, family := range Families() {
		line, err := WashOver(
			p,
			family,
			WashLine,
			p.Anchors.Field,
		)
		if err != nil {
			return Washes{}, err
		}

		emphasis, err := WashOver(
			p,
			family,
			WashEmphasis,
			p.Anchors.Field,
		)
		if err != nil {
			return Washes{}, err
		}

		*targets[family] = Wash{
			Line:     line,
			Emphasis: emphasis,
		}
	}

	return washes, nil
}

func washParameters(
	mode Mode,
	tier WashTier,
) (deltaL float64, distance float64, err error) {
	switch {
	case mode == ModeDay && tier == WashLine:
		return 0, dayWashLineDistance, nil

	case mode == ModeNight && tier == WashLine:
		return 0, nightWashLineDistance, nil

	case mode == ModeDay && tier == WashEmphasis:
		return dayWashEmphasisDeltaL, dayWashEmphasisDistance, nil

	case mode == ModeNight && tier == WashEmphasis:
		return nightWashEmphasisDeltaL, nightWashEmphasisDistance, nil
	}

	return 0, 0, fmt.Errorf(
		"derive wash: unsupported mode %q or tier %q",
		mode,
		tier,
	)
}

func familyAnchor(
	p Palette,
	family Family,
) (color.SRGB, error) {
	switch family {
	case FamilyMineral:
		return p.Anchors.Mineral, nil

	case FamilyMoss:
		return p.Anchors.Moss, nil

	case FamilyOchre:
		return p.Anchors.Ochre, nil

	case FamilyRust:
		return p.Anchors.Rust, nil

	default:
		return color.SRGB{}, fmt.Errorf(
			"unknown chromatic family %q",
			family,
		)
	}
}
