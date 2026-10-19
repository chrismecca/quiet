// @(#)quiet / go / cmd/quiet/main.go / 2026-10-18

// Command quiet is the reference command-line implementation for inspecting,
// deriving, validating, and porting Quiet Core color data.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"

	"github.com/chrismecca/quiet/internal/port/base16"
	"github.com/chrismecca/quiet/internal/port/base24"
	"github.com/chrismecca/quiet/internal/port/machine"
	"github.com/chrismecca/quiet/internal/port/web"

	"github.com/chrismecca/quiet/internal/port/x11"

	"github.com/chrismecca/quiet/internal/port/terminal"
)

const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

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
			fmt.Fprintf(
				stderr,
				"quiet: %v\n",
				err,
			)
			return exitError
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
			fmt.Fprintf(
				stderr,
				"quiet: %v\n",
				err,
			)
			return exitError
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
			fmt.Fprintf(
				stderr,
				"quiet: %v\n",
				err,
			)
			return exitError
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
			fmt.Fprintf(
				stderr,
				"quiet: %v\n",
				err,
			)
			return exitError
		}

		if !passed {
			return exitError
		}

		return exitOK

	case "port":
		if len(args) < 2 {
			usage(stderr)
			return exitUsage
		}

		switch args[1] {
		case "base16":
			if len(args) != 3 {
				usage(stderr)
				return exitUsage
			}

			if err := portBase16(
				stdout,
				args[2],
			); err != nil {
				fmt.Fprintf(
					stderr,
					"quiet: %v\n",
					err,
				)
				return exitError
			}

			return exitOK

		case "base24":
			if len(args) != 3 {
				usage(stderr)
				return exitUsage
			}

			if err := portBase24(
				stdout,
				args[2],
			); err != nil {
				fmt.Fprintf(
					stderr,
					"quiet: %v\n",
					err,
				)
				return exitError
			}

			return exitOK

		case "machine":
			if len(args) != 4 {
				usage(stderr)
				return exitUsage
			}

			if err := portMachine(
				stdout,
				args[2],
				args[3],
			); err != nil {
				fmt.Fprintf(
					stderr,
					"quiet: %v\n",
					err,
				)
				return exitError
			}

			return exitOK

		case "web":
			if len(args) != 4 {
				usage(stderr)
				return exitUsage
			}

			if err := portWeb(
				stdout,
				args[2],
				args[3],
			); err != nil {
				fmt.Fprintf(
					stderr,
					"quiet: %v\n",
					err,
				)
				return exitError
			}

			return exitOK

		case "x11":
			if len(args) != 3 {
				usage(stderr)
				return exitUsage
			}

			if err := portX11(
				stdout,
				args[2],
			); err != nil {
				fmt.Fprintf(
					stderr,
					"quiet: %v\n",
					err,
				)
				return exitError
			}

			return exitOK

		case "terminal":
			if len(args) != 3 {
				usage(stderr)
				return exitUsage
			}

			if err := portTerminal(
				stdout,
				args[2],
			); err != nil {
				fmt.Fprintf(
					stderr,
					"quiet: %v\n",
					err,
				)
				return exitError
			}

			return exitOK

		default:
			fmt.Fprintf(
				stderr,
				"quiet: unknown port %q\n",
				args[1],
			)
			usage(stderr)
			return exitUsage
		}

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
		"  selection     %s\n",
		derived.Selection.MustHex(),
	)

	fmt.Fprintf(
		w,
		"  callout       %s\n",
		derived.Callout.MustHex(),
	)

	fmt.Fprintf(
		w,
		"  subtle/field  %s\n",
		subtleOnField.MustHex(),
	)

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

func usage(w io.Writer) {
	fmt.Fprintln(
		w,
		"usage:",
	)

	fmt.Fprintln(
		w,
		"  quiet inspect HEX",
	)

	fmt.Fprintln(
		w,
		"  quiet contrast HEX HEX",
	)

	fmt.Fprintln(
		w,
		"  quiet derive PALETTE",
	)

	fmt.Fprintln(
		w,
		"  quiet check CONTRACT PALETTE [PALETTE ...]",
	)

	fmt.Fprintln(
		w,
		"  quiet port base16 PALETTE",
	)

	fmt.Fprintln(
		w,
		"  quiet port base24 PALETTE",
	)

	fmt.Fprintln(
		w,
		"  quiet port machine DAY_PALETTE NIGHT_PALETTE",
	)

	fmt.Fprintln(
		w,
		"  quiet port web DAY_PALETTE NIGHT_PALETTE",
	)

	fmt.Fprintln(
		w,
		"  quiet port x11 PALETTE",
	)

	fmt.Fprintln(
		w,
		"  quiet port terminal PALETTE",
	)
}
