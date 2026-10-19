// @(#)quiet / go / internal/port/vim/vim_test.go / 2026-10-18

package vim

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

func TestCanonicalSchemesMatchCommittedArtifacts(t *testing.T) {
	tests := []struct {
		palette  string
		artifact string
	}{
		{
			palette:  "quiet-day.json",
			artifact: "quiet-day.vim",
		},
		{
			palette:  "quiet-night.json",
			artifact: "quiet-night.vim",
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

			theme, err := Resolve(palette)
			if err != nil {
				t.Fatal(err)
			}

			var got bytes.Buffer

			if err := Encode(
				&got,
				theme,
			); err != nil {
				t.Fatal(err)
			}

			want, err := os.ReadFile(
				repoPath(
					t,
					"ports",
					"vim",
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
					"generated Vim scheme differs from %s",
					tt.artifact,
				)
			}
		})
	}
}

func TestTerminalPaletteMatchesGenericANSI(t *testing.T) {
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

			vimTheme, err := Resolve(palette)
			if err != nil {
				t.Fatal(err)
			}

			terminalTheme, err := terminal.Resolve(palette)
			if err != nil {
				t.Fatal(err)
			}

			for i := 0; i < 16; i++ {
				got := vimTheme.ANSI[i].MustHex()
				want := terminalTheme.ANSI[i].MustHex()

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

func TestNoTypographyPolicy(t *testing.T) {
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

	theme, err := Resolve(palette)
	if err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer

	if err := Encode(
		&output,
		theme,
	); err != nil {
		t.Fatal(err)
	}

	for _, forbidden := range []string{
		"gui=bold",
		"gui=italic",
		"gui=underline",
		"gui=undercurl",
		"cterm=bold",
		"cterm=italic",
		"cterm=underline",
	} {
		if strings.Contains(
			output.String(),
			forbidden,
		) {
			t.Errorf(
				"Vim port unexpectedly defines typography policy %q",
				forbidden,
			)
		}
	}
}

func TestSemanticStateMappings(t *testing.T) {
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

	derived, err := core.Derive(palette)
	if err != nil {
		t.Fatal(err)
	}

	theme, err := Resolve(palette)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"Visual":     derived.Selection.MustHex(),
		"CursorLine": derived.Callout.MustHex(),
		"WarningMsg": palette.Anchors.Ochre.MustHex(),
		"ErrorMsg":   palette.Anchors.Rust.MustHex(),
		"DiffAdd":    palette.Anchors.Moss.MustHex(),
		"DiffDelete": palette.Anchors.Rust.MustHex(),
	}

	for _, h := range theme.Highlights {
		expected, ok := want[h.Name]
		if !ok {
			continue
		}

		var got string

		switch h.Name {
		case "Visual", "CursorLine":
			if h.Background == nil {
				t.Errorf(
					"%s has no background",
					h.Name,
				)
				continue
			}

			got = h.Background.MustHex()

		default:
			if h.Foreground == nil {
				t.Errorf(
					"%s has no foreground",
					h.Name,
				)
				continue
			}

			got = h.Foreground.MustHex()
		}

		if got != expected {
			t.Errorf(
				"%s = %s, want %s",
				h.Name,
				got,
				expected,
			)
		}

		delete(
			want,
			h.Name,
		)
	}

	for name := range want {
		t.Errorf(
			"missing Vim group %q",
			name,
		)
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
