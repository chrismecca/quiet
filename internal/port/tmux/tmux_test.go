// @(#)quiet / go / internal/port/tmux/tmux_test.go / 2026-10-19

package tmux

import (
	"bytes"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chrismecca/quiet/internal/core"
)

func TestResolveCanonicalPalettes(t *testing.T) {
	tests := []struct {
		name         string
		file         string
		artifactPath string
	}{
		{
			name:         "day",
			file:         "quiet-day.json",
			artifactPath: "ports/tmux/quiet-day.conf",
		},
		{
			name:         "night",
			file:         "quiet-night.json",
			artifactPath: "ports/tmux/quiet-night.conf",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				palette, err := core.LoadPaletteFile(
					repoPath(
						t,
						"spec",
						tt.file,
					),
				)
				if err != nil {
					t.Fatalf(
						"LoadPaletteFile: %v",
						err,
					)
				}

				derived, err := core.Derive(palette)
				if err != nil {
					t.Fatalf(
						"Derive: %v",
						err,
					)
				}

				theme, err := Resolve(palette)
				if err != nil {
					t.Fatalf(
						"Resolve: %v",
						err,
					)
				}

				if theme.Name != palette.Name {
					t.Errorf(
						"Name = %q, want %q",
						theme.Name,
						palette.Name,
					)
				}

				if theme.ArtifactPath != tt.artifactPath {
					t.Errorf(
						"ArtifactPath = %q, want %q",
						theme.ArtifactPath,
						tt.artifactPath,
					)
				}

				want := map[string]string{
					"Field":     palette.Anchors.Field.MustHex(),
					"Surface":   palette.Anchors.Surface.MustHex(),
					"Structure": palette.Anchors.Structure.MustHex(),
					"Ink":       palette.Anchors.Ink.MustHex(),
					"Muted":     palette.Anchors.Muted.MustHex(),
					"Line":      palette.Anchors.Line.MustHex(),
					"Mineral":   palette.Anchors.Mineral.MustHex(),
					"Ochre":     palette.Anchors.Ochre.MustHex(),
					"Rust":      palette.Anchors.Rust.MustHex(),
					"Selection": derived.Selection.MustHex(),
					"Callout":   derived.Callout.MustHex(),
				}

				got := map[string]string{
					"Field":     theme.Field,
					"Surface":   theme.Surface,
					"Structure": theme.Structure,
					"Ink":       theme.Ink,
					"Muted":     theme.Muted,
					"Line":      theme.Line,
					"Mineral":   theme.Mineral,
					"Ochre":     theme.Ochre,
					"Rust":      theme.Rust,
					"Selection": theme.Selection,
					"Callout":   theme.Callout,
				}

				for role, wantValue := range want {
					if got[role] != wantValue {
						t.Errorf(
							"%s = %s, want %s",
							role,
							got[role],
							wantValue,
						)
					}
				}
			},
		)
	}
}

func TestEncodeUsesAppendOnlyStyleAssignments(
	t *testing.T,
) {
	theme := loadTheme(
		t,
		"quiet-night.json",
	)

	var buf bytes.Buffer

	if err := Encode(
		&buf,
		theme,
	); err != nil {
		t.Fatalf(
			"Encode: %v",
			err,
		)
	}

	for _, line := range strings.Split(
		buf.String(),
		"\n",
	) {
		if !strings.HasPrefix(
			line,
			"set-option ",
		) &&
			!strings.HasPrefix(
				line,
				"set-window-option ",
			) {
			continue
		}

		if !strings.Contains(
			line,
			" -ag ",
		) {
			t.Errorf(
				"style assignment is not append-only: %q",
				line,
			)
		}
	}
}

