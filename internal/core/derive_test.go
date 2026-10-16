// @(#)quiet / go / internal/core/derive_test.go / 2026-10-16

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
			selection: "#D6E0E4",
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
			want: "#A3A79F",
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
