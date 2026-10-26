// @(#)quiet / go / internal/port/chrome/chrome.go / 2026-10-25

// Package chrome maps quiet-core semantics onto a Chrome theme.
//
// A Chrome theme is an extension whose manifest names browser chrome colors.
// The port colors Chrome's interface only; web content keeps its own colors.
package chrome

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
	"github.com/chrismecca/quiet/internal/release"
)

// role names a quiet color used by the Chrome mapping.
type role int

const (
	field role = iota
	surface
	ink
	muted
	line
	mineral
)

type entry struct {
	key  string
	role role
}

// mapping is the complete Chrome chrome policy. It follows the Firefox port
// where Chrome exposes the same element: the tab strip is Surface, the
// selected tab and toolbar are Field and join the page.
//
// Chrome draws the selected tab in the toolbar color and derives the address
// bar, menus, and hover states from the theme colors itself; the theme cannot
// set them. Incognito keys are omitted: Chrome's color pipeline reads no
// Incognito theme colors.
var mapping = []entry{
	// Tab strip.
	{"frame", surface},
	{"frame_inactive", surface},
	{"background_tab", surface},
	{"background_tab_inactive", surface},
	{"tab_background_text", muted},
	{"tab_background_text_inactive", muted},
	{"tab_text", ink},

	// Toolbar and bookmarks bar.
	{"toolbar", field},
	{"toolbar_text", ink},
	{"toolbar_button_icon", muted},
	{"bookmark_text", ink},

	// Address bar suggestions.
	{"omnibox_background", surface},
	{"omnibox_text", ink},

	// New tab page.
	{"ntp_background", field},
	{"ntp_text", ink},
	{"ntp_link", mineral},
	{"ntp_header", line},
}

func resolve(p core.Palette) map[role]color.SRGB {
	return map[role]color.SRGB{
		field:   p.Anchors.Field,
		surface: p.Anchors.Surface,
		ink:     p.Anchors.Ink,
		muted:   p.Anchors.Muted,
		line:    p.Anchors.Line,
		mineral: p.Anchors.Mineral,
	}
}

// Encode writes the Chrome theme manifest for p.
//
// The manifest is written directly rather than through encoding/json, so
// each color stays a one-line [r, g, b] array in mapping order.
func Encode(
	w io.Writer,
	p core.Palette,
) error {
	if p.Mode != core.ModeDay && p.Mode != core.ModeNight {
		return fmt.Errorf(
			"unsupported Chrome palette mode %q",
			p.Mode,
		)
	}

	resolved := resolve(p)

	name, err := json.Marshal(p.Name)
	if err != nil {
		return fmt.Errorf(
			"encode Chrome theme name: %w",
			err,
		)
	}

	var b strings.Builder

	fmt.Fprintf(
		&b,
		`{
  "manifest_version": 3,
  "name": %s,
  "version": %q,
  "description": "A restrained color system for long-duration work.",
  "theme": {
    "colors": {
`,
		name,
		release.Version,
	)

	for i, e := range mapping {
		r, g, bl, err := resolved[e.role].RGB8()
		if err != nil {
			return fmt.Errorf(
				"Chrome %s: %w",
				e.key,
				err,
			)
		}

		separator := ","
		if i == len(mapping)-1 {
			separator = ""
		}

		fmt.Fprintf(
			&b,
			"      %q: [%d, %d, %d]%s\n",
			e.key,
			r,
			g,
			bl,
			separator,
		)
	}

	b.WriteString(`    }
  }
}
`)

	if _, err := io.WriteString(
		w,
		b.String(),
	); err != nil {
		return fmt.Errorf(
			"write Chrome theme: %w",
			err,
		)
	}

	return nil
}
