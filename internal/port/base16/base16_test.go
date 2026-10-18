// @(#)quiet / go / internal/port/base16/base16_test.go / 2026-10-18

package base16

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/chrismecca/quiet/internal/core"
)

func TestCanonicalProjections(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		variant string
		want    [16]string
	}{
		{
			name:    "day",
			file:    "quiet-day.json",
			variant: "light",
			want: [16]string{
				"#EEEADF",
				"#E4EAED",
				"#D4DDE1",
				"#616B65",
				"#3C433F",
				"#1B1F1C",
				"#1B1F1C",
				"#1B1F1C",

				"#8F5A42",
				"#85602E",
				"#85602E",
				"#586E5C",
				"#4C706B",
				"#4A6E7A",
				"#755E82",
				"#8F5A42",
			},
		},
		{
			name:    "night",
			file:    "quiet-night.json",
			variant: "dark",
			want: [16]string{
				"#1C1E1C",
				"#222A2D",
				"#344246",
				"#9AA29C",
				"#BDC1B9",
				"#E2E0D7",
				"#E2E0D7",
				"#E2E0D7",

				"#C47D6B",
				"#B49A67",
				"#B49A67",
				"#879A83",
				"#7A9C93",
				"#789AA3",
				"#9C89B4",
				"#C47D6B",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			palette, err := core.LoadPaletteFile(
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

			resolved, err := Resolve(palette)
			if err != nil {
				t.Fatalf(
					"Resolve: %v",
					err,
				)
			}

			if resolved.Variant != tt.variant {
				t.Errorf(
					"variant = %q, want %q",
					resolved.Variant,
					tt.variant,
				)
			}

			for i, want := range tt.want {
				got := resolved.Palette[i].MustHex()

				if got != want {
					t.Errorf(
						"base%02X = %s, want %s",
						i,
						got,
						want,
					)
				}
			}
		})
	}
}

func TestCanonicalYAMLMatchesCommittedArtifacts(t *testing.T) {
	tests := []struct {
		palette  string
		artifact string
	}{
		{
			palette:  "quiet-day.json",
			artifact: "quiet-day.yaml",
		},
		{
			palette:  "quiet-night.json",
			artifact: "quiet-night.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.palette, func(t *testing.T) {
			palette, err := core.LoadPaletteFile(
				repoPath(
					t,
					"spec",
					tt.palette,
				),
			)
			if err != nil {
				t.Fatalf(
					"LoadPaletteFile: %v",
					err,
				)
			}

			scheme, err := Resolve(palette)
			if err != nil {
				t.Fatalf(
					"Resolve: %v",
					err,
				)
			}

			var got bytes.Buffer

			if err := EncodeYAML(
				&got,
				scheme,
			); err != nil {
				t.Fatalf(
					"EncodeYAML: %v",
					err,
				)
			}

			want, err := os.ReadFile(
				repoPath(
					t,
					"ports",
					"base16",
					tt.artifact,
				),
			)
			if err != nil {
				t.Fatalf(
					"ReadFile: %v",
					err,
				)
			}

			if !bytes.Equal(
				got.Bytes(),
				want,
			) {
				t.Fatalf(
					"generated Base16 differs from %s",
					tt.artifact,
				)
			}
		})
	}
}

func repoPath(
	t *testing.T,
	elem ...string,
) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal(
			"runtime.Caller could not locate test source",
		)
	}

	root := filepath.Clean(
		filepath.Join(
			filepath.Dir(file),
			"..",
			"..",
			"..",
		),
	)

	parts := append(
		[]string{root},
		elem...,
	)

	return filepath.Join(parts...)
}
