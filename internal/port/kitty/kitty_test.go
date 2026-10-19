// @(#)quiet / go / internal/port/kitty/kitty_test.go / 2026-10-19

package kitty

import (
	"bytes"
	"fmt"
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
			artifact: "Quiet Day.conf",
		},
		{
			palette:  "quiet-night.json",
			artifact: "Quiet Night.conf",
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
					"kitty",
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
					"generated kitty theme differs from %q",
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
		"foreground            " + resolved.Text.MustHex(),
		"background            " + resolved.Background.MustHex(),
		"cursor                " + resolved.Cursor.MustHex(),
		"cursor_text_color     " + resolved.Background.MustHex(),
		"selection_foreground  " + resolved.Text.MustHex(),
		"selection_background  " + resolved.Selection.MustHex(),
		"url_color             " + resolved.Cursor.MustHex(),
	}

	for i, value := range resolved.ANSI {
		want = append(
			want,
			fmt.Sprintf(
				"color%-2d               %s",
				i,
				value.MustHex(),
			),
		)
	}

	for _, line := range want {
		if !strings.Contains(
			got,
			line,
		) {
			t.Errorf(
				"kitty theme missing %q",
				line,
			)
		}
	}
}

func TestThemeContainsOnlyColorPolicy(t *testing.T) {
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
		"font_",
		"background_opacity",
		"macos_titlebar_color",
		"tab_",
		"window_",
		"cursor_shape",
		"keymap",
		"map ",
	} {
		if strings.Contains(
			output.String(),
			forbidden,
		) {
			t.Errorf(
				"kitty port unexpectedly defines non-color policy %q",
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
