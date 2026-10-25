// @(#)quiet / go / internal/port/delta/delta_test.go / 2026-10-24

package delta

import (
	"bytes"
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
	"github.com/chrismecca/quiet/internal/port/tmtheme"
)

func TestCanonicalPalettesMatchCommittedArtifact(t *testing.T) {
	day, night := palettes(t)

	got := encode(t, day, night)

	want, err := os.ReadFile(
		repoPath(
			t,
			"ports",
			"delta",
			"quiet.gitconfig",
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, want) {
		t.Fatal(
			"generated delta configuration differs from quiet.gitconfig",
		)
	}
}

// Under simulated deuteranopia the Moss and Rust washes nearly coincide, so
// every feature must keep delta's +/- markers.
func TestEveryFeatureKeepsPlusMinusMarkers(t *testing.T) {
	day, night := palettes(t)
	got := string(encode(t, day, night))

	features := strings.Count(got, "\n[delta \"")
	markers := strings.Count(got, "keep-plus-minus-markers = true")

	if features != 2 || markers != features {
		t.Fatalf(
			"features = %d, keep-plus-minus-markers = %d; want 2 and 2",
			features,
			markers,
		)
	}
}

func TestEncodeRejectsSwappedPalettes(t *testing.T) {
	day, night := palettes(t)

	var b bytes.Buffer

	if err := Encode(&b, night, day); err == nil {
		t.Fatal(
			"Encode accepted night and day in the wrong order",
		)
	}
}

// delta renders a sample diff with each feature, using an isolated home
// directory and bat cache, so the user's configuration is never read. The output must keep the - and + markers and
// paint removed and added lines with the Rust and Moss washes.
func TestDeltaRendersWithFeatures(t *testing.T) {
	deltaBinary, err := exec.LookPath("delta")
	if err != nil {
		t.Skip("delta is not installed")
	}

	batBinary, err := exec.LookPath("bat")
	if err != nil {
		t.Skip("bat is not installed")
	}

	day, night := palettes(t)
	root := t.TempDir()

	config := filepath.Join(root, "quiet.gitconfig")
	if err := os.WriteFile(config, encode(t, day, night), 0o644); err != nil {
		t.Fatal(err)
	}

	cache := buildBatCache(t, batBinary, root, day, night)

	diff := "--- a/x.go\n+++ b/x.go\n@@ -1 +1 @@\n" +
		"-package main // old\n" +
		"+package main // new\n"

	for _, p := range []core.Palette{day, night} {
		// delta reads git configuration through libgit2, which ignores
		// GIT_CONFIG_GLOBAL, so the test isolates HOME instead.
		home := filepath.Join(root, "home-"+p.Name)

		if err := os.MkdirAll(home, 0o755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(
			filepath.Join(home, ".gitconfig"),
			[]byte("[include]\n\tpath = "+config+"\n[delta]\n\tfeatures = "+p.Name+"\n"),
			0o644,
		); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(deltaBinary)
		cmd.Env = append(
			os.Environ(),
			"HOME="+home,
			"XDG_CONFIG_HOME="+filepath.Join(home, ".config"),
			"GIT_CONFIG_NOSYSTEM=1",
			"XDG_CACHE_HOME="+cache,
		)
		cmd.Stdin = strings.NewReader(diff)
		cmd.Dir = root

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf(
				"%s: delta: %v\n%s",
				p.Name,
				err,
				output,
			)
		}

		derived, err := core.Derive(p)
		if err != nil {
			t.Fatal(err)
		}

		for _, want := range []string{
			background(derived.Wash.Rust.Line) + "-",
			background(derived.Wash.Moss.Line) + "+",
			foreground(p.Anchors.Mineral) + "package",
		} {
			if !bytes.Contains(output, []byte(want)) {
				t.Errorf(
					"%s: delta output lacks %q:\n%q",
					p.Name,
					want,
					output,
				)
			}
		}
	}
}

// buildBatCache compiles the quiet tmThemes into a bat cache laid out the
// way delta reads it: $XDG_CACHE_HOME/bat.
func buildBatCache(
	t *testing.T,
	bat string,
	root string,
	day core.Palette,
	night core.Palette,
) string {
	t.Helper()

	themes := filepath.Join(root, "bat-config", "themes")
	if err := os.MkdirAll(themes, 0o755); err != nil {
		t.Fatal(err)
	}

	for _, p := range []core.Palette{day, night} {
		var b bytes.Buffer

		if err := tmtheme.Encode(&b, p); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(
			filepath.Join(themes, p.Name+".tmTheme"),
			b.Bytes(),
			0o644,
		); err != nil {
			t.Fatal(err)
		}
	}

	xdg := filepath.Join(root, "xdg-cache")

	build := exec.Command(
		bat, "cache", "--build",
		"--source", filepath.Join(root, "bat-config"),
		"--target", filepath.Join(xdg, "bat"),
	)

	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf(
			"bat cache --build: %v\n%s",
			err,
			output,
		)
	}

	return xdg
}

func background(c color.SRGB) string {
	return "\x1b[48;2;" + channels(c) + "m"
}

func foreground(c color.SRGB) string {
	return "\x1b[38;2;" + channels(c) + "m"
}

func channels(c color.SRGB) string {
	return strings.Join(
		[]string{
			strconv.Itoa(int(math.Round(c.R * 255))),
			strconv.Itoa(int(math.Round(c.G * 255))),
			strconv.Itoa(int(math.Round(c.B * 255))),
		},
		";",
	)
}

func palettes(t *testing.T) (core.Palette, core.Palette) {
	t.Helper()

	day, err := core.LoadPaletteFile(
		repoPath(
			t,
			"spec",
			"quiet-day.json",
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	night, err := core.LoadPaletteFile(
		repoPath(
			t,
			"spec",
			"quiet-night.json",
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	return day, night
}

func encode(
	t *testing.T,
	day core.Palette,
	night core.Palette,
) []byte {
	t.Helper()

	var b bytes.Buffer

	if err := Encode(&b, day, night); err != nil {
		t.Fatal(err)
	}

	return b.Bytes()
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
