// @(#)quiet / go / internal/report/wash.go / 2026-10-24

package report

import (
	"fmt"
	"math"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
)

// The rejected alternatives are evaluated at the parameters that motivated
// the Wash design. They are evidence, not part of quiet-core.
var interpolationAmounts = []float64{0.04, 0.06, 0.08, 0.10, 0.15, 0.20}

const (
	// fixedChroma is the absolute OKLCH chroma evaluated for the rejected
	// fixed-chroma alternative. It lies near the chosen line distance so the
	// two methods are compared at similar strength.
	fixedChroma = 0.025

	// gamutScanStep and gamutScanLimit bound the search for the largest
	// in-gamut line distance. The step is finer than the printed precision.
	gamutScanStep  = 0.0001
	gamutScanLimit = 0.2
)

type family struct {
	name   string
	family core.Family
	anchor color.SRGB
}

func families(p core.Palette) []family {
	return []family{
		{"Mineral", core.FamilyMineral, p.Anchors.Mineral},
		{"Moss", core.FamilyMoss, p.Anchors.Moss},
		{"Ochre", core.FamilyOchre, p.Anchors.Ochre},
		{"Rust", core.FamilyRust, p.Anchors.Rust},
	}
}

// readable lists the foregrounds that must reach 4.5:1 on a line wash. Ink
// is excluded because its 7:1 requirement is never the binding one here.
func readable(p core.Palette) []family {
	return append(
		[]family{{"Muted", "", p.Anchors.Muted}},
		families(p)...,
	)
}

// lowest returns the lowest 4.5-class foreground contrast on background,
// formatted with the foreground that produced it.
func lowest(
	p core.Palette,
	background color.SRGB,
) (float64, string) {
	worst := math.Inf(1)
	label := ""

	for _, fg := range readable(p) {
		if r := color.Contrast(fg.anchor, background); r < worst {
			worst = r
			label = fg.name
		}
	}

	return worst, label
}

func washTables(day, night core.Palette) ([]Table, error) {
	tables := []Table{
		interpolationTable(day, night),
		fixedChromaTable(day, night),
	}

	displacement, err := displacementTable(day, night)
	if err != nil {
		return nil, err
	}

	grounds, err := groundTable(day, night)
	if err != nil {
		return nil, err
	}

	return append(
		tables,
		displacement,
		gamutTable(day, night),
		grounds,
	), nil
}

// interpolationTable evaluates the rejected Selection-style derivation:
// mixing Field toward each family anchor in OKLab.
func interpolationTable(day, night core.Palette) Table {
	t := orgTable{
		header: []string{"Amount", day.Name + " lowest", night.Name + " lowest"},
	}

	cell := func(p core.Palette, amount float64) string {
		worst := math.Inf(1)
		label := ""

		for _, f := range families(p) {
			mixed := color.MixOKLab(p.Anchors.Field, f.anchor, amount)

			if r, fg := lowest(p, mixed); r < worst {
				worst = r
				label = fmt.Sprintf("%s on %s mix", fg, f.name)
			}
		}

		return ratio(worst) + " " + label
	}

	for _, amount := range interpolationAmounts {
		t.rows = append(t.rows, []string{
			fmt.Sprintf("%.2f", amount),
			cell(day, amount),
			cell(night, amount),
		})
	}

	return t.render("wash-interpolation")
}

// fixedChromaTable evaluates the rejected fixed-chroma derivation: Field
// lightness with an absolute chroma at each family hue. The spread of
// distances from Field is the defect.
func fixedChromaTable(day, night core.Palette) Table {
	t := orgTable{
		header: []string{"Family", day.Name + " ΔE", night.Name + " ΔE"},
	}

	cell := func(p core.Palette, f family) string {
		field := p.Anchors.Field.OKLab().OKLCH()

		tinted := color.OKLCH{
			L: field.L,
			C: fixedChroma,
			H: f.anchor.OKLab().OKLCH().H,
		}.OKLab().SRGB()

		return distance(color.DeltaEOK(tinted, p.Anchors.Field))
	}

	dayFamilies := families(day)
	nightFamilies := families(night)

	for i := range dayFamilies {
		t.rows = append(t.rows, []string{
			dayFamilies[i].name,
			cell(day, dayFamilies[i]),
			cell(night, nightFamilies[i]),
		})
	}

	return t.render("wash-fixed-chroma")
}

