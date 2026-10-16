// @(#)quiet / go / internal/core/palette.go / 2026-10-16

// Package core defines Quiet's semantic palette model and the rules that
// operate on it.
//
// The color package knows color science but nothing about Quiet. Package core
// is the boundary at which generic color values acquire Quiet semantics.
package core

import "github.com/chrismecca/quiet/internal/color"

const (
	// PaletteSchema identifies the on-disk representation understood by this
	// version of the reference implementation. It is a data-format version,
	// not a Quiet Core release version.
	PaletteSchema = "quiet.palette/v1"

	// AnchorCount is the number of canonical colors in a Quiet palette.
	// Derived semantic colors are deliberately excluded.
	AnchorCount = 10
)

// Mode identifies the adaptation environment represented by a palette.
type Mode string

const (
	ModeDay   Mode = "day"
	ModeNight Mode = "night"
)

// Status describes the lifecycle state of a canonical palette document.
type Status string

const (
	// StatusCandidate marks palette data that is frozen for evaluation but has
	// not yet been published as part of a stable Quiet release.
	StatusCandidate Status = "candidate"
)

// Palette is one realization of Quiet Core.
//
// Anchors are the canonical inputs. Selection, Callout, Subtle, and future
// semantic display colors are derived from these anchors and therefore do not
// belong in this structure.
type Palette struct {
	Schema  string
	Name    string
	Mode    Mode
	Status  Status
	Anchors Anchors
}

// Anchors contains the ten canonical Quiet colors.
//
// The fields are explicit rather than map-based so missing, misspelled, and
// invented anchor names cannot silently enter the domain model.
type Anchors struct {
	Field     color.SRGB
	Surface   color.SRGB
	Structure color.SRGB
	Ink       color.SRGB
	Muted     color.SRGB
	Line      color.SRGB
	Mineral   color.SRGB
	Moss      color.SRGB
	Ochre     color.SRGB
	Rust      color.SRGB
}

// Anchor associates a canonical semantic name with its color value. It is
// primarily useful to validators and reports that need to traverse all anchors
// without discarding their semantic names.
type Anchor struct {
	Name  string
	Color color.SRGB
}

// All returns the canonical anchors in stable semantic order.
//
// The order is intentionally structural colors first and chromatic signal
// colors second. It is for deterministic reporting and iteration; it does not
// imply visual priority.
func (a Anchors) All() []Anchor {
	return []Anchor{
		{Name: "field", Color: a.Field},
		{Name: "surface", Color: a.Surface},
		{Name: "structure", Color: a.Structure},
		{Name: "ink", Color: a.Ink},
		{Name: "muted", Color: a.Muted},
		{Name: "line", Color: a.Line},
		{Name: "mineral", Color: a.Mineral},
		{Name: "moss", Color: a.Moss},
		{Name: "ochre", Color: a.Ochre},
		{Name: "rust", Color: a.Rust},
	}
}
