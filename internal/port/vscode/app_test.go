// @(#)quiet / go / internal/port/vscode/app_test.go / 2026-10-25

package vscode

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/chrismecca/quiet/internal/core"
)

// The test in this file reads an installed VS Code and skips when none is
// found. Set QUIET_VSCODE_APP to a VS Code .app bundle or its resources
// directory to choose one.

// TestThemeKeysExistInVSCode checks every color key against the color IDs
// VS Code registers: in its workbench bundle, or contributed by its built-in
// extensions.
func TestThemeKeysExistInVSCode(t *testing.T) {
	app := vscodeApp(t)

	bundle, err := os.ReadFile(filepath.Join(app, "out", "vs", "workbench", "workbench.desktop.main.js"))
	if err != nil {
		t.Fatal(err)
	}

	contributed := map[string]bool{}

	manifests, err := filepath.Glob(filepath.Join(app, "extensions", "*", "package.json"))
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range manifests {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		var manifest struct {
			Contributes struct {
				Colors []struct {
					ID string `json:"id"`
				} `json:"colors"`
			} `json:"contributes"`
		}

		if err := json.Unmarshal(data, &manifest); err != nil {
			continue
		}

		for _, c := range manifest.Contributes.Colors {
			contributed[c.ID] = true
		}
	}

	palette, err := core.LoadPaletteFile(repoPath(t, "spec", "quiet-day.json"))
	if err != nil {
		t.Fatal(err)
	}

	for key := range generatedDocument(t, palette).Colors {
		if contributed[key] || bytes.Contains(bundle, []byte(`"`+key+`"`)) {
			continue
		}

		t.Errorf("%s is not a VS Code color", key)
	}
}

func vscodeApp(t *testing.T) string {
	t.Helper()

	candidates := []string{os.Getenv("QUIET_VSCODE_APP")}

	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, "Applications", "Visual Studio Code.app"))
	}

	candidates = append(candidates, "/Applications/Visual Studio Code.app")

	for _, c := range candidates {
		if c == "" {
			continue
		}

		for _, dir := range []string{
			filepath.Join(c, "Contents", "Resources", "app"),
			c,
		} {
			if _, err := os.Stat(filepath.Join(dir, "out", "vs", "workbench")); err == nil {
				return dir
			}
		}
	}

	t.Skip("no VS Code installation found")
	return ""
}
