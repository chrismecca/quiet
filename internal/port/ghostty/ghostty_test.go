// @(#)quiet / go / internal/port/ghostty/ghostty_test.go / 2026-10-24

package ghostty

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
			artifact: "quiet-day",
		},
		{
			palette:  "quiet-night.json",
			artifact: "quiet-night",
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
					"ghostty",
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
					"generated Ghostty theme differs from %q",
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
		"background = " + resolved.Background.MustHex(),
		"foreground = " + resolved.Text.MustHex(),
		"cursor-color = " + resolved.Cursor.MustHex(),
		"cursor-text = " + resolved.Background.MustHex(),
		"selection-background = " + resolved.Selection.MustHex(),
		"selection-foreground = " + resolved.Text.MustHex(),
	}

	for i, value := range resolved.ANSI {
		want = append(
			want,
			fmt.Sprintf(
				"palette = %d=%s",
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
				"Ghostty theme missing %q",
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
		"font-",
		"window-",
		"cursor-style",
		"background-opacity",
		"keybind",
		"shell-integration",
	} {
		if strings.Contains(
			output.String(),
			forbidden,
		) {
			t.Errorf(
				"Ghostty port unexpectedly defines non-color policy %q",
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
