// @(#)quiet / go / internal/core/derive_test.go / 2026-10-24

package core

import (
	"testing"

	"github.com/chrismecca/quiet/internal/color"
)

// These outputs were established by the independent prototype and audit
// pipeline before the Go reference implementation existed. They therefore act
// as cross-implementation fixtures rather than expectations derived from the
// implementation under test.
func TestCanonicalDerivedColors(t *testing.T) {
	tests := []struct {
		name      string
		file      string
		selection string
		callout   string
	}{
		{
			name:      "day",
			file:      "quiet-day.json",
			selection: "#D4DDE1",
			callout:   "#E8ECE2",
		},
		{
			name:      "night",
			file:      "quiet-night.json",
			selection: "#344246",
			callout:   "#262D28",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			palette, err := LoadPaletteFile(
				repoPath(
					t,
					"spec",
					tt.file,
				),
			)
			if err != nil {
				t.Fatalf(
					"LoadPaletteFile: %v",
					err,
				)
			}

			derived, err := Derive(palette)
			if err != nil {
				t.Fatalf(
					"Derive: %v",
					err,
				)
			}

			assertHex(
				t,
				"selection",
				derived.Selection.MustHex(),
				tt.selection,
			)

			assertHex(
				t,
				"callout",
				derived.Callout.MustHex(),
				tt.callout,
			)
		})
	}
}

// The historical Subtle fixtures are specifically the appearance of Subtle
// over Field. They are not independent canonical colors.
func TestCanonicalSubtleOverField(t *testing.T) {
	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "day",
			file: "quiet-day.json",
			want: "#A0A49C",
		},
		{
			name: "night",
			file: "quiet-night.json",
			want: "#616762",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			palette, err := LoadPaletteFile(
				repoPath(
					t,
					"spec",
					tt.file,
				),
			)
			if err != nil {
				t.Fatalf(
					"LoadPaletteFile: %v",
					err,
				)
			}

			subtle, err := SubtleOver(
				palette,
				palette.Anchors.Field,
			)
			if err != nil {
				t.Fatalf(
					"SubtleOver: %v",
					err,
				)
			}

			assertHex(
				t,
				"subtle over field",
				subtle.MustHex(),
				tt.want,
			)
		})
	}
}

func TestSubtleDependsOnBackground(t *testing.T) {
	palette, err := LoadPaletteFile(
		repoPath(
			t,
			"spec",
			"quiet-day.json",
		),
	)
	if err != nil {
		t.Fatalf(
			"LoadPaletteFile: %v",
			err,
		)
	}

	onField, err := SubtleOver(
		palette,
		palette.Anchors.Field,
	)
	if err != nil {
		t.Fatalf(
			"SubtleOver field: %v",
			err,
		)
	}

	onSurface, err := SubtleOver(
		palette,
		palette.Anchors.Surface,
	)
	if err != nil {
		t.Fatalf(
			"SubtleOver surface: %v",
			err,
		)
	}

	if onField.MustHex() == onSurface.MustHex() {
		t.Fatal(
			"Subtle did not vary with its local background",
		)
	}
}

