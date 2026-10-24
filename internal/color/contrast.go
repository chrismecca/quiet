// @(#)quiet / go / internal/color/contrast.go / 2026-10-24

package color

import "math"

// RelativeLuminance returns the WCAG 2.x relative luminance of an sRGB color.
// The calculation is performed on linear-light sRGB components.
func RelativeLuminance(c SRGB) float64 {
	r := srgbToLinear(c.R)
	g := srgbToLinear(c.G)
	b := srgbToLinear(c.B)

	return 0.2126*r +
		0.7152*g +
		0.0722*b
}

// Contrast returns the WCAG 2.x contrast ratio between a and b. The result is
// symmetric and ranges from 1:1 through 21:1 for in-gamut sRGB colors.
func Contrast(a, b SRGB) float64 {
	yA := RelativeLuminance(a)
	yB := RelativeLuminance(b)

	hi := yA
	lo := yB

	if lo > hi {
		hi, lo = lo, hi
	}

	return (hi + 0.05) /
		(lo + 0.05)
}

// DeltaEOK returns Euclidean distance in OKLab. quiet uses the metric for
// relative/minimum-change comparisons; the project does not assign a universal
// just-noticeable-difference threshold to this value.
func DeltaEOK(a, b SRGB) float64 {
	x := a.OKLab()
	y := b.OKLab()

	dL := x.L - y.L
	dA := x.A - y.A
	dB := x.B - y.B

	return math.Sqrt(
		dL*dL +
			dA*dA +
			dB*dB,
	)
}
