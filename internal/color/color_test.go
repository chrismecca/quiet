// @(#)quiet / go / internal/color/color_test.go / 2026-10-24

package color

import (
	"math"
	"testing"
)

// The quiet-specific fixtures below were frozen from the independent Python
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

func TestRGB8MatchesHex(t *testing.T) {
	r, g, b, err := MustParseHex(quietDayMineral).RGB8()
	if err != nil {
		t.Fatalf("RGB8: %v", err)
	}

	if r != 0x4A || g != 0x6E || b != 0x7A {
		t.Fatalf(
			"RGB8 = %d, %d, %d, want 74, 110, 122",
			r,
			g,
			b,
		)
	}
}

func TestRGB8RejectsOutOfGamut(t *testing.T) {
	_, _, _, err := (SRGB{
		R: 0.5,
		G: -0.01,
		B: 0.5,
	}).RGB8()

	if err == nil {
		t.Fatal(
			"RGB8 accepted an out-of-gamut color",
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

// The displacement contract is geometric: lightness moves only by deltaL and
// the OKLab distance equals hypot(deltaL, distance), whatever the hue. The
// tolerance matches TestOKLabRoundTrip because measuring the result converts
// back through sRGB with the published ten-digit matrix coefficients.
func TestDisplaceOKLabGeometry(t *testing.T) {
	base := MustParseHex(quietDaySurface)

	for _, hue := range []float64{
		0,
		72.4,
		220.8,
	} {
		got := DisplaceOKLab(
			base,
			-0.03,
			0.04,
			hue,
		)

		assertNear(
			t,
			"L",
			got.OKLab().L,
			base.OKLab().L-0.03,
			1e-7,
		)

		assertNear(
			t,
			"distance",
			DeltaEOK(got, base),
			0.05,
			1e-7,
		)
	}
}

func TestDisplaceOKLabZeroIsIdentity(t *testing.T) {
	want := MustParseHex(quietDayMineral)

	got := DisplaceOKLab(
		want,
		0,
		0,
		123,
	)

	if got.MustHex() != quietDayMineral {
		t.Fatalf(
			"zero displacement = %s, want %s",
			got.MustHex(),
			quietDayMineral,
		)
	}
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

func TestCompositeLegacySRGBEndpoints(t *testing.T) {
	foreground := MustParseHex("#123456")
	background := MustParseHex("#ABCDEF")

	atZero, err := CompositeLegacySRGB(
		foreground,
		background,
		0,
	)
	if err != nil {
		t.Fatalf(
			"CompositeLegacySRGB alpha=0: %v",
			err,
		)
	}

	if got := atZero.MustHex(); got != "#ABCDEF" {
		t.Fatalf(
			"alpha=0 = %s, want #ABCDEF",
			got,
		)
	}

	atOne, err := CompositeLegacySRGB(
		foreground,
		background,
		1,
	)
	if err != nil {
		t.Fatalf(
			"CompositeLegacySRGB alpha=1: %v",
			err,
		)
	}

	if got := atOne.MustHex(); got != "#123456" {
		t.Fatalf(
			"alpha=1 = %s, want #123456",
			got,
		)
	}
}

func TestCompositeLegacySRGBRejectsInvalidAlpha(t *testing.T) {
	for _, alpha := range []float64{
		-0.01,
		1.01,
	} {
		if _, err := CompositeLegacySRGB(
			MustParseHex("#000000"),
			MustParseHex("#FFFFFF"),
			alpha,
		); err == nil {
			t.Errorf(
				"alpha %.2f succeeded; want error",
				alpha,
			)
		}
	}
}