// The Wash fixtures were computed by an independent Python implementation of
// the specified derivation from the published OKLab reference transform. They
// cross-check the Go implementation rather than restating its output.
func TestCanonicalWashes(t *testing.T) {
	type pair struct {
		line     string
		emphasis string
	}

	tests := []struct {
		name string
		file string
		want map[Family]pair
	}{
		{
			name: "day",
			file: "quiet-day.json",
			want: map[Family]pair{
				FamilyMineral: {"#E1EEE9", "#C9E7E9"},
				FamilyMoss:    {"#E5EED8", "#D3E8C7"},
				FamilyOchre:   {"#F7E8D1", "#F5DCB8"},
				FamilyRust:    {"#FAE6D4", "#FCD8C0"},
			},
		},
		{
			name: "night",
			file: "quiet-night.json",
			want: map[Family]pair{
				FamilyMineral: {"#102122", "#042A30"},
				FamilyMoss:    {"#172015", "#182A15"},
				FamilyOchre:   {"#211D11", "#2D2309"},
				FamilyRust:    {"#251B16", "#341E15"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			palette, err := LoadPaletteFile(
				repoPath(
					t,
					"spec",
					tt.file,
				),
			)
			if err != nil {
				t.Fatalf(
					"LoadPaletteFile: %v",
					err,
				)
			}

			derived, err := Derive(palette)
			if err != nil {
				t.Fatalf(
					"Derive: %v",
					err,
				)
			}

			for _, family := range Families() {
				wash, err := derived.Wash.For(family)
				if err != nil {
					t.Fatalf(
						"Wash.For(%s): %v",
						family,
						err,
					)
				}

				assertHex(
					t,
					string(family)+" wash",
					wash.Line.MustHex(),
					tt.want[family].line,
				)

				assertHex(
					t,
					string(family)+" wash emphasis",
					wash.Emphasis.MustHex(),
					tt.want[family].emphasis,
				)
			}
		})
	}
}

// The line wash is defined to preserve the ground's lightness; that property
// is what keeps every Field foreground readable on it.
func TestLineWashPreservesGroundLightness(t *testing.T) {
	for _, file := range []string{
		"quiet-day.json",
		"quiet-night.json",
	} {
		palette, err := LoadPaletteFile(
			repoPath(
				t,
				"spec",
				file,
			),
		)
		if err != nil {
			t.Fatalf(
				"LoadPaletteFile: %v",
				err,
			)
		}

		for _, family := range Families() {
			wash, err := WashOver(
				palette,
				family,
				WashLine,
				palette.Anchors.Field,
			)
			if err != nil {
				t.Fatalf(
					"WashOver %s: %v",
					family,
					err,
				)
			}

			got := wash.OKLab().L
			want := palette.Anchors.Field.OKLab().L

			if diff := got - want; diff > 1e-7 || diff < -1e-7 {
				t.Errorf(
					"%s %s wash L = %.9f, want %.9f",
					palette.Name,
					family,
					got,
					want,
				)
			}
		}
	}
}

func TestWashDependsOnGround(t *testing.T) {
	palette, err := LoadPaletteFile(
		repoPath(
			t,
			"spec",
			"quiet-day.json",
		),
	)
	if err != nil {
		t.Fatalf(
			"LoadPaletteFile: %v",
			err,
		)
	}

	onField, err := WashOver(
		palette,
		FamilyMoss,
		WashLine,
		palette.Anchors.Field,
	)
	if err != nil {
		t.Fatalf(
			"WashOver field: %v",
			err,
		)
	}

	onSurface, err := WashOver(
		palette,
		FamilyMoss,
		WashLine,
		palette.Anchors.Surface,
	)
	if err != nil {
		t.Fatalf(
			"WashOver surface: %v",
			err,
		)
	}

	if onField.MustHex() == onSurface.MustHex() {
		t.Fatal(
			"Wash did not vary with its ground",
		)
	}
}

func TestWashRejectsNeutralFamily(t *testing.T) {
	palette := Palette{
		Mode: ModeDay,
		Anchors: Anchors{
			Field: mustTestColor("#EEEADF"),
			Moss:  mustTestColor("#666666"),
		},
	}

	if _, err := WashOver(
		palette,
		FamilyMoss,
		WashLine,
		palette.Anchors.Field,
	); err == nil {
		t.Fatal(
			"WashOver accepted a neutral family anchor",
		)
	}
}

// Wash must fail rather than clip when the displaced ground leaves sRGB. A
// white ground has no room for added chroma at constant lightness.
func TestWashRejectsOutOfGamutResult(t *testing.T) {
	palette := Palette{
		Mode: ModeDay,
		Anchors: Anchors{
			Rust: mustTestColor("#8F5A42"),
		},
	}

	if _, err := WashOver(
		palette,
		FamilyRust,
		WashLine,
		mustTestColor("#FFFFFF"),
	); err == nil {
		t.Fatal(
			"WashOver accepted an out-of-gamut result",
		)
	}
}

func TestWashRejectsUnknownFamilyAndTier(t *testing.T) {
	palette, err := LoadPaletteFile(
		repoPath(
			t,
			"spec",
			"quiet-day.json",
		),
	)
	if err != nil {
		t.Fatalf(
			"LoadPaletteFile: %v",
			err,
		)
	}

	if _, err := WashOver(
		palette,
		Family("cyan"),
		WashLine,
		palette.Anchors.Field,
	); err == nil {
		t.Error(
			"WashOver accepted an unknown family",
		)
	}

	if _, err := WashOver(
		palette,
		FamilyMoss,
		WashTier("heavy"),
		palette.Anchors.Field,
	); err == nil {
		t.Error(
			"WashOver accepted an unknown tier",
		)
	}
}

func TestDeriveRejectsUnsupportedMode(t *testing.T) {
	palette := Palette{
		Mode: Mode("dusk"),
		Anchors: Anchors{
			Field:     mustTestColor("#F3F0E6"),
			Surface:   mustTestColor("#E6E9E2"),
			Structure: mustTestColor("#E7EDF0"),
			Muted:     mustTestColor("#616B65"),
			Mineral:   mustTestColor("#4A6E7A"),
		},
	}

	if _, err := Derive(palette); err == nil {
		t.Fatal(
			"Derive accepted an unsupported mode",
		)
	}
}

func assertHex(
	t *testing.T,
	name string,
	got string,
	want string,
) {
	t.Helper()

	if got != want {
		t.Fatalf(
			"%s = %s, want %s",
			name,
			got,
			want,
		)
	}
}

func mustTestColor(s string) color.SRGB {
	return color.MustParseHex(s)
}
