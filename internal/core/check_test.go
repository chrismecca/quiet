// @(#)quiet / go / internal/core/check_test.go / 2026-10-16

package core

import (
	"math"
	"strings"
	"testing"
)

func TestCanonicalPalettesConformToCore(t *testing.T) {
	contract, err := LoadContractFile(
		repoPath(
			t,
			"spec",
			"quiet-core.json",
		),
	)
	if err != nil {
		t.Fatalf(
			"LoadContractFile: %v",
			err,
		)
	}

	if got := len(contract.ContrastConstraints); got != 19 {
		t.Fatalf(
			"contrast constraint count = %d, want 19",
			got,
		)
	}

	for _, file := range []string{
		"quiet-day.json",
		"quiet-night.json",
	} {
		t.Run(file, func(t *testing.T) {
			palette, err := LoadPaletteFile(
				repoPath(
					t,
					"spec",
					file,
				),
			)
			if err != nil {
				t.Fatalf(
					"LoadPaletteFile: %v",
					err,
				)
			}

			report, err := Check(
				contract,
				palette,
			)
			if err != nil {
				t.Fatalf(
					"Check: %v",
					err,
				)
			}

			if !report.Passed() {
				for _, result := range report.Results {
					if !result.Pass {
						t.Errorf(
							"%s: %.6f < %.6f",
							result.ID,
							result.Actual,
							result.Minimum,
						)
					}
				}

				t.FailNow()
			}

			if got := report.PassedCount(); got != 19 {
				t.Fatalf(
					"passed constraints = %d, want 19",
					got,
				)
			}
		})
	}
}

// Night Rust on Structure is currently the closest canonical hard contrast
// relationship to the 4.5 floor. Freeze its independently measured value so a
// future numerical or palette change cannot silently move the boundary.
func TestNightRustStructureFixture(t *testing.T) {
	contract, err := LoadContractFile(
		repoPath(
			t,
			"spec",
			"quiet-core.json",
		),
	)
	if err != nil {
		t.Fatalf(
			"LoadContractFile: %v",
			err,
		)
	}

	palette, err := LoadPaletteFile(
		repoPath(
			t,
			"spec",
			"quiet-night.json",
		),
	)
	if err != nil {
		t.Fatalf(
			"LoadPaletteFile: %v",
			err,
		)
	}

	report, err := Check(
		contract,
		palette,
	)
	if err != nil {
		t.Fatalf(
			"Check: %v",
			err,
		)
	}

	for _, result := range report.Results {
		if result.ID != "rust-on-structure" {
			continue
		}

		if math.Abs(
			result.Actual-4.5180793187,
		) > 1e-9 {
			t.Fatalf(
				"rust-on-structure = %.10f, want %.10f",
				result.Actual,
				4.5180793187,
			)
		}

		return
	}

	t.Fatal(
		"rust-on-structure constraint not found",
	)
}

func TestDecodeContractRejectsUnknownRole(t *testing.T) {
	const document = `{
		"schema": "quiet.core/v1",
		"name": "Quiet Core",
		"status": "candidate",
		"contrast_constraints": [
			{
				"id": "bad-role",
				"foreground": "ultraviolet",
				"background": "field",
				"minimum": 4.5
			}
		]
	}`

	_, err := DecodeContract(
		strings.NewReader(document),
	)

	if err == nil {
		t.Fatal(
			"DecodeContract accepted an unknown role",
		)
	}
}

func TestDecodeContractRejectsDuplicateIDs(t *testing.T) {
	const document = `{
		"schema": "quiet.core/v1",
		"name": "Quiet Core",
		"status": "candidate",
		"contrast_constraints": [
			{
				"id": "duplicate",
				"foreground": "ink",
				"background": "field",
				"minimum": 7.0
			},
			{
				"id": "duplicate",
				"foreground": "muted",
				"background": "field",
				"minimum": 4.5
			}
		]
	}`

	_, err := DecodeContract(
		strings.NewReader(document),
	)

	if err == nil {
		t.Fatal(
			"DecodeContract accepted duplicate ids",
		)
	}
}

func TestDecodeContractRejectsImpossibleContrast(t *testing.T) {
	const document = `{
		"schema": "quiet.core/v1",
		"name": "Quiet Core",
		"status": "candidate",
		"contrast_constraints": [
			{
				"id": "impossible",
				"foreground": "ink",
				"background": "field",
				"minimum": 22.0
			}
		]
	}`

	_, err := DecodeContract(
		strings.NewReader(document),
	)

	if err == nil {
		t.Fatal(
			"DecodeContract accepted an impossible contrast ratio",
		)
	}
}
