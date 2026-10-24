// @(#)quiet / go / internal/port/fzf/fzf_test.go / 2026-10-24

package fzf

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
)

var canonical = []struct {
	palette  string
	artifact string
}{
	{"quiet-day.json", "quiet-day.fzfrc"},
	{"quiet-night.json", "quiet-night.fzfrc"},
}

func TestCanonicalPalettesMatchCommittedArtifacts(t *testing.T) {
	for _, tt := range canonical {
		t.Run(tt.palette, func(t *testing.T) {
			got := encode(t, tt.palette)

			want, err := os.ReadFile(
				repoPath(
					t,
					"ports",
					"fzf",
					tt.artifact,
				),
			)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(got, want) {
				t.Fatalf(
					"generated fzf options differ from %q",
					tt.artifact,
				)
			}
		})
	}
}

// Every element that draws readable text must reach 4.5:1 on the background
// it is drawn on. Inherited names resolve through fzf's own hierarchy: for
// example selected-hl inherits hl and selected-fg inherits fg.
func TestTextElementsAreReadableOnTheirBackgrounds(t *testing.T) {
	pairs := []struct {
		foreground string
		background string
	}{
		{"fg", "bg"},
		{"hl", "bg"},
		{"query", "bg"},
		{"prompt", "bg"},
		{"disabled", "bg"},
		{"info", "bg"},
		{"header", "bg"},
		{"footer", "bg"},
		{"label", "bg"},
		{"fg+", "bg+"},
		{"hl+", "bg+"},
		{"fg", "selected-bg"},
		{"hl", "selected-bg"},
		{"preview-fg", "preview-bg"},
		{"label", "preview-bg"},
	}

	for _, tt := range canonical {
		colors := parse(t, encode(t, tt.palette))

		for _, p := range pairs {
			ratio := color.Contrast(
				lookup(t, colors, p.foreground),
				lookup(t, colors, p.background),
			)

			if ratio < 4.5 {
				t.Errorf(
					"%s: %s on %s = %.2f, want at least 4.5",
					tt.palette,
					p.foreground,
					p.background,
					ratio,
				)
			}
		}
	}
}

func TestCurrentLineMatchesUseTypography(t *testing.T) {
	for _, tt := range canonical {
		for _, line := range strings.Split(string(encode(t, tt.palette)), "\n") {
			if strings.HasPrefix(line, "--color=hl+:") &&
				!strings.HasSuffix(line, ":regular:bold") {
				t.Errorf(
					"%s: %q does not mark current-line matches in bold",
					tt.palette,
					line,
				)
			}
		}
	}
}

// fzf validates color names and values when it loads an options file. The
// smoke test runs only where fzf is installed; exit status 2 means fzf
// rejected the file.
func TestFzfAcceptsGeneratedOptions(t *testing.T) {
	binary, err := exec.LookPath("fzf")
	if err != nil {
		t.Skip("fzf is not installed")
	}

	for _, tt := range canonical {
		path := filepath.Join(t.TempDir(), tt.artifact)

		if err := os.WriteFile(path, encode(t, tt.palette), 0o644); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(binary, "--filter", "quiet")
		cmd.Env = append(os.Environ(), "FZF_DEFAULT_OPTS_FILE="+path)
		cmd.Stdin = strings.NewReader("quiet\n")

		output, err := cmd.CombinedOutput()

		var exit *exec.ExitError
		if err != nil && !(errors.As(err, &exit) && exit.ExitCode() == 1) {
			t.Errorf(
				"%s: fzf rejected options: %v\n%s",
				tt.artifact,
				err,
				output,
			)
		}
	}
}

func encode(t *testing.T, file string) []byte {
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

	return b.Bytes()
}

// parse returns each --color name and its hexadecimal value.
func parse(t *testing.T, options []byte) map[string]string {
	t.Helper()

	colors := make(map[string]string)

	for _, line := range strings.Split(string(options), "\n") {
		spec, ok := strings.CutPrefix(line, "--color=")
		if !ok {
			continue
		}

		fields := strings.Split(spec, ":")
		if len(fields) >= 2 {
			colors[fields[0]] = fields[1]
		}
	}

	return colors
}

func lookup(
	t *testing.T,
	colors map[string]string,
	name string,
) color.SRGB {
	t.Helper()

	value, ok := colors[name]
	if !ok {
		t.Fatalf("no fzf color %q", name)
	}

	c, err := color.ParseHex(value)
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
