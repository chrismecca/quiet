// @(#)quiet / go / internal/port/firefox/firefox.go / 2026-10-25

// Package firefox maps quiet-core semantics onto a Firefox theme.
//
// A Firefox theme is a WebExtension whose manifest names browser chrome
// colors. The port colors Firefox's interface only; web content keeps its
// own colors.
package firefox

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
	"github.com/chrismecca/quiet/internal/release"
)

// minimumVersion is the first Firefox release with the color_scheme and
// content_color_scheme theme properties.
const minimumVersion = "106.0"

// role names a quiet color used by the Firefox mapping.
type role int

const (
	field role = iota
	surface
	ink
	muted
	line
	mineral
	selection
	callout
)

type entry struct {
	key  string
	role role
}

// mapping is the complete Firefox chrome policy. It follows the VS Code
// port: the tab strip is Surface, the selected tab and toolbar are Field and
// join the page, inputs and menus are Surface, and Selection and Callout mark
// selected and hovered items.
var mapping = []entry{
	// Tab strip.
	{"frame", surface},
	{"frame_inactive", surface},
	{"tab_background_text", muted},
	{"tab_selected", field},
	{"tab_text", ink},
	{"tab_line", mineral},
	{"tab_loading", mineral},

	// Toolbar.
	{"toolbar", field},
	{"toolbar_text", ink},
	{"icons", muted},
	{"icons_attention", mineral},
	{"button_background_hover", callout},
	{"button_background_active", selection},
	{"toolbar_top_separator", line},
	{"toolbar_bottom_separator", line},
	{"toolbar_vertical_separator", line},

	// Address bar.
	{"toolbar_field", surface},
	{"toolbar_field_text", ink},
	{"toolbar_field_border", line},
	{"toolbar_field_focus", surface},
	{"toolbar_field_text_focus", ink},
	{"toolbar_field_border_focus", mineral},
	{"toolbar_field_highlight", selection},
	{"toolbar_field_highlight_text", ink},

	// Menus and panels.
	{"popup", surface},
	{"popup_text", ink},
	{"popup_border", line},
	{"popup_highlight", selection},
	{"popup_highlight_text", ink},

	// Sidebar.
	{"sidebar", surface},
	{"sidebar_text", ink},
	{"sidebar_border", line},
	{"sidebar_highlight", selection},
	{"sidebar_highlight_text", ink},

	// New tab page.
	{"ntp_background", field},
	{"ntp_text", ink},
	{"ntp_card_background", surface},
}

type manifest struct {
	ManifestVersion int             `json:"manifest_version"`
	Name            string          `json:"name"`
	Version         string          `json:"version"`
	Description     string          `json:"description"`
	Browser         browserSettings `json:"browser_specific_settings"`
	Theme           theme           `json:"theme"`
}

type browserSettings struct {
	Gecko gecko `json:"gecko"`
}

type gecko struct {
	ID               string `json:"id"`
	StrictMinVersion string `json:"strict_min_version"`
}

type theme struct {
	Colors     orderedColors `json:"colors"`
	Properties properties    `json:"properties"`
}

// properties sets Firefox's own menus and built-in pages to the adaptation's
// appearance. color_scheme alone would also set the preferred color scheme
// of web pages; content_color_scheme overrides that back to the system
// setting, because quiet colors Firefox's interface, not web content.
type properties struct {
	ColorScheme        string `json:"color_scheme"`
	ContentColorScheme string `json:"content_color_scheme"`
}

// orderedColors keeps manifest keys in mapping order. A Go map would sort
// them alphabetically and scatter the groups the mapping documents.
type orderedColors []colorEntry

type colorEntry struct {
	key   string
	value string
}

func (c orderedColors) MarshalJSON() ([]byte, error) {
	var b []byte

	b = append(b, '{')

	for i, e := range c {
		if i > 0 {
			b = append(b, ',')
		}

		key, err := json.Marshal(e.key)
		if err != nil {
			return nil, err
		}

		b = append(b, key...)
		b = append(b, ':')
		b = append(b, '"')
		b = append(b, e.value...)
		b = append(b, '"')
	}

	return append(b, '}'), nil
}

func resolve(p core.Palette) (map[role]color.SRGB, error) {
	derived, err := core.Derive(p)
	if err != nil {
		return nil, fmt.Errorf(
			"derive quiet roles: %w",
			err,
		)
	}

	return map[role]color.SRGB{
		field:     p.Anchors.Field,
		surface:   p.Anchors.Surface,
		ink:       p.Anchors.Ink,
		muted:     p.Anchors.Muted,
		line:      p.Anchors.Line,
		mineral:   p.Anchors.Mineral,
		selection: derived.Selection,
		callout:   derived.Callout,
	}, nil
}

// Encode writes the Firefox theme manifest for p.
func Encode(
	w io.Writer,
	p core.Palette,
) error {
	resolved, err := resolve(p)
	if err != nil {
		return err
	}

	var scheme string

	switch p.Mode {
	case core.ModeDay:
		scheme = "light"

	case core.ModeNight:
		scheme = "dark"

	default:
		return fmt.Errorf(
			"unsupported Firefox palette mode %q",
			p.Mode,
		)
	}

	colors := make(orderedColors, 0, len(mapping))

	for _, e := range mapping {
		hex, err := resolved[e.role].Hex()
		if err != nil {
			return fmt.Errorf(
				"Firefox %s: %w",
				e.key,
				err,
			)
		}

		colors = append(colors, colorEntry{e.key, hex})
	}

	doc := manifest{
		ManifestVersion: 3,
		Name:            p.Name,
		Version:         release.Version,
		Description:     "A restrained color system for long-duration work.",
		Browser: browserSettings{
			Gecko: gecko{
				ID:               p.Name + "@chrismecca.com",
				StrictMinVersion: minimumVersion,
			},
		},
		Theme: theme{
			Colors: colors,
			Properties: properties{
				ColorScheme:        scheme,
				ContentColorScheme: "system",
			},
		},
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(doc); err != nil {
		return fmt.Errorf(
			"encode Firefox theme: %w",
			err,
		)
	}

	return nil
}
