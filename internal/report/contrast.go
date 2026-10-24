// @(#)quiet / go / internal/report/contrast.go / 2026-10-24

package report

import (
	"fmt"
	"math"
	"sort"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
)

// meaningfulHueChroma is the OKLCH chroma below which a hue angle is shown as
// a dash. It matches the Wash source threshold: below it, a numerically
// defined hue does not carry semantic meaning.
const meaningfulHueChroma = 0.020

// tightestCount is the number of lowest-margin constraints listed across both
// adaptations.
const tightestCount = 8

func anchorTable(p core.Palette) Table {
	t := orgTable{
		header: []string{"Anchor", "Swatch", "sRGB", "L", "C", "h", "Y"},
	}

	for _, a := range p.Anchors.All() {
		lch := a.Color.OKLab().OKLCH()

		hue := "—"
		if lch.C >= meaningfulHueChroma && !math.IsNaN(lch.H) {
			hue = fmt.Sprintf("%.1f", lch.H)
		}

		t.rows = append(t.rows, []string{
			prose(a.Name),
			swatch(a.Color),
			a.Color.MustHex(),
			fmt.Sprintf("%.4f", lch.L),
			fmt.Sprintf("%.4f", lch.C),
			hue,
			fmt.Sprintf("%.4f", color.RelativeLuminance(a.Color)),
		})
	}

	return t.render("anchors-" + string(p.Mode))
}

func derivedTable(p core.Palette) (Table, error) {
	derived, err := core.Derive(p)
	if err != nil {
		return Table{}, err
	}

	subtle, err := core.SubtleOver(
		p,
		p.Anchors.Field,
	)
	if err != nil {
		return Table{}, err
	}

	type entry struct {
		name  string
		value color.SRGB
		rule  string
	}

	entries := []entry{
		{"Selection", derived.Selection, "OKLab mix, Structure toward Mineral"},
		{"Callout", derived.Callout, "OKLCH adjustment of Surface"},
		{"Subtle over Field", subtle, "sRGB composite of Muted over Field"},
	}

	for _, family := range core.Families() {
		wash, err := derived.Wash.For(family)
		if err != nil {
			return Table{}, err
		}

		name := prose(string(family))

		entries = append(
			entries,
			entry{name + " wash", wash.Line, "OKLab displacement of Field"},
			entry{name + " wash emphasis", wash.Emphasis, "OKLab displacement of Field"},
		)
	}

	t := orgTable{
		header: []string{"Role", "Swatch", "sRGB", "Rule"},
	}

	for _, e := range entries {
		t.rows = append(t.rows, []string{
			e.name,
			swatch(e.value),
			e.value.MustHex(),
			e.rule,
		})
	}

	return t.render("derived-" + string(p.Mode)), nil
}

// constraintTables evaluates contract with core.Check, so the report and
// `quiet check` cannot disagree about any ratio.
func constraintTables(
	contract core.Contract,
	day core.Palette,
	night core.Palette,
) ([]Table, error) {
	dayReport, err := core.Check(contract, day)
	if err != nil {
		return nil, err
	}

	nightReport, err := core.Check(contract, night)
	if err != nil {
		return nil, err
	}

	all := orgTable{
		header: []string{
			"Constraint",
			"Minimum",
			day.Name,
			"Margin",
			night.Name,
			"Margin",
		},
	}

	type scored struct {
		palette string
		result  core.ContrastResult
	}

	var everything []scored

	for i, d := range dayReport.Results {
		n := nightReport.Results[i]

		all.rows = append(all.rows, []string{
			constraintLabel(d),
			ratio(d.Minimum),
			ratio(d.Actual),
			margin(d.Actual - d.Minimum),
			ratio(n.Actual),
			margin(n.Actual - n.Minimum),
		})

		everything = append(
			everything,
			scored{day.Name, d},
			scored{night.Name, n},
		)
	}

	// Stable ordering by margin, then by palette and contract order, keeps
	// the generated artifact deterministic when margins tie at print
	// precision.
	sort.SliceStable(everything, func(i, j int) bool {
		a := everything[i].result.Actual - everything[i].result.Minimum
		b := everything[j].result.Actual - everything[j].result.Minimum

		return a < b
	})

	tight := orgTable{
		header: []string{"Palette", "Constraint", "Ratio", "Minimum", "Margin"},
	}

	// Three decimals: at two, a margin of +0.002 prints as +0.00 and reads
	// as a failure at the boundary.
	for _, s := range everything[:tightestCount] {
		tight.rows = append(tight.rows, []string{
			s.palette,
			constraintLabel(s.result),
			fmt.Sprintf("%.3f", s.result.Actual),
			ratio(s.result.Minimum),
			fmt.Sprintf("%+.3f", s.result.Actual-s.result.Minimum),
		})
	}

	return []Table{
		all.render("constraints"),
		tight.render("constraints-tightest"),
	}, nil
}

func constraintLabel(r core.ContrastResult) string {
	return prose(string(r.Foreground)) + " on " + prose(string(r.Background))
}
