// @(#)quiet / go / cmd/quiet/main.go / 2026-10-26

// Command quiet is the reference command-line implementation for inspecting,
// deriving, validating, and porting quiet-core color data.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
	"github.com/chrismecca/quiet/internal/port/base16"
	"github.com/chrismecca/quiet/internal/port/base24"
	"github.com/chrismecca/quiet/internal/port/chrome"
	"github.com/chrismecca/quiet/internal/port/delta"
	"github.com/chrismecca/quiet/internal/port/dircolors"
	"github.com/chrismecca/quiet/internal/port/emacs"
	"github.com/chrismecca/quiet/internal/port/firefox"
	"github.com/chrismecca/quiet/internal/port/fzf"
	"github.com/chrismecca/quiet/internal/port/ghostty"
	"github.com/chrismecca/quiet/internal/port/helix"
	"github.com/chrismecca/quiet/internal/port/iterm2"
	"github.com/chrismecca/quiet/internal/port/jetbrains"
	"github.com/chrismecca/quiet/internal/port/kitty"
	"github.com/chrismecca/quiet/internal/port/machine"
	"github.com/chrismecca/quiet/internal/port/neovim"
	"github.com/chrismecca/quiet/internal/port/terminal"
	"github.com/chrismecca/quiet/internal/port/tmtheme"
	"github.com/chrismecca/quiet/internal/port/tmux"
	"github.com/chrismecca/quiet/internal/port/vim"
	"github.com/chrismecca/quiet/internal/port/vscode"
	"github.com/chrismecca/quiet/internal/port/web"
	"github.com/chrismecca/quiet/internal/port/wezterm"
	"github.com/chrismecca/quiet/internal/port/x11"
	"github.com/chrismecca/quiet/internal/port/xcode"
	"github.com/chrismecca/quiet/internal/port/zed"
	"github.com/chrismecca/quiet/internal/release"
	"github.com/chrismecca/quiet/internal/report"
)

const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

type portCommand struct {
	name  string
	usage string
	argc  int
	run   func(io.Writer, []string) error
}

