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
		t.Run(
			tt.palette,
			func(t *testing.T) {
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
			},
		)
	}
}

func TestThemeDeclaresNativeVariantMetadata(
	t *testing.T,
) {
	tests := []struct {
		name           string
		palette        string
		theme          string
		backgroundMode string
	}{
		{
			name:           "day",
			palette:        "quiet-day.json",
			theme:          "quiet-day",
			backgroundMode: "light",
		},
		{
			name:           "night",
			palette:        "quiet-night.json",
			theme:          "quiet-night",
			backgroundMode: "dark",
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
						tt.palette,
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
					";;;###autoload\n(deftheme " + tt.theme,
					":background-mode '" + tt.backgroundMode,
					":kind 'color-scheme",
					":family 'quiet",
				} {
					if !strings.Contains(
						got,
						want,
					) {
						t.Errorf(
							"theme missing native variant metadata %q",
							want,
						)
					}
				}
			},
		)
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

	got := output.String()

	start := strings.Index(
		got,
		"(custom-theme-set-faces",
	)
	if start < 0 {
		t.Fatal(
			"generated Emacs theme has no custom-theme-set-faces",
		)
	}

	faceDefinitions := got[start:]

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
			faceDefinitions,
			forbidden,
		) {
			t.Errorf(
				"Emacs port unexpectedly defines typography/decorative policy %q",
				forbidden,
			)
		}
	}
}

func TestThemeBackgroundModeRejectsUnknownMode(
	t *testing.T,
) {
	if _, err := themeBackgroundMode(
		core.Mode("unknown"),
	); err == nil {
		t.Fatal(
			"themeBackgroundMode accepted an unknown mode",
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
