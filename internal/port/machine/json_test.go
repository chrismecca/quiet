// @(#)quiet / go / internal/port/machine/json_test.go / 2026-10-24

package machine

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/chrismecca/quiet/internal/core"
)

func TestCanonicalJSONMatchesCommittedArtifact(t *testing.T) {
	day, night := canonicalPalettes(t)

	var got bytes.Buffer

	if err := EncodeJSON(
		&got,
		day,
		night,
	); err != nil {
		t.Fatalf(
			"EncodeJSON: %v",
			err,
		)
	}

	want, err := os.ReadFile(
		repoPath(
			t,
			"ports",
			"machine",
			"quiet.json",
		),
	)
	if err != nil {
		t.Fatalf(
			"ReadFile: %v",
			err,
		)
	}

	if !bytes.Equal(
		got.Bytes(),
		want,
	) {
		t.Fatal(
			"generated machine document differs from ports/machine/quiet.json",
		)
	}
}

func TestMachineDocumentPreservesContextualSubtle(t *testing.T) {
	day, night := canonicalPalettes(t)

	var output bytes.Buffer

	if err := EncodeJSON(
		&output,
		day,
		night,
	); err != nil {
		t.Fatalf(
			"EncodeJSON: %v",
			err,
		)
	}

	var doc struct {
		Modes struct {
			Day struct {
				Derived map[string]json.RawMessage `json:"derived"`
			} `json:"day"`

			Night struct {
				Derived map[string]json.RawMessage `json:"derived"`
			} `json:"night"`
		} `json:"modes"`

		Contextual struct {
			Subtle struct {
				Foreground  string  `json:"foreground"`
				Opacity     float64 `json:"opacity"`
				Background  string  `json:"background"`
				Compositing string  `json:"compositing"`
			} `json:"subtle"`
		} `json:"contextual"`
	}

	if err := json.Unmarshal(
		output.Bytes(),
		&doc,
	); err != nil {
		t.Fatalf(
			"json.Unmarshal: %v",
			err,
		)
	}

	if _, ok := doc.Modes.Day.Derived["subtle"]; ok {
		t.Fatal(
			"Day derived roles must not contain an opaque Subtle color",
		)
	}

	if _, ok := doc.Modes.Night.Derived["subtle"]; ok {
		t.Fatal(
			"Night derived roles must not contain an opaque Subtle color",
		)
	}

	got := doc.Contextual.Subtle

	if got.Foreground != "muted" {
		t.Errorf(
			"Subtle foreground = %q, want %q",
			got.Foreground,
			"muted",
		)
	}

	if got.Opacity != 0.55 {
		t.Errorf(
			"Subtle opacity = %.6f, want 0.55",
			got.Opacity,
		)
	}

	if got.Background != "local" {
		t.Errorf(
			"Subtle background = %q, want %q",
			got.Background,
			"local",
		)
	}

	if got.Compositing != "legacy-srgb" {
		t.Errorf(
			"Subtle compositing = %q, want %q",
			got.Compositing,
			"legacy-srgb",
		)
	}
}

func canonicalPalettes(
	t *testing.T,
) (core.Palette, core.Palette) {
	t.Helper()

	day, err := core.LoadPaletteFile(
		repoPath(
			t,
			"spec",
			"quiet-day.json",
		),
	)
	if err != nil {
		t.Fatalf(
			"LoadPaletteFile(day): %v",
			err,
		)
	}

	night, err := core.LoadPaletteFile(
		repoPath(
			t,
			"spec",
			"quiet-night.json",
		),
	)
	if err != nil {
		t.Fatalf(
			"LoadPaletteFile(night): %v",
			err,
		)
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

func TestMachineDocumentExportsFieldWashes(t *testing.T) {
	day, night := canonicalPalettes(t)

	var output bytes.Buffer

	if err := EncodeJSON(
		&output,
		day,
		night,
	); err != nil {
		t.Fatalf(
			"EncodeJSON: %v",
			err,
		)
	}

	type washes map[string]struct {
		Line     string `json:"line"`
		Emphasis string `json:"emphasis"`
	}

	var doc struct {
		Modes map[string]struct {
			Derived struct {
				Wash washes `json:"wash"`
			} `json:"derived"`
		} `json:"modes"`
	}

	if err := json.Unmarshal(
		output.Bytes(),
		&doc,
	); err != nil {
		t.Fatalf(
			"json.Unmarshal: %v",
			err,
		)
	}

	for mode, palette := range map[string]core.Palette{
		"day":   day,
		"night": night,
	} {
		derived, err := core.Derive(palette)
		if err != nil {
			t.Fatalf(
				"Derive: %v",
				err,
			)
		}

		for _, family := range core.Families() {
			want, err := derived.Wash.For(family)
			if err != nil {
				t.Fatalf(
					"Wash.For: %v",
					err,
				)
			}

			got, ok := doc.Modes[mode].Derived.Wash[string(family)]
			if !ok {
				t.Fatalf(
					"%s: missing %s wash",
					mode,
					family,
				)
			}

			if got.Line != want.Line.MustHex() ||
				got.Emphasis != want.Emphasis.MustHex() {
				t.Errorf(
					"%s %s wash = %s/%s, want %s/%s",
					mode,
					family,
					got.Line,
					got.Emphasis,
					want.Line.MustHex(),
					want.Emphasis.MustHex(),
				)
			}
		}
	}
}
