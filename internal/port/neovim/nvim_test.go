// @(#)quiet / go / internal/port/neovim/nvim_test.go / 2026-10-25

package neovim

import (
	"encoding/json"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

// The tests in this file run an installed Neovim and skip when none is on
// PATH. Set QUIET_NVIM to choose a binary.

// stateCues lists the non-color attributes that mark state, such as a
// problem or a misspelling. quiet keeps these where Neovim's defaults have
// them, and drops typographic emphasis such as bold and reverse.
var stateCues = []string{
	"underline",
	"undercurl",
	"underdouble",
	"underdotted",
	"underdashed",
	"strikethrough",
}

func TestSchemesLoadInNeovim(t *testing.T) {
	nvim := nvimPath(t)

	for _, name := range []string{"quiet-day", "quiet-night"} {
		out, err := exec.Command(
			nvim, "--clean", "--headless",
			"+luafile "+repoPath(t, "ports", "neovim", name+".lua"),
			"+lua assert(vim.g.colors_name == '"+name+"')",
			"+lua assert(vim.g.terminal_color_15 ~= nil)",
			"+qa!",
		).CombinedOutput()
		if err != nil || len(strings.TrimSpace(string(out))) > 0 {
			t.Errorf("%s: %v\n%s", name, err, out)
		}
	}
}

// TestCuesMatchNeovimDefaults checks that every group quiet defines keeps
// exactly the state cues Neovim's own defaults give it.
func TestCuesMatchNeovimDefaults(t *testing.T) {
	nvim := nvimPath(t)

	var all []highlight
	all = append(all, baseHighlights()...)
	all = append(all, treeSitterHighlights()...)
	all = append(all, diagnosticHighlights()...)
	all = append(all, lspHighlights()...)

	var groups []string
	for _, h := range all {
		if h.link == "" {
			groups = append(groups, h.group)
		}
	}

	encoded, err := json.Marshal(groups)
	if err != nil {
		t.Fatal(err)
	}

	script := `local out = {}
for _, g in ipairs(vim.json.decode(vim.env.QUIET_GROUPS)) do
  out[g] = vim.api.nvim_get_hl(0, { name = g, link = false })
end
io.stdout:write(vim.json.encode(out))`

	cmd := exec.Command(nvim, "--clean", "--headless", "+lua "+script, "+qa!")
	cmd.Env = append(os.Environ(), "QUIET_GROUPS="+string(encoded))

	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}

	var defaults map[string]map[string]any
	if err := json.Unmarshal(out, &defaults); err != nil {
		t.Fatalf("decode Neovim defaults: %v\n%s", err, out)
	}

	for _, h := range all {
		if h.link != "" {
			continue
		}

		var want []string
		for _, cue := range stateCues {
			if v, ok := defaults[h.group][cue].(bool); ok && v {
				want = append(want, cue)
			}
		}

		var got []string
		if h.cue != "" {
			got = append(got, h.cue)
		}

		if !slices.Equal(got, want) {
			t.Errorf("%s keeps cues %v, Neovim's default has %v", h.group, got, want)
		}
	}
}

func nvimPath(t *testing.T) string {
	t.Helper()

	if path := os.Getenv("QUIET_NVIM"); path != "" {
		return path
	}

	path, err := exec.LookPath("nvim")
	if err != nil {
		t.Skip("no Neovim installation found")
	}

	return path
}
