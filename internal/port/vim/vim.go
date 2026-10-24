// @(#)quiet / go / internal/port/vim/vim.go / 2026-10-24

// Package vim maps quiet-core semantics onto Vim highlight groups.
package vim

import (
	"fmt"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
	"github.com/chrismecca/quiet/internal/port/terminal"
)

type Highlight struct {
	Name string

	Foreground *color.SRGB
	Background *color.SRGB

	CTermForeground string
	CTermBackground string
}

type Link struct {
	From string
	To   string
}

type Theme struct {
	Name       string
	Background string

	Highlights []Highlight
	Links      []Link

	ANSI [16]color.SRGB
}

func Resolve(
	p core.Palette,
) (Theme, error) {
	var (
		name       string
		background string

		fieldCTerm string
		inkCTerm   string
		mutedCTerm string
		lineCTerm  string
	)

	switch fmt.Sprint(p.Mode) {
	case "day":
		name = "quiet-day"
		background = "light"

		fieldCTerm = "15"
		inkCTerm = "0"
		mutedCTerm = "8"
		lineCTerm = "7"

	case "night":
		name = "quiet-night"
		background = "dark"

		fieldCTerm = "0"
		inkCTerm = "15"
		mutedCTerm = "7"
		lineCTerm = "8"

	default:
		return Theme{}, fmt.Errorf(
			"unsupported quiet mode %q",
			fmt.Sprint(p.Mode),
		)
	}

	derived, err := core.Derive(p)
	if err != nil {
		return Theme{}, fmt.Errorf(
			"derive quiet roles: %w",
			err,
		)
	}

	subtle, err := core.SubtleOver(
		p,
		p.Anchors.Field,
	)
	if err != nil {
		return Theme{}, fmt.Errorf(
			"derive Subtle over Field: %w",
			err,
		)
	}

	ansi, err := terminal.Resolve(p)
	if err != nil {
		return Theme{}, fmt.Errorf(
			"resolve terminal palette: %w",
			err,
		)
	}

	a := p.Anchors

	field := role(
		a.Field,
		fieldCTerm,
	)

	ink := role(
		a.Ink,
		inkCTerm,
	)

	muted := role(
		a.Muted,
		mutedCTerm,
	)

	line := role(
		a.Line,
		lineCTerm,
	)

	subtleRole := role(
		subtle,
		lineCTerm,
	)

	mineral := role(
		a.Mineral,
		"4",
	)

	moss := role(
		a.Moss,
		"2",
	)

	ochre := role(
		a.Ochre,
		"3",
	)

	rust := role(
		a.Rust,
		"1",
	)

	surface := role(
		a.Surface,
		fieldCTerm,
	)

	structure := role(
		a.Structure,
		lineCTerm,
	)

	selection := role(
		derived.Selection,
		lineCTerm,
	)

	callout := role(
		derived.Callout,
		fieldCTerm,
	)

	highlights := []Highlight{
		highlight(
			"Normal",
			ink,
			field,
		),

		highlight(
			"NormalNC",
			ink,
			field,
		),

		// Syntax hierarchy.
		highlight(
			"Comment",
			muted,
			field,
		),

		highlight(
			"Constant",
			ink,
			field,
		),

		highlight(
			"Identifier",
			ink,
			field,
		),

		highlight(
			"Statement",
			mineral,
			field,
		),

		highlight(
			"PreProc",
			mineral,
			field,
		),

		highlight(
			"Type",
			mineral,
			field,
		),

		highlight(
			"Special",
			ink,
			field,
		),

		highlight(
			"Underlined",
			mineral,
			field,
		),

		highlight(
			"Ignore",
			subtleRole,
			field,
		),

		highlight(
			"Error",
			rust,
			field,
		),

		highlight(
			"Todo",
			ochre,
			field,
		),

		// Editor structure.
		highlight(
			"ColorColumn",
			noneRole(),
			structure,
		),

		highlight(
			"Conceal",
			muted,
			field,
		),

		highlight(
			"Cursor",
			field,
			mineral,
		),

		highlight(
			"lCursor",
			field,
			mineral,
		),

		highlight(
			"CursorColumn",
			noneRole(),
			callout,
		),

		highlight(
			"CursorLine",
			noneRole(),
			callout,
		),

		highlight(
			"CursorLineNr",
			mineral,
			callout,
		),

		highlight(
			"LineNr",
			subtleRole,
			field,
		),

		highlight(
			"Directory",
			mineral,
			field,
		),

		highlight(
			"EndOfBuffer",
			line,
			field,
		),

		highlight(
			"FoldColumn",
			subtleRole,
			field,
		),

		highlight(
			"Folded",
			muted,
			surface,
		),

		highlight(
			"NonText",
			line,
			field,
		),

		highlight(
			"SignColumn",
			muted,
			field,
		),

		highlight(
			"SpecialKey",
			line,
			field,
		),

		highlight(
			"WinSeparator",
			line,
			field,
		),

		highlight(
			"VertSplit",
			line,
			field,
		),

		// Selection and navigation.
		highlight(
			"Visual",
			ink,
			selection,
		),

		highlight(
			"VisualNOS",
			ink,
			selection,
		),

		highlight(
			"Search",
			ink,
			selection,
		),

		highlight(
			"IncSearch",
			field,
			mineral,
		),

		highlight(
			"MatchParen",
			field,
			mineral,
		),

		highlight(
			"QuickFixLine",
			ink,
			selection,
		),

		highlight(
			"WildMenu",
			ink,
			selection,
		),

		// Messages and state.
		highlight(
			"ErrorMsg",
			rust,
			field,
		),

		highlight(
			"WarningMsg",
			ochre,
			field,
		),

		highlight(
			"ModeMsg",
			mineral,
			field,
		),

		highlight(
			"MoreMsg",
			mineral,
			field,
		),

		highlight(
			"Question",
			mineral,
			field,
		),

		highlight(
			"Title",
			ink,
			field,
		),

		// Menus and auxiliary surfaces.
		highlight(
			"Pmenu",
			ink,
			surface,
		),

		highlight(
			"PmenuSel",
			ink,
			selection,
		),

		highlight(
			"PmenuSbar",
			noneRole(),
			structure,
		),

		highlight(
			"PmenuThumb",
			noneRole(),
			mineral,
		),

		highlight(
			"NormalFloat",
			ink,
			surface,
		),

		highlight(
			"FloatBorder",
			line,
			surface,
		),

		highlight(
			"FloatTitle",
			ink,
			surface,
		),

		// Status and tabs.
		highlight(
			"StatusLine",
			ink,
			structure,
		),

		highlight(
			"StatusLineNC",
			muted,
			surface,
		),

		highlight(
			"TabLine",
			muted,
			surface,
		),

		highlight(
			"TabLineFill",
			line,
			surface,
		),

		highlight(
			"TabLineSel",
			ink,
			structure,
		),

		// Diff semantics.
		highlight(
			"DiffAdd",
			moss,
			surface,
		),

		highlight(
			"DiffChange",
			ochre,
			surface,
		),

		highlight(
			"DiffDelete",
			rust,
			surface,
		),

		highlight(
			"DiffText",
			ink,
			selection,
		),

		highlight(
			"Added",
			moss,
			field,
		),

		highlight(
			"Changed",
			ochre,
			field,
		),

		highlight(
			"Removed",
			rust,
			field,
		),
	}

	links := []Link{
		{"Character", "Constant"},
		{"Number", "Constant"},
		{"Boolean", "Constant"},
		{"Float", "Constant"},

		{"Function", "Identifier"},

		{"Conditional", "Statement"},
		{"Repeat", "Statement"},
		{"Label", "Statement"},
		{"Operator", "Statement"},
		{"Keyword", "Statement"},
		{"Exception", "Statement"},

		{"Include", "PreProc"},
		{"Define", "PreProc"},
		{"Macro", "PreProc"},
		{"PreCondit", "PreProc"},

		{"StorageClass", "Type"},
		{"Structure", "Type"},
		{"Typedef", "Type"},

		{"SpecialChar", "Special"},
		{"Tag", "Special"},
		{"Delimiter", "Special"},
		{"SpecialComment", "Comment"},
		{"Debug", "Special"},
	}

	return Theme{
		Name:       name,
		Background: background,
		Highlights: highlights,
		Links:      links,
		ANSI:       ansi.ANSI,
	}, nil
}

type resolvedRole struct {
	color color.SRGB
	cterm string
	set   bool
}

func role(
	c color.SRGB,
	cterm string,
) resolvedRole {
	return resolvedRole{
		color: c,
		cterm: cterm,
		set:   true,
	}
}

func noneRole() resolvedRole {
	return resolvedRole{}
}

func highlight(
	name string,
	fg resolvedRole,
	bg resolvedRole,
) Highlight {
	result := Highlight{
		Name:            name,
		CTermForeground: "NONE",
		CTermBackground: "NONE",
	}

	if fg.set {
		value := fg.color
		result.Foreground = &value
		result.CTermForeground = fg.cterm
	}

	if bg.set {
		value := bg.color
		result.Background = &value
		result.CTermBackground = bg.cterm
	}

	return result
}
