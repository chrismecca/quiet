// @(#)quiet / go / internal/color/composite.go / 2026-10-24

package color

import "fmt"

// CompositeLegacySRGB composites foreground over an opaque background using
// alpha in gamma-encoded sRGB.
//
// quiet uses this operation for legacy/display semantics such as Subtle, where
// the intended result corresponds to conventional sRGB alpha compositing rather
// than interpolation in a perceptual color space.
//
// alpha must be in [0, 1]. The result may still be checked with InGamut, though
// compositing two in-gamut opaque sRGB colors at a valid alpha necessarily
// remains in gamut.
func CompositeLegacySRGB(
	foreground SRGB,
	background SRGB,
	alpha float64,
) (SRGB, error) {
	if alpha < 0 || alpha > 1 {
		return SRGB{}, fmt.Errorf(
			"alpha %.6f outside [0, 1]",
			alpha,
		)
	}

	return SRGB{
		R: foreground.R*alpha + background.R*(1-alpha),
		G: foreground.G*alpha + background.G*(1-alpha),
		B: foreground.B*alpha + background.B*(1-alpha),
	}, nil
}
