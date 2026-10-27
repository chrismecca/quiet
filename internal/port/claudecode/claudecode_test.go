// @(#)quiet / go / internal/port/claudecode/claudecode_test.go / 2026-10-27

package claudecode

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
)

var canonical = []struct {
	palette  string
	artifact string
	base     string
}{
	{"quiet-day.json", "quiet-day.json", "light"},
	{"quiet-night.json", "quiet-night.json", "dark"},
}

// theme is the custom theme format Claude Code reads.
type theme struct {
	Name      string            `json:"name"`
	Base      string            `json:"base"`
	Overrides map[string]string `json:"overrides"`
}

// identity lists the palette keys the theme leaves to the base theme on
// purpose: Claude Code's own identity colors, the categorical colors users
// assign to subagents, and decorative effects. None carries state.
var identity = map[string]bool{
	"claude":                    true,
	"claudeShimmer":             true,
	"clawd_body":                true,
	"clawd_background":          true,
	"briefLabelClaude":          true,
	"chromeYellow":              true,
	"red_FOR_SUBAGENTS_ONLY":    true,
	"blue_FOR_SUBAGENTS_ONLY":   true,
	"green_FOR_SUBAGENTS_ONLY":  true,
	"yellow_FOR_SUBAGENTS_ONLY": true,
	"purple_FOR_SUBAGENTS_ONLY": true,
	"orange_FOR_SUBAGENTS_ONLY": true,
	"pink_FOR_SUBAGENTS_ONLY":   true,
	"cyan_FOR_SUBAGENTS_ONLY":   true,
	"rainbow_red":               true,
	"rainbow_orange":            true,
	"rainbow_yellow":            true,
	"rainbow_green":             true,
	"rainbow_blue":              true,
	"rainbow_indigo":            true,
	"rainbow_violet":            true,
	"rainbow_red_shimmer":       true,
	"rainbow_orange_shimmer":    true,
	"rainbow_yellow_shimmer":    true,
	"rainbow_green_shimmer":     true,
	"rainbow_blue_shimmer":      true,
	"rainbow_indigo_shimmer":    true,
	"rainbow_violet_shimmer":    true,
}

func TestCanonicalPalettesMatchCommittedArtifacts(t *testing.T) {
	for _, tt := range canonical {
		t.Run(tt.palette, func(t *testing.T) {
			got := encode(t, tt.palette)

			want, err := os.ReadFile(
				repoPath(
					t,
					"ports",
					"claude-code",
					tt.artifact,
				),
			)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(got, want) {
				t.Fatalf(
					"generated Claude Code theme differs from %q",
					tt.artifact,
				)
			}
		})
	}
}

// Claude Code lists a theme under its file name and accepts these color
// forms; a value it does not accept is dropped silently.
func TestThemeIsWellFormed(t *testing.T) {
	hex := regexp.MustCompile(`^#[0-9A-F]{6}$`)

	for _, tt := range canonical {
		th := decode(t, tt.palette)

		if want := strings.TrimSuffix(tt.artifact, ".json"); th.Name != want {
			t.Errorf("%s: name = %q, want %q", tt.palette, th.Name, want)
		}

		if th.Base != tt.base {
			t.Errorf("%s: base = %q, want %q", tt.palette, th.Base, tt.base)
		}

		for key, value := range th.Overrides {
			if !hex.MatchString(value) {
				t.Errorf("%s: %s = %q is not #RRGGBB", tt.palette, key, value)
			}
		}
	}
}

// Claude Code ignores override keys its palette does not define.
func TestKeysExistInClaudeCode(t *testing.T) {
	keys := readList(t, "palette-keys.txt")

	for _, tt := range canonical {
		for key := range decode(t, tt.palette).Overrides {
			if !keys[key] {
				t.Errorf(
					"%s: %s is not a Claude Code palette key",
					tt.palette,
					key,
				)
			}
		}
	}
}

func TestEveryKeyIsSetOrLeftToTheBase(t *testing.T) {
	keys := readList(t, "palette-keys.txt")

	for key := range identity {
		if !keys[key] {
			t.Errorf("identity key %s is not a Claude Code palette key", key)
		}
	}

	for _, tt := range canonical {
		overrides := decode(t, tt.palette).Overrides

		for key := range keys {
			_, set := overrides[key]

			switch {
			case set && identity[key]:
				t.Errorf("%s sets identity key %s", tt.palette, key)

			case !set && !identity[key]:
				t.Errorf("%s does not set %s", tt.palette, key)
			}
		}
	}
}

