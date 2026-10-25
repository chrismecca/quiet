// @(#)quiet / go / internal/port/jetbrains/scheme.go / 2026-10-25

package jetbrains

import (
	"fmt"
	"io"
	"strings"

	"github.com/chrismecca/quiet/internal/core"
)

// schemeVersion is the newest editor scheme format the IntelliJ Platform
// reads (AbstractColorsScheme.CURR_VERSION).
const schemeVersion = 142

type colorOption struct {
	name string
	role role
}

type colorSection struct {
	comment string
	options []colorOption
}

// schemeColors is the editor's color-key policy. It follows the VS Code
// editor mapping: Mineral caret and current line number, Callout caret row,
// Selection with Ink text, Line for guides and separators.
var schemeColors = []colorSection{
	{
		comment: "Caret and selection.",
		options: []colorOption{
			{"CARET_COLOR", mineral},
			{"CARET_ROW_COLOR", callout},
			{"SELECTION_BACKGROUND", selection},
			{"SELECTION_BACKGROUND_INACTIVE", callout},
			{"SELECTION_FOREGROUND", ink},
		},
	},
	{
		comment: "Gutter.",
		options: []colorOption{
			{"GUTTER_BACKGROUND", field},
			{"EDITOR_GUTTER_BACKGROUND", field},
			{"LINE_NUMBERS_COLOR", muted},
			{"LINE_NUMBER_ON_CARET_ROW_COLOR", mineral},
			{"ANNOTATIONS_COLOR", muted},
			{"ANNOTATIONS_LAST_COMMIT_COLOR", ink},
		},
	},
	{
		comment: "Guides, separators, and invisible characters.",
		options: []colorOption{
			{"WHITESPACES", line},
			{"TABS", line},
			{"INDENT_GUIDE", line},
			{"STRING_CONTENT_INDENT_GUIDE", line},
			{"VISUAL_INDENT_GUIDE", line},
			{"SELECTED_INDENT_GUIDE", mineral},
			{"MATCHED_BRACES_INDENT_GUIDE_COLOR", mineral},
			{"RIGHT_MARGIN_COLOR", line},
			{"SOFT_WRAP_SIGN_COLOR", line},
			{"METHOD_SEPARATORS_COLOR", line},
			{"TEARLINE_COLOR", line},
			{"SELECTED_TEARLINE_COLOR", muted},
			{"FOLDED_TEXT_BORDER_COLOR", line},
			{"BREADCRUMBS_BORDER_COLOR", line},
			{"CODE_LENS_BORDER_COLOR", line},
		},
	},
	{
		comment: "Changed lines in the gutter.",
		options: []colorOption{
			{"ADDED_LINES_COLOR", moss},
			{"MODIFIED_LINES_COLOR", ochre},
			{"DELETED_LINES_COLOR", rust},
			{"WHITESPACES_MODIFIED_LINES_COLOR", muted},
			{"IGNORED_ADDED_LINES_BORDER_COLOR", moss},
			{"IGNORED_MODIFIED_LINES_BORDER_COLOR", ochre},
			{"IGNORED_DELETED_LINES_BORDER_COLOR", rust},
			{"BORDER_LINES_COLOR", line},
		},
	},
	{
		comment: "Popups, panels, and sticky lines.",
		options: []colorOption{
			{"DOCUMENTATION_COLOR", surface},
			{"LOOKUP_COLOR", surface},
			{"NOTIFICATION_BACKGROUND", surface},
			{"READONLY_FRAGMENT_BACKGROUND", surface},
			{"PREVIEW_BACKGROUND", surface},
			{"PREVIEW_BORDER_COLOR", line},
			{"STICKY_LINES_BACKGROUND", surface},
			{"STICKY_LINES_HOVERED_COLOR", callout},
			{"STICKY_LINES_BORDER_COLOR", line},
		},
	},
	{
		comment: "Console.",
		options: []colorOption{
			{"CONSOLE_BACKGROUND_KEY", terminalBackground},
		},
	},
}

type fontType int

// Font type codes, as JetBrains writes them.
const (
	plain  fontType = 0
	bold   fontType = 1
	italic fontType = 2
)

type effectKind int

const (
	noEffect effectKind = iota
	boxed
	underline
	wave
	strikeout
)