// displacementTable reports the chosen derivation: the canonical line washes.
func displacementTable(day, night core.Palette) (Table, error) {
	t := orgTable{
		header: []string{
			"Family",
			day.Name + " ΔE",
			day.Name + " lowest",
			night.Name + " ΔE",
			night.Name + " lowest",
		},
	}

	cells := func(p core.Palette, f family) ([]string, error) {
		wash, err := core.WashOver(p, f.family, core.WashLine, p.Anchors.Field)
		if err != nil {
			return nil, err
		}

		r, fg := lowest(p, wash)

		return []string{
			distance(color.DeltaEOK(wash, p.Anchors.Field)),
			ratio(r) + " " + fg,
		}, nil
	}

	dayFamilies := families(day)
	nightFamilies := families(night)

	for i := range dayFamilies {
		d, err := cells(day, dayFamilies[i])
		if err != nil {
			return Table{}, err
		}

		n, err := cells(night, nightFamilies[i])
		if err != nil {
			return Table{}, err
		}

		t.rows = append(
			t.rows,
			append(append([]string{dayFamilies[i].name}, d...), n...),
		)
	}

	return t.render("wash-displacement"), nil
}

// gamutTable reports the largest line distance over Field that stays inside
// sRGB for each family. The smallest value bounds any shared line distance.
func gamutTable(day, night core.Palette) Table {
	t := orgTable{
		header: []string{"Family", day.Name + " ceiling", night.Name + " ceiling"},
	}

	ceiling := func(p core.Palette, f family) string {
		hue := f.anchor.OKLab().OKLCH().H
		last := 0.0

		// Integer steps avoid accumulating floating-point error in d.
		for i := 1; float64(i)*gamutScanStep <= gamutScanLimit; i++ {
			d := float64(i) * gamutScanStep

			if !color.DisplaceOKLab(p.Anchors.Field, 0, d, hue).InGamut() {
				break
			}

			last = d
		}

		return fmt.Sprintf("%.4f", last)
	}

	dayFamilies := families(day)
	nightFamilies := families(night)

	for i := range dayFamilies {
		t.rows = append(t.rows, []string{
			dayFamilies[i].name,
			ceiling(day, dayFamilies[i]),
			ceiling(night, nightFamilies[i]),
		})
	}

	return t.render("wash-gamut")
}

// groundTable shows why Field is the canonical ground: the lowest 4.5-class
// foreground contrast over every family's line wash on each candidate ground.
func groundTable(day, night core.Palette) (Table, error) {
	t := orgTable{
		header: []string{"Ground", day.Name + " lowest", night.Name + " lowest"},
	}

	cell := func(p core.Palette, ground color.SRGB) (string, error) {
		worst := math.Inf(1)
		label := ""

		for _, f := range families(p) {
			wash, err := core.WashOver(p, f.family, core.WashLine, ground)
			if err != nil {
				return "", err
			}

			if r, fg := lowest(p, wash); r < worst {
				worst = r
				label = fmt.Sprintf("%s on %s wash", fg, f.name)
			}
		}

		return ratio(worst) + " " + label, nil
	}

	grounds := []struct {
		name  string
		day   color.SRGB
		night color.SRGB
	}{
		{"Field", day.Anchors.Field, night.Anchors.Field},
		{"Surface", day.Anchors.Surface, night.Anchors.Surface},
		{"Structure", day.Anchors.Structure, night.Anchors.Structure},
	}

	for _, g := range grounds {
		d, err := cell(day, g.day)
		if err != nil {
			return Table{}, err
		}

		n, err := cell(night, g.night)
		if err != nil {
			return Table{}, err
		}

		t.rows = append(t.rows, []string{g.name, d, n})
	}

	return t.render("wash-grounds"), nil
}
