// @(#)quiet / go / internal/color/cvd_test.go / 2026-10-24

package color

import "testing"

// Pure linear primaries select one column of each simulation matrix, so the
// expected values below are read directly from the published Machado 2009
// full-severity table rather than computed by the code under test.
func TestSimulateCVDPrimariesMatchPublishedColumns(t *testing.T) {
	tests := []struct {
		deficiency Deficiency
		input      SRGB
		want       [3]float64
	}{
		{Protanopia, SRGB{R: 1}, [3]float64{0.152286, 0.114503, -0.003882}},
		{Protanopia, SRGB{G: 1}, [3]float64{1.052583, 0.786281, -0.048116}},
		{Deuteranopia, SRGB{R: 1}, [3]float64{0.367322, 0.280085, -0.011820}},
		{Deuteranopia, SRGB{B: 1}, [3]float64{-0.227968, 0.047413, 0.968881}},
		{Tritanopia, SRGB{G: 1}, [3]float64{-0.076749, 0.930809, 0.691367}},
		{Tritanopia, SRGB{B: 1}, [3]float64{-0.178779, 0.147602, 0.303900}},
	}

	for _, tt := range tests {
		got, err := SimulateCVD(
			tt.input,
			tt.deficiency,
		)
		if err != nil {
			t.Fatalf(
				"SimulateCVD: %v",
				err,
			)
		}

		for i, component := range []float64{
			srgbToLinear(got.R),
			srgbToLinear(got.G),
			srgbToLinear(got.B),
		} {
			assertNear(
				t,
				string(tt.deficiency),
				component,
				tt.want[i],
				1e-9,
			)
		}
	}
}

// Every matrix row sums to 1, so the model leaves neutral colors unchanged.
func TestSimulateCVDPreservesNeutrals(t *testing.T) {
	gray := MustParseHex("#808080")

	for _, d := range Deficiencies() {
		got, err := SimulateCVD(
			gray,
			d,
		)
		if err != nil {
			t.Fatalf(
				"SimulateCVD: %v",
				err,
			)
		}

		if got.MustHex() != "#808080" {
			t.Errorf(
				"%s gray = %s, want #808080",
				d,
				got.MustHex(),
			)
		}
	}
}

func TestSimulateCVDRejectsUnknownDeficiency(t *testing.T) {
	if _, err := SimulateCVD(
		SRGB{},
		Deficiency("achromatopsia"),
	); err == nil {
		t.Fatal(
			"SimulateCVD accepted an unknown deficiency",
		)
	}
}
