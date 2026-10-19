// @(#)quiet / go / internal/port/emacs/emacs_test.go / 2026-10-19

package emacs

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chrismecca/quiet/internal/core"
)

func TestCanonicalThemesMatchCommittedArtifacts(t *testing.T) {
	tests := []struct {
		palette  string
		artifact string
	}{
		{
			palette:  "quiet-day.json",
			artifact: "quiet-day-theme.el",
		},
		{
			palette:  "quiet-night.json",
			artifact: "quiet-night-theme.el",
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

			var got bytes.Buffer

			if err := Encode(
				&got,
				palette,
			); err != nil {
				t.Fatal(err)
			}

			want, err := os.ReadFile(
				repoPath(
					t,
					"ports",
					"emacs",
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
					"generated Emacs theme differs from %q",
					tt.artifact,
				)
			}
		})
	}
}

func TestThemeReservesStatusColorsForStatus(t *testing.T) {
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

	var output bytes.Buffer

	if err := Encode(
		&output,
		palette,
	); err != nil {
		t.Fatal(err)
	}

	got := output.String()

	for _, want := range []string{
		"`(success ((t (:foreground ,moss))))",
		"`(warning ((t (:foreground ,ochre))))",
		"`(error ((t (:foreground ,rust))))",
		"`(font-lock-warning-face ((t (:foreground ,ochre))))",
	} {
		if !strings.Contains(
			got,
			want,
		) {
			t.Errorf(
				"theme missing semantic mapping %q",
				want,
			)
		}
	}
}

func TestThemeUsesRestrainedSyntaxHierarchy(t *testing.T) {
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

	var output bytes.Buffer

	if err := Encode(
		&output,
		palette,
	); err != nil {
		t.Fatal(err)
	}

	got := output.String()

	for _, want := range []string{
		"`(font-lock-comment-face ((t (:foreground ,muted))))",
		"`(font-lock-keyword-face ((t (:foreground ,mineral))))",
		"`(font-lock-type-face ((t (:foreground ,mineral))))",
		"`(font-lock-builtin-face ((t (:foreground ,mineral))))",
		"`(font-lock-string-face ((t (:foreground ,ink))))",
		"`(font-lock-function-name-face ((t (:foreground ,ink))))",
		"`(font-lock-variable-name-face ((t (:foreground ,ink))))",
		"`(font-lock-constant-face ((t (:foreground ,ink))))",
	} {
		if !strings.Contains(
			got,
			want,
		) {
			t.Errorf(
				"theme missing syntax mapping %q",
				want,
			)
		}
	}
}

func TestThemeContainsNoTypographyPolicy(t *testing.T) {
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

	var output bytes.Buffer

	if err := Encode(
		&output,
		palette,
	); err != nil {
		t.Fatal(err)
	}

	for _, forbidden := range []string{
		":family",
		":height",
		":weight",
		":slant",
		":underline",
		":overline",
		":strike-through",
		":box",
	} {
		if strings.Contains(
			output.String(),
			forbidden,
		) {
			t.Errorf(
				"Emacs port unexpectedly defines typography/decorative policy %q",
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
