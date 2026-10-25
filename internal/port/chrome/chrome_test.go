// @(#)quiet / go / internal/port/chrome/chrome_test.go / 2026-10-24

package chrome

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
)

var canonical = []string{
	"quiet-day",
	"quiet-night",
}

// overwritableKeys lists the theme colors Chrome reads from a manifest, from
// kOverwritableColorTable in Chromium's browser_theme_pack.cc, excluding the
// Incognito keys, which Chrome's color pipeline never reads, and
// button_background, which colors Windows caption buttons only. Chrome
// ignores unknown keys silently, so a misspelling would otherwise go
// unnoticed.
var overwritableKeys = map[string]bool{
	"background_tab":               true,
	"background_tab_inactive":      true,
	"bookmark_text":                true,
	"frame":                        true,
	"frame_inactive":               true,
	"ntp_background":               true,
	"ntp_header":                   true,
	"ntp_link":                     true,
	"ntp_text":                     true,
	"omnibox_background":           true,
	"omnibox_text":                 true,
	"tab_background_text":          true,
	"tab_background_text_inactive": true,
	"tab_text":                     true,
	"toolbar":                      true,
	"toolbar_button_icon":          true,
	"toolbar_text":                 true,
}

type themeManifest struct {
	Theme struct {
		Colors map[string][]int `json:"colors"`
	} `json:"theme"`
}

func TestCanonicalPalettesMatchCommittedArtifacts(t *testing.T) {
	for _, name := range canonical {
		t.Run(name, func(t *testing.T) {
			want, err := os.ReadFile(
				repoPath(
					t,
					"ports",
					"chrome",
					name,
					"manifest.json",
				),
			)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(encode(t, name), want) {
				t.Fatalf(
					"generated Chrome manifest differs from %s/manifest.json",
					name,
				)
			}
		})
	}
}

func TestManifestUsesOnlyOverwritableUniqueKeys(t *testing.T) {
	seen := make(map[string]bool)

	for _, e := range mapping {
		if !overwritableKeys[e.key] {
			t.Errorf("Chrome does not read theme key %q", e.key)
		}

		if seen[e.key] {
			t.Errorf("duplicate Chrome theme key %q", e.key)
		}

		seen[e.key] = true
	}
}

// Chrome's extension packer accepts any color value and Chrome drops invalid
// ones silently when it builds the theme, so the manifest is checked here.
func TestColorsAreRGBTriples(t *testing.T) {
	for _, name := range canonical {
		doc := decode(t, name)

		if len(doc.Theme.Colors) != len(mapping) {
			t.Errorf(
				"%s: %d colors, want %d",
				name,
				len(doc.Theme.Colors),
				len(mapping),
			)
		}

		for key, value := range doc.Theme.Colors {
			if len(value) != 3 {
				t.Errorf(
					"%s: %s has %d components, want 3",
					name,
					key,
					len(value),
				)

				continue
			}

			for _, component := range value {
				if component < 0 || component > 255 {
					t.Errorf(
						"%s: %s component %d out of range",
						name,
						key,
						component,
					)
				}
			}
		}
	}
}

// Every element that draws readable text must reach 4.5:1 on the background
// Chrome draws it on.
func TestTextIsReadableOnItsBackground(t *testing.T) {
	pairs := [][2]string{
		{"tab_background_text", "frame"},
		{"tab_background_text", "background_tab"},
		{"tab_background_text_inactive", "frame_inactive"},
		{"tab_background_text_inactive", "background_tab_inactive"},
		{"tab_text", "toolbar"},
		{"toolbar_text", "toolbar"},
		{"toolbar_button_icon", "toolbar"},
		{"bookmark_text", "toolbar"},
		{"omnibox_text", "omnibox_background"},
		{"ntp_text", "ntp_background"},
		{"ntp_link", "ntp_background"},
	}

	for _, name := range canonical {
		colors := decode(t, name).Theme.Colors

		for _, pair := range pairs {
			ratio := color.Contrast(
				srgb(colors[pair[0]]),
				srgb(colors[pair[1]]),
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

// Chrome's extension packer loads the extension and rejects an invalid
// manifest, so packing each theme validates it against the installed Chrome.
// The packer runs before Chrome starts a browser window, with a temporary
// profile.
func TestChromePacksTheme(t *testing.T) {
	binary := chromeBinary()
	if binary == "" {
		t.Skip("Chrome not installed")
	}

	for _, name := range canonical {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			extension := filepath.Join(dir, name)

			if err := os.Mkdir(extension, 0o755); err != nil {
				t.Fatal(err)
			}

			if err := os.WriteFile(
				filepath.Join(extension, "manifest.json"),
				encode(t, name),
				0o644,
			); err != nil {
				t.Fatal(err)
			}

			ctx, cancel := context.WithTimeout(
				context.Background(),
				2*time.Minute,
			)
			defer cancel()

			out, err := exec.CommandContext(
				ctx,
				binary,
				"--user-data-dir="+filepath.Join(dir, "profile"),
				"--no-first-run",
				"--no-message-box",
				"--pack-extension="+extension,
			).CombinedOutput()
			if err != nil {
				t.Fatalf(
					"Chrome rejected the theme: %v\n%s",
					err,
					out,
				)
			}

			if _, err := os.Stat(extension + ".crx"); err != nil {
				t.Fatalf(
					"Chrome produced no package: %v\n%s",
					err,
					out,
				)
			}
		})
	}
}

func chromeBinary() string {
	if runtime.GOOS == "darwin" {
		app := "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
		if _, err := os.Stat(app); err == nil {
			return app
		}
	}

	for _, name := range []string{
		"google-chrome",
		"google-chrome-stable",
		"chromium",
		"chromium-browser",
	} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}

	return ""
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

func decode(t *testing.T, name string) themeManifest {
	t.Helper()

	var doc themeManifest

	if err := json.Unmarshal(encode(t, name), &doc); err != nil {
		t.Fatalf("%s: manifest is not valid JSON: %v", name, err)
	}

	return doc
}

func srgb(c []int) color.SRGB {
	return color.SRGB{
		R: float64(c[0]) / 255,
		G: float64(c[1]) / 255,
		B: float64(c[2]) / 255,
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

	return filepath.Join(
		append(
			[]string{root},
			elem...,
		)...,
	)
}
