// @(#)quiet / go / internal/port/zed/zed_test.go / 2026-10-25

package zed

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
)

func TestCommittedFamilyMatchesGenerator(t *testing.T) {
	got := encodeFamily(t)

	want, err := os.ReadFile(repoPath("ports", "zed", "quiet.json"))
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, want) {
		t.Error("ports/zed/quiet.json is stale; run make zed")
	}
}

type family struct {
	Schema string  `json:"$schema"`
	Name   string  `json:"name"`
	Themes []theme `json:"themes"`
}

type theme struct {
	Name       string         `json:"name"`
	Appearance string         `json:"appearance"`
	Style      map[string]any `json:"style"`
}

func TestFamilyIsValidJSON(t *testing.T) {
	f := decodeFamily(t)

	if len(f.Themes) != 2 {
		t.Fatalf("family has %d themes, want 2", len(f.Themes))
	}

	want := map[string]string{"quiet-day": "light", "quiet-night": "dark"}

	for _, th := range f.Themes {
		if want[th.Name] != th.Appearance {
			t.Errorf("%s has appearance %q, want %q", th.Name, th.Appearance, want[th.Name])
		}
	}
}

// Zed 1.25.3's schema and bundled themes are recorded in testdata; every
// key and syntax name must be one Zed knows.
func TestKeysExistInZed(t *testing.T) {
	keys := readList(t, "style-keys.txt")
	names := readList(t, "syntax-names.txt")

	for _, th := range decodeFamily(t).Themes {
		for key := range th.Style {
			switch key {
			case "accents", "players", "syntax":
				continue
			}

			if !keys[key] {
				t.Errorf("%s: %s is not a Zed style key", th.Name, key)
			}
		}

		syntax, _ := th.Style["syntax"].(map[string]any)
		for name := range syntax {
			if !names[name] {
				t.Errorf("%s: %s is not a Zed syntax name", th.Name, name)
			}
		}
	}
}

func TestEveryStyleKeyIsSet(t *testing.T) {
	// Keys Zed accepts that the theme leaves unset on purpose.
	unset := map[string]bool{
		"experimental.theme_overrides": true,
		"unstable.ui_density":          true,
		"scrollbar_thumb.background":   true, // Deprecated alias.
	}

	for _, th := range decodeFamily(t).Themes {
		for key := range readList(t, "style-keys.txt") {
			if _, ok := th.Style[key]; !ok && !unset[key] {
				t.Errorf("%s does not set %s", th.Name, key)
			}
		}
	}
}

// TestTextIsReadable checks the backgrounds Zed draws beneath source, and
// interface text on the surfaces it is drawn on.
func TestTextIsReadable(t *testing.T) {
	underSource := []string{
		"editor.background",
		"editor.highlighted_line.background",
		"editor.debugger_active_line.background",
		"editor.document_highlight.read_background",
		"editor.document_highlight.write_background",
		"editor.diff_hunk.added.background",
		"editor.diff_hunk.deleted.background",
		"search.match_background",
		"search.active_match_background",
		"version_control.conflict_marker.ours",
		"version_control.conflict_marker.theirs",
		"vim.yank.background",
	}

	underInk := []string{
		"editor.document_highlight.bracket_background",
		"version_control.word_added",
		"version_control.word_deleted",
		"element.selected",
	}

	surfaces := []string{
		"background",
		"surface.background",
		"elevated_surface.background",
		"panel.background",
		"status_bar.background",
		"title_bar.background",
		"tab_bar.background",
		"tab.active_background",
		"editor.background",
		"element.background",
	}

	for _, th := range decodeFamily(t).Themes {
		get := func(key string) color.SRGB {
			t.Helper()

			s, ok := th.Style[key].(string)
			if !ok {
				t.Fatalf("%s: %s is not set", th.Name, key)
			}

			return color.MustParseHex(s)
		}

		syntax, _ := th.Style["syntax"].(map[string]any)

		var foregrounds []color.SRGB
		seen := map[string]bool{}
		for _, v := range syntax {
			hex := v.(map[string]any)["color"].(string)
			if !seen[hex] {
				seen[hex] = true
				foregrounds = append(foregrounds, color.MustParseHex(hex))
			}
		}

		ink := get("editor.foreground")

		for _, key := range underSource {
			bg := get(key)

			for _, fg := range foregrounds {
				minimum := 4.5
				if fg == ink {
					minimum = 7
				}

				if got := color.Contrast(fg, bg); got < minimum {
					t.Errorf("%s: %s on %s is %.2f, want %.1f", th.Name, fg.MustHex(), key, got, minimum)
				}
			}
		}

		for _, key := range underInk {
			if got := color.Contrast(ink, get(key)); got < 4.5 {
				t.Errorf("%s: Ink on %s is %.2f, want 4.5", th.Name, key, got)
			}
		}

		for _, key := range surfaces {
			for _, text := range []string{"text", "text.muted", "text.accent"} {
				if got := color.Contrast(get(text), get(key)); got < 4.5 {
					t.Errorf("%s: %s on %s is %.2f, want 4.5", th.Name, text, key, got)
				}
			}
		}
	}
}

func TestDimColorsAreDimmer(t *testing.T) {
	for _, th := range decodeFamily(t).Themes {
		bg := color.MustParseHex(th.Style["terminal.background"].(string))

		for _, name := range ansiNames {
			normal := color.MustParseHex(th.Style["terminal.ansi."+name].(string))
			dim := color.MustParseHex(th.Style["terminal.ansi.dim_"+name].(string))

			if color.Contrast(dim, bg) > color.Contrast(normal, bg) {
				t.Errorf("%s: dim %s is not dimmer than %s", th.Name, name, name)
			}
		}
	}
}

func encodeFamily(t *testing.T) []byte {
	t.Helper()

	var b bytes.Buffer
	if err := Encode(&b, loadPalette(t, "quiet-day"), loadPalette(t, "quiet-night")); err != nil {
		t.Fatal(err)
	}

	return b.Bytes()
}

func decodeFamily(t *testing.T) family {
	t.Helper()

	var f family
	if err := json.Unmarshal(encodeFamily(t), &f); err != nil {
		t.Fatalf("generated family is not valid JSON: %v", err)
	}

	return f
}

func readList(t *testing.T, name string) map[string]bool {
	t.Helper()

	file, err := os.Open(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	result := map[string]bool{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			result[line] = true
		}
	}

	return result
}

func loadPalette(t *testing.T, name string) core.Palette {
	t.Helper()

	p, err := core.LoadPaletteFile(repoPath("spec", name+".json"))
	if err != nil {
		t.Fatal(err)
	}

	return p
}

func repoPath(parts ...string) string {
	return filepath.Join(append([]string{"..", "..", ".."}, parts...)...)
}
