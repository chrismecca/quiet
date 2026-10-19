// @(#)quiet / go / internal/port/cde/cde.go / 2026-10-19

// Package cde maps Quiet Core semantics onto the CDE high-color palette model.
//
// CDE palette entries are base colors. CDE/Motif derives foreground,
// top-shadow, bottom-shadow, and selection colors from those bases.
//
// Direct Quiet Day and Quiet Night architectural colors were validated on a
// real CDE desktop and found to operate too near the ends of CDE's usable
// lightness range. This port therefore applies a documented contextual OKLab
// lightness adaptation before serialization.
package cde

import (
	"fmt"
	"io"
	"strings"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
)

const (
	dayLightnessDelta   = -0.08
	nightLightnessDelta = +0.10
)

// Palette is the resolved eight-color CDE high-color palette.
type Palette struct {
	Name      string
	ColorSets [8]color.SRGB
}

// Resolve maps one canonical Quiet palette onto CDE's eight high-color sets.
//
// Mapping:
//
//	1  active window frame       Selection
//	2  inactive window frame     Structure
//	3  workspace 1               Field
//	4  text/list entry           Field
//	5  application background    Surface
//	6  menus/dialogs/dtterm      Surface
//	7  workspace 4               Field
//	8  Front Panel               Structure
func Resolve(p core.Palette) (Palette, error) {
	derived, err := core.Derive(p)
	if err != nil {
		return Palette{}, fmt.Errorf(
			"derive Quiet roles: %w",
			err,
		)
	}

	delta, err := lightnessDelta(p.Name)
	if err != nil {
		return Palette{}, err
	}

	selection, err := adaptLightness(
		"Selection",
		derived.Selection,
		delta,
	)
	if err != nil {
		return Palette{}, err
	}

	structure, err := adaptLightness(
		"Structure",
		p.Anchors.Structure,
		delta,
	)
	if err != nil {
		return Palette{}, err
	}

	field, err := adaptLightness(
		"Field",
		p.Anchors.Field,
		delta,
	)
	if err != nil {
		return Palette{}, err
	}

	surface, err := adaptLightness(
		"Surface",
		p.Anchors.Surface,
		delta,
	)
	if err != nil {
		return Palette{}, err
	}

	return Palette{
		Name: p.Name,
		ColorSets: [8]color.SRGB{
			selection,
			structure,
			field,
			field,
			surface,
			surface,
			field,
			structure,
		},
	}, nil
}

// Encode writes a native CDE .dp palette.
//
// CDE accepts sixteen-bit X color components. Quiet deliberately exports the
// validated eight-bit sRGB result into the high byte of each component:
//
//	#RRGGBB -> #RR00GG00BB00
//
// This reproduces the representation validated on the reference CDE system.
func Encode(
	w io.Writer,
	p Palette,
) error {
	var b strings.Builder

	for i, value := range p.ColorSets {
		encoded, err := encodeColor(value)
		if err != nil {
			return fmt.Errorf(
				"color set %d: %w",
				i+1,
				err,
			)
		}

		fmt.Fprintln(
			&b,
			encoded,
		)
	}

	if _, err := io.WriteString(
		w,
		b.String(),
	); err != nil {
		return fmt.Errorf(
			"write CDE palette: %w",
			err,
		)
	}

	return nil
}

func lightnessDelta(name string) (float64, error) {
	switch name {
	case "Quiet Day":
		return dayLightnessDelta, nil

	case "Quiet Night":
		return nightLightnessDelta, nil

	default:
		return 0, fmt.Errorf(
			"unsupported CDE palette %q",
			name,
		)
	}
}

func adaptLightness(
	name string,
	value color.SRGB,
	delta float64,
) (color.SRGB, error) {
	lab := value.OKLab()
	lab.L += delta

	result := lab.SRGB()

	if !result.InGamut() {
		return color.SRGB{}, fmt.Errorf(
			"%s CDE adaptation is outside sRGB gamut",
			name,
		)
	}

	return result, nil
}

func encodeColor(value color.SRGB) (string, error) {
	hex, err := value.Hex()
	if err != nil {
		return "", err
	}

	if len(hex) != 7 {
		return "", fmt.Errorf(
			"unexpected sRGB serialization %q",
			hex,
		)
	}

	return fmt.Sprintf(
		"#%s00%s00%s00",
		hex[1:3],
		hex[3:5],
		hex[5:7],
	), nil
}
