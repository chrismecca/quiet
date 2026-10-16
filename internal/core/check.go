// @(#)quiet / go / internal/core/check.go / 2026-10-16

package core

import (
	"fmt"

	"github.com/chrismecca/quiet/internal/color"
)

// ContrastResult records one evaluated contrast constraint.
type ContrastResult struct {
	ID         string
	Foreground Role
	Background Role
	Minimum    float64
	Actual     float64
	Pass       bool
}

// CheckReport contains the deterministic conformance results for one palette.
type CheckReport struct {
	Palette string
	Mode    Mode
	Results []ContrastResult
}

// Passed reports whether every evaluated constraint passed.
func (r CheckReport) Passed() bool {
	for _, result := range r.Results {
		if !result.Pass {
			return false
		}
	}

	return true
}

// PassedCount returns the number of successful constraints.
func (r CheckReport) PassedCount() int {
	count := 0

	for _, result := range r.Results {
		if result.Pass {
			count++
		}
	}

	return count
}

// Check evaluates contract against palette.
//
// Derived semantic colors are recomputed from the canonical anchors before any
// rules are evaluated. Conformance therefore covers both the stored palette and
// the derivation engine rather than trusting cached derived values.
func Check(
	contract Contract,
	palette Palette,
) (CheckReport, error) {
	if contract.Schema != ContractSchema {
		return CheckReport{}, fmt.Errorf(
			"unsupported contract schema %q",
			contract.Schema,
		)
	}

	derived, err := Derive(palette)
	if err != nil {
		return CheckReport{}, fmt.Errorf(
			"derive %q: %w",
			palette.Name,
			err,
		)
	}

	results := make(
		[]ContrastResult,
		0,
		len(contract.ContrastConstraints),
	)

	for _, constraint := range contract.ContrastConstraints {
		foreground, err := resolveRole(
			palette,
			derived,
			constraint.Foreground,
		)
		if err != nil {
			return CheckReport{}, err
		}

		background, err := resolveRole(
			palette,
			derived,
			constraint.Background,
		)
		if err != nil {
			return CheckReport{}, err
		}

		actual := color.Contrast(
			foreground,
			background,
		)

		results = append(
			results,
			ContrastResult{
				ID:         constraint.ID,
				Foreground: constraint.Foreground,
				Background: constraint.Background,
				Minimum:    constraint.Minimum,
				Actual:     actual,
				Pass:       actual >= constraint.Minimum,
			},
		)
	}

	return CheckReport{
		Palette: palette.Name,
		Mode:    palette.Mode,
		Results: results,
	}, nil
}

func resolveRole(
	palette Palette,
	derived Derived,
	role Role,
) (color.SRGB, error) {
	switch role {
	case RoleField:
		return palette.Anchors.Field, nil

	case RoleSurface:
		return palette.Anchors.Surface, nil

	case RoleStructure:
		return palette.Anchors.Structure, nil

	case RoleInk:
		return palette.Anchors.Ink, nil

	case RoleMuted:
		return palette.Anchors.Muted, nil

	case RoleLine:
		return palette.Anchors.Line, nil

	case RoleMineral:
		return palette.Anchors.Mineral, nil

	case RoleMoss:
		return palette.Anchors.Moss, nil

	case RoleOchre:
		return palette.Anchors.Ochre, nil

	case RoleRust:
		return palette.Anchors.Rust, nil

	case RoleSelection:
		return derived.Selection, nil

	case RoleCallout:
		return derived.Callout, nil

	default:
		return color.SRGB{}, fmt.Errorf(
			"resolve unknown role %q",
			role,
		)
	}
}
