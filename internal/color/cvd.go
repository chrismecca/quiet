// @(#)quiet / go / internal/color/cvd.go / 2026-10-24

package color

import "fmt"

// Deficiency names a simulated dichromatic color-vision deficiency.
type Deficiency string

const (
	Protanopia   Deficiency = "protanopia"
	Deuteranopia Deficiency = "deuteranopia"
	Tritanopia   Deficiency = "tritanopia"
)

// Deficiencies returns the simulated deficiencies in conventional order.
func Deficiencies() []Deficiency {
	return []Deficiency{
		Protanopia,
		Deuteranopia,
		Tritanopia,
	}
}

// machadoSeverity1 holds the full-severity simulation matrices published by
// Machado, Oliveira, and Fernandes, "A Physiologically-based Model for
// Simulation of Color Vision Deficiency", IEEE TVCG 15(6), 2009. Each matrix
// maps linear sRGB to simulated linear sRGB, and each row sums to 1 within
// the published six-decimal precision, so neutrals are preserved.
var machadoSeverity1 = map[Deficiency][3][3]float64{
	Protanopia: {
		{0.152286, 1.052583, -0.204868},
		{0.114503, 0.786281, 0.099216},
		{-0.003882, -0.048116, 1.051998},
	},
	Deuteranopia: {
		{0.367322, 0.860646, -0.227968},
		{0.280085, 0.672501, 0.047413},
		{-0.011820, 0.042940, 0.968881},
	},
	Tritanopia: {
		{1.255528, -0.076749, -0.178779},
		{-0.078411, 0.930809, 0.147602},
		{0.004733, 0.691367, 0.303900},
	},
}

// SimulateCVD returns c as perceived under deficiency d at full severity,
// according to the Machado 2009 model.
//
// The result is an analysis value, not a display color. It is neither clipped
// nor gamut-checked: a simulated color may fall slightly outside sRGB, and the
// OKLab conversion accepts such values, so comparisons remain exact.
//
// Simulation is a screening tool. It does not replace evaluation with people
// who have color-vision deficiencies.
func SimulateCVD(c SRGB, d Deficiency) (SRGB, error) {
	m, ok := machadoSeverity1[d]
	if !ok {
		return SRGB{}, fmt.Errorf(
			"unknown color-vision deficiency %q",
			d,
		)
	}

	r := srgbToLinear(c.R)
	g := srgbToLinear(c.G)
	b := srgbToLinear(c.B)

	return SRGB{
		R: linearToSRGB(m[0][0]*r + m[0][1]*g + m[0][2]*b),
		G: linearToSRGB(m[1][0]*r + m[1][1]*g + m[1][2]*b),
		B: linearToSRGB(m[2][0]*r + m[2][1]*g + m[2][2]*b),
	}, nil
}
