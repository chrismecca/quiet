// @(#)quiet / go / internal/port/vscode/vscode_test.go / 2026-10-25

package vscode

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chrismecca/quiet/internal/color"
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
			artifact: "quiet-day-color-theme.json",
		},
		{
			palette:  "quiet-night.json",
			artifact: "quiet-night-color-theme.json",
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
					"vscode",
					"themes",
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
					"generated VS Code theme differs from %q",
					tt.artifact,
				)
			}
		})
	}
}

func TestIntegratedTerminalMatchesReferencePalette(t *testing.T) {
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

	reference, err := terminal.Resolve(palette)
	if err != nil {
		t.Fatal(err)
	}

	doc := generatedDocument(
		t,
		palette,
	)

	ids := []string{
		"terminal.ansiBlack",
		"terminal.ansiRed",
		"terminal.ansiGreen",
		"terminal.ansiYellow",
		"terminal.ansiBlue",
		"terminal.ansiMagenta",
		"terminal.ansiCyan",
		"terminal.ansiWhite",
		"terminal.ansiBrightBlack",
		"terminal.ansiBrightRed",
		"terminal.ansiBrightGreen",
		"terminal.ansiBrightYellow",
		"terminal.ansiBrightBlue",
		"terminal.ansiBrightMagenta",
		"terminal.ansiBrightCyan",
		"terminal.ansiBrightWhite",
	}

	for i, id := range ids {
		got := doc.Colors[id]
		want := reference.ANSI[i].MustHex()

		if got != want {
			t.Errorf(
				"%s = %s, want %s",
				id,
				got,
				want,
			)
		}
	}
}

func TestRestrainedSemanticHierarchy(t *testing.T) {
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

	c, err := resolve(palette)
	if err != nil {
		t.Fatal(err)
	}

	doc := generatedDocument(
		t,
		palette,
	)

	want := map[string]string{
		"comment":      c.muted,
		"variable":     c.ink,
		"parameter":    c.ink,
		"property":     c.ink,
		"function":     c.ink,
		"method":       c.ink,
		"string":       c.ink,
		"number":       c.ink,
		"type":         c.mineral,
		"class":        c.mineral,
		"interface":    c.mineral,
		"struct":       c.mineral,
		"keyword":      c.mineral,
		"operator":     c.mineral,
		"*.deprecated": c.ochre,
	}

	for token, color := range want {
		if got := doc.SemanticTokenColors[token]; got != color {
			t.Errorf(
				"semantic token %s = %s, want %s",
				token,
				got,
				color,
			)
		}
	}
}

func TestSemanticHighlightingEnabled(t *testing.T) {
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

	doc := generatedDocument(
		t,
		palette,
	)

	if !doc.SemanticHighlighting {
		t.Fatal(
			"semantic highlighting is disabled",
		)
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
		`"fontStyle"`,
		`"fontFamily"`,
		`"fontWeight"`,
		`"italic"`,
		`"bold"`,
		`"underline"`,
	} {
		if strings.Contains(
			output.String(),
			forbidden,
		) {
			t.Errorf(
				"VS Code port unexpectedly defines style policy %q",
				forbidden,
			)
		}
	}
}

// underSource lists the backgrounds VS Code draws beneath syntax-colored
// source. Each must keep every syntax color at its contract minimum.
var underSource = []string{
	"editor.background",
	"editor.findMatchHighlightBackground",
	"editor.findRangeHighlightBackground",
	"editor.rangeHighlightBackground",
	"editor.selectionHighlightBackground",
	"editor.wordHighlightBackground",
	"editor.wordHighlightTextBackground",
	"editor.wordHighlightStrongBackground",
	"editor.hoverHighlightBackground",
	"editor.linkedEditingBackground",
	"editor.snippetTabstopHighlightBackground",
	"editor.stackFrameHighlightBackground",
	"editor.focusedStackFrameHighlightBackground",
	"diffEditor.insertedLineBackground",
	"diffEditor.insertedTextBackground",
	"diffEditor.removedLineBackground",
	"diffEditor.removedTextBackground",
	"diffEditor.unchangedCodeBackground",
	"merge.currentContentBackground",
	"merge.incomingContentBackground",
	"merge.commonContentBackground",
	"mergeEditor.change.background",
	"mergeEditor.changeBase.background",
	"mergeEditor.conflict.input1.background",
	"mergeEditor.conflict.input2.background",
	"mergeEditor.conflictingLines.background",
	"peekViewEditor.background",
	"peekViewEditor.matchHighlightBackground",
	"editorStickyScroll.background",
	"editorStickyScrollHover.background",
}

// underInk lists backgrounds that guarantee only Ink: Selection and the
// emphasis washes, used where a span needs more weight than a line wash.
var underInk = []string{
	"editor.selectionBackground",
	"editor.findMatchBackground",
	"editorBracketMatch.background",
	"merge.currentHeaderBackground",
	"merge.incomingHeaderBackground",
	"mergeEditor.change.word.background",
	"mergeEditor.changeBase.word.background",
}

func TestSourceIsReadableOnEditorBackgrounds(t *testing.T) {
	for _, name := range []string{"quiet-day.json", "quiet-night.json"} {
		palette, err := core.LoadPaletteFile(repoPath(t, "spec", name))
		if err != nil {
			t.Fatal(err)
		}

		doc := generatedDocument(t, palette)
		ink := palette.Anchors.Ink.MustHex()

		foregrounds := map[string]bool{}
		for _, rule := range doc.TokenColors {
			foregrounds[rule.Settings.Foreground] = true
		}
		for _, fg := range doc.SemanticTokenColors {
			foregrounds[fg] = true
		}

		for _, key := range underSource {
			bg, ok := doc.Colors[key]
			if !ok {
				t.Fatalf("%s: %s is not set", name, key)
			}

			for fg := range foregrounds {
				minimum := 4.5
				if fg == ink {
					minimum = 7
				}

				if got := contrast(t, fg, bg); got < minimum {
					t.Errorf("%s: %s on %s (%s) is %.2f, want %.1f", name, fg, key, bg, got, minimum)
				}
			}
		}

		for _, key := range underInk {
			bg, ok := doc.Colors[key]
			if !ok {
				t.Fatalf("%s: %s is not set", name, key)
			}

			if got := contrast(t, ink, bg); got < 4.5 {
				t.Errorf("%s: Ink on %s (%s) is %.2f, want 4.5", name, key, bg, got)
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

func generatedDocument(
	t *testing.T,
	palette core.Palette,
) document {
	t.Helper()

	var output bytes.Buffer

	if err := Encode(
		&output,
		palette,
	); err != nil {
		t.Fatal(err)
	}

	var doc document

	if err := json.Unmarshal(
		output.Bytes(),
		&doc,
	); err != nil {
		t.Fatalf(
			"decode generated VS Code theme: %v",
			err,
		)
	}

	return doc
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
