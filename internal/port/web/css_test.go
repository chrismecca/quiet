// @(#)quiet / go / internal/port/web/css_test.go / 2026-10-18

package web

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chrismecca/quiet/internal/core"
)

func TestCanonicalCSSMatchesCommittedArtifact(t *testing.T) {
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

	var got bytes.Buffer

	if err := EncodeCSS(
		&got,
		day,
		night,
	); err != nil {
		t.Fatalf(
			"EncodeCSS: %v",
			err,
		)
	}

	want, err := os.ReadFile(
		repoPath(
			t,
			"ports",
			"web",
			"quiet.css",
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
		t.Fatalf(
			"generated CSS differs from ports/web/quiet.css",
		)
	}
}

func TestCSSDoesNotCreateGlobalSubtleToken(t *testing.T) {
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

	var css bytes.Buffer

	if err := EncodeCSS(
		&css,
		day,
		night,
	); err != nil {
		t.Fatalf(
			"EncodeCSS: %v",
			err,
		)
	}

	if strings.Contains(
		css.String(),
		"--quiet-subtle:",
	) {
		t.Fatal(
			"web export must not create a global Subtle token",
		)
	}
}

func TestCSSProvidesExplicitDayAndNightOverrides(t *testing.T) {
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

	var css bytes.Buffer

	if err := EncodeCSS(
		&css,
		day,
		night,
	); err != nil {
		t.Fatalf(
			"EncodeCSS: %v",
			err,
		)
	}

	for _, want := range []string{
		`[data-quiet-theme="day"]`,
		`[data-quiet-theme="night"]`,
		`@media (prefers-color-scheme: dark)`,
		`:root:not([data-quiet-theme])`,
	} {
		if !strings.Contains(
			css.String(),
			want,
		) {
			t.Errorf(
				"generated CSS missing %q",
				want,
			)
		}
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

	parts := append(
		[]string{root},
		elem...,
	)

	return filepath.Join(parts...)
}
