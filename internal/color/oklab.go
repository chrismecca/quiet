// @(#)quiet / go / internal/color/oklab.go / 2026-10-16

package color

import "math"

// OKLab is the rectangular OKLab representation. Quiet uses OKLab for
// perceptually even interpolation of derived colors such as Selection.
type OKLab struct {
	L float64
	A float64
	B float64
}

// OKLCH is the cylindrical form of OKLab. L is perceptual lightness, C is
// chroma, and H is hue in degrees. H is NaN for effectively achromatic colors,
// where a hue angle is not meaningful.
type OKLCH struct {
	L float64
	C float64
	H float64
}

// OKLab converts gamma-encoded sRGB to OKLab through linear sRGB. The matrix
// coefficients are the published OKLab reference transform; keeping them here
// makes the reference implementation dependency-free and auditable.
func (c SRGB) OKLab() OKLab {
	r := srgbToLinear(c.R)
	g := srgbToLinear(c.G)
	b := srgbToLinear(c.B)

	l := 0.4122214708*r +
		0.5363325363*g +
		0.0514459929*b

	m := 0.2119034982*r +
		0.6806995451*g +
		0.1073969566*b

	s := 0.0883024619*r +
		0.2817188376*g +
		0.6299787005*b

	l_ := math.Cbrt(l)
	m_ := math.Cbrt(m)
	s_ := math.Cbrt(s)

	return OKLab{
		L: 0.2104542553*l_ +
			0.7936177850*m_ -
			0.0040720468*s_,

		A: 1.9779984951*l_ -
			2.4285922050*m_ +
			0.4505937099*s_,

		B: 0.0259040371*l_ +
			0.7827717662*m_ -
			0.8086757660*s_,
	}
}

// SRGB converts OKLab back to gamma-encoded sRGB. The result is not clipped;
// callers must use InGamut before treating it as a display/export color.
func (c OKLab) SRGB() SRGB {
	l_ := c.L +
		0.3963377774*c.A +
		0.2158037573*c.B

	m_ := c.L -
		0.1055613458*c.A -
		0.0638541728*c.B

	s_ := c.L -
		0.0894841775*c.A -
		1.2914855480*c.B

	l := l_ * l_ * l_
	m := m_ * m_ * m_
	s := s_ * s_ * s_

	r := 4.0767416621*l -
		3.3077115913*m +
		0.2309699292*s

	g := -1.2684380046*l +
		2.6097574011*m -
		0.3413193965*s

	b := -0.0041960863*l -
		0.7034186147*m +
		1.7076147010*s

	return SRGB{
		R: linearToSRGB(r),
		G: linearToSRGB(g),
		B: linearToSRGB(b),
	}
}

// OKLCH converts rectangular OKLab to cylindrical OKLCH. Hue is undefined for
// an effectively neutral color, so the method returns NaN rather than inventing
// a semantically meaningful angle near the neutral axis.
func (c OKLab) OKLCH() OKLCH {
	chroma := math.Hypot(c.A, c.B)
	hue := math.NaN()

	if chroma > 1e-12 {
		hue = math.Atan2(c.B, c.A) * 180 / math.Pi
		if hue < 0 {
			hue += 360
		}
	}

	return OKLCH{
		L: c.L,
		C: chroma,
		H: hue,
	}
}

// OKLab converts cylindrical OKLCH coordinates to rectangular OKLab.
func (c OKLCH) OKLab() OKLab {
	h := c.H * math.Pi / 180

	return OKLab{
		L: c.L,
		A: c.C * math.Cos(h),
		B: c.C * math.Sin(h),
	}
}

// MixOKLab linearly interpolates between a and b in OKLab. amount=0 returns a
// and amount=1 returns b. Values outside [0, 1] intentionally extrapolate;
// callers that require interpolation-only semantics should validate the input.
func MixOKLab(a, b SRGB, amount float64) SRGB {
	x := a.OKLab()
	y := b.OKLab()

	return OKLab{
		L: x.L + (y.L-x.L)*amount,
		A: x.A + (y.A-x.A)*amount,
		B: x.B + (y.B-x.B)*amount,
	}.SRGB()
}