// portCommands is the single canonical ordering for port dispatch and usage.
// Keep this list alphabetic. Imports and helper functions mirror the same order.
var portCommands = []portCommand{
	{
		name:  "base16",
		usage: "base16 PALETTE",
		argc:  1,
		run: func(w io.Writer, args []string) error {
			return portBase16(w, args[0])
		},
	},
	{
		name:  "base24",
		usage: "base24 PALETTE",
		argc:  1,
		run: func(w io.Writer, args []string) error {
			return portBase24(w, args[0])
		},
	},
	{
		name:  "chrome",
		usage: "chrome PALETTE",
		argc:  1,
		run: func(w io.Writer, args []string) error {
			return portChrome(w, args[0])
		},
	},
	{
		name:  "delta",
		usage: "delta DAY_PALETTE NIGHT_PALETTE",
		argc:  2,
		run: func(w io.Writer, args []string) error {
			return portDelta(w, args[0], args[1])
		},
	},
	{
		name:  "dircolors",
		usage: "dircolors",
		argc:  0,
		run: func(w io.Writer, _ []string) error {
			return portDircolors(w)
		},
	},
	{
		name:  "emacs",
		usage: "emacs PALETTE",
		argc:  1,
		run: func(w io.Writer, args []string) error {
			return portEmacs(w, args[0])
		},
	},
	{
		name:  "firefox",
		usage: "firefox PALETTE",
		argc:  1,
		run: func(w io.Writer, args []string) error {
			return portFirefox(w, args[0])
		},
	},
	{
		name:  "fzf",
		usage: "fzf PALETTE",
		argc:  1,
		run: func(w io.Writer, args []string) error {
			return portFzf(w, args[0])
		},
	},
	{
		name:  "ghostty",
		usage: "ghostty PALETTE",
		argc:  1,
		run: func(w io.Writer, args []string) error {
			return portGhostty(w, args[0])
		},
	},
	{
		name:  "helix",
		usage: "helix PALETTE",
		argc:  1,
		run: func(w io.Writer, args []string) error {
			return portHelix(w, args[0])
		},
	},
	{
		name:  "iterm2",
		usage: "iterm2 DAY NIGHT",
		argc:  2,
		run: func(w io.Writer, args []string) error {
			return portITerm2(w, args[0], args[1])
		},
	}, {
		name:  "iterm2-variant",
		usage: "iterm2-variant PALETTE",
		argc:  1,
		run: func(w io.Writer, args []string) error {
			return portITerm2Variant(w, args[0])
		},
	},

	{
		name:  "jetbrains-plugin",
		usage: "jetbrains-plugin DAY_PALETTE NIGHT_PALETTE",
		argc:  2,
		run: func(w io.Writer, args []string) error {
			return portJetBrainsPlugin(w, args[0], args[1])
		},
	},
	{
		name:  "jetbrains-scheme",
		usage: "jetbrains-scheme PALETTE",
		argc:  1,
		run: func(w io.Writer, args []string) error {
			return portJetBrainsScheme(w, args[0])
		},
	},
	{
		name:  "jetbrains-theme",
		usage: "jetbrains-theme PALETTE",
		argc:  1,
		run: func(w io.Writer, args []string) error {
			return portJetBrainsTheme(w, args[0])
		},
	},
	{
		name:  "kitty",
		usage: "kitty PALETTE",
		argc:  1,
		run: func(w io.Writer, args []string) error {
			return portKitty(w, args[0])
		},
	},
	{
		name:  "machine",
		usage: "machine DAY_PALETTE NIGHT_PALETTE",
		argc:  2,
		run: func(w io.Writer, args []string) error {
			return portMachine(w, args[0], args[1])
		},
	},
	{
		name:  "neovim",
		usage: "neovim PALETTE",
		argc:  1,
		run: func(w io.Writer, args []string) error {
			return portNeovim(w, args[0])
		},
	},
	{
		name:  "terminal",
		usage: "terminal PALETTE",
		argc:  1,
		run: func(w io.Writer, args []string) error {
			return portTerminal(w, args[0])
		},
	},
	{
		name:  "tmtheme",
		usage: "tmtheme PALETTE",
		argc:  1,
		run: func(w io.Writer, args []string) error {
			return portTmTheme(w, args[0])
		},
	},
	{
		name:  "tmux",
		usage: "tmux PALETTE",
		argc:  1,
		run: func(w io.Writer, args []string) error {
			return portTmux(w, args[0])
		},
	},
	{
		name:  "vim",
		usage: "vim PALETTE",
		argc:  1,
		run: func(w io.Writer, args []string) error {
			return portVim(w, args[0])
		},
	},
	{
		name:  "vscode",
		usage: "vscode PALETTE",
		argc:  1,
		run: func(w io.Writer, args []string) error {
			return portVSCode(w, args[0])
		},
	},
	{
		name:  "web",
		usage: "web DAY_PALETTE NIGHT_PALETTE",
		argc:  2,
		run: func(w io.Writer, args []string) error {
			return portWeb(w, args[0], args[1])
		},
	},
	{
		name:  "wezterm",
		usage: "wezterm PALETTE",
		argc:  1,
		run: func(w io.Writer, args []string) error {
			return portWezTerm(w, args[0])
		},
	},
	{
		name:  "x11",
		usage: "x11 PALETTE",
		argc:  1,
		run: func(w io.Writer, args []string) error {
			return portX11(w, args[0])
		},
	},
	{
		name:  "xcode",
		usage: "xcode PALETTE",
		argc:  1,
		run: func(w io.Writer, args []string) error {
			return portXcode(w, args[0])
		},
	},
	{
		name:  "zed",
		usage: "zed DAY_PALETTE NIGHT_PALETTE",
		argc:  2,
		run: func(w io.Writer, args []string) error {
			return portZed(w, args[0], args[1])
		},
	},
}

func main() {
	os.Exit(
		run(
			os.Args[1:],
			os.Stdout,
			os.Stderr,
		),
	)
}

