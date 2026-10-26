// @(#)quiet / go / internal/port/helix/hx_test.go / 2026-10-25

package helix

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The test in this file reads an installed Helix runtime and skips when none
// is found. Set QUIET_HELIX_RUNTIME to a Helix runtime directory to choose
// one.

var (
	capture  = regexp.MustCompile(`@([a-z][a-z0-9_.-]*)`)
	themeKey = regexp.MustCompile(`(?m)^"?([a-z][a-z0-9_.-]*)"?\s*=`)
)

// TestScopesExistInHelix checks every scope against the scopes the installed
// Helix knows: the captures in its highlight queries, and the keys its
// bundled themes set.
func TestScopesExistInHelix(t *testing.T) {
	runtime := helixRuntime(t)

	known := map[string]bool{}

	queries, err := filepath.Glob(filepath.Join(runtime, "queries", "*", "highlights.scm"))
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range queries {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		for _, m := range capture.FindAllSubmatch(data, -1) {
			known[string(m[1])] = true
		}
	}

	themes, err := filepath.Glob(filepath.Join(runtime, "themes", "*.toml"))
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range themes {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		for _, m := range themeKey.FindAllSubmatch(data, -1) {
			known[string(m[1])] = true
		}
	}

	if len(queries) == 0 || len(themes) == 0 {
		t.Fatalf("%s holds no queries or themes", runtime)
	}

	for _, s := range styles() {
		if !known[s.scope] {
			t.Errorf("%s is not a Helix scope", s.scope)
		}
	}
}

func helixRuntime(t *testing.T) string {
	t.Helper()

	if dir := os.Getenv("QUIET_HELIX_RUNTIME"); dir != "" {
		return dir
	}

	hx, err := exec.LookPath("hx")
	if err != nil {
		t.Skip("no Helix installation found")
	}

	if resolved, err := filepath.EvalSymlinks(hx); err == nil {
		hx = resolved
	}

	// Package managers install the runtime beside the binary or under
	// libexec.
	root := filepath.Dir(filepath.Dir(hx))
	for _, dir := range []string{
		filepath.Join(filepath.Dir(hx), "runtime"),
		filepath.Join(root, "libexec", "runtime"),
		filepath.Join(root, "share", "helix", "runtime"),
		filepath.Join(root, "lib", "helix", "runtime"),
	} {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	t.Skip("Helix runtime not found near " + strings.TrimSpace(hx))
	return ""
}