// effectCodes are the EFFECT_TYPE codes JetBrains reads
// (AttributesFlyweight.toEffectType).
var effectCodes = map[effectKind]int{
	boxed:     0,
	underline: 1,
	wave:      2,
	strikeout: 3,
}

// attributes is one JetBrains text-attributes value. A role of none leaves
// that part unset.
type attributes struct {
	fg          role
	bg          role
	font        fontType
	effect      effectKind
	effectColor role
	stripe      role
}

type attributeOption struct {
	name       string
	attributes attributes
}

type attributeSection struct {
	comment string
	options []attributeOption
}

// schemeAttributes is the editor's text-attributes policy.
//
// Syntax follows quiet's TextMate scope policy: Ink for ordinary source,
// Muted for comments, Mineral for language structure. Language-specific keys
// fall back to these defaults, so every language shares the hierarchy.
//
// Font styles and effects follow the non-color cues JetBrains' Default and
// Darcula schemes share: underlined reassigned variables, italic static
// members, struck-out deprecated names. quiet changes only their colors.
//
// Backgrounds drawn under highlighted source use the line-tier washes, where
// quiet-core guarantees every foreground.
var schemeAttributes = []attributeSection{
	{
		comment: "Text.",
		options: []attributeOption{
			{"TEXT", attributes{fg: ink, bg: field}},
			{"BAD_CHARACTER", attributes{effect: wave, effectColor: rust}},
		},
	},
	{
		comment: "Comments.",
		options: []attributeOption{
			{"DEFAULT_LINE_COMMENT", attributes{fg: muted}},
			{"DEFAULT_BLOCK_COMMENT", attributes{fg: muted}},
			{"DEFAULT_DOC_COMMENT", attributes{fg: muted, font: italic}},
			{"DEFAULT_DOC_MARKUP", attributes{fg: muted}},
			{"DEFAULT_DOC_COMMENT_TAG", attributes{fg: muted, font: bold, effect: underline, effectColor: muted}},
			{"DEFAULT_DOC_COMMENT_TAG_VALUE", attributes{fg: muted}},
		},
	},
	{
		comment: "Ordinary source.",
		options: []attributeOption{
			{"DEFAULT_IDENTIFIER", attributes{fg: ink}},
			{"DEFAULT_NUMBER", attributes{fg: ink}},
			{"DEFAULT_STRING", attributes{fg: ink}},
			{"DEFAULT_VALID_STRING_ESCAPE", attributes{fg: ink}},
			{"DEFAULT_CONSTANT", attributes{fg: ink, font: italic}},
			{"DEFAULT_LOCAL_VARIABLE", attributes{fg: ink}},
			{"DEFAULT_REASSIGNED_LOCAL_VARIABLE", attributes{fg: ink, effect: underline, effectColor: muted}},
			{"DEFAULT_GLOBAL_VARIABLE", attributes{fg: ink}},
			{"DEFAULT_PARAMETER", attributes{fg: ink}},
			{"DEFAULT_REASSIGNED_PARAMETER", attributes{fg: ink, effect: underline, effectColor: muted}},
			{"DEFAULT_FUNCTION_DECLARATION", attributes{fg: ink}},
			{"DEFAULT_FUNCTION_CALL", attributes{fg: ink}},
			{"DEFAULT_INSTANCE_METHOD", attributes{fg: ink}},
			{"DEFAULT_INSTANCE_FIELD", attributes{fg: ink}},
			{"DEFAULT_STATIC_METHOD", attributes{fg: ink, font: italic}},
			{"DEFAULT_STATIC_FIELD", attributes{fg: ink, font: italic}},
			{"DEFAULT_ATTRIBUTE", attributes{fg: ink}},
			{"DEFAULT_ENTITY", attributes{fg: ink}},
			{"DEFAULT_LABEL", attributes{fg: ink}},
			{"DEFAULT_HIGHLIGHTED_REFERENCE", attributes{fg: ink, effect: underline, effectColor: muted}},
			{"DEFAULT_BRACES", attributes{fg: ink}},
			{"DEFAULT_BRACKETS", attributes{fg: ink}},
			{"DEFAULT_PARENTHS", attributes{fg: ink}},
			{"DEFAULT_DOT", attributes{fg: ink}},
			{"DEFAULT_COMMA", attributes{fg: ink}},
			{"DEFAULT_SEMICOLON", attributes{fg: ink}},
		},
	},
	{
		comment: "Language structure.",
		options: []attributeOption{
			{"DEFAULT_KEYWORD", attributes{fg: mineral}},
			{"DEFAULT_OPERATION_SIGN", attributes{fg: mineral}},
			{"DEFAULT_CLASS_NAME", attributes{fg: mineral}},
			{"DEFAULT_INTERFACE_NAME", attributes{fg: mineral}},
			{"DEFAULT_CLASS_REFERENCE", attributes{fg: mineral}},
			{"DEFAULT_PREDEFINED_SYMBOL", attributes{fg: mineral}},
			{"DEFAULT_METADATA", attributes{fg: mineral}},
			{"DEFAULT_TAG", attributes{fg: mineral}},
		},
	},
	{
		comment: "Invalid source.",
		options: []attributeOption{
			{"DEFAULT_INVALID_STRING_ESCAPE", attributes{fg: rust, effect: wave, effectColor: rust}},
		},
	},
	{
		comment: "Diagnostics.",
		options: []attributeOption{
			{"ERRORS_ATTRIBUTES", attributes{effect: wave, effectColor: rust, stripe: rust}},
			{"RUNTIME_ERROR", attributes{effect: wave, effectColor: rust, stripe: rust}},
			{"WRONG_REFERENCES_ATTRIBUTES", attributes{fg: rust}},
			{"WARNING_ATTRIBUTES", attributes{effect: wave, effectColor: ochre, stripe: ochre}},
			{"GENERIC_SERVER_ERROR_OR_WARNING", attributes{effect: wave, effectColor: ochre, stripe: ochre}},
			{"INFO_ATTRIBUTES", attributes{effect: wave, effectColor: mineral, stripe: mineral}},
			{"NOT_USED_ELEMENT_ATTRIBUTES", attributes{fg: muted}},
			{"DEPRECATED_ATTRIBUTES", attributes{effect: strikeout, effectColor: ochre}},
			{"MARKED_FOR_REMOVAL_ATTRIBUTES", attributes{effect: strikeout, effectColor: rust}},
			{"TODO_DEFAULT_ATTRIBUTES", attributes{fg: ochre, font: italic, stripe: ochre}},
		},
	},
	{
		comment: "Search and navigation.",
		options: []attributeOption{
			{"SEARCH_RESULT_ATTRIBUTES", attributes{bg: mineralWash, stripe: mineral}},
			{"TEXT_SEARCH_RESULT_ATTRIBUTES", attributes{bg: mineralWash, stripe: mineral}},
			{"WRITE_SEARCH_RESULT_ATTRIBUTES", attributes{bg: ochreWash, stripe: ochre}},
			{"IDENTIFIER_UNDER_CARET_ATTRIBUTES", attributes{bg: mineralWash, stripe: muted}},
			{"WRITE_IDENTIFIER_UNDER_CARET_ATTRIBUTES", attributes{bg: ochreWash, stripe: ochre}},
			{"MATCHED_BRACE_ATTRIBUTES", attributes{fg: field, bg: mineral}},
			{"UNMATCHED_BRACE_ATTRIBUTES", attributes{fg: field, bg: rust}},
			{"HYPERLINK_ATTRIBUTES", attributes{fg: mineral, effect: underline, effectColor: mineral}},
			{"FOLLOWED_HYPERLINK_ATTRIBUTES", attributes{fg: mineral, effect: underline, effectColor: mineral}},
			{"INACTIVE_HYPERLINK_ATTRIBUTES", attributes{effect: underline, effectColor: muted}},
			{"BOOKMARKS_ATTRIBUTES", attributes{stripe: mineral}},
		},
	},
	{
		comment: "Editor regions.",
		options: []attributeOption{
			{"FOLDED_TEXT_ATTRIBUTES", attributes{fg: muted, bg: surface}},
			{"DELETED_TEXT_ATTRIBUTES", attributes{effect: strikeout, effectColor: muted}},
			{"INJECTED_LANGUAGE_FRAGMENT", attributes{}},
			{"LIVE_TEMPLATE_ATTRIBUTES", attributes{effect: boxed, effectColor: mineral}},
			{"LIVE_TEMPLATE_INACTIVE_SEGMENT", attributes{effect: boxed, effectColor: line}},
			{"BREADCRUMBS_DEFAULT", attributes{fg: muted}},
			{"BREADCRUMBS_HOVERED", attributes{fg: ink, bg: callout}},
			{"BREADCRUMBS_CURRENT", attributes{fg: ink, bg: callout}},
			{"BREADCRUMBS_INACTIVE", attributes{fg: muted}},
		},
	},
	{
		comment: "Inlay hints and suggestions.",
		options: []attributeOption{
			{"INLAY_DEFAULT", attributes{fg: muted, bg: surface}},
			{"INLAY_TEXT_WITHOUT_BACKGROUND", attributes{fg: muted}},
			{"INLINE_PARAMETER_HINT", attributes{fg: muted, bg: surface}},
			{"INLINE_PARAMETER_HINT_HIGHLIGHTED", attributes{fg: ink, bg: selection}},
			{"INLINE_PARAMETER_HINT_CURRENT", attributes{fg: ink, bg: selection}},
			{"INLINE_SUGGESTION", attributes{fg: muted}},
		},
	},
	{
		comment: "Diff and merge.",
		options: []attributeOption{
			{"DIFF_INSERTED", attributes{bg: mossWash, stripe: moss}},
			{"DIFF_MODIFIED", attributes{bg: ochreWash, stripe: ochre}},
			{"DIFF_DELETED", attributes{bg: structure, stripe: muted}},
			{"DIFF_CONFLICT", attributes{bg: rustWash, stripe: rust}},
		},
	},
	{
		comment: "Code coverage.",
		options: []attributeOption{
			{"LINE_FULL_COVERAGE", attributes{fg: mossWashEmphasis, font: bold}},
			{"LINE_PARTIAL_COVERAGE", attributes{fg: ochreWashEmphasis, font: bold}},
			{"LINE_NONE_COVERAGE", attributes{fg: rustWashEmphasis, font: bold}},
		},
	},
	{
		comment: "Console and logs.",
		options: []attributeOption{
			{"CONSOLE_NORMAL_OUTPUT", attributes{fg: terminalText}},
			{"CONSOLE_ERROR_OUTPUT", attributes{fg: rust}},
			{"CONSOLE_SYSTEM_OUTPUT", attributes{fg: muted}},
			{"CONSOLE_USER_INPUT", attributes{fg: mineral, font: italic}},
			{"LOG_ERROR_OUTPUT", attributes{fg: rust}},
			{"LOG_WARNING_OUTPUT", attributes{fg: ochre}},
			{"LOG_INFO_OUTPUT", attributes{fg: ink}},
			{"LOG_DEBUG_OUTPUT", attributes{fg: muted}},
			{"LOG_VERBOSE_OUTPUT", attributes{fg: muted}},
			{"LOG_EXPIRED_ENTRY", attributes{fg: muted}},
		},
	},
	{
		comment: "Console ANSI colors, from quiet's terminal palette.",
		options: consoleANSI(),
	},
}