// run contains command dispatch so behavior can be exercised without coupling
// command tests to process-wide os.Exit calls.
func run(
	args []string,
	stdout io.Writer,
	stderr io.Writer,
) int {
	if len(args) == 0 {
		usage(stderr)
		return exitUsage
	}

	switch args[0] {
	case "help", "-h", "--help":
		usage(stdout)
		return exitOK

	case "inspect":
		if len(args) != 2 {
			usage(stderr)
			return exitUsage
		}

		if err := inspect(
			stdout,
			args[1],
		); err != nil {
			return commandError(stderr, err)
		}

		return exitOK

	case "contrast":
		if len(args) != 3 {
			usage(stderr)
			return exitUsage
		}

		if err := printContrast(
			stdout,
			args[1],
			args[2],
		); err != nil {
			return commandError(stderr, err)
		}

		return exitOK

	case "derive":
		if len(args) != 2 {
			usage(stderr)
			return exitUsage
		}

		if err := derive(
			stdout,
			args[1],
		); err != nil {
			return commandError(stderr, err)
		}

		return exitOK

	case "check":
		if len(args) < 3 {
			usage(stderr)
			return exitUsage
		}

		passed, err := check(
			stdout,
			args[1],
			args[2:],
		)
		if err != nil {
			return commandError(stderr, err)
		}

		if !passed {
			return exitError
		}

		return exitOK

	case "report":
		if len(args) != 5 {
			usage(stderr)
			return exitUsage
		}

		if err := writeReport(
			args[1],
			args[2],
			args[3],
			args[4],
		); err != nil {
			return commandError(stderr, err)
		}

		return exitOK

	case "version":
		if len(args) != 1 {
			usage(stderr)
			return exitUsage
		}

		fmt.Fprintln(stdout, release.Version)
		return exitOK

	case "release-notes":
		if len(args) != 3 {
			usage(stderr)
			return exitUsage
		}

		if err := releaseNotes(
			stdout,
			args[1],
			args[2],
		); err != nil {
			return commandError(stderr, err)
		}

		return exitOK

	case "port":
		return runPort(
			args[1:],
			stdout,
			stderr,
		)

	default:
		fmt.Fprintf(
			stderr,
			"quiet: unknown command %q\n",
			args[0],
		)
		usage(stderr)
		return exitUsage
	}
}

func runPort(
	args []string,
	stdout io.Writer,
	stderr io.Writer,
) int {
	if len(args) == 0 {
		usage(stderr)
		return exitUsage
	}

	for _, command := range portCommands {
		if args[0] != command.name {
			continue
		}

		portArgs := args[1:]
		if len(portArgs) != command.argc {
			usage(stderr)
			return exitUsage
		}

		if err := command.run(
			stdout,
			portArgs,
		); err != nil {
			return commandError(stderr, err)
		}

		return exitOK
	}

	fmt.Fprintf(
		stderr,
		"quiet: unknown port %q\n",
		args[0],
	)
	usage(stderr)
	return exitUsage
}

func commandError(
	stderr io.Writer,
	err error,
) int {
	fmt.Fprintf(
		stderr,
		"quiet: %v\n",
		err,
	)

	return exitError
}

func portBase16(
	w io.Writer,
	path string,
) error {
	palette, err := core.LoadPaletteFile(path)
	if err != nil {
		return err
	}

	scheme, err := base16.Resolve(palette)
	if err != nil {
		return err
	}

	return base16.EncodeYAML(
		w,
		scheme,
	)
}

func portBase24(
	w io.Writer,
	path string,
) error {
	palette, err := core.LoadPaletteFile(path)
	if err != nil {
		return err
	}

	scheme, err := base24.Resolve(palette)
	if err != nil {
		return err
	}

	return base24.EncodeYAML(
		w,
		scheme,
	)
}

func portChrome(
	w io.Writer,
	path string,
) error {
	palette, err := core.LoadPaletteFile(path)
	if err != nil {
		return err
	}

	return chrome.Encode(
		w,
		palette,
	)
}

func portDelta(
	w io.Writer,
	dayPath string,
	nightPath string,
) error {
	day, err := core.LoadPaletteFile(dayPath)
	if err != nil {
		return fmt.Errorf(
			"load Day palette: %w",
			err,
		)
	}

	night, err := core.LoadPaletteFile(nightPath)
	if err != nil {
		return fmt.Errorf(
			"load Night palette: %w",
			err,
		)
	}

	return delta.Encode(
		w,
		day,
		night,
	)
}

func portDircolors(
	w io.Writer,
) error {
	return dircolors.Encode(w)
}

func portEmacs(
	w io.Writer,
	path string,
) error {
	palette, err := core.LoadPaletteFile(path)
	if err != nil {
		return err
	}

	return emacs.Encode(
		w,
		palette,
	)
}

