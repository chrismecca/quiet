// @(#)quiet / go / internal/port/neovim/neovim_test.go / 2026-10-25

package neovim

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
			artifact: "quiet-day.lua",
		},
		{
			palette:  "quiet-night.json",
			artifact: "quiet-night.lua",
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
					"neovim",
					"colors",
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
					"generated Neovim theme differs from %q",
					tt.artifact,
				)
			}
		})
	}
}

func TestRestrainedSyntaxHierarchy(t *testing.T) {
	got := generateDay(t)

	for _, want := range []string{
		`set(0, "Comment", { fg = c.muted })`,
		`set(0, "String", { fg = c.ink })`,
		`set(0, "Number", { fg = c.ink })`,
		`set(0, "Function", { fg = c.ink })`,
		`set(0, "Identifier", { fg = c.ink })`,
		`set(0, "Keyword", { fg = c.mineral })`,
		`set(0, "Type", { fg = c.mineral })`,
		`set(0, "PreProc", { fg = c.mineral })`,
	} {
		if !strings.Contains(
			got,
			want,
		) {
			t.Errorf(
				"missing syntax mapping %q",
				want,
			)
		}
	}
}

func TestSemanticStateColors(t *testing.T) {
	got := generateDay(t)

	for _, want := range []string{
		`set(0, "DiagnosticError", { fg = c.rust })`,
		`set(0, "DiagnosticWarn", { fg = c.ochre })`,
		`set(0, "DiagnosticInfo", { fg = c.mineral })`,
		`set(0, "DiagnosticOk", { fg = c.moss })`,
		`set(0, "DiffAdd", { bg = c.moss_wash })`,
		`set(0, "DiffChange", { bg = c.ochre_wash })`,
		`set(0, "DiffDelete", { fg = c.rust, bg = c.rust_wash })`,
		`set(0, "DiffText", { fg = c.ink, bg = c.ochre_wash_emphasis })`,
		`set(0, "Added", { fg = c.moss })`,
		`set(0, "Changed", { fg = c.ochre })`,
		`set(0, "Removed", { fg = c.rust })`,
	} {
		if !strings.Contains(
			got,
			want,
		) {
			t.Errorf(
				"missing semantic mapping %q",
				want,
			)
		}
	}
}

func TestTreeSitterAndLSPMappings(t *testing.T) {
	got := generateDay(t)

	for _, want := range []string{
		`set(0, "@comment", { link = "Comment" })`,
		`set(0, "@string", { link = "String" })`,
		`set(0, "@variable", { link = "Identifier" })`,
		`set(0, "@keyword", { link = "Keyword" })`,
		`set(0, "@type", { link = "Type" })`,
		`set(0, "@diff.plus", { link = "Added" })`,
		`set(0, "@lsp.type.variable", { link = "Identifier" })`,
		`set(0, "@lsp.type.function", { link = "Function" })`,
		`set(0, "@lsp.type.class", { link = "Type" })`,
		`set(0, "@lsp.mod.deprecated", { link = "WarningMsg" })`,
	} {
		if !strings.Contains(
			got,
			want,
		) {
			t.Errorf(
				"missing native Neovim mapping %q",
				want,
			)
		}
	}
}

func TestThemeContainsNoTypographyPolicy(t *testing.T) {
	got := generateDay(t)

	for _, forbidden := range []string{
		"italic =",
		"reverse =",
		"font",
	} {
		if strings.Contains(
			got,
			forbidden,
		) {
			t.Errorf(
				"Neovim port unexpectedly defines style policy %q",
				forbidden,
			)
		}
	}
}

// TestTextIsReadable checks every highlight that draws a background:
// backgrounds without a foreground keep every syntax color at its contract
// minimum, and explicit pairs reach 4.5:1.
func TestTextIsReadable(t *testing.T) {
	exempt := map[string]bool{
		// Known limits shared with quiet's other editor ports: Callout
		// guarantees only Muted, and Selection only Ink.
		"CursorLine":    true,
		"CursorColumn":  true,
		"CursorLineNr":  true,
		"PmenuMatchSel": true,

		// Drawn without text: a border and a scrollbar thumb.
		"FloatBorder": true,
		"PmenuThumb":  true,
	}

	for _, name := range []string{"quiet-day.json", "quiet-night.json"} {
		palette, err := core.LoadPaletteFile(repoPath(t, "spec", name))
		if err != nil {
			t.Fatal(err)
		}

		c, err := resolveColors(palette)
		if err != nil {
			t.Fatal(err)
		}

		byName := map[string]string{
			"field":               c.field,
			"surface":             c.surface,
			"structure":           c.structure,
			"ink":                 c.ink,
			"muted":               c.muted,
			"line":                c.line,
			"mineral":             c.mineral,
			"moss":                c.moss,
			"ochre":               c.ochre,
			"rust":                c.rust,
			"selection":           c.selection,
			"callout":             c.callout,
			"mineral_wash":        c.mineralWash,
			"moss_wash":           c.mossWash,
			"ochre_wash":          c.ochreWash,
			"rust_wash":           c.rustWash,
			"ochre_wash_emphasis": c.ochreWashEmphasis,
		}

		syntax := map[string]float64{
			"ink":     7,
			"muted":   4.5,
			"mineral": 4.5,
			"moss":    4.5,
			"ochre":   4.5,
			"rust":    4.5,
		}

		var all []highlight
		all = append(all, baseHighlights()...)
		all = append(all, diagnosticHighlights()...)

		for _, h := range all {
			if h.bg == "" || exempt[h.group] {
				continue
			}

			bg := byName[h.bg]

			if h.fg != "" {
				if h.fg == "field" && h.group == "EndOfBuffer" {
					continue
				}

				if got := contrast(t, byName[h.fg], bg); got < 4.5 {
					t.Errorf("%s: %s %s on %s is %.2f, want 4.5", name, h.group, h.fg, h.bg, got)
				}

				continue
			}

			for fg, minimum := range syntax {
				if got := contrast(t, byName[fg], bg); got < minimum {
					t.Errorf("%s: %s on %s (%s) is %.2f, want %.1f", name, fg, h.group, h.bg, got, minimum)
				}
			}
		}
	}
}

func contrast(t *testing.T, a, b string) float64 {
	t.Helper()

	ca, err := color.ParseHex(a)
	if err != nil {
		t.Fatal(err)
	}

	cb, err := color.ParseHex(b)
	if err != nil {
		t.Fatal(err)
	}

	return color.Contrast(ca, cb)
}

func generateDay(
	t *testing.T,
) string {
	t.Helper()

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

	return output.String()
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
