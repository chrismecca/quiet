// @(#)quiet / go / internal/core/contract.go / 2026-10-24

package core

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

const (
	// ContractSchema identifies the machine-readable Quiet Core contract
	// understood by this version of the reference implementation. It versions
	// the serialized format, not a public Quiet release.
	ContractSchema = "quiet.core/v1"
)

// Role names a color-bearing semantic role that may participate in a Quiet
// Core rule. Roles include both canonical anchors and deterministic derived
// colors.
type Role string

const (
	RoleField     Role = "field"
	RoleSurface   Role = "surface"
	RoleStructure Role = "structure"
	RoleInk       Role = "ink"
	RoleMuted     Role = "muted"
	RoleLine      Role = "line"
	RoleMineral   Role = "mineral"
	RoleMoss      Role = "moss"
	RoleOchre     Role = "ochre"
	RoleRust      Role = "rust"

	RoleSelection Role = "selection"
	RoleCallout   Role = "callout"

	// Wash roles are the canonical Field-grounded washes.
	RoleMineralWash         Role = "mineral-wash"
	RoleMineralWashEmphasis Role = "mineral-wash-emphasis"
	RoleMossWash            Role = "moss-wash"
	RoleMossWashEmphasis    Role = "moss-wash-emphasis"
	RoleOchreWash           Role = "ochre-wash"
	RoleOchreWashEmphasis   Role = "ochre-wash-emphasis"
	RoleRustWash            Role = "rust-wash"
	RoleRustWashEmphasis    Role = "rust-wash-emphasis"
)

// Contract is the machine-readable portion of the candidate Quiet Core
// specification.
//
// Human rationale belongs in the normative AsciiDoc specification. Contract
// contains only rules that the reference implementation can evaluate
// deterministically.
type Contract struct {
	Schema              string
	Name                string
	Status              Status
	ContrastConstraints []ContrastConstraint
}

// ContrastConstraint requires a minimum WCAG contrast ratio between two Quiet
// semantic roles.
type ContrastConstraint struct {
	ID         string
	Foreground Role
	Background Role
	Minimum    float64
}

type contractDocument struct {
	Schema              string                       `json:"schema"`
	Name                string                       `json:"name"`
	Status              string                       `json:"status"`
	ContrastConstraints []contrastConstraintDocument `json:"contrast_constraints"`
}

type contrastConstraintDocument struct {
	ID         string  `json:"id"`
	Foreground string  `json:"foreground"`
	Background string  `json:"background"`
	Minimum    float64 `json:"minimum"`
}

// LoadContractFile reads and validates a machine-readable Quiet Core contract
// from path.
func LoadContractFile(path string) (Contract, error) {
	f, err := os.Open(path)
	if err != nil {
		return Contract{}, fmt.Errorf(
			"open contract %q: %w",
			path,
			err,
		)
	}
	defer f.Close()

	contract, err := DecodeContract(f)
	if err != nil {
		return Contract{}, fmt.Errorf(
			"load contract %q: %w",
			path,
			err,
		)
	}

	return contract, nil
}

// DecodeContract decodes exactly one Quiet Core contract.
//
// Unknown fields are rejected so misspellings and unsupported schema
// extensions cannot silently weaken conformance checks.
func DecodeContract(r io.Reader) (Contract, error) {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	var doc contractDocument

	if err := decoder.Decode(&doc); err != nil {
		return Contract{}, fmt.Errorf(
			"decode contract: %w",
			err,
		)
	}

	// A normative contract consists of exactly one JSON value.
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return Contract{}, fmt.Errorf(
				"decode contract: multiple JSON values",
			)
		}

		return Contract{}, fmt.Errorf(
			"decode contract trailing data: %w",
			err,
		)
	}

	if doc.Schema != ContractSchema {
		return Contract{}, fmt.Errorf(
			"unsupported contract schema %q; want %q",
			doc.Schema,
			ContractSchema,
		)
	}

	if strings.TrimSpace(doc.Name) == "" {
		return Contract{}, fmt.Errorf(
			"contract name must not be empty",
		)
	}

	status := Status(doc.Status)

	switch status {
	case StatusCandidate:
	default:
		return Contract{}, fmt.Errorf(
			"unsupported contract status %q",
			doc.Status,
		)
	}

	if len(doc.ContrastConstraints) == 0 {
		return Contract{}, fmt.Errorf(
			"contract must define at least one contrast constraint",
		)
	}

	constraints := make(
		[]ContrastConstraint,
		0,
		len(doc.ContrastConstraints),
	)

	ids := make(map[string]struct{})

	for i, item := range doc.ContrastConstraints {
		constraint, err := decodeContrastConstraint(
			item,
		)
		if err != nil {
			return Contract{}, fmt.Errorf(
				"contrast constraint %d: %w",
				i,
				err,
			)
		}

		if _, exists := ids[constraint.ID]; exists {
			return Contract{}, fmt.Errorf(
				"duplicate contrast constraint id %q",
				constraint.ID,
			)
		}

		ids[constraint.ID] = struct{}{}

		constraints = append(
			constraints,
			constraint,
		)
	}

	return Contract{
		Schema:              doc.Schema,
		Name:                doc.Name,
		Status:              status,
		ContrastConstraints: constraints,
	}, nil
}

func decodeContrastConstraint(
	doc contrastConstraintDocument,
) (ContrastConstraint, error) {
	id := strings.TrimSpace(doc.ID)

	if id == "" {
		return ContrastConstraint{}, fmt.Errorf(
			"id must not be empty",
		)
	}

	foreground := Role(doc.Foreground)
	background := Role(doc.Background)

	if !knownRole(foreground) {
		return ContrastConstraint{}, fmt.Errorf(
			"unknown foreground role %q",
			doc.Foreground,
		)
	}

	if !knownRole(background) {
		return ContrastConstraint{}, fmt.Errorf(
			"unknown background role %q",
			doc.Background,
		)
	}

	if foreground == background {
		return ContrastConstraint{}, fmt.Errorf(
			"foreground and background must differ",
		)
	}

	// WCAG contrast ratios for opaque sRGB colors are bounded by 1:1 and
	// 21:1. Rejecting values outside that range catches malformed normative
	// data rather than producing an impossible contract.
	if math.IsNaN(doc.Minimum) ||
		doc.Minimum < 1 ||
		doc.Minimum > 21 {
		return ContrastConstraint{}, fmt.Errorf(
			"minimum contrast %.6f outside [1, 21]",
			doc.Minimum,
		)
	}

	return ContrastConstraint{
		ID:         id,
		Foreground: foreground,
		Background: background,
		Minimum:    doc.Minimum,
	}, nil
}

func knownRole(role Role) bool {
	switch role {
	case RoleField,
		RoleSurface,
		RoleStructure,
		RoleInk,
		RoleMuted,
		RoleLine,
		RoleMineral,
		RoleMoss,
		RoleOchre,
		RoleRust,
		RoleSelection,
		RoleCallout,
		RoleMineralWash,
		RoleMineralWashEmphasis,
		RoleMossWash,
		RoleMossWashEmphasis,
		RoleOchreWash,
		RoleOchreWashEmphasis,
		RoleRustWash,
		RoleRustWashEmphasis:
		return true

	default:
		return false
	}
}