// consoleKeys names the console's sixteen ANSI colors in slot order.
var consoleKeys = [16]string{
	"CONSOLE_BLACK_OUTPUT",
	"CONSOLE_RED_OUTPUT",
	"CONSOLE_GREEN_OUTPUT",
	"CONSOLE_YELLOW_OUTPUT",
	"CONSOLE_BLUE_OUTPUT",
	"CONSOLE_MAGENTA_OUTPUT",
	"CONSOLE_CYAN_OUTPUT",
	"CONSOLE_GRAY_OUTPUT",
	"CONSOLE_DARKGRAY_OUTPUT",
	"CONSOLE_RED_BRIGHT_OUTPUT",
	"CONSOLE_GREEN_BRIGHT_OUTPUT",
	"CONSOLE_YELLOW_BRIGHT_OUTPUT",
	"CONSOLE_BLUE_BRIGHT_OUTPUT",
	"CONSOLE_MAGENTA_BRIGHT_OUTPUT",
	"CONSOLE_CYAN_BRIGHT_OUTPUT",
	"CONSOLE_WHITE_OUTPUT",
}

// consoleANSI sets each ANSI key's foreground and background, which the
// console uses for ANSI foreground and background sequences.
func consoleANSI() []attributeOption {
	options := make([]attributeOption, 0, len(consoleKeys))

	for i, key := range consoleKeys {
		options = append(options, attributeOption{
			key,
			attributes{fg: ansi(i), bg: ansi(i)},
		})
	}

	return options
}

