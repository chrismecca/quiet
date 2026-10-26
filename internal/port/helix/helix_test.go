// @(#)quiet / go / internal/port/helix/helix_test.go / 2026-10-25

package helix

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
)

var paletteNames = []string{"quiet-day", "quiet-night"}

func TestCommittedThemesMatchGenerator(t *testing.T) {
	for _, name := range paletteNames {
		var got bytes.Buffer
		if err := Encode(&got, loadPalette(t, name)); err != nil {
			t.Fatal(err)
		}

		want, err := os.ReadFile(repoPath("ports", "helix", name+".toml"))
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(got.Bytes(), want) {
			t.Errorf("ports/helix/%s.toml is stale; run make helix", name)
		}
	}
}

var (
	styleLine   = regexp.MustCompile(`^"[a-z][a-z0-9.-]*" = \{ .+ \}$`)
	paletteLine = regexp.MustCompile(`^[a-z][a-z-]* = "#[0-9A-F]{6}"$`)
)

func TestThemeIsWellFormed(t *testing.T) {
	var out bytes.Buffer
	if err := Encode(&out, loadPalette(t, "quiet-day")); err != nil {
		t.Fatal(err)
	}

	inPalette := false

	for _, line := range strings.Split(out.String(), "\n") {
		switch {
		case line == "" || strings.HasPrefix(line, "#"):
		case line == "[palette]":
			inPalette = true
		case inPalette && !paletteLine.MatchString(line):
			t.Errorf("malformed palette line %q", line)
		case !inPalette && !styleLine.MatchString(line):
			t.Errorf("malformed style line %q", line)
		}
	}
}

func TestStylesUseKnownColors(t *testing.T) {
	entries, err := palette(loadPalette(t, "quiet-day"))
	if err != nil {
		t.Fatal(err)
	}

	known := map[string]bool{}
	for _, e := range entries {
		known[e[0]] = true
	}

	seen := map[string]bool{}

	for _, s := range styles() {
		if seen[s.scope] {
			t.Errorf("%s is set twice", s.scope)
		}

		seen[s.scope] = true

		for _, c := range []string{s.fg, s.bg, s.curl} {
			if c != "" && !known[c] {
				t.Errorf("%s uses unknown color %q", s.scope, c)
			}
		}
	}
}

// TestTextIsReadable checks every scope that draws a background: scopes
// without a foreground keep every syntax color at its contract minimum, and
// explicit pairs reach 4.5:1.
func TestTextIsReadable(t *testing.T) {
	exempt := map[string]bool{
		// Callout guarantees only Muted, as in every quiet editor port.
		"ui.cursorline.primary":     true,
		"ui.cursorline.secondary":   true,
		"ui.cursorcolumn.primary":   true,
		"ui.cursorcolumn.secondary": true,
		"ui.gutter.selected":        true,

		// A scrollbar: its foreground is the thumb.
		"ui.menu.scroll": true,
	}

	syntax := map[string]float64{
		"ink":     7,
		"muted":   4.5,
		"mineral": 4.5,
		"moss":    4.5,
		"ochre":   4.5,
		"rust":    4.5,
	}

	for _, name := range paletteNames {
		entries, err := palette(loadPalette(t, name))
		if err != nil {
			t.Fatal(err)
		}

		colors := map[string]color.SRGB{}
		for _, e := range entries {
			colors[e[0]] = color.MustParseHex(e[1])
		}

		for _, s := range styles() {
			if s.bg == "" || exempt[s.scope] {
				continue
			}

			bg := colors[s.bg]

			if s.fg != "" {
				if got := color.Contrast(colors[s.fg], bg); got < 4.5 {
					t.Errorf("%s: %s %s on %s is %.2f, want 4.5", name, s.scope, s.fg, s.bg, got)
				}

				continue
			}

			for fg, minimum := range syntax {
				if got := color.Contrast(colors[fg], bg); got < minimum {
					t.Errorf("%s: %s on %s (%s) is %.2f, want %.1f", name, fg, s.scope, s.bg, got, minimum)
				}
			}
		}
	}
}

func loadPalette(t *testing.T, name string) core.Palette {
	t.Helper()

	p, err := core.LoadPaletteFile(repoPath("spec", name+".json"))
	if err != nil {
		t.Fatal(err)
	}

	return p
}

func repoPath(parts ...string) string {
	return filepath.Join(append([]string{"..", "..", ".."}, parts...)...)
}
