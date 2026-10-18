// @(#)quiet / go / internal/port/base16/yaml.go / 2026-10-18

package base16

import (
	"fmt"
	"io"
	"strconv"
)

// EncodeYAML writes a Base16 scheme in the current Tinted Theming common
// scheme format.
//
// A dedicated YAML dependency would be disproportionate here: the document
// has a fixed schema, scalar metadata, and sixteen deterministic color keys.
func EncodeYAML(
	w io.Writer,
	s Scheme,
) error {
	if _, err := fmt.Fprintf(
		w,
		"system: %s\n"+
			"name: %s\n"+
			"slug: %s\n"+
			"author: %s\n"+
			"description: %s\n"+
			"variant: %s\n"+
			"palette:\n",
		strconv.Quote("base16"),
		strconv.Quote(s.Name),
		strconv.Quote(s.Slug),
		strconv.Quote(s.Author),
		strconv.Quote(s.Description),
		strconv.Quote(s.Variant),
	); err != nil {
		return fmt.Errorf(
			"write Base16 metadata: %w",
			err,
		)
	}

	for i, value := range s.Palette {
		hex, err := value.Hex()
		if err != nil {
			return fmt.Errorf(
				"Base16 base%02X: %w",
				i,
				err,
			)
		}

		if _, err := fmt.Fprintf(
			w,
			"  base%02X: %s\n",
			i,
			strconv.Quote(hex),
		); err != nil {
			return fmt.Errorf(
				"write Base16 base%02X: %w",
				i,
				err,
			)
		}
	}

	return nil
}
