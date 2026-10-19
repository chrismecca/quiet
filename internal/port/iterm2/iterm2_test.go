// @(#)quiet / go / internal/port/iterm2/iterm2_test.go / 2026-10-19

package iterm2

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chrismecca/quiet/internal/core"
	"github.com/chrismecca/quiet/internal/port/terminal"
)

func TestCanonicalPresetMatchesCommittedArtifact(t *testing.T) {
	day, night := canonicalPalettes(t)

	var got bytes.Buffer

	if err := Encode(
		&got,
		day,
		night,
	); err != nil {
		t.Fatal(err)
	}

	want, err := os.ReadFile(
		repoPath(
			t,
			"ports",
			"iterm2",
			"Quiet.itermcolors",
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(
		got.Bytes(),
		want,
	) {
		t.Fatal(
			"generated iTerm2 preset differs from committed artifact",
		)
	}
}

func TestPresetIsWellFormedXML(t *testing.T) {
	day, night := canonicalPalettes(t)

	var output bytes.Buffer

	if err := Encode(
		&output,
		day,
		night,
	); err != nil {
		t.Fatal(err)
	}

	decoder := xml.NewDecoder(
		bytes.NewReader(output.Bytes()),
	)

	for {
		if _, err := decoder.Token(); err != nil {
			if err.Error() == "EOF" {
				break
			}

			t.Fatalf(
				"parse generated XML: %v",
				err,
			)
		}
	}
}

func TestPresetContainsAdaptiveRoles(t *testing.T) {
	day, night := canonicalPalettes(t)

	var output bytes.Buffer

	if err := Encode(
		&output,
		day,
		night,
	); err != nil {
		t.Fatal(err)
	}

	got := output.String()

	for _, role := range []string{
		"Background Color",
		"Foreground Color",
		"Bold Color",
		"Cursor Color",
		"Cursor Text Color",
		"Selection Color",
		"Selected Text Color",
		"Link Color",
		"Active Pane Border Color",
		"IME Cursor Color",
		"Cursor Guide Color",
		"Badge Color",
		"Match Background Color",
	} {
		for _, suffix := range []string{
			"",
			" (Dark)",
			" (Light)",
		} {
			want := "<key>" +
				role +
				suffix +
				"</key>"

			if !strings.Contains(
				got,
				want,
			) {
				t.Errorf(
					"missing %q",
					want,
				)
			}
		}
	}

	for i := 0; i < 16; i++ {
		for _, suffix := range []string{
			"",
			" (Dark)",
			" (Light)",
		} {
			want := fmt.Sprintf(
				"<key>Ansi %d Color%s</key>",
				i,
				suffix,
			)

			if !strings.Contains(
				got,
				want,
			) {
				t.Errorf(
					"missing %q",
					want,
				)
			}
		}
	}
}

func TestPresetUsesOnlySRGB(t *testing.T) {
	day, night := canonicalPalettes(t)

	var output bytes.Buffer

	if err := Encode(
		&output,
		day,
		night,
	); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(
		output.String(),
		"<string>P3</string>",
	) {
		t.Fatal(
			"iTerm2 preset unexpectedly contains P3 colors",
		)
	}
}

func canonicalPalettes(
	t *testing.T,
) (terminal.Palette, terminal.Palette) {
	t.Helper()

	dayCore, err := core.LoadPaletteFile(
		repoPath(
			t,
			"spec",
			"quiet-day.json",
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	nightCore, err := core.LoadPaletteFile(
		repoPath(
			t,
			"spec",
			"quiet-night.json",
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	day, err := terminal.Resolve(dayCore)
	if err != nil {
		t.Fatal(err)
	}

	night, err := terminal.Resolve(nightCore)
	if err != nil {
		t.Fatal(err)
	}

	return day, night
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
