// @(#)quiet / go / internal/port/machine/json.go / 2026-10-24

// Package machine exports quiet as a generic machine-readable semantic
// document.
package machine

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
)

type document struct {
	Schema     string             `json:"schema"`
	Reference  referenceDocument  `json:"reference"`
	Modes      modesDocument      `json:"modes"`
	Contextual contextualDocument `json:"contextual"`
}

type referenceDocument struct {
	ColorSpace   string `json:"color_space"`
	WhitePoint   string `json:"white_point"`
	DynamicRange string `json:"dynamic_range"`
}

type modesDocument struct {
	Day   modeDocument `json:"day"`
	Night modeDocument `json:"night"`
}

type modeDocument struct {
	Name    string          `json:"name"`
	Status  string          `json:"status"`
	Anchors anchorDocument  `json:"anchors"`
	Derived derivedDocument `json:"derived"`
}

type anchorDocument struct {
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

type derivedDocument struct {
	Selection string         `json:"selection"`
	Callout   string         `json:"callout"`
	Wash      washesDocument `json:"wash"`
}

// washesDocument contains the canonical Field-grounded washes. Washes over
// other grounds are different colors and are not exported.
type washesDocument struct {
	Mineral washDocument `json:"mineral"`
	Moss    washDocument `json:"moss"`
	Ochre   washDocument `json:"ochre"`
	Rust    washDocument `json:"rust"`
}

type washDocument struct {
	Line     string `json:"line"`
	Emphasis string `json:"emphasis"`
}

type contextualDocument struct {
	Subtle subtleDocument `json:"subtle"`
}

type subtleDocument struct {
	Foreground  string  `json:"foreground"`
	Opacity     float64 `json:"opacity"`
	Background  string  `json:"background"`
	Compositing string  `json:"compositing"`
}

// EncodeJSON writes the generic quiet machine-readable interchange document.
//
// The document contains canonical anchors and opaque derived roles for both
// adaptations. Contextual roles are represented as derivation metadata rather
// than flattened into universal opaque colors.
func EncodeJSON(
	w io.Writer,
	day core.Palette,
	night core.Palette,
) error {
	if got := fmt.Sprint(day.Mode); got != "day" {
		return fmt.Errorf(
			"first machine palette mode is %q, want %q",
			got,
			"day",
		)
	}

	if got := fmt.Sprint(night.Mode); got != "night" {
		return fmt.Errorf(
			"second machine palette mode is %q, want %q",
			got,
			"night",
		)
	}

	dayDocument, err := resolveMode(day)
	if err != nil {
		return fmt.Errorf(
			"resolve quiet-day machine document: %w",
			err,
		)
	}

	nightDocument, err := resolveMode(night)
	if err != nil {
		return fmt.Errorf(
			"resolve quiet-night machine document: %w",
			err,
		)
	}

	doc := document{
		Schema: "quiet.export/v1",

		Reference: referenceDocument{
			ColorSpace:   "sRGB",
			WhitePoint:   "D65",
			DynamicRange: "SDR",
		},

		Modes: modesDocument{
			Day:   dayDocument,
			Night: nightDocument,
		},

		Contextual: contextualDocument{
			Subtle: subtleDocument{
				Foreground:  "muted",
				Opacity:     0.55,
				Background:  "local",
				Compositing: "legacy-srgb",
			},
		},
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(doc); err != nil {
		return fmt.Errorf(
			"encode machine document: %w",
			err,
		)
	}

	return nil
}

func resolveMode(
	p core.Palette,
) (modeDocument, error) {
	derived, err := core.Derive(p)
	if err != nil {
		return modeDocument{}, err
	}

	anchors, err := resolveAnchors(p)
	if err != nil {
		return modeDocument{}, err
	}

	selection, err := derived.Selection.Hex()
	if err != nil {
		return modeDocument{}, fmt.Errorf(
			"selection: %w",
			err,
		)
	}

	callout, err := derived.Callout.Hex()
	if err != nil {
		return modeDocument{}, fmt.Errorf(
			"callout: %w",
			err,
		)
	}

	wash, err := resolveWashes(derived.Wash)
	if err != nil {
		return modeDocument{}, err
	}

	return modeDocument{
		Name:   p.Name,
		Status: fmt.Sprint(p.Status),

		Anchors: anchors,

		Derived: derivedDocument{
			Selection: selection,
			Callout:   callout,
			Wash:      wash,
		},
	}, nil
}

func resolveWashes(
	w core.Washes,
) (washesDocument, error) {
	var doc washesDocument

	targets := map[core.Family]*washDocument{
		core.FamilyMineral: &doc.Mineral,
		core.FamilyMoss:    &doc.Moss,
		core.FamilyOchre:   &doc.Ochre,
		core.FamilyRust:    &doc.Rust,
	}

	for _, family := range core.Families() {
		wash, err := w.For(family)
		if err != nil {
			return washesDocument{}, err
		}

		line, err := hex(
			string(family)+" wash",
			wash.Line,
		)
		if err != nil {
			return washesDocument{}, err
		}

		emphasis, err := hex(
			string(family)+" wash emphasis",
			wash.Emphasis,
		)
		if err != nil {
			return washesDocument{}, err
		}

		*targets[family] = washDocument{
			Line:     line,
			Emphasis: emphasis,
		}
	}

	return doc, nil
}

func resolveAnchors(
	p core.Palette,
) (anchorDocument, error) {
	field, err := hex(
		"field",
		p.Anchors.Field,
	)
	if err != nil {
		return anchorDocument{}, err
	}

	surface, err := hex(
		"surface",
		p.Anchors.Surface,
	)
	if err != nil {
		return anchorDocument{}, err
	}

	structure, err := hex(
		"structure",
		p.Anchors.Structure,
	)
	if err != nil {
		return anchorDocument{}, err
	}

	ink, err := hex(
		"ink",
		p.Anchors.Ink,
	)
	if err != nil {
		return anchorDocument{}, err
	}

	muted, err := hex(
		"muted",
		p.Anchors.Muted,
	)
	if err != nil {
		return anchorDocument{}, err
	}

	line, err := hex(
		"line",
		p.Anchors.Line,
	)
	if err != nil {
		return anchorDocument{}, err
	}

	mineral, err := hex(
		"mineral",
		p.Anchors.Mineral,
	)
	if err != nil {
		return anchorDocument{}, err
	}

	moss, err := hex(
		"moss",
		p.Anchors.Moss,
	)
	if err != nil {
		return anchorDocument{}, err
	}

	ochre, err := hex(
		"ochre",
		p.Anchors.Ochre,
	)
	if err != nil {
		return anchorDocument{}, err
	}

	rust, err := hex(
		"rust",
		p.Anchors.Rust,
	)
	if err != nil {
		return anchorDocument{}, err
	}

	return anchorDocument{
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

func hex(
	name string,
	value color.SRGB,
) (string, error) {
	result, err := value.Hex()
	if err != nil {
		return "", fmt.Errorf(
			"%s: %w",
			name,
			err,
		)
	}

	return result, nil
}
