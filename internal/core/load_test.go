// @(#)quiet / go / internal/core/load_test.go / 2026-10-24

package core

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCanonicalPalettes(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		mode     Mode
		title    string
		expected map[string]string
	}{
		{
			name:  "day",
			file:  "quiet-day.json",
			mode:  ModeDay,
			title: "quiet-day",
			expected: map[string]string{
				"field":     "#EEEADF",
				"surface":   "#E6E9E2",
				"structure": "#E4EAED",
				"ink":       "#1B1F1C",
				"muted":     "#616B65",
				"line":      "#CDD1CA",
				"mineral":   "#4A6E7A",
				"moss":      "#586E5C",
				"ochre":     "#85602E",
				"rust":      "#8F5A42",
			},
		},
		{
			name:  "night",
			file:  "quiet-night.json",
			mode:  ModeNight,
			title: "quiet-night",
			expected: map[string]string{
				"field":     "#1C1E1C",
				"surface":   "#242925",
				"structure": "#222A2D",
				"ink":       "#E2E0D7",
				"muted":     "#9AA29C",
				"line":      "#343B38",
				"mineral":   "#789AA3",
				"moss":      "#879A83",
				"ochre":     "#B49A67",
				"rust":      "#C47D6B",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			palette, err := LoadPaletteFile(
				repoPath(t, "spec", tt.file),
			)
			if err != nil {
				t.Fatalf(
					"LoadPaletteFile: %v",
					err,
				)
			}

			if palette.Schema != PaletteSchema {
				t.Errorf(
					"schema = %q, want %q",
					palette.Schema,
					PaletteSchema,
				)
			}

			if palette.Name != tt.title {
				t.Errorf(
					"name = %q, want %q",
					palette.Name,
					tt.title,
				)
			}

			if palette.Mode != tt.mode {
				t.Errorf(
					"mode = %q, want %q",
					palette.Mode,
					tt.mode,
				)
			}

			if palette.Status != StatusCandidate {
				t.Errorf(
					"status = %q, want %q",
					palette.Status,
					StatusCandidate,
				)
			}

			anchors := palette.Anchors.All()

			if len(anchors) != AnchorCount {
				t.Fatalf(
					"anchor count = %d, want %d",
					len(anchors),
					AnchorCount,
				)
			}

			for _, anchor := range anchors {
				want, ok := tt.expected[anchor.Name]
				if !ok {
					t.Errorf(
						"unexpected anchor %q",
						anchor.Name,
					)
					continue
				}

				got := anchor.Color.MustHex()

				if got != want {
					t.Errorf(
						"%s = %s, want %s",
						anchor.Name,
						got,
						want,
					)
				}
			}
		})
	}
}

func TestDecodePaletteRejectsUnknownFields(t *testing.T) {
	const document = `{
		"schema": "quiet.palette/v1",
		"name": "quiet-day",
		"mode": "day",
		"status": "candidate",
		"unexpected": true,
		"anchors": {}
	}`

	_, err := DecodePalette(
		strings.NewReader(document),
	)

	if err == nil {
		t.Fatal(
			"DecodePalette accepted an unknown field",
		)
	}
}

func TestDecodePaletteRejectsUnknownMode(t *testing.T) {
	const document = `{
		"schema": "quiet.palette/v1",
		"name": "quiet-dusk",
		"mode": "dusk",
		"status": "candidate",
		"anchors": {}
	}`

	_, err := DecodePalette(
		strings.NewReader(document),
	)

	if err == nil {
		t.Fatal(
			"DecodePalette accepted an unknown mode",
		)
	}
}

func TestDecodePaletteRejectsNonCanonicalAnchorEncoding(
	t *testing.T,
) {
	const document = `{
		"schema": "quiet.palette/v1",
		"name": "Test",
		"mode": "day",
		"status": "candidate",
		"anchors": {
			"field": "#ffffff",
			"surface": "#000000",
			"structure": "#000000",
			"ink": "#000000",
			"muted": "#000000",
			"line": "#000000",
			"mineral": "#000000",
			"moss": "#000000",
			"ochre": "#000000",
			"rust": "#000000"
		}
	}`

	_, err := DecodePalette(
		strings.NewReader(document),
	)

	if err == nil {
		t.Fatal(
			"DecodePalette accepted non-canonical anchor encoding",
		)
	}
}

// repoPath locates repository-owned test data without depending on the process
// working directory used by `go test`.
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
		),
	)

	parts := append(
		[]string{root},
		elem...,
	)

	return filepath.Join(parts...)
}
