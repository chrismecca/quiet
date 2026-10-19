// @(#)quiet / go / internal/port/wezterm/wezterm_test.go / 2026-10-19

package wezterm

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chrismecca/quiet/internal/core"
	"github.com/chrismecca/quiet/internal/port/terminal"
)

func TestCanonicalThemesMatchCommittedArtifacts(t *testing.T) {
	tests := []struct {
		palette  string
		artifact string
	}{
		{
			palette:  "quiet-day.json",
			artifact: "Quiet Day.toml",
		},
		{
			palette:  "quiet-night.json",
			artifact: "Quiet Night.toml",
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
				t.Fatal(err)
			}

			resolved, err := terminal.Resolve(palette)
			if err != nil {
				t.Fatal(err)
			}

			var got bytes.Buffer

			if err := Encode(
				&got,
				resolved,
			); err != nil {
				t.Fatal(err)
			}

			want, err := os.ReadFile(
				repoPath(
					t,
					"ports",
					"wezterm",
					tt.artifact,
				),
			)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(
				got.Bytes(),
				want,
			) {
				t.Fatalf(
					"generated WezTerm theme differs from %q",
					tt.artifact,
				)
			}
		})
	}
}

func TestThemePreservesGenericTerminalPalette(t *testing.T) {
	palette, err := core.LoadPaletteFile(
		repoPath(
			t,
			"spec",
			"quiet-day.json",
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	resolved, err := terminal.Resolve(palette)
	if err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer

	if err := Encode(
		&output,
		resolved,
	); err != nil {
		t.Fatal(err)
	}

	got := output.String()

	want := []string{
		`foreground = "` + resolved.Text.MustHex() + `"`,
		`background = "` + resolved.Background.MustHex() + `"`,
		`cursor_bg = "` + resolved.Cursor.MustHex() + `"`,
		`cursor_border = "` + resolved.Cursor.MustHex() + `"`,
		`cursor_fg = "` + resolved.Background.MustHex() + `"`,
		`selection_bg = "` + resolved.Selection.MustHex() + `"`,
		`selection_fg = "` + resolved.Text.MustHex() + `"`,
	}

	for _, line := range want {
		if !strings.Contains(
			got,
			line,
		) {
			t.Errorf(
				"WezTerm theme missing %q",
				line,
			)
		}
	}

	for i, value := range resolved.ANSI {
		if !strings.Contains(
			got,
			value.MustHex(),
		) {
			t.Errorf(
				"WezTerm theme missing ANSI %d %s",
				i,
				value.MustHex(),
			)
		}
	}
}

func TestThemeContainsExactlyTwoANSIRanges(t *testing.T) {
	palette, err := core.LoadPaletteFile(
		repoPath(
			t,
			"spec",
			"quiet-day.json",
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	resolved, err := terminal.Resolve(palette)
	if err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer

	if err := Encode(
		&output,
		resolved,
	); err != nil {
		t.Fatal(err)
	}

	got := output.String()

	for _, key := range []string{
		"ansi = [",
		"brights = [",
	} {
		if strings.Count(
			got,
			key,
		) != 1 {
			t.Errorf(
				"expected exactly one %q",
				key,
			)
		}
	}
}

func TestThemeDoesNotSkinWezTerm(t *testing.T) {
	palette, err := core.LoadPaletteFile(
		repoPath(
			t,
			"spec",
			"quiet-day.json",
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	resolved, err := terminal.Resolve(palette)
	if err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer

	if err := Encode(
		&output,
		resolved,
	); err != nil {
		t.Fatal(err)
	}

	for _, forbidden := range []string{
		"tab_bar",
		"scrollbar_thumb",
		"split =",
		"visual_bell",
		"window_frame",
		"font",
	} {
		if strings.Contains(
			output.String(),
			forbidden,
		) {
			t.Errorf(
				"WezTerm port unexpectedly defines %q",
				forbidden,
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