func portFirefox(
	w io.Writer,
	path string,
) error {
	palette, err := core.LoadPaletteFile(path)
	if err != nil {
		return err
	}

	return firefox.Encode(
		w,
		palette,
	)
}

func portFzf(
	w io.Writer,
	path string,
) error {
	palette, err := core.LoadPaletteFile(path)
	if err != nil {
		return err
	}

	return fzf.Encode(
		w,
		palette,
	)
}

func portGhostty(
	w io.Writer,
	path string,
) error {
	palette, err := core.LoadPaletteFile(path)
	if err != nil {
		return err
	}

	resolved, err := terminal.Resolve(palette)
	if err != nil {
		return err
	}

	return ghostty.Encode(
		w,
		resolved,
	)
}

func portHelix(
	w io.Writer,
	path string,
) error {
	palette, err := core.LoadPaletteFile(path)
	if err != nil {
		return err
	}

	return helix.Encode(
		w,
		palette,
	)
}

func portITerm2(
	w io.Writer,
	dayPath string,
	nightPath string,
) error {
	dayCore, err := core.LoadPaletteFile(dayPath)
	if err != nil {
		return err
	}

	nightCore, err := core.LoadPaletteFile(nightPath)
	if err != nil {
		return err
	}

	day, err := terminal.Resolve(dayCore)
	if err != nil {
		return err
	}

	night, err := terminal.Resolve(nightCore)
	if err != nil {
		return err
	}

	return iterm2.Encode(
		w,
		day,
		night,
	)
}

func portITerm2Variant(
	w io.Writer,
	path string,
) error {
	palette, err := core.LoadPaletteFile(path)
	if err != nil {
		return err
	}

	resolved, err := terminal.Resolve(palette)
	if err != nil {
		return err
	}

	return iterm2.EncodeVariant(
		w,
		resolved,
	)
}

func portJetBrainsPlugin(
	w io.Writer,
	dayPath string,
	nightPath string,
) error {
	day, err := core.LoadPaletteFile(dayPath)
	if err != nil {
		return fmt.Errorf(
			"load Day palette: %w",
			err,
		)
	}

	night, err := core.LoadPaletteFile(nightPath)
	if err != nil {
		return fmt.Errorf(
			"load Night palette: %w",
			err,
		)
	}

	return jetbrains.EncodePlugin(
		w,
		day,
		night,
	)
}

func portJetBrainsScheme(
	w io.Writer,
	path string,
) error {
	palette, err := core.LoadPaletteFile(path)
	if err != nil {
		return err
	}

	return jetbrains.EncodeScheme(
		w,
		palette,
	)
}

func portJetBrainsTheme(
	w io.Writer,
	path string,
) error {
	palette, err := core.LoadPaletteFile(path)
	if err != nil {
		return err
	}

	return jetbrains.EncodeTheme(
		w,
		palette,
	)
}

func portKitty(
	w io.Writer,
	path string,
) error {
	palette, err := core.LoadPaletteFile(path)
	if err != nil {
		return err
	}

	resolved, err := terminal.Resolve(palette)
	if err != nil {
		return err
	}

	return kitty.Encode(
		w,
		resolved,
	)
}

func portMachine(
	w io.Writer,
	dayPath string,
	nightPath string,
) error {
	day, err := core.LoadPaletteFile(dayPath)
	if err != nil {
		return fmt.Errorf(
			"load Day palette: %w",
			err,
		)
	}

	night, err := core.LoadPaletteFile(nightPath)
	if err != nil {
		return fmt.Errorf(
			"load Night palette: %w",
			err,
		)
	}

	return machine.EncodeJSON(
		w,
		day,
		night,
	)
}

func portNeovim(
	w io.Writer,
	path string,
) error {
	palette, err := core.LoadPaletteFile(path)
	if err != nil {
		return err
	}

	return neovim.Encode(
		w,
		palette,
	)
}

func portTerminal(
	w io.Writer,
	path string,
) error {
	palette, err := core.LoadPaletteFile(path)
	if err != nil {
		return err
	}

	resolved, err := terminal.Resolve(palette)
	if err != nil {
		return err
	}

	return terminal.EncodeJSON(
		w,
		resolved,
	)
}

