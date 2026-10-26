// @(#)quiet / go / internal/color/composite.go / 2026-10-26

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
		R: mix(foreground.R, background.R, alpha),
		G: mix(foreground.G, background.G, alpha),
		B: mix(foreground.B, background.B, alpha),
	}, nil
}

// mix returns a*alpha + b*(1-alpha), rounding each product.
//
// Go may fuse a multiply and an add into one instruction on some
// architectures, such as arm64, and not on others, such as amd64. Fused and
// unfused results can differ in the last bit, enough to change a quantized
// 8-bit channel, so generated artifacts would depend on the machine that
// generated them. An explicit conversion rounds and prevents fusion.
func mix(a, b, alpha float64) float64 {
	return float64(a*alpha) + float64(b*(1-alpha))
}
