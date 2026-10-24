// @(#)quiet / go / internal/port/tmtheme/tmtheme_test.go / 2026-10-24

package tmtheme

import (
	"bytes"
	"encoding/xml"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
)

var canonical = []struct {
	palette  string
	artifact string
}{
	{"quiet-day.json", "quiet-day.tmTheme"},
	{"quiet-night.json", "quiet-night.tmTheme"},
}

func TestCanonicalPalettesMatchCommittedArtifacts(t *testing.T) {
	for _, tt := range canonical {
		t.Run(tt.palette, func(t *testing.T) {
			got, _ := encode(t, tt.palette)

			want, err := os.ReadFile(
				repoPath(
					t,
					"ports",
					"tmtheme",
					tt.artifact,
				),
			)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(got, want) {
				t.Fatalf(
					"generated tmTheme differs from %q",
					tt.artifact,
				)
			}
		})
	}
}

func TestThemeIsWellFormedXML(t *testing.T) {
	for _, tt := range canonical {
		got, _ := encode(t, tt.palette)

		decoder := xml.NewDecoder(bytes.NewReader(got))
		decoder.Strict = true

		for {
			_, err := decoder.Token()
			if err == io.EOF {
				break
			}

			if err != nil {
				t.Fatalf(
					"%s: %v",
					tt.artifact,
					err,
				)
			}
		}
	}
}

// quiet-core guarantees only Ink on Selection, so selected text must be
// forced to Ink rather than keep its syntax color.
func TestSelectedTextIsInk(t *testing.T) {
	for _, tt := range canonical {
		got, palette := encode(t, tt.palette)

		want := "<key>selectionForeground</key>\n\t\t\t\t<string>" +
			palette.Anchors.Ink.MustHex() + "</string>"

		if !strings.Contains(string(got), want) {
			t.Errorf(
				"%s: selectionForeground is not Ink",
				tt.artifact,
			)
		}
	}
}

func TestPlutilAcceptsTheme(t *testing.T) {
	plutil, err := exec.LookPath("plutil")
	if err != nil {
		t.Skip("plutil is not installed")
	}

	for _, tt := range canonical {
		path := writeTheme(t, t.TempDir(), tt)

		if output, err := exec.Command(plutil, "-lint", path).CombinedOutput(); err != nil {
			t.Errorf(
				"%s: plutil rejected theme: %v\n%s",
				tt.artifact,
				err,
				output,
			)
		}
	}
}

// bat compiles tmThemes into its cache and fails on themes it cannot parse.
// The cache is isolated in a temporary directory so the test never touches
// the user's bat configuration.
func TestBatRendersWithTheme(t *testing.T) {
	bat, err := exec.LookPath("bat")
	if err != nil {
		t.Skip("bat is not installed")
	}

	root := t.TempDir()
	themes := filepath.Join(root, "config", "themes")
	cache := filepath.Join(root, "cache")

	if err := os.MkdirAll(themes, 0o755); err != nil {
		t.Fatal(err)
	}

	for _, tt := range canonical {
		writeTheme(t, themes, tt)
	}

	build := exec.Command(
		bat, "cache", "--build",
		"--source", filepath.Join(root, "config"),
		"--target", cache,
	)

	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf(
			"bat cache --build: %v\n%s",
			err,
			output,
		)
	}

	for _, tt := range canonical {
		name := strings.TrimSuffix(tt.artifact, ".tmTheme")

		render := exec.Command(
			bat, "--theme", name,
			"--color", "always",
			"--style", "plain",
			"--language", "go",
		)
		render.Env = append(os.Environ(), "BAT_CACHE_PATH="+cache)
		render.Stdin = strings.NewReader("package main // quiet\n")

		output, err := render.CombinedOutput()
		if err != nil {
			t.Fatalf(
				"bat --theme %s: %v\n%s",
				name,
				err,
				output,
			)
		}

		// bat emits one 24-bit escape per token, so the keyword must start
		// with Mineral and the comment marker with Muted.
		_, palette := encode(t, tt.palette)

		for _, expect := range []struct {
			role  string
			value [3]float64
			text  string
		}{
			{"Mineral", rgb(palette.Anchors.Mineral), "package"},
			{"Muted", rgb(palette.Anchors.Muted), "//"},
		} {
			want := []byte(
				"\x1b[38;2;" +
					channel(expect.value[0]) + ";" +
					channel(expect.value[1]) + ";" +
					channel(expect.value[2]) + "m" + expect.text,
			)

			if !bytes.Contains(output, want) {
				t.Errorf(
					"bat --theme %s did not render %q in %s:\n%q",
					name,
					expect.text,
					expect.role,
					output,
				)
			}
		}
	}
}

func encode(t *testing.T, file string) ([]byte, core.Palette) {
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

	var b bytes.Buffer

	if err := Encode(&b, palette); err != nil {
		t.Fatal(err)
	}

	return b.Bytes(), palette
}

func writeTheme(
	t *testing.T,
	dir string,
	tt struct {
		palette  string
		artifact string
	},
) string {
	t.Helper()

	got, _ := encode(t, tt.palette)
	path := filepath.Join(dir, tt.artifact)

	if err := os.WriteFile(path, got, 0o644); err != nil {
		t.Fatal(err)
	}

	return path
}

func rgb(c color.SRGB) [3]float64 {
	return [3]float64{c.R, c.G, c.B}
}

// channel converts a gamma-encoded component to the 8-bit decimal that bat
// emits in 24-bit escape sequences.
func channel(x float64) string {
	return strconv.Itoa(int(math.Round(x * 255)))
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
