// @(#)quiet / go / internal/port/terminal/palette_test.go / 2026-10-17

package terminal

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
)

// Expected values preserve continuous precision through all derivations and
// quantize to 8-bit sRGB only at final export.
//
// In particular, bright cyan and magenta are derived from the continuous
// terminal-local extension colors, not from their already-quantized
// hexadecimal representations.
func TestCanonicalANSIReferencePalettes(t *testing.T) {
	tests := []struct {
		name string
		file string
		want [16]string
	}{
		{
			name: "day",
			file: "quiet-day.json",
			want: [16]string{
				"#1B1F1C",
				"#8F5A42",
				"#586E5C",
				"#85602E",
				"#4A6E7A",
				"#755E82",
				"#4C706B",
				"#CDD1CA",

				"#616B65",
				"#99593C",
				"#56705B",
				"#8D611F",
				"#44707F",
				"#7A5E8A",
				"#45726C",
				"#EEEADF",
			},
		},
		{
			name: "night",
			file: "quiet-night.json",
			want: [16]string{
				"#1C1E1C",
				"#C47D6B",
				"#879A83",
				"#B49A67",
				"#789AA3",
				"#9C89B4",
				"#7A9C93",
				"#9AA29C",

				"#343B38",
				"#D07B65",
				"#879E83",
				"#BC9C5D",
				"#759EA9",
				"#A189BD",
				"#77A196",
				"#E2E0D7",
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
				got := resolved.ANSI[i].MustHex()

				if got != want {
					t.Errorf(
						"ANSI %d = %s, want %s",
						i,
						got,
						want,
					)
				}
			}
		})
	}
}

func TestBrightChromaticsIncreaseIntensity(t *testing.T) {
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

			for i := 1; i <= 6; i++ {
				base := resolved.ANSI[i].OKLab().OKLCH()
				bright := resolved.ANSI[i+8].OKLab().OKLCH()

				if bright.L <= base.L {
					t.Errorf(
						"ANSI %d L %.6f <= ANSI %d L %.6f",
						i+8,
						bright.L,
						i,
						base.L,
					)
				}

				if bright.C <= base.C {
					t.Errorf(
						"ANSI %d C %.6f <= ANSI %d C %.6f",
						i+8,
						bright.C,
						i,
						base.C,
					)
				}

				actual := color.Contrast(
					resolved.ANSI[i+8],
					resolved.Background,
				)

				if actual < minimumChromaticContrast {
					t.Errorf(
						"ANSI %d contrast %.6f < %.6f",
						i+8,
						actual,
						minimumChromaticContrast,
					)
				}
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
