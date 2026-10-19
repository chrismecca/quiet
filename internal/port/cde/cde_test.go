// @(#)quiet / go / internal/port/cde/cde_test.go / 2026-10-19

package cde

import (
	"bytes"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chrismecca/quiet/internal/core"
)

func TestCanonicalCDEPalettes(t *testing.T) {
	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "day",
			file: "quiet-day.json",
			want: strings.Join(
				[]string{
					"#BA00C300C700",
					"#CA00D000D300",
					"#D400D000C500",
					"#D400D000C500",
					"#CC00CF00C800",
					"#CC00CF00C800",
					"#D400D000C500",
					"#CA00D000D300",
					"",
				},
				"\n",
			),
		},
		{
			name: "night",
			file: "quiet-night.json",
			want: strings.Join(
				[]string{
					"#4F005D006100",
					"#3B0044004700",
					"#340037003400",
					"#340037003400",
					"#3D0043003E00",
					"#3D0043003E00",
					"#340037003400",
					"#3B0044004700",
					"",
				},
				"\n",
			),
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
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

				var buf bytes.Buffer

				if err := Encode(
					&buf,
					resolved,
				); err != nil {
					t.Fatalf(
						"Encode: %v",
						err,
					)
				}

				if got := buf.String(); got != tt.want {
					t.Fatalf(
						"generated CDE palette:\n%s\nwant:\n%s",
						got,
						tt.want,
					)
				}
			},
		)
	}
}

func TestCDEColorSetMapping(t *testing.T) {
	palette, err := core.LoadPaletteFile(
		repoPath(
			t,
			"spec",
			"quiet-night.json",
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

	want := [8]string{
		"#4F5D61",
		"#3B4447",
		"#343734",
		"#343734",
		"#3D433E",
		"#3D433E",
		"#343734",
		"#3B4447",
	}

	for i, value := range resolved.ColorSets {
		got, err := value.Hex()
		if err != nil {
			t.Fatalf(
				"ColorSets[%d].Hex: %v",
				i,
				err,
			)
		}

		if got != want[i] {
			t.Errorf(
				"ColorSets[%d] = %s, want %s",
				i,
				got,
				want[i],
			)
		}
	}
}

func TestContextualLightnessAdaptation(
	t *testing.T,
) {
	tests := []struct {
		name      string
		file      string
		wantDelta float64
	}{
		{
			name:      "day",
			file:      "quiet-day.json",
			wantDelta: dayLightnessDelta,
		},
		{
			name:      "night",
			file:      "quiet-night.json",
			wantDelta: nightLightnessDelta,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
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

				got, err := adaptLightness(
					"Field",
					palette.Anchors.Field,
					tt.wantDelta,
				)
				if err != nil {
					t.Fatalf(
						"adaptLightness: %v",
						err,
					)
				}

				before := palette.Anchors.Field.OKLab()
				after := got.OKLab()

				const tolerance = 1e-6

				if diff := after.L -
					before.L -
					tt.wantDelta; diff < -tolerance ||
					diff > tolerance {
					t.Errorf(
						"L delta = %.9f, want %.9f",
						after.L-before.L,
						tt.wantDelta,
					)
				}

				if diff := after.A - before.A; diff < -tolerance ||
					diff > tolerance {
					t.Errorf(
						"A changed by %.9f",
						diff,
					)
				}

				if diff := after.B - before.B; diff < -tolerance ||
					diff > tolerance {
					t.Errorf(
						"B changed by %.9f",
						diff,
					)
				}
			},
		)
	}
}

func TestEncodeProducesEightColorSets(
	t *testing.T,
) {
	palette, err := core.LoadPaletteFile(
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

	resolved, err := Resolve(palette)
	if err != nil {
		t.Fatalf(
			"Resolve: %v",
			err,
		)
	}

	var buf bytes.Buffer

	if err := Encode(
		&buf,
		resolved,
	); err != nil {
		t.Fatalf(
			"Encode: %v",
			err,
		)
	}

	lines := strings.Split(
		strings.TrimSpace(buf.String()),
		"\n",
	)

	if len(lines) != 8 {
		t.Fatalf(
			"got %d color sets, want 8",
			len(lines),
		)
	}

	for i, line := range lines {
		if len(line) != 13 {
			t.Errorf(
				"color set %d has malformed value %q",
				i+1,
				line,
			)
		}
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
