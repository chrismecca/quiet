// @(#)quiet / go / internal/port/xcode/theme_test.go / 2026-10-18

package xcode

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
			artifact: "Quiet Day.xcworkspacecolortheme",
		},
		{
			palette:  "quiet-night.json",
			artifact: "Quiet Night.xcworkspacecolortheme",
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
				t.Fatalf(
					"LoadPaletteFile: %v",
					err,
				)
			}

			theme, err := Resolve(palette)
			if err != nil {
				t.Fatalf(
					"Resolve: %v",
					err,
				)
			}

			var got bytes.Buffer

			if err := EncodeJSON(
				&got,
				theme,
			); err != nil {
				t.Fatalf(
					"EncodeJSON: %v",
					err,
				)
			}

			want, err := os.ReadFile(
				repoPath(
					t,
					"ports",
					"xcode",
					tt.artifact,
				),
			)
			if err != nil {
				t.Fatalf(
					"ReadFile: %v",
					err,
				)
			}

			if !bytes.Equal(
				got.Bytes(),
				want,
			) {
				t.Fatalf(
					"generated Xcode theme differs from %s",
					tt.artifact,
				)
			}
		})
	}
}

func TestCanonicalThemeHasCompleteRoleSet(t *testing.T) {
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

	if got, want := len(theme.Overrides), 38; got != want {
		t.Fatalf(
			"Xcode overrides = %d, want %d",
			got,
			want,
		)
	}

	seen := make(
		map[string]bool,
		len(theme.Overrides),
	)

	for _, override := range theme.Overrides {
		if seen[override.Role] {
			t.Errorf(
				"duplicate Xcode role %q",
				override.Role,
			)
		}

		seen[override.Role] = true
	}
}

func TestSyntaxTaxonomyUsesQuietHierarchy(t *testing.T) {
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

	allowed := map[string]bool{
		palette.Anchors.Ink.MustHex():     true,
		palette.Anchors.Muted.MustHex():   true,
		palette.Anchors.Mineral.MustHex(): true,
	}

	stateRoles := map[string]bool{
		"selectedTextBackgroundColor": true,
		"insertionPointColor":         true,
		"currentLineHighlight":        true,
		"invisibles":                  true,
		"diffAddition":                true,
		"diffDeletion":                true,
		"debugger":                    true,
	}

	for _, override := range theme.Overrides {
		if stateRoles[override.Role] {
			continue
		}

		got := override.Color.MustHex()

		if !allowed[got] {
			t.Errorf(
				"syntax role %s uses %s outside Ink/Muted/Mineral hierarchy",
				override.Role,
				got,
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
		"selectedTextBackgroundColor": derived.Selection.MustHex(),
		"insertionPointColor":         palette.Anchors.Mineral.MustHex(),
		"currentLineHighlight":        derived.Callout.MustHex(),
		"invisibles":                  palette.Anchors.Line.MustHex(),
		"diffAddition":                palette.Anchors.Moss.MustHex(),
		"diffDeletion":                palette.Anchors.Rust.MustHex(),
		"debugger":                    palette.Anchors.Mineral.MustHex(),
	}

	for _, override := range theme.Overrides {
		expected, ok := want[override.Role]
		if !ok {
			continue
		}

		got := override.Color.MustHex()

		if got != expected {
			t.Errorf(
				"%s = %s, want %s",
				override.Role,
				got,
				expected,
			)
		}

		delete(
			want,
			override.Role,
		)
	}

	for role := range want {
		t.Errorf(
			"missing Xcode semantic role %q",
			role,
		)
	}
}

func TestNativeThemesDeclareSRGB(t *testing.T) {
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

	if err := EncodeJSON(
		&output,
		theme,
	); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(
		output.String(),
		`"gamut": "P3"`,
	) {
		t.Fatal(
			"Xcode port must not promote canonical sRGB colors to P3",
		)
	}

	if !strings.Contains(
		output.String(),
		`"gamut": "sRGB"`,
	) {
		t.Fatal(
			"Xcode port does not declare sRGB",
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
