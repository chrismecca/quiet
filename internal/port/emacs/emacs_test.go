// @(#)quiet / go / internal/port/emacs/emacs_test.go / 2026-10-25

package emacs

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chrismecca/quiet/internal/color"
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

// TestTextIsReadable checks every face that draws a background: faces
// without a foreground keep every syntax color at its contract minimum, and
// explicit pairs reach 4.5:1.
func TestTextIsReadable(t *testing.T) {
	exempt := map[string]bool{
		// Callout guarantees only Muted, as in every quiet editor port.
		"highlight": true,

		// Emacs draws the character under the cursor in the default
		// background.
		"cursor": true,
	}

	syntax := map[string]float64{
		"ink":     7,
		"muted":   4.5,
		"mineral": 4.5,
		"moss":    4.5,
		"ochre":   4.5,
		"rust":    4.5,
	}

	for _, name := range []string{"quiet-day.json", "quiet-night.json"} {
		palette, err := core.LoadPaletteFile(repoPath(t, "spec", name))
		if err != nil {
			t.Fatal(err)
		}

		roles, err := resolveRoles(palette)
		if err != nil {
			t.Fatal(err)
		}

		byName := map[string]color.SRGB{}
		for _, r := range roles {
			byName[r.name] = r.value
		}

		for _, f := range faces() {
			if f.background == "" || exempt[f.name] || strings.HasPrefix(f.name, "ansi-color-") {
				continue
			}

			bg, ok := byName[f.background]
			if !ok {
				t.Fatalf("%s: unknown role %q", f.name, f.background)
			}

			if f.foreground != "" {
				if got := color.Contrast(byName[f.foreground], bg); got < 4.5 {
					t.Errorf("%s: %s %s on %s is %.2f, want 4.5", name, f.name, f.foreground, f.background, got)
				}

				continue
			}

			for fg, minimum := range syntax {
				if got := color.Contrast(byName[fg], bg); got < minimum {
					t.Errorf("%s: %s on %s (%s) is %.2f, want %.1f", name, fg, f.name, f.background, got, minimum)
				}
			}
		}
	}
}

func TestFacesUseKnownRoles(t *testing.T) {
	palette, err := core.LoadPaletteFile(repoPath(t, "spec", "quiet-day.json"))
	if err != nil {
		t.Fatal(err)
	}

	roles, err := resolveRoles(palette)
	if err != nil {
		t.Fatal(err)
	}

	known := map[string]bool{}
	for _, r := range roles {
		known[r.name] = true
	}

	seen := map[string]bool{}

	for _, f := range faces() {
		if seen[f.name] {
			t.Errorf("%s is set twice", f.name)
		}

		seen[f.name] = true

		for _, r := range []string{f.foreground, f.background, f.underlineColor} {
			if r != "" && !known[r] {
				t.Errorf("%s uses unknown role %q", f.name, r)
			}
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
