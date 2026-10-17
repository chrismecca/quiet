// @(#)quiet / go / internal/port/terminal/json.go / 2026-10-16

package terminal

import (
	"encoding/json"
	"fmt"
	"io"
)

// document is the interchange representation consumed by host adapters such
// as the Terminal.app Swift serializer.
//
// It is intentionally generic ANSI-terminal data: the receiving adapter does
// not need knowledge of Quiet Core roles or derivation rules.
type document struct {
	Name       string   `json:"name"`
	Background string   `json:"background"`
	Text       string   `json:"text"`
	Bold       string   `json:"bold"`
	Selection  string   `json:"selection"`
	Cursor     string   `json:"cursor"`
	ANSI       []string `json:"ansi"`
}

// EncodeJSON writes a resolved terminal palette as deterministic, indented
// JSON suitable for inspection, versioned fixtures, or host serialization.
func EncodeJSON(
	w io.Writer,
	p Palette,
) error {
	doc := document{
		Name: p.Name,
	}

	var err error

	if doc.Background, err = p.Background.Hex(); err != nil {
		return fmt.Errorf("background: %w", err)
	}

	if doc.Text, err = p.Text.Hex(); err != nil {
		return fmt.Errorf("text: %w", err)
	}

	if doc.Bold, err = p.Bold.Hex(); err != nil {
		return fmt.Errorf("bold: %w", err)
	}

	if doc.Selection, err = p.Selection.Hex(); err != nil {
		return fmt.Errorf("selection: %w", err)
	}

	if doc.Cursor, err = p.Cursor.Hex(); err != nil {
		return fmt.Errorf("cursor: %w", err)
	}

	doc.ANSI = make(
		[]string,
		len(p.ANSI),
	)

	for i, value := range p.ANSI {
		doc.ANSI[i], err = value.Hex()
		if err != nil {
			return fmt.Errorf(
				"ANSI %d: %w",
				i,
				err,
			)
		}
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(doc); err != nil {
		return fmt.Errorf(
			"encode terminal palette: %w",
			err,
		)
	}

	return nil
}