// TestTextIsReadable checks every key Claude Code draws text in against the
// ground it is drawn on. Claude Code draws on the terminal background, which
// is Field under a quiet terminal port.
//
// subtle is excluded: it marks nonessential hints, and Subtle is not held to
// 4.5:1. So are diffAddedWord and diffRemovedWord as text: see mapping.
func TestTextIsReadable(t *testing.T) {
	onField := []string{
		"text",
		"inactive",
		"autoAccept",
		"autoAcceptShimmer",
		"skill",
		"bashBorder",
		"claudeBlue_FOR_SYSTEM_SPINNER",
		"claudeBlueShimmer_FOR_SYSTEM_SPINNER",
		"permission",
		"permissionShimmer",
		"planMode",
		"ide",
		"inactiveShimmer",
		"suggestion",
		"remember",
		"background",
		"success",
		"error",
		"warning",
		"merged",
		"warningShimmer",
		"professionalBlue",
		"fastMode",
		"fastModeShimmer",
		"effortUltra",
		"briefLabelYou",
	}

	// Grounds Claude Code draws ordinary text on.
	underText := []string{
		"userMessageBackground",
		"userMessageBackgroundHover",
		"composerSidebarBackground",
		"selectionBg",
		"bashMessageBackgroundColor",
		"memoryBackgroundColor",
		"diffAdded",
		"diffRemoved",
		"diffAddedDimmed",
		"diffRemovedDimmed",
		"diffAddedWord",
		"diffRemovedWord",
	}

	// Badges draw inverseText on a state color.
	badges := []string{
		"autoAccept",
		"permission",
		"planMode",
		"suggestion",
		"background",
		"success",
		"error",
		"warning",
		"merged",
		"professionalBlue",
		"fastMode",
	}

	for _, tt := range canonical {
		th := decode(t, tt.palette)
		field := loadPalette(t, tt.palette).Anchors.Field

		get := func(key string) color.SRGB {
			t.Helper()

			value, ok := th.Overrides[key]
			if !ok {
				t.Fatalf("%s: %s is not set", tt.palette, key)
			}

			return color.MustParseHex(value)
		}

		check := func(fg, bg string, f, b color.SRGB) {
			t.Helper()

			if ratio := color.Contrast(f, b); ratio < 4.5 {
				t.Errorf(
					"%s: %s on %s = %.2f, want at least 4.5",
					tt.palette,
					fg,
					bg,
					ratio,
				)
			}
		}

		for _, key := range onField {
			check(key, "Field", get(key), field)
		}

		for _, key := range underText {
			check("text", key, get("text"), get(key))
		}

		for _, key := range badges {
			check("inverseText", key, get("inverseText"), get(key))
		}

		// The memory marker is drawn in remember on its own ground.
		check(
			"remember",
			"memoryBackgroundColor",
			get("remember"),
			get("memoryBackgroundColor"),
		)
	}
}

// A shimmer key is the highlight Claude Code sweeps across an animated
// label; it must differ from the label for the animation to show.
func TestShimmersDifferFromTheirLabels(t *testing.T) {
	pairs := map[string]string{
		"autoAcceptShimmer":                    "autoAccept",
		"claudeBlueShimmer_FOR_SYSTEM_SPINNER": "claudeBlue_FOR_SYSTEM_SPINNER",
		"permissionShimmer":                    "permission",
		"promptBorderShimmer":                  "promptBorder",
		"inactiveShimmer":                      "inactive",
		"warningShimmer":                       "warning",
		"fastModeShimmer":                      "fastMode",
	}

	for _, tt := range canonical {
		overrides := decode(t, tt.palette).Overrides

		for shimmer, label := range pairs {
			if overrides[shimmer] == overrides[label] {
				t.Errorf(
					"%s: %s matches %s",
					tt.palette,
					shimmer,
					label,
				)
			}
		}
	}
}

func encode(t *testing.T, file string) []byte {
	t.Helper()

	var b bytes.Buffer

	if err := Encode(&b, loadPalette(t, file)); err != nil {
		t.Fatal(err)
	}

	return b.Bytes()
}

func decode(t *testing.T, file string) theme {
	t.Helper()

	var th theme

	if err := json.Unmarshal(encode(t, file), &th); err != nil {
		t.Fatalf("%s: %v", file, err)
	}

	return th
}

func loadPalette(t *testing.T, file string) core.Palette {
	t.Helper()

	palette, err := core.LoadPaletteFile(
		repoPath(
			t,
			"spec",
			file,
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	return palette
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

	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}

	return result
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
