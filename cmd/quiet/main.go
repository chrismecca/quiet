// @(#)quiet / go / cmd/quiet/main.go / 2026-10-16

// Command quiet is the reference command-line implementation for inspecting,
// deriving, and eventually validating Quiet Core color data.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
)

const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run contains command dispatch so behavior can be exercised without coupling
// command tests to process-wide os.Exit calls.
func run(args []string, stdout, stderr io.Writer) int {
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

		if err := inspect(stdout, args[1]); err != nil {
			fmt.Fprintf(stderr, "quiet: %v\n", err)
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
			fmt.Fprintf(stderr, "quiet: %v\n", err)
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
			fmt.Fprintf(stderr, "quiet: %v\n", err)
			return exitError
		}

		return exitOK

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

func inspect(w io.Writer, s string) error {
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
		color.Contrast(a, b),
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

	fmt.Fprintf(
		w,
		"%s\n",
		palette.Name,
	)

	fmt.Fprintf(
		w,
		"  selection  %s\n",
		derived.Selection.MustHex(),
	)

	fmt.Fprintf(
		w,
		"  callout    %s\n",
		derived.Callout.MustHex(),
	)

	fmt.Fprintf(
		w,
		"  subtle     %s\n",
		derived.Subtle.MustHex(),
	)

	return nil
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "usage:")
	fmt.Fprintln(w, "  quiet inspect HEX")
	fmt.Fprintln(w, "  quiet contrast HEX HEX")
	fmt.Fprintln(w, "  quiet derive PALETTE")
}
