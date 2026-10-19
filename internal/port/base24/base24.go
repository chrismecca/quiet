// @(#)quiet / go / internal/port/base24/base24.go / 2026-10-18

// Package base24 projects Quiet Core onto the Base24 compatibility model.
package base24

import (
	"fmt"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
	"github.com/chrismecca/quiet/internal/port/base16"
	"github.com/chrismecca/quiet/internal/port/terminal"
)

const (
	// Base24 requires two background colors beyond base00.
	//
	// Quiet has no canonical roles beyond Field in that direction. These
	// values therefore extrapolate from Field away from Ink in OKLab. The
	// modest distances preserve Field's neutral character while creating the
	// additional background levels required by Base24.
	backgroundNear = -0.03
	backgroundFar  = -0.06
)

// Scheme is a resolved Base24 compatibility scheme.
//
// Base24 extends Base16 with two additional background levels and six bright
// chromatic colors. It is a compatibility representation, not Quiet Core.
type Scheme struct {
	Name        string
	Slug        string
	Author      string
	Description string
	Variant     string
	Palette     [24]color.SRGB
}

// Resolve projects one Quiet palette onto Base24.
func Resolve(p core.Palette) (Scheme, error) {
	base, err := base16.Resolve(p)
	if err != nil {
		return Scheme{}, fmt.Errorf(
			"resolve Base16 compatibility projection: %w",
			err,
		)
	}

	ansi, err := terminal.Resolve(p)
	if err != nil {
		return Scheme{}, fmt.Errorf(
			"resolve terminal compatibility palette: %w",
			err,
		)
	}

	base10 := color.MixOKLab(
		p.Anchors.Field,
		p.Anchors.Ink,
		backgroundNear,
	)

	base11 := color.MixOKLab(
		p.Anchors.Field,
		p.Anchors.Ink,
		backgroundFar,
	)

	if !base10.InGamut() {
		return Scheme{}, fmt.Errorf(
			"Base24 base10 is outside sRGB gamut",
		)
	}

	if !base11.InGamut() {
		return Scheme{}, fmt.Errorf(
			"Base24 base11 is outside sRGB gamut",
		)
	}

	result := Scheme{
		Name:        base.Name,
		Slug:        base.Slug,
		Author:      base.Author,
		Description: "A Base24 compatibility projection of " + p.Name + ".",
		Variant:     base.Variant,
	}

	copy(
		result.Palette[:16],
		base.Palette[:],
	)

	result.Palette[16] = base10
	result.Palette[17] = base11

	// Base24's six additional chromatic slots are the conventional bright
	// ANSI families. Reuse the generic terminal reference mapping rather than
	// defining a second bright-color derivation.
	result.Palette[18] = ansi.ANSI[9]  // base12: bright red
	result.Palette[19] = ansi.ANSI[11] // base13: bright yellow
	result.Palette[20] = ansi.ANSI[10] // base14: bright green
	result.Palette[21] = ansi.ANSI[14] // base15: bright cyan
	result.Palette[22] = ansi.ANSI[12] // base16: bright blue
	result.Palette[23] = ansi.ANSI[13] // base17: bright magenta

	return result, nil
}
