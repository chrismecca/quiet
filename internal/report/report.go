// @(#)quiet / go / internal/report/report.go / 2026-10-24

package report

import (
	"fmt"

	"github.com/chrismecca/quiet/internal/core"
)

// Build computes every report table from the contract and both canonical
// adaptations. Table order is stable so generation is deterministic.
func Build(
	contract core.Contract,
	day core.Palette,
	night core.Palette,
) ([]Table, error) {
	if day.Mode != core.ModeDay || night.Mode != core.ModeNight {
		return nil, fmt.Errorf(
			"report palettes are %q and %q, want %q and %q",
			day.Mode,
			night.Mode,
			core.ModeDay,
			core.ModeNight,
		)
	}

	var tables []Table

	for _, p := range []core.Palette{day, night} {
		derived, err := derivedTable(p)
		if err != nil {
			return nil, fmt.Errorf("%s derived roles: %w", p.Name, err)
		}

		tables = append(tables, anchorTable(p), derived)
	}

	constraints, err := constraintTables(contract, day, night)
	if err != nil {
		return nil, fmt.Errorf("constraints: %w", err)
	}

	wash, err := washTables(day, night)
	if err != nil {
		return nil, fmt.Errorf("wash: %w", err)
	}

	tables = append(tables, constraints...)
	tables = append(tables, wash...)

	for _, p := range []core.Palette{day, night} {
		cvd, err := cvdTables(p)
		if err != nil {
			return nil, fmt.Errorf("%s CVD: %w", p.Name, err)
		}

		tables = append(tables, cvd...)
	}

	return tables, nil
}
