// @(#)quiet / go / internal/report/cvd.go / 2026-10-24

package report

import (
	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
)

func cvdHeader(first string) []string {
	header := []string{first, "Normal"}

	for _, d := range color.Deficiencies() {
		header = append(header, prose(string(d)))
	}

	return header
}

// simulate returns c under each deficiency, preceded by c itself.
func simulate(c color.SRGB) ([]color.SRGB, error) {
	out := []color.SRGB{c}

	for _, d := range color.Deficiencies() {
		s, err := color.SimulateCVD(c, d)
		if err != nil {
			return nil, err
		}

		out = append(out, s)
	}

	return out, nil
}

// pairTable reports OKLab distance between every pair of the given colors,
// with normal vision and under each simulated deficiency. Both colors of a
// pair are simulated, as a viewer with the deficiency would see both.
func pairTable(
	name string,
	labels []string,
	colors []color.SRGB,
) (Table, error) {
	simulated := make([][]color.SRGB, len(colors))

	for i, c := range colors {
		s, err := simulate(c)
		if err != nil {
			return Table{}, err
		}

		simulated[i] = s
	}

	t := orgTable{header: cvdHeader("Pair")}

	for i := range colors {
		for j := i + 1; j < len(colors); j++ {
			row := []string{labels[i] + " / " + labels[j]}

			for k := range simulated[i] {
				row = append(row, distance(color.DeltaEOK(
					simulated[i][k],
					simulated[j][k],
				)))
			}

			t.rows = append(t.rows, row)
		}
	}

	return t.render(name), nil
}

func cvdTables(p core.Palette) ([]Table, error) {
	derived, err := core.Derive(p)
	if err != nil {
		return nil, err
	}

	var (
		names   []string
		anchors []color.SRGB
		washes  []color.SRGB
	)

	for _, f := range families(p) {
		wash, err := derived.Wash.For(f.family)
		if err != nil {
			return nil, err
		}

		names = append(names, f.name)
		anchors = append(anchors, f.anchor)
		washes = append(washes, wash.Line)
	}

	mode := string(p.Mode)

	anchorPairs, err := pairTable("cvd-anchors-"+mode, names, anchors)
	if err != nil {
		return nil, err
	}

	washPairs, err := pairTable("cvd-washes-"+mode, names, washes)
	if err != nil {
		return nil, err
	}

	field, err := simulate(p.Anchors.Field)
	if err != nil {
		return nil, err
	}

	contrast := orgTable{header: cvdHeader("Anchor on Field")}

	for i, c := range anchors {
		s, err := simulate(c)
		if err != nil {
			return nil, err
		}

		row := []string{names[i]}

		for k := range s {
			row = append(row, ratio(color.Contrast(s[k], field[k])))
		}

		contrast.rows = append(contrast.rows, row)
	}

	return []Table{
		anchorPairs,
		washPairs,
		contrast.render("cvd-contrast-" + mode),
	}, nil
}
