// @(#)quiet / go / internal/port/firefox/firefox_test.go / 2026-10-24

package firefox

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
)

var canonical = []string{
	"quiet-day",
	"quiet-night",
}

// documentedKeys lists the supported theme colors documented on MDN's
// manifest.json theme page, excluding keys unsupported since Firefox 89 and
// the bookmark_text alias of toolbar_text. Firefox ignores unknown keys
// silently, so a misspelling would otherwise go unnoticed.
var documentedKeys = map[string]bool{
	"button_background_active":     true,
	"button_background_hover":      true,
	"frame":                        true,
	"frame_inactive":               true,
	"icons":                        true,
	"icons_attention":              true,
	"ntp_background":               true,
	"ntp_card_background":          true,
	"ntp_text":                     true,
	"popup":                        true,
	"popup_border":                 true,
	"popup_highlight":              true,
	"popup_highlight_text":         true,
	"popup_text":                   true,
	"sidebar":                      true,
	"sidebar_border":               true,
	"sidebar_highlight":            true,
	"sidebar_highlight_text":       true,
	"sidebar_text":                 true,
	"tab_background_text":          true,
	"tab_line":                     true,
	"tab_loading":                  true,
	"tab_selected":                 true,
	"tab_text":                     true,
	"toolbar":                      true,
	"toolbar_bottom_separator":     true,
	"toolbar_field":                true,
	"toolbar_field_border":         true,
	"toolbar_field_border_focus":   true,
	"toolbar_field_focus":          true,
	"toolbar_field_highlight":      true,
	"toolbar_field_highlight_text": true,
	"toolbar_field_text":           true,
	"toolbar_field_text_focus":     true,
	"toolbar_text":                 true,
	"toolbar_top_separator":        true,
	"toolbar_vertical_separator":   true,
}

func TestCanonicalPalettesMatchCommittedArtifacts(t *testing.T) {
	for _, name := range canonical {
		t.Run(name, func(t *testing.T) {
			want, err := os.ReadFile(
				repoPath(
					t,
					"ports",
					"firefox",
					name,
					"manifest.json",
				),
			)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(encode(t, name), want) {
				t.Fatalf(
					"generated Firefox manifest differs from %s/manifest.json",
					name,
				)
			}
		})
	}
}

func TestManifestUsesOnlyDocumentedUniqueKeys(t *testing.T) {
	seen := make(map[string]bool)

	for _, e := range mapping {
		if !documentedKeys[e.key] {
			t.Errorf("undocumented Firefox theme key %q", e.key)
		}

		if seen[e.key] {
			t.Errorf("duplicate Firefox theme key %q", e.key)
		}

		seen[e.key] = true
	}
}

// Every element that draws readable text must reach 4.5:1 on the background
// Firefox draws it on.
// Web content keeps following the system appearance; only Firefox's own
// interface follows the adaptation.
func TestWebContentFollowsSystem(t *testing.T) {
	for _, name := range canonical {
		var doc struct {
			Theme struct {
				Properties map[string]string `json:"properties"`
			} `json:"theme"`
		}

		if err := json.Unmarshal(encode(t, name), &doc); err != nil {
			t.Fatal(err)
		}

		if got := doc.Theme.Properties["content_color_scheme"]; got != "system" {
			t.Errorf(
				"%s: content_color_scheme = %q, want %q",
				name,
				got,
				"system",
			)
		}
	}
}

func TestTextIsReadableOnItsBackground(t *testing.T) {
	pairs := [][2]string{
		{"tab_background_text", "frame"},
		{"tab_background_text", "frame_inactive"},
		{"tab_text", "tab_selected"},
		{"toolbar_text", "toolbar"},
		{"icons", "toolbar"},
		{"icons_attention", "toolbar"},
		{"toolbar_field_text", "toolbar_field"},
		{"toolbar_field_text_focus", "toolbar_field_focus"},
		{"toolbar_field_highlight_text", "toolbar_field_highlight"},
		{"popup_text", "popup"},
		{"popup_highlight_text", "popup_highlight"},
		{"sidebar_text", "sidebar"},
		{"sidebar_highlight_text", "sidebar_highlight"},
		{"ntp_text", "ntp_background"},
		{"ntp_text", "ntp_card_background"},
	}

	for _, name := range canonical {
		var doc struct {
			Theme struct {
				Colors map[string]string `json:"colors"`
			} `json:"theme"`
		}

		if err := json.Unmarshal(encode(t, name), &doc); err != nil {
			t.Fatal(err)
		}

		for _, pair := range pairs {
			ratio := color.Contrast(
				parse(t, doc.Theme.Colors[pair[0]]),
				parse(t, doc.Theme.Colors[pair[1]]),
			)

			if ratio < 4.5 {
				t.Errorf(
					"%s: %s on %s = %.2f, want at least 4.5",
					name,
					pair[0],
					pair[1],
					ratio,
				)
			}
		}
	}
}

func encode(t *testing.T, name string) []byte {
	t.Helper()

	palette, err := core.LoadPaletteFile(
		repoPath(
			t,
			"spec",
			name+".json",
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	var b bytes.Buffer

	if err := Encode(&b, palette); err != nil {
		t.Fatal(err)
	}

	return b.Bytes()
}

func parse(t *testing.T, hex string) color.SRGB {
	t.Helper()

	c, err := color.ParseHex(hex)
	if err != nil {
		t.Fatal(err)
	}

	return c
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

	return filepath.Join(
		append(
			[]string{root},
			elem...,
		)...,
	)
}