func portTmTheme(
	w io.Writer,
	path string,
) error {
	palette, err := core.LoadPaletteFile(path)
	if err != nil {
		return err
	}

	return tmtheme.Encode(
		w,
		palette,
	)
}

func portTmux(
	w io.Writer,
	path string,
) error {
	palette, err := core.LoadPaletteFile(path)
	if err != nil {
		return err
	}

	theme, err := tmux.Resolve(palette)
	if err != nil {
		return err
	}

	return tmux.Encode(
		w,
		theme,
	)
}

func portVim(
	w io.Writer,
	path string,
) error {
	palette, err := core.LoadPaletteFile(path)
	if err != nil {
		return err
	}

	theme, err := vim.Resolve(palette)
	if err != nil {
		return err
	}

	return vim.Encode(
		w,
		theme,
	)
}

func portVSCode(
	w io.Writer,
	path string,
) error {
	palette, err := core.LoadPaletteFile(path)
	if err != nil {
		return err
	}

	return vscode.Encode(
		w,
		palette,
	)
}

func portWeb(
	w io.Writer,
	dayPath string,
	nightPath string,
) error {
	day, err := core.LoadPaletteFile(dayPath)
	if err != nil {
		return fmt.Errorf(
			"load Day palette: %w",
			err,
		)
	}

	night, err := core.LoadPaletteFile(nightPath)
	if err != nil {
		return fmt.Errorf(
			"load Night palette: %w",
			err,
		)
	}

	return web.EncodeCSS(
		w,
		day,
		night,
	)
}

func portWezTerm(
	w io.Writer,
	path string,
) error {
	palette, err := core.LoadPaletteFile(path)
	if err != nil {
		return err
	}

	resolved, err := terminal.Resolve(palette)
	if err != nil {
		return err
	}

	return wezterm.Encode(
		w,
		resolved,
	)
}

func portX11(
	w io.Writer,
	path string,
) error {
	palette, err := core.LoadPaletteFile(path)
	if err != nil {
		return err
	}

	return x11.EncodeXResources(
		w,
		x11.Resolve(palette),
	)
}

func portXcode(
	w io.Writer,
	path string,
) error {
	palette, err := core.LoadPaletteFile(path)
	if err != nil {
		return err
	}

	theme, err := xcode.Resolve(palette)
	if err != nil {
		return err
	}

	return xcode.EncodeJSON(
		w,
		theme,
	)
}

func portZed(
	w io.Writer,
	dayPath string,
	nightPath string,
) error {
	day, err := core.LoadPaletteFile(dayPath)
	if err != nil {
		return fmt.Errorf(
			"load Day palette: %w",
			err,
		)
	}

	night, err := core.LoadPaletteFile(nightPath)
	if err != nil {
		return fmt.Errorf(
			"load Night palette: %w",
			err,
		)
	}

	return zed.Encode(
		w,
		day,
		night,
	)
}

func inspect(
	w io.Writer,
	s string,
) error {
	c, err := color.ParseHex(s)
	if err != nil {
		return err
	}

	hex, err := c.Hex()
	if err != nil {
		return err
	}

	p := c.OKLab().OKLCH()

	fmt.Fprintf(
		w,
		"hex:       %s\n",
		hex,
	)

	fmt.Fprintf(
		w,
		"oklch:     L=%.6f C=%.6f h=%.3f°\n",
		p.L,
		p.C,
		p.H,
	)

	fmt.Fprintf(
		w,
		"luminance: %.6f\n",
		color.RelativeLuminance(c),
	)

	return nil
}

func printContrast(
	w io.Writer,
	aHex string,
	bHex string,
) error {
	a, err := color.ParseHex(aHex)
	if err != nil {
		return fmt.Errorf(
			"first color: %w",
			err,
		)
	}

	b, err := color.ParseHex(bHex)
	if err != nil {
		return fmt.Errorf(
			"second color: %w",
			err,
		)
	}

	fmt.Fprintf(
		w,
		"%.6f:1\n",
		color.Contrast(
			a,
			b,
		),
	)

	return nil
}

