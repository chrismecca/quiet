// @(#)quiet / go / internal/port/alacritty/alacritty_test.go / 2026-10-25

package alacritty

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
	"github.com/chrismecca/quiet/internal/port/terminal"
)

var paletteNames = []string{"quiet-day", "quiet-night"}

func TestCommittedThemesMatchGenerator(t *testing.T) {
	for _, name := range paletteNames {
		got := encode(t, loadPalette(t, name))

		want, err := os.ReadFile(repoPath("ports", "alacritty", name+".toml"))
		if err != nil {
			t.Fatal(err)
		}

		if got != string(want) {
			t.Errorf("ports/alacritty/%s.toml is stale; run make alacritty", name)
		}
	}
}

func TestThemePreservesGenericTerminalPalette(t *testing.T) {
	for _, name := range paletteNames {
		p := loadPalette(t, name)

		term, err := terminal.Resolve(p)
		if err != nil {
			t.Fatal(err)
		}

		got := encode(t, p)

		for bank, table := range []string{"normal", "bright"} {
			section := tableBody(t, got, "colors."+table)

			for i, ansi := range ansiNames {
				want := fmt.Sprintf("%s = %q", ansi, term.ANSI[bank*8+i].MustHex())
				if !strings.Contains(section, want) {
					t.Errorf("%s [colors.%s] lacks %s", name, table, want)
				}
			}
		}

		primary := tableBody(t, got, "colors.primary")
		for _, want := range []string{
			fmt.Sprintf("background = %q", term.Background.MustHex()),
			fmt.Sprintf("foreground = %q", term.Text.MustHex()),
		} {
			if !strings.Contains(primary, want) {
				t.Errorf("%s [colors.primary] lacks %s", name, want)
			}
		}
	}
}

// Search matches keep each cell's color, so every terminal foreground must
// stay readable on the Mineral wash.
func TestCellColorsAreReadableOnSearchMatches(t *testing.T) {
	for _, name := range paletteNames {
		p := loadPalette(t, name)

		term, err := terminal.Resolve(p)
		if err != nil {
			t.Fatal(err)
		}

		derived, err := core.Derive(p)
		if err != nil {
			t.Fatal(err)
		}

		wash := derived.Wash.Mineral.Line

		if got := color.Contrast(term.Text, wash); got < 7 {
			t.Errorf("%s: text on search matches is %.2f, want 7", name, got)
		}

		for i, c := range term.ANSI {
			if i%8 == 0 || i%8 == 7 {
				continue // Neutral slots double as backgrounds.
			}

			if got := color.Contrast(c, wash); got < 4.5 {
				t.Errorf("%s: ANSI %d on search matches is %.2f, want 4.5", name, i, got)
			}
		}
	}
}

func TestThemeContainsOnlyColors(t *testing.T) {
	for _, line := range strings.Split(encode(t, loadPalette(t, "quiet-day")), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") {
			if !strings.HasPrefix(line, "[colors.") {
				t.Errorf("table outside [colors]: %s", line)
			}

			continue
		}

		key, value, ok := strings.Cut(line, " = ")
		if !ok {
			t.Errorf("malformed line %q", line)
			continue
		}

		if value != `"`+cellForeground+`"` && !strings.HasPrefix(value, `"#`) {
			t.Errorf("%s is not a color: %s", key, value)
		}
	}
}

// tableBody returns the lines of one TOML table.
func tableBody(t *testing.T, doc, table string) string {
	t.Helper()

	header := "[" + table + "]\n"

	start := strings.Index(doc, header)
	if start < 0 {
		t.Fatalf("no [%s] table", table)
	}

	body := doc[start+len(header):]
	if end := strings.Index(body, "\n["); end >= 0 {
		body = body[:end]
	}

	return body
}

func encode(t *testing.T, p core.Palette) string {
	t.Helper()

	var b bytes.Buffer
	if err := Encode(&b, p); err != nil {
		t.Fatal(err)
	}

	return b.String()
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
