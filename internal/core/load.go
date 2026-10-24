// @(#)quiet / go / internal/core/load.go / 2026-10-24

package core

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chrismecca/quiet/internal/color"
)

// paletteDocument is the wire representation of a canonical palette.
//
// The wire model remains separate from Palette so serialized strings and
// parsing concerns do not leak into the domain model.
type paletteDocument struct {
	Schema  string          `json:"schema"`
	Name    string          `json:"name"`
	Mode    string          `json:"mode"`
	Status  string          `json:"status"`
	Anchors anchorsDocument `json:"anchors"`
}

type anchorsDocument struct {
	Field     string `json:"field"`
	Surface   string `json:"surface"`
	Structure string `json:"structure"`
	Ink       string `json:"ink"`
	Muted     string `json:"muted"`
	Line      string `json:"line"`
	Mineral   string `json:"mineral"`
	Moss      string `json:"moss"`
	Ochre     string `json:"ochre"`
	Rust      string `json:"rust"`
}

// LoadPaletteFile reads and validates a canonical quiet palette document from
// path.
func LoadPaletteFile(path string) (Palette, error) {
	f, err := os.Open(path)
	if err != nil {
		return Palette{}, fmt.Errorf(
			"open palette %q: %w",
			path,
			err,
		)
	}
	defer f.Close()

	p, err := DecodePalette(f)
	if err != nil {
		return Palette{}, fmt.Errorf(
			"load palette %q: %w",
			path,
			err,
		)
	}

	return p, nil
}

// DecodePalette decodes one canonical quiet palette document.
//
// Unknown JSON fields are rejected. quiet's normative data should fail loudly
// when a field is misspelled or when a newer schema is passed to an older
// implementation.
func DecodePalette(r io.Reader) (Palette, error) {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	var doc paletteDocument

	if err := decoder.Decode(&doc); err != nil {
		return Palette{}, fmt.Errorf(
			"decode palette: %w",
			err,
		)
	}

	// A normative file contains exactly one JSON value. Reject concatenated
	// documents and other trailing data rather than silently ignoring it.
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return Palette{}, fmt.Errorf(
				"decode palette: multiple JSON values",
			)
		}

		return Palette{}, fmt.Errorf(
			"decode palette trailing data: %w",
			err,
		)
	}

	if doc.Schema != PaletteSchema {
		return Palette{}, fmt.Errorf(
			"unsupported palette schema %q; want %q",
			doc.Schema,
			PaletteSchema,
		)
	}

	if strings.TrimSpace(doc.Name) == "" {
		return Palette{}, fmt.Errorf(
			"palette name must not be empty",
		)
	}

	mode := Mode(doc.Mode)

	switch mode {
	case ModeDay, ModeNight:
	default:
		return Palette{}, fmt.Errorf(
			"unsupported palette mode %q",
			doc.Mode,
		)
	}

	status := Status(doc.Status)

	switch status {
	case StatusCandidate:
	default:
		return Palette{}, fmt.Errorf(
			"unsupported palette status %q",
			doc.Status,
		)
	}

	anchors, err := decodeAnchors(doc.Anchors)
	if err != nil {
		return Palette{}, err
	}

	return Palette{
		Schema:  doc.Schema,
		Name:    doc.Name,
		Mode:    mode,
		Status:  status,
		Anchors: anchors,
	}, nil
}

func decodeAnchors(doc anchorsDocument) (Anchors, error) {
	field, err := decodeAnchor("field", doc.Field)
	if err != nil {
		return Anchors{}, err
	}

	surface, err := decodeAnchor("surface", doc.Surface)
	if err != nil {
		return Anchors{}, err
	}

	structure, err := decodeAnchor("structure", doc.Structure)
	if err != nil {
		return Anchors{}, err
	}

	ink, err := decodeAnchor("ink", doc.Ink)
	if err != nil {
		return Anchors{}, err
	}

	muted, err := decodeAnchor("muted", doc.Muted)
	if err != nil {
		return Anchors{}, err
	}

	line, err := decodeAnchor("line", doc.Line)
	if err != nil {
		return Anchors{}, err
	}

	mineral, err := decodeAnchor("mineral", doc.Mineral)
	if err != nil {
		return Anchors{}, err
	}

	moss, err := decodeAnchor("moss", doc.Moss)
	if err != nil {
		return Anchors{}, err
	}

	ochre, err := decodeAnchor("ochre", doc.Ochre)
	if err != nil {
		return Anchors{}, err
	}

	rust, err := decodeAnchor("rust", doc.Rust)
	if err != nil {
		return Anchors{}, err
	}

	return Anchors{
		Field:     field,
		Surface:   surface,
		Structure: structure,
		Ink:       ink,
		Muted:     muted,
		Line:      line,
		Mineral:   mineral,
		Moss:      moss,
		Ochre:     ochre,
		Rust:      rust,
	}, nil
}

// decodeAnchor applies the stricter representation rules used by canonical
// quiet data. The generic color parser accepts reasonable user input; spec
// files must use one deterministic uppercase #RRGGBB spelling.
func decodeAnchor(name, value string) (color.SRGB, error) {
	c, err := color.ParseHex(value)
	if err != nil {
		return color.SRGB{}, fmt.Errorf(
			"anchor %q: %w",
			name,
			err,
		)
	}

	canonical, err := c.Hex()
	if err != nil {
		return color.SRGB{}, fmt.Errorf(
			"anchor %q: %w",
			name,
			err,
		)
	}

	if value != canonical {
		return color.SRGB{}, fmt.Errorf(
			"anchor %q must use canonical uppercase #RRGGBB; got %q, want %q",
			name,
			value,
			canonical,
		)
	}

	return c, nil
}
