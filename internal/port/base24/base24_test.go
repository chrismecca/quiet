// @(#)quiet / go / internal/port/base24/base24_test.go / 2026-10-18

package base24

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/chrismecca/quiet/internal/core"

	"github.com/chrismecca/quiet/internal/port/base16"
)

func TestCanonicalProjections(t *testing.T) {
	tests := []struct {
		name string
		file string
		want [24]string
	}{
		{
			name: "day",
			file: "quiet-day.json",
			want: [24]string{
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

				"#F5F1E6",
				"#FDF8EC",

				"#99593C",
				"#8D611F",
				"#56705B",
				"#45726C",
				"#44707F",
				"#7A5E8A",
			},
		},
		{
			name: "night",
			file: "quiet-night.json",
			want: [24]string{
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

				"#171918",
				"#131513",

				"#D07B65",
				"#BC9C5D",
				"#879E83",
				"#77A196",
				"#759EA9",
				"#A189BD",
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

func TestBase16PrefixIsPreserved(t *testing.T) {
	for _, file := range []string{
		"quiet-day.json",
		"quiet-night.json",
	} {
		t.Run(file, func(t *testing.T) {
			palette, err := core.LoadPaletteFile(
				repoPath(
					t,
					"spec",
					file,
				),
			)
			if err != nil {
				t.Fatal(err)
			}

			base16Scheme, err := base16.Resolve(palette)
			if err != nil {
				t.Fatal(err)
			}

			base24Scheme, err := Resolve(palette)
			if err != nil {
				t.Fatal(err)
			}

			for i := 0; i < 16; i++ {
				got := base24Scheme.Palette[i].MustHex()
				want := base16Scheme.Palette[i].MustHex()

				if got != want {
					t.Errorf(
						"base%02X = %s, want Base16 %s",
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
					"base24",
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
					"generated Base24 differs from %s",
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
