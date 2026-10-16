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

// Derived contains opaque Quiet colors produced deterministically from the
// canonical anchors.
//
// Subtle is intentionally absent. Unlike Selection and Callout, Subtle is not
// one opaque derived color: it is a compositing rule whose displayed value
// depends on the local background.
type Derived struct {
	Selection color.SRGB
	Callout   color.SRGB
}

// Derive applies the opaque candidate Quiet Core derivation rules to palette.
//
// The derivations intentionally use different color operations because the
// roles have different perceptual purposes:
//
//   - Selection interpolates Structure toward Mineral in OKLab.
//   - Callout adjusts Surface lightness and chroma in OKLCH.
//
// Subtle is evaluated separately with SubtleOver because its output depends on
// the local background.
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

	return Derived{
		Selection: selection,
		Callout:   callout,
	}, nil
}

// SubtleOver returns the displayed Subtle color when Muted is composited over
// background at the Quiet Core Subtle opacity.
//
// The operation intentionally uses conventional gamma-encoded sRGB
// compositing. Subtle models ordinary interface opacity behavior; it is not a
// perceptual interpolation.
//
// The background is supplied by the caller because Subtle is local to the
// surface on which it appears. There is therefore no single canonical opaque
// hexadecimal value for Subtle.
func SubtleOver(
	p Palette,
	background color.SRGB,
) (color.SRGB, error) {
	result, err := color.CompositeLegacySRGB(
		p.Anchors.Muted,
		background,
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
