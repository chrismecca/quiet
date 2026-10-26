// @(#)quiet / go / internal/port/iterm2/iterm2_test.go / 2026-10-26

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
			"quiet.itermcolors",
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

func TestVariantPresetsMatchCommittedArtifacts(t *testing.T) {
	day, night := canonicalPalettes(t)

	for name, p := range map[string]terminal.Palette{
		"quiet-day":   day,
		"quiet-night": night,
	} {
		var got bytes.Buffer
		if err := EncodeVariant(&got, p); err != nil {
			t.Fatal(err)
		}

		want, err := os.ReadFile(repoPath(t, "ports", "iterm2", name+".itermcolors"))
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(got.Bytes(), want) {
			t.Errorf("ports/iterm2/%s.itermcolors is stale; run make iterm2", name)
		}
	}
}

// A variant preset holds each key once, without (Light) or (Dark), with
// the values the adaptive preset gives that variant.
func TestVariantPresetsMatchAdaptivePreset(t *testing.T) {
	day, night := canonicalPalettes(t)

	var adaptive bytes.Buffer
	if err := Encode(&adaptive, day, night); err != nil {
		t.Fatal(err)
	}

	for suffix, p := range map[string]terminal.Palette{
		" (Light)": day,
		" (Dark)":  night,
	} {
		var variant bytes.Buffer
		if err := EncodeVariant(&variant, p); err != nil {
			t.Fatal(err)
		}

		if strings.Contains(variant.String(), "(Light)") || strings.Contains(variant.String(), "(Dark)") {
			t.Errorf("%s variant preset has suffixed keys", strings.TrimSpace(suffix))
		}

		want := strings.ReplaceAll(
			keyedSection(adaptive.String(), suffix),
			suffix+"</key>",
			"</key>",
		)

		if got := keyedSection(variant.String(), ""); got != want {
			t.Errorf("%s variant preset differs from the adaptive preset", strings.TrimSpace(suffix))
		}
	}
}

// keyedSection returns the key/value blocks whose keys end in suffix.
func keyedSection(preset, suffix string) string {
	var b strings.Builder

	// Top-level keys sit at one tab; the components inside each color sit
	// at two.
	blocks := strings.Split(preset, "\n\t<key>")
	for _, block := range blocks[1:] {
		key, _, _ := strings.Cut(block, "</key>")
		hasSuffix := strings.HasSuffix(key, " (Light)") || strings.HasSuffix(key, " (Dark)")

		if (suffix == "" && !hasSuffix) || (suffix != "" && strings.HasSuffix(key, suffix)) {
			entry, _, _ := strings.Cut(block, "\n\t</dict>")
			b.WriteString("\t<key>" + entry + "\n\t</dict>\n")
		}
	}

	return b.String()
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
