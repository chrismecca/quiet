// @(#)quiet / go / internal/port/base24/yaml.go / 2026-10-18

package base24

import (
	"fmt"
	"io"
	"strconv"
)

// EncodeYAML writes a deterministic Base24 compatibility scheme.
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
		strconv.Quote("base24"),
		strconv.Quote(s.Name),
		strconv.Quote(s.Slug),
		strconv.Quote(s.Author),
		strconv.Quote(s.Description),
		strconv.Quote(s.Variant),
	); err != nil {
		return fmt.Errorf(
			"write Base24 metadata: %w",
			err,
		)
	}

	for i, value := range s.Palette {
		hex, err := value.Hex()
		if err != nil {
			return fmt.Errorf(
				"Base24 base%02X: %w",
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
				"write Base24 base%02X: %w",
				i,
				err,
			)
		}
	}

	return nil
}