func TestEncodeDoesNotConfigureTmuxBehavior(
	t *testing.T,
) {
	theme := loadTheme(
		t,
		"quiet-night.json",
	)

	var buf bytes.Buffer

	if err := Encode(
		&buf,
		theme,
	); err != nil {
		t.Fatalf(
			"Encode: %v",
			err,
		)
	}

	forbidden := map[string]bool{
		"window-style":                 true,
		"window-active-style":          true,
		"status-left":                  true,
		"status-right":                 true,
		"status-format":                true,
		"window-status-format":         true,
		"window-status-current-format": true,
		"pane-border-format":           true,
		"pane-border-lines":            true,
		"pane-border-status":           true,
		"status-position":              true,
		"status-justify":               true,
		"mouse":                        true,
		"prefix":                       true,
	}

	for _, line := range strings.Split(
		buf.String(),
		"\n",
	) {
		if !strings.HasPrefix(
			line,
			"set-option ",
		) &&
			!strings.HasPrefix(
				line,
				"set-window-option ",
			) {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			t.Fatalf(
				"malformed generated tmux line: %q",
				line,
			)
		}

		option := fields[2]

		if forbidden[option] {
			t.Errorf(
				"generated configuration modifies forbidden option %q",
				option,
			)
		}
	}
}

func TestEncodeSemanticMappings(t *testing.T) {
	theme := loadTheme(
		t,
		"quiet-night.json",
	)

	var buf bytes.Buffer

	if err := Encode(
		&buf,
		theme,
	); err != nil {
		t.Fatalf(
			"Encode: %v",
			err,
		)
	}

	got := buf.String()

	want := []string{
		"set-option -ag status-style 'bg=" +
			theme.Structure +
			",fg=" +
			theme.Ink +
			"'",

		"set-window-option -ag window-status-current-style 'bg=" +
			theme.Selection +
			",fg=" +
			theme.Ink +
			"'",

		"set-window-option -ag pane-border-style 'fg=" +
			theme.Line +
			"'",

		"set-window-option -ag mode-style 'bg=" +
			theme.Selection +
			",fg=" +
			theme.Ink +
			"'",

		"set-window-option -ag copy-mode-match-style 'bg=" +
			theme.Callout +
			",fg=" +
			theme.Ink +
			"'",

		"set-window-option -ag copy-mode-current-match-style 'bg=" +
			theme.Selection +
			",fg=" +
			theme.Ink +
			"'",

		"set-window-option -ag menu-style 'bg=" +
			theme.Surface +
			",fg=" +
			theme.Ink +
			"'",

		"set-window-option -ag popup-style 'bg=" +
			theme.Surface +
			",fg=" +
			theme.Ink +
			"'",
	}

	for _, expected := range want {
		if !strings.Contains(
			got,
			expected,
		) {
			t.Errorf(
				"generated configuration missing:\n%s",
				expected,
			)
		}
	}
}

func TestEncodeActivePaneSemanticStates(
	t *testing.T,
) {
	theme := loadTheme(
		t,
		"quiet-night.json",
	)

	var buf bytes.Buffer

	if err := Encode(
		&buf,
		theme,
	); err != nil {
		t.Fatalf(
			"Encode: %v",
			err,
		)
	}

	want := "set-window-option -ag pane-active-border-style " +
		"'fg=#{?pane_modal_flag," +
		theme.Mineral +
		",#{?pane_marked," +
		theme.Ochre +
		",#{?synchronize-panes," +
		theme.Rust +
		",#{?pane_in_mode," +
		theme.Ochre +
		"," +
		theme.Mineral +
		"}}}}'"

	if !strings.Contains(
		buf.String(),
		want,
	) {
		t.Errorf(
			"active pane state mapping missing:\n%s",
			want,
		)
	}
}

func loadTheme(
	t *testing.T,
	file string,
) Theme {
	t.Helper()

	palette, err := core.LoadPaletteFile(
		repoPath(
			t,
			"spec",
			file,
		),
	)
	if err != nil {
		t.Fatalf(
			"LoadPaletteFile: %v",
			err,
		)
	}

	theme, err := Resolve(palette)
	if err != nil {
		t.Fatalf(
			"Resolve: %v",
			err,
		)
	}

	return theme
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
