// @(#)quiet / go / internal/port/emacs/emacs_app_test.go / 2026-10-25

package emacs

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// The test in this file runs an installed Emacs and skips when none is on
// PATH. Set QUIET_EMACS to choose a binary.

// libraries define the faces the theme sets beyond Emacs's preloaded ones.
var libraries = []string{
	"ansi-color",
	"diff-mode",
	"ediff",
	"eglot",
	"flymake",
	"flyspell",
	"smerge-mode",
	"tab-line",
}

// TestThemesLoadInEmacs loads both themes in a clean Emacs and checks that
// every face the theme sets exists, that every face it underlines is
// underlined in Emacs's own default, and that it keeps every wavy underline
// Emacs's defaults draw.
func TestThemesLoadInEmacs(t *testing.T) {
	emacs := emacsPath(t)

	var underlined, waved, plain []string
	for _, f := range faces() {
		switch f.underline {
		case "":
			plain = append(plain, f.name)

		case "wave":
			waved = append(waved, f.name)
			underlined = append(underlined, f.name)

		default:
			underlined = append(underlined, f.name)
		}
	}

	all := append(append([]string{}, plain...), underlined...)

	script := `(progn
  (dolist (lib '(` + strings.Join(libraries, " ") + `)) (require lib))
  (add-to-list 'custom-theme-load-path "` + repoPath(t, "ports", "emacs") + `")
  (dolist (theme '(quiet-day quiet-night)) (load-theme theme t))
  (dolist (f '(` + strings.Join(all, " ") + `))
    (unless (facep f) (princ (format "unknown face %s\n" f))))
  (dolist (f '(` + strings.Join(underlined, " ") + `))
    (unless (string-match-p ":underline" (format "%S" (face-default-spec f)))
      (princ (format "%s has no underline in Emacs's default\n" f))))
  (dolist (f '(` + strings.Join(plain, " ") + `))
    (when (string-match-p ":style wave" (format "%S" (face-default-spec f)))
      (princ (format "%s drops the wavy underline in Emacs's default\n" f)))))`

	out, err := exec.Command(emacs, "--batch", "-Q", "--eval", script).CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}

	for _, line := range strings.Split(string(out), "\n") {
		if line != "" && !strings.HasPrefix(line, "Loading") {
			t.Error(line)
		}
	}
}

func emacsPath(t *testing.T) string {
	t.Helper()

	if path := os.Getenv("QUIET_EMACS"); path != "" {
		return path
	}

	path, err := exec.LookPath("emacs")
	if err != nil {
		t.Skip("no Emacs installation found")
	}

	return path
}
