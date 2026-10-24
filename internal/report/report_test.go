// @(#)quiet / go / internal/report/report_test.go / 2026-10-24

package report

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/chrismecca/quiet/internal/core"
)

// lineDistance mirrors the candidate Wash line distance in quiet-core. It is
// repeated here so the gamut test fails if either value moves alone.
const lineDistance = 0.020

func TestCanonicalTablesMatchCommittedData(t *testing.T) {
	tables := canonicalTables(t)

	want := make(map[string]bool)

	for _, table := range tables {
		want[table.Name+".org"] = true

		committed, err := os.ReadFile(
			repoPath(
				t,
				"reports",
				"data",
				table.Name+".org",
			),
		)
		if err != nil {
			t.Fatal(err)
		}

		if string(committed) != table.Org {
			t.Errorf(
				"generated report table differs from %q",
				table.Name+".org",
			)
		}
	}

	entries, err := os.ReadDir(
		repoPath(
			t,
			"reports",
			"data",
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if !want[entry.Name()] {
			t.Errorf(
				"stale report table %q is not generated",
				entry.Name(),
			)
		}
	}
}

func TestConstraintTableCoversContract(t *testing.T) {
	contract, err := core.LoadContractFile(
		repoPath(
			t,
			"spec",
			"quiet-core.json",
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	table := findTable(
		t,
		canonicalTables(t),
		"constraints",
	)

	rows := dataRows(table)

	if len(rows) != len(contract.ContrastConstraints) {
		t.Fatalf(
			"constraint rows = %d, want %d",
			len(rows),
			len(contract.ContrastConstraints),
		)
	}
}

// Every family's gamut ceiling must lie above the shared line distance, or
// Derive itself would fail for that family.
func TestWashGamutCeilingsExceedLineDistance(t *testing.T) {
	table := findTable(
		t,
		canonicalTables(t),
		"wash-gamut",
	)

	for _, row := range dataRows(table) {
		for _, cell := range row[1:] {
			ceiling, err := strconv.ParseFloat(cell, 64)
			if err != nil {
				t.Fatal(err)
			}

			if ceiling <= lineDistance {
				t.Errorf(
					"%s ceiling %.4f does not exceed line distance %.3f",
					row[0],
					ceiling,
					lineDistance,
				)
			}
		}
	}
}

func canonicalTables(t *testing.T) []Table {
	t.Helper()

	contract, err := core.LoadContractFile(
		repoPath(
			t,
			"spec",
			"quiet-core.json",
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	day, err := core.LoadPaletteFile(
		repoPath(
			t,
			"spec",
			"quiet-day.json",
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	night, err := core.LoadPaletteFile(
		repoPath(
			t,
			"spec",
			"quiet-night.json",
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	tables, err := Build(
		contract,
		day,
		night,
	)
	if err != nil {
		t.Fatal(err)
	}

	return tables
}

func findTable(
	t *testing.T,
	tables []Table,
	name string,
) Table {
	t.Helper()

	for _, table := range tables {
		if table.Name == name {
			return table
		}
	}

	t.Fatalf(
		"no report table %q",
		name,
	)

	return Table{}
}

// dataRows returns the cells of every body row, skipping the header row,
// the rule, and the trailing generated-file comment.
func dataRows(table Table) [][]string {
	var rows [][]string

	lines := strings.Split(
		strings.TrimSpace(table.Org),
		"\n",
	)

	for _, line := range lines[2 : len(lines)-1] {
		var cells []string

		for _, cell := range strings.Split(strings.Trim(line, "|"), "|") {
			cells = append(cells, strings.TrimSpace(cell))
		}

		rows = append(rows, cells)
	}

	return rows
}

func repoPath(
	t *testing.T,
	elem ...string,
) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal(
			"runtime.Caller could not locate test source",
		)
	}

	root := filepath.Clean(
		filepath.Join(
			filepath.Dir(file),
			"..",
			"..",
		),
	)

	return filepath.Join(
		append(
			[]string{root},
			elem...,
		)...,
	)
}
