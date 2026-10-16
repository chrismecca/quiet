// @(#)quiet / go / internal/core/derive.go / 2026-10-16

package core

import (
	"fmt"

	"github.com/chrismecca/quiet/internal/color"
)

// The constants below are part of the current candidate derivation contract.
//
// They are named rather than embedded in calculations so the relationship
// between implementation, specification, and future audit output remains
// explicit.
const (
	daySelectionMix   = 0.10
	nightSelectionMix = 0.23

	dayCalloutDeltaL = 0.008
	dayCalloutDeltaC = 0.004

	nightCalloutDeltaL = 0.014
	nightCalloutDeltaC = 0.003

	subtleAlpha = 0.55
)

// Derived contains Quiet colors produced deterministically from canonical
// anchors.
//
// These values are outputs, not additional canonical anchors. A conforming
// implementation should reproduce them from the derivation rules rather than
// treating the resulting hexadecimal values as independent source data.
type Derived struct {
	Selection color.SRGB
	Callout   color.SRGB
	Subtle    color.SRGB
}

// Derive applies the candidate Quiet Core derivation rules to palette.
//
// The derivations intentionally use different color operations because the
// roles have different perceptual purposes:
//
//   - Selection interpolates Structure toward Mineral in OKLab.
//   - Callout adjusts Surface lightness and chroma in OKLCH.
//   - Subtle composites Muted over Field in legacy sRGB at 55% opacity.
//
// Every resulting display color must remain representable in sRGB. Derivation
// fails rather than silently clipping an out-of-gamut result.
func Derive(p Palette) (Derived, error) {
	selection, err := deriveSelection(p)
	if err != nil {
		return Derived{}, err
	}

	callout, err := deriveCallout(p)
	if err != nil {
		return Derived{}, err
	}

	subtle, err := deriveSubtle(p)
	if err != nil {
		return Derived{}, err
	}

	return Derived{
		Selection: selection,
		Callout:   callout,
		Subtle:    subtle,
	}, nil
}

func deriveSelection(p Palette) (color.SRGB, error) {
	var amount float64

	switch p.Mode {
	case ModeDay:
		amount = daySelectionMix

	case ModeNight:
		amount = nightSelectionMix

	default:
		return color.SRGB{}, fmt.Errorf(
			"derive selection: unsupported mode %q",
			p.Mode,
		)
	}

	result := color.MixOKLab(
		p.Anchors.Structure,
		p.Anchors.Mineral,
		amount,
	)

	if err := requireDisplayGamut(
		"selection",
		result,
	); err != nil {
		return color.SRGB{}, err
	}

	return result, nil
}

func deriveCallout(p Palette) (color.SRGB, error) {
	source := p.Anchors.Surface.OKLab().OKLCH()

	var deltaL float64
	var deltaC float64

	switch p.Mode {
	case ModeDay:
		deltaL = dayCalloutDeltaL
		deltaC = dayCalloutDeltaC

	case ModeNight:
		deltaL = nightCalloutDeltaL
		deltaC = nightCalloutDeltaC

	default:
		return color.SRGB{}, fmt.Errorf(
			"derive callout: unsupported mode %q",
			p.Mode,
		)
	}

	// Preserve Surface hue while increasing only lightness and chroma. Surface
	// is not exactly neutral in either canonical palette, so its hue is defined.
	adjusted := color.OKLCH{
		L: source.L + deltaL,
		C: source.C + deltaC,
		H: source.H,
	}

	result := adjusted.OKLab().SRGB()

	if err := requireDisplayGamut(
		"callout",
		result,
	); err != nil {
		return color.SRGB{}, err
	}

	return result, nil
}

func deriveSubtle(p Palette) (color.SRGB, error) {
	result, err := color.CompositeLegacySRGB(
		p.Anchors.Muted,
		p.Anchors.Field,
		subtleAlpha,
	)
	if err != nil {
		return color.SRGB{}, fmt.Errorf(
			"derive subtle: %w",
			err,
		)
	}

	if err := requireDisplayGamut(
		"subtle",
		result,
	); err != nil {
		return color.SRGB{}, err
	}

	return result, nil
}

func requireDisplayGamut(
	name string,
	c color.SRGB,
) error {
	if c.InGamut() {
		return nil
	}

	return fmt.Errorf(
		"derived color %q is outside sRGB gamut",
		name,
	)
}
