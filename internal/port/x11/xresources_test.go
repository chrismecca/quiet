// @(#)quiet / go / internal/port/x11/xresources_test.go / 2026-10-18

package x11

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chrismecca/quiet/internal/core"
)

func TestCanonicalXResourcesMatchCommittedArtifacts(t *testing.T) {
	tests := []struct {
		palette  string
		artifact string
	}{
		{
			palette:  "quiet-day.json",
			artifact: "quiet-day.Xresources",
		},
		{
			palette:  "quiet-night.json",
			artifact: "quiet-night.Xresources",
		},
	}

	for _, tt := range tests {
		t.Run(tt.palette, func(t *testing.T) {
			palette, err := core.LoadPaletteFile(
				repoPath(
					t,
					"spec",
					tt.palette,
				),
			)
			if err != nil {
				t.Fatalf(
					"LoadPaletteFile: %v",
					err,
				)
			}

			resolved := Resolve(palette)

			var got bytes.Buffer

			if err := EncodeXResources(
				&got,
				resolved,
			); err != nil {
				t.Fatalf(
					"EncodeXResources: %v",
					err,
				)
			}

			want, err := os.ReadFile(
				repoPath(
					t,
					"ports",
					"x11",
					tt.artifact,
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
					"generated X resources differ from %s",
					tt.artifact,
				)
			}
		})
	}
}

func TestXResourcesDoNotInventSemanticFamilies(t *testing.T) {
	palette, err := core.LoadPaletteFile(
		repoPath(
			t,
			"spec",
			"quiet-day.json",
		),
	)
	if err != nil {
		t.Fatalf(
			"LoadPaletteFile: %v",
			err,
		)
	}

	var output bytes.Buffer

	if err := EncodeXResources(
		&output,
		Resolve(palette),
	); err != nil {
		t.Fatalf(
			"EncodeXResources: %v",
			err,
		)
	}

	for _, forbidden := range []string{
		palette.Anchors.Moss.MustHex(),
		palette.Anchors.Ochre.MustHex(),
		palette.Anchors.Rust.MustHex(),
	} {
		if strings.Contains(
			output.String(),
			forbidden,
		) {
			t.Errorf(
				"generic X11 port unexpectedly emits unused semantic color %s",
				forbidden,
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