// EncodeScheme writes the JetBrains editor color scheme for p.
func EncodeScheme(
	w io.Writer,
	p core.Palette,
) error {
	resolved, err := resolve(p)
	if err != nil {
		return err
	}

	parent, err := parentScheme(p)
	if err != nil {
		return err
	}

	var b strings.Builder

	fmt.Fprintf(
		&b,
		"<scheme name=\"%s\" version=\"%d\" parent_scheme=\"%s\">\n"+
			"  <!-- Generated by quiet port jetbrains-scheme. Do not edit. -->\n"+
			"  <colors>\n",
		p.Name,
		schemeVersion,
		parent,
	)

	for i, s := range schemeColors {
		if i > 0 {
			b.WriteString("\n")
		}

		fmt.Fprintf(&b, "    <!-- %s -->\n", s.comment)

		for _, o := range s.options {
			hex, err := resolved.hex(o.role)
			if err != nil {
				return fmt.Errorf(
					"%s %s: %w",
					p.Name,
					o.name,
					err,
				)
			}

			fmt.Fprintf(
				&b,
				"    <option name=\"%s\" value=\"%s\" />\n",
				o.name,
				hex,
			)
		}
	}

	b.WriteString("  </colors>\n  <attributes>\n")

	for i, s := range schemeAttributes {
		if i > 0 {
			b.WriteString("\n")
		}

		fmt.Fprintf(&b, "    <!-- %s -->\n", s.comment)

		for _, o := range s.options {
			if err := writeAttributes(&b, resolved, o); err != nil {
				return fmt.Errorf(
					"%s %s: %w",
					p.Name,
					o.name,
					err,
				)
			}
		}
	}

	b.WriteString("  </attributes>\n</scheme>\n")

	if _, err := io.WriteString(
		w,
		b.String(),
	); err != nil {
		return fmt.Errorf(
			"write JetBrains scheme: %w",
			err,
		)
	}

	return nil
}

