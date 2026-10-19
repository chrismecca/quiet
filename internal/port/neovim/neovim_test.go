// @(#)quiet / go / internal/port/neovim/neovim_test.go / 2026-10-19

package neovim

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
		`set(0, "DiffAdd", { fg = c.moss })`,
		`set(0, "DiffChange", { fg = c.ochre })`,
		`set(0, "DiffDelete", { fg = c.rust })`,
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
		`set(0, "@diff.plus", { link = "DiffAdd" })`,
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
		"bold =",
		"italic =",
		"underline =",
		"undercurl =",
		"strikethrough =",
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
