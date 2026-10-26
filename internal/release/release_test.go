// @(#)quiet / go / internal/release/release_test.go / 2026-10-25

package release

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestVersionIsSemantic(t *testing.T) {
	if !regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(Version) {
		t.Fatalf("Version %q is not MAJOR.MINOR.PATCH", Version)
	}
}

func TestVSCodeManifestMatchesVersion(t *testing.T) {
	data, err := os.ReadFile(repoPath("ports", "vscode", "package.json"))
	if err != nil {
		t.Fatal(err)
	}

	var manifest struct {
		Version string `json:"version"`
	}

	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}

	if manifest.Version != Version {
		t.Errorf("ports/vscode/package.json version is %s, want %s", manifest.Version, Version)
	}
}

func TestEmacsPackageMatchesVersion(t *testing.T) {
	data, err := os.ReadFile(repoPath("ports", "emacs", "quiet-themes.el"))
	if err != nil {
		t.Fatal(err)
	}

	m := regexp.MustCompile(`(?m)^;; Version: (\S+)$`).FindSubmatch(data)
	if m == nil {
		t.Fatal("ports/emacs/quiet-themes.el has no Version header")
	}

	if string(m[1]) != Version {
		t.Errorf("ports/emacs/quiet-themes.el version is %s, want %s", m[1], Version)
	}
}

// The newest released section of the changelog names Version.
func TestChangelogNamesVersion(t *testing.T) {
	data, err := os.ReadFile(repoPath("CHANGELOG.adoc"))
	if err != nil {
		t.Fatal(err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "== ") || line == "== Unreleased" {
			continue
		}

		if !strings.HasPrefix(line, "== "+Version+" ") {
			t.Errorf("newest changelog release is %q, want %s", line, Version)
		}

		return
	}

	t.Errorf("CHANGELOG.adoc has no section for %s", Version)
}

func repoPath(parts ...string) string {
	return filepath.Join(append([]string{"..", ".."}, parts...)...)
}
