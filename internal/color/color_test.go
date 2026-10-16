// @(#)quiet / go / internal/color/color_test.go / 2026-10-16

package color

import (
	"math"
	"testing"
)

// The Quiet-specific fixtures below were frozen from the independent Python
// prototype. They intentionally cross-check this Go implementation rather than
// deriving expected values from the code under test.
const (
	quietDayMineral = "#4A6E7A"
	quietDayMuted   = "#616B65"
	quietDaySurface = "#E6E9E2"
)

func TestParseHexRoundTrip(t *testing.T) {
	got, err := ParseHex("#4a6e7a")
	if err != nil {
		t.Fatalf("ParseHex: %v", err)
	}

	hex, err := got.Hex()
	if err != nil {
		t.Fatalf("Hex: %v", err)
	}

	if hex != quietDayMineral {
		t.Fatalf(
			"round trip = %s, want %s",
			hex,
			quietDayMineral,
		)
	}
}

func TestParseHexRejectsNonCanonicalLengths(t *testing.T) {
	for _, input := range []string{
		"#abc",
		"#11223344",
		"",
	} {
		if _, err := ParseHex(input); err == nil {
			t.Errorf(
				"ParseHex(%q) succeeded; want error",
				input,
			)
		}
	}
}

func TestHexRejectsOutOfGamut(t *testing.T) {
	_, err := (SRGB{
		R: 1.01,
		G: 0.5,
		B: 0.5,
	}).Hex()

	if err == nil {
		t.Fatal(
			"Hex accepted an out-of-gamut color",
		)
	}
}

func TestBlackWhiteContrast(t *testing.T) {
	got := Contrast(
		MustParseHex("#000000"),
		MustParseHex("#FFFFFF"),
	)

	assertNear(
		t,
		"contrast",
		got,
		21.0,
		1e-12,
	)
}

func TestQuietDayMineralOKLCH(t *testing.T) {
	got := MustParseHex(
		quietDayMineral,
	).OKLab().OKLCH()

	assertNear(
		t,
		"L",
		got.L,
		0.5156759,
		1e-6,
	)

	assertNear(
		t,
		"C",
		got.C,
		0.0450841,
		1e-6,
	)

	assertNear(
		t,
		"H",
		got.H,
		220.7693,
		1e-3,
	)
}

func TestQuietDayMutedSurfaceContrast(t *testing.T) {
	got := Contrast(
		MustParseHex(quietDayMuted),
		MustParseHex(quietDaySurface),
	)

	assertNear(
		t,
		"contrast",
		got,
		4.5056946,
		1e-6,
	)
}

func TestQuietDaySelectionMix(t *testing.T) {
	got := MixOKLab(
		MustParseHex("#E7EDF0"),
		MustParseHex(quietDayMineral),
		0.10,
	).MustHex()

	if got != "#D6E0E4" {
		t.Fatalf(
			"selection = %s, want #D6E0E4",
			got,
		)
	}
}

func TestOKLabRoundTrip(t *testing.T) {
	want := MustParseHex(
		quietDayMineral,
	)

	got := want.OKLab().SRGB()

	assertNear(
		t,
		"R",
		got.R,
		want.R,
		1e-7,
	)

	assertNear(
		t,
		"G",
		got.G,
		want.G,
		1e-7,
	)

	assertNear(
		t,
		"B",
		got.B,
		want.B,
		1e-7,
	)
}

func assertNear(
	t *testing.T,
	name string,
	got float64,
	want float64,
	tolerance float64,
) {
	t.Helper()

	if math.Abs(got-want) > tolerance {
		t.Fatalf(
			"%s = %.10f, want %.10f ± %.1e",
			name,
			got,
			want,
			tolerance,
		)
	}
}
