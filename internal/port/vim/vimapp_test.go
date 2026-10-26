// @(#)quiet / go / internal/port/vim/vimapp_test.go / 2026-10-25

package vim

import (
	"os"
	"os/exec"
	"testing"
)

// TestSchemesLoadInVim loads each scheme the way a plugin manager does: with
// ports/vim on the runtimepath and :colorscheme. It runs an installed Vim
// and skips when none is on PATH. Set QUIET_VIM to choose a binary.
func TestSchemesLoadInVim(t *testing.T) {
	vim := os.Getenv("QUIET_VIM")
	if vim == "" {
		path, err := exec.LookPath("vim")
		if err != nil {
			t.Skip("no Vim installation found")
		}

		vim = path
	}

	for _, name := range []string{"quiet-day", "quiet-night"} {
		out, err := exec.Command(
			vim, "-Nu", "NONE", "-es",
			"+set runtimepath^="+repoPath(t, "ports", "vim"),
			"+colorscheme "+name,
			"+if g:colors_name !=# '"+name+"' | cquit | endif",
			"+if len(g:terminal_ansi_colors) != 16 | cquit | endif",
			"+qa!",
		).CombinedOutput()
		if err != nil {
			t.Errorf("%s: %v\n%s", name, err, out)
		}
	}
}