func derive(
	w io.Writer,
	path string,
) error {
	palette, err := core.LoadPaletteFile(path)
	if err != nil {
		return err
	}

	derived, err := core.Derive(palette)
	if err != nil {
		return err
	}

	subtleOnField, err := core.SubtleOver(
		palette,
		palette.Anchors.Field,
	)
	if err != nil {
		return err
	}

	fmt.Fprintf(
		w,
		"%s\n",
		palette.Name,
	)

	fmt.Fprintf(
		w,
		"  selection              %s\n",
		derived.Selection.MustHex(),
	)

	fmt.Fprintf(
		w,
		"  callout                %s\n",
		derived.Callout.MustHex(),
	)

	fmt.Fprintf(
		w,
		"  subtle/field           %s\n",
		subtleOnField.MustHex(),
	)

	for _, family := range core.Families() {
		wash, err := derived.Wash.For(family)
		if err != nil {
			return err
		}

		fmt.Fprintf(
			w,
			"  %-22s %s\n",
			string(family)+"-wash",
			wash.Line.MustHex(),
		)

		fmt.Fprintf(
			w,
			"  %-22s %s\n",
			string(family)+"-wash-emphasis",
			wash.Emphasis.MustHex(),
		)
	}

	return nil
}

func check(
	w io.Writer,
	contractPath string,
	palettePaths []string,
) (bool, error) {
	contract, err := core.LoadContractFile(
		contractPath,
	)
	if err != nil {
		return false, err
	}

	allPassed := true

	for _, path := range palettePaths {
		palette, err := core.LoadPaletteFile(path)
		if err != nil {
			return false, err
		}

		report, err := core.Check(
			contract,
			palette,
		)
		if err != nil {
			return false, err
		}

		passed := report.PassedCount()
		total := len(report.Results)

		if report.Passed() {
			fmt.Fprintf(
				w,
				"%s: PASS (%d/%d)\n",
				report.Palette,
				passed,
				total,
			)
			continue
		}

		allPassed = false

		fmt.Fprintf(
			w,
			"%s: FAIL (%d/%d)\n",
			report.Palette,
			passed,
			total,
		)

		for _, result := range report.Results {
			if result.Pass {
				continue
			}

			fmt.Fprintf(
				w,
				"  FAIL %s on %s: %.6f < %.6f [%s]\n",
				result.Foreground,
				result.Background,
				result.Actual,
				result.Minimum,
				result.ID,
			)
		}
	}

	return allPassed, nil
}

// writeReport regenerates every report table under dir. Each table is one
// Org file so authored report prose can include exactly the tables it cites.
func writeReport(
	dir string,
	contractPath string,
	dayPath string,
	nightPath string,
) error {
	contract, err := core.LoadContractFile(contractPath)
	if err != nil {
		return err
	}

	day, err := core.LoadPaletteFile(dayPath)
	if err != nil {
		return err
	}

	night, err := core.LoadPaletteFile(nightPath)
	if err != nil {
		return err
	}

	tables, err := report.Build(
		contract,
		day,
		night,
	)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf(
			"create report directory: %w",
			err,
		)
	}

	for _, t := range tables {
		path := filepath.Join(
			dir,
			t.Name+".org",
		)

		if err := os.WriteFile(
			path,
			[]byte(t.Org),
			0o644,
		); err != nil {
			return fmt.Errorf(
				"write report table: %w",
				err,
			)
		}
	}

	return nil
}

// releaseNotes writes the changelog section for version as the Markdown body
// of a GitHub release.
func releaseNotes(
	w io.Writer,
	changelogPath string,
	version string,
) error {
	data, err := os.ReadFile(changelogPath)
	if err != nil {
		return fmt.Errorf(
			"read changelog: %w",
			err,
		)
	}

	notes, err := release.Notes(
		string(data),
		version,
	)
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, notes)
	return err
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "usage:")
	fmt.Fprintln(w, "  quiet inspect HEX")
	fmt.Fprintln(w, "  quiet contrast HEX HEX")
	fmt.Fprintln(w, "  quiet derive PALETTE")
	fmt.Fprintln(w, "  quiet check CONTRACT PALETTE [PALETTE ...]")
	fmt.Fprintln(w, "  quiet report DIR CONTRACT DAY_PALETTE NIGHT_PALETTE")
	fmt.Fprintln(w, "  quiet release-notes CHANGELOG VERSION")
	fmt.Fprintln(w, "  quiet version")

	for _, command := range portCommands {
		fmt.Fprintf(
			w,
			"  quiet port %s\n",
			command.usage,
		)
	}
}