// writeAttributes writes one attributes option, with its fields in the order
// JetBrains writes them (AttributesFlyweight.writeExternal).
func writeAttributes(
	b *strings.Builder,
	p palette,
	o attributeOption,
) error {
	a := o.attributes

	var fields []string

	addColor := func(name string, r role) error {
		if r == none {
			return nil
		}

		hex, err := p.hex(r)
		if err != nil {
			return err
		}

		fields = append(fields, fmt.Sprintf(
			"<option name=\"%s\" value=\"%s\" />",
			name,
			hex,
		))

		return nil
	}

	if err := addColor("FOREGROUND", a.fg); err != nil {
		return err
	}

	if err := addColor("BACKGROUND", a.bg); err != nil {
		return err
	}

	if a.font != plain {
		fields = append(fields, fmt.Sprintf(
			"<option name=\"FONT_TYPE\" value=\"%d\" />",
			a.font,
		))
	}

	if (a.effect == noEffect) != (a.effectColor == none) {
		return fmt.Errorf("effect and effect color must be set together")
	}

	if err := addColor("EFFECT_COLOR", a.effectColor); err != nil {
		return err
	}

	if err := addColor("ERROR_STRIPE_COLOR", a.stripe); err != nil {
		return err
	}

	if a.effect != noEffect {
		fields = append(fields, fmt.Sprintf(
			"<option name=\"EFFECT_TYPE\" value=\"%d\" />",
			effectCodes[a.effect],
		))
	}

	if len(fields) == 0 {
		fmt.Fprintf(
			b,
			"    <option name=\"%s\">\n      <value />\n    </option>\n",
			o.name,
		)

		return nil
	}

	fmt.Fprintf(b, "    <option name=\"%s\">\n      <value>\n", o.name)

	for _, f := range fields {
		fmt.Fprintf(b, "        %s\n", f)
	}

	b.WriteString("      </value>\n    </option>\n")

	return nil
}
