// @(#)quiet / go / internal/port/jetbrains/theme.go / 2026-10-25

package jetbrains

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/chrismecca/quiet/internal/core"
)

// roleNames are the named colors the theme defines for quiet's roles. Every
// other color in the theme refers to one of them by name.
var roleNames = []struct {
	role role
	name string
}{
	{field, "quiet-field"},
	{surface, "quiet-surface"},
	{structure, "quiet-structure"},
	{ink, "quiet-ink"},
	{muted, "quiet-muted"},
	{line, "quiet-line"},
	{mineral, "quiet-mineral"},
	{moss, "quiet-moss"},
	{ochre, "quiet-ochre"},
	{rust, "quiet-rust"},
	{selection, "quiet-selection"},
	{callout, "quiet-callout"},
	{mineralWash, "quiet-mineral-wash"},
	{mossWash, "quiet-moss-wash"},
	{ochreWash, "quiet-ochre-wash"},
	{rustWash, "quiet-rust-wash"},
}

// token assigns a quiet role to one of the parent theme's named colors. Most
// tokens mean the same in both adaptations; the layer tokens do not, because
// the Islands themes order their layers differently in light and dark.
type token struct {
	name  string
	day   role
	night role
}

func same(name string, r role) token {
	return token{name, r, r}
}

type tokenSection struct {
	comment string
	tokens  []token
}

// themeTokens is the UI theme's policy, expressed over the semantic named
// colors of JetBrains' Islands themes. The parent theme's own keys refer to
// these names, so overriding a name recolors every element that uses it.
//
// The mapping follows the VS Code workbench: Field for the editor, Surface
// for tool windows, popups, and dialogs, Structure for the window frame,
// Field for inputs, Selection and Callout for selected and hovered items,
// Mineral for focus and primary actions.
var themeTokens = []tokenSection{
	{
		comment: "Text",
		tokens: []token{
			same("text-default", ink),
			same("editor-text", ink),
			same("text-muted", muted),
			same("text-secondary", muted),
			same("text-disabled", muted),
			same("text-over-accent", field),
			same("text-over-accent-inverted", ink),
			same("text-link", mineral),
			same("text-error", rust),
			same("text-warning", ochre),
			same("text-success", moss),
		},
	},
	{
		comment: "Layers, as the Islands themes order them in each adaptation",
		tokens: []token{
			{"layer-0-bg", structure, field},
			{"layer-0-bg-inline", surface, surface},
			{"layer-0-border", line, line},
			{"layer-0-border-inline", line, line},
			{"layer-1-bg", surface, structure},
			{"layer-1-bg-inline", field, surface},
			{"layer-1-border", line, line},
			{"layer-1-border-inline", line, line},
			{"layer-2-bg", field, surface},
			{"layer-2-bg-inline", surface, field},
			{"layer-2-border", line, line},
			{"layer-2-border-inline", line, line},
		},
	},
	{
		comment: "Regions",
		tokens: []token{
			same("main-window-bg", structure),
			same("main-window-bg-alt", structure),
			same("main-window-border", line),
			same("editor-bg", field),
			same("editor-bg-inline", surface),
			same("editor-border", line),
			same("editor-border-inline", line),
			same("editor-border-alt", line),
			same("tool-window-bg", surface),
			same("tool-window-bg-inline", field),
			same("tool-window-bg-alt", surface),
			same("tool-window-border", line),
			same("tool-window-border-inline", line),
			same("dialog-bg", surface),
			same("dialog-bg-inline", field),
			same("dialog-border", line),
			same("popup-bg", surface),
			same("popup-bg-inline", field),
			same("popup-border", line),
			same("popup-border-inline", line),
			same("popup-border-windows", line),
			same("editor-floating-toolbar-bg", surface),
			same("presentation-assistant-bg", structure),
		},
	},
	{
		comment: "Accents",
		tokens: []token{
			same("accent-brand-bg", mineral),
			same("accent-brand-border", mineral),
			same("accent-brand-bg-secondary", mineralWash),
			same("accent-brand-border-secondary", mineral),
			same("accent-error-bg", rust),
			same("accent-error-border", rust),
			same("accent-error-bg-secondary", rustWash),
			same("accent-error-border-secondary", rust),
			same("accent-warning-bg", ochre),
			same("accent-warning-border", ochre),
			same("accent-warning-bg-secondary", ochreWash),
			same("accent-warning-border-secondary", ochre),
			same("accent-success-bg", moss),
			same("accent-success-border", moss),
			same("accent-success-bg-secondary", mossWash),
			same("accent-success-border-secondary", moss),
			same("accent-neutral-bg", muted),
			same("accent-ai-bg", mineralWash),
			same("accent-ai-border", mineral),
		},
	},
	{
		comment: "Hover and pressed states",
		tokens: []token{
			same("core-bg-transparent-hovered", callout),
			same("core-bg-transparent-pressed", selection),
			same("core-border-transparent", line),
		},
	},
	{
		comment: "Controls",
		tokens: []token{
			same("control-bg", field),
			same("control-bg-disabled", surface),
			same("control-bg-raised", structure),
			same("control-border", line),
			same("control-border-disabled", line),
			same("control-border-raised", line),
			same("control-border-over-accent", mineral),
			same("control-bg-small", structure),
			same("control-bg-small-disabled", surface),
			same("control-border-small", line),
			same("control-brand-bg", mineral),
			same("control-brand-border", mineral),
			same("control-error-bg", rust),
			same("control-error-border", rust),
			same("control-error-border-secondary", rust),
			same("control-warning-bg", ochre),
			same("control-warning-border", ochre),
			same("control-warning-border-secondary", ochre),
			same("control-success-bg", moss),
			same("control-success-border", moss),
		},
	},
	{
		comment: "Toggles",
		tokens: []token{
			same("toggle-off-bg", line),
			same("toggle-border", line),
			same("toggle-button-bg", muted),
		},
	},
	{
		comment: "Toolbar",
		tokens: []token{
			same("toolbar-bg-hovered", callout),
			same("toolbar-bg-pressed", selection),
			same("toolbar-border", line),
			same("toolbar-selected-bg", selection),
			same("toolbar-selected-bg-hovered", selection),
			same("toolbar-selected-bg-active", mineral),
			same("toolbar-run-bg", moss),
			same("toolbar-run-bg-hovered", moss),
			same("toolbar-stop-bg", rust),
			same("toolbar-stop-bg-hovered", rust),
		},
	},
	{
		comment: "Selection and tabs",
		tokens: []token{
			same("selection-bg-active", selection),
			same("selection-bg-active-muted", callout),
			same("selection-bg-inactive", callout),
			same("selection-bg-hovered", callout),
			same("tab-selected-bg-active", selection),
			same("tab-selected-bg-inactive", callout),
			same("tab-selected-border-active", line),
			same("tab-selected-border-inactive", line),
			same("tab-bg-hovered", callout),
			same("search-match-bg", mineralWash),
			same("popup-completion-match-text", mineral),
			same("tree-indent-guide-border", line),
		},
	},
	{
		comment: "Notifications and banners",
		tokens: []token{
			same("feedback-bg", surface),
			same("feedback-border", line),
			same("feedback-bg-inline", field),
			same("feedback-control-border", line),
			same("feedback-brand-bg", mineralWash),
			same("feedback-brand-border", mineral),
			same("feedback-success-bg", mossWash),
			same("feedback-success-border", moss),
			same("feedback-warning-bg", ochreWash),
			same("feedback-warning-border", ochre),
			same("feedback-error-bg", rustWash),
			same("feedback-error-border", rust),
			same("feedback-ai-bg", mineralWash),
			same("feedback-ai-border", mineral),
			same("got-it-bg", mineralWash),
			same("got-it-border", mineral),
			same("got-it-text-link", mineral),
			same("got-it-text-step", muted),
			same("got-it-shortcut-bg", structure),
			same("got-it-code-border", line),
			same("got-it-contrast-button-bg", mineral),
		},
	},
	{
		comment: "Classic Look and Feel colors, which older components still use",
		tokens: []token{
			same("foreground", ink),
			same("infoPanelForeground", muted),
			same("disabledForeground", muted),
			same("selectionForeground", ink),
			same("linkForeground", mineral),
			same("border", line),
			same("componentBorder", line),
			same("lightBorder", line),
			same("separatorForeground", muted),
			same("panel", surface),
			same("contentBackground", field),
		},
	},
	{
		comment: "Icons and inlays",
		tokens: []token{
			same("icon-default-stroke", muted),
			same("icon-over-accent", field),
			same("icon-green-stroke", moss),
			same("inlay-bg", surface),
			same("inlay-border", line),
		},
	},
}

type uiKey struct {
	key  string
	role role
}

type uiSection struct {
	comment string
	keys    []uiKey
}

// themeKeys sets interface keys the named colors do not reach: keys the
// Islands themes inherit from older themes with fixed colors. Each takes the
// role the VS Code workbench gives the same element.
//
// Translucent overlays and shadows are left to the parent theme, because
// they tint whatever quiet color lies beneath them. Categorical colors that
// identify something the user chose, such as project, file, and
// collaborator colors, are left as JetBrains defines them.
var themeKeys = []uiSection{
	{
		comment: "Focus",
		keys: []uiKey{
			{"Component.focusColor", mineral},
			{"Button.default.focusColor", mineral},
			{"*.caretForeground", mineral},
		},
	},
	{
		comment: "Text and inputs",
		keys: []uiKey{
			{"*.textBackground", field},
			{"*.textForeground", ink},
			{"*.inactiveBackground", surface},
			{"*.selectionBackgroundInactive", callout},
			{"*.selectionForegroundInactive", ink},
			{"Label.foreground", ink},
			{"Label.selectedForeground", ink},
			{"Label.infoForeground", muted},
			{"Label.disabledText", muted},
			{"Label.selectedDisabledForeground", muted},
			{"Component.infoForeground", muted},
			{"Component.disabledBorderColor", line},
			{"Component.iconColor", muted},
			{"Component.hoverIconColor", ink},
			{"TextField.background", field},
			{"TextField.borderColor", line},
			{"TextField.disabledBackground", surface},
			{"TextArea.background", field},
			{"TextArea.selectionForeground", ink},
			{"FormattedTextField.background", field},
			{"PasswordField.background", field},
			{"Spinner.background", field},
			{"SearchFieldWithExtension.background", field},
			{"ComboBox.nonEditableBackground", field},
			{"ComboBox.disabledForeground", muted},
			{"ComboBox.modifiedItemForeground", mineral},
			{"ComboBox.ArrowButton.iconColor", muted},
			{"ComboBox.ArrowButton.disabledIconColor", line},
			{"ComboBoxButton.background", field},
			{"Editor.background", field},
			{"Editor.foreground", ink},
			{"Editor.shortcutForeground", mineral},
			{"EditorPane.background", field},
			{"EditorPane.inactiveForeground", ink},
			{"EditorPane.inactiveBackground", surface},
			{"Content.background", field},
		},
	},
	{
		comment: "Buttons and toggles",
		keys: []uiKey{
			{"Button.startBackground", structure},
			{"Button.endBackground", structure},
			{"Button.disabledBorderColor", line},
			{"Button.disabledText", muted},
			{"ToggleButton.onBackground", mineral},
			{"ToggleButton.onForeground", field},
			{"ToggleButton.offBackground", line},
			{"ToggleButton.offForeground", muted},
			{"ToggleButton.buttonColor", field},
			{"ToggleButton.borderColor", line},
			{"Slider.trackColor", line},
			{"Slider.buttonColor", mineral},
			{"Slider.buttonBorderColor", mineral},
			{"Slider.tickColor", muted},
			{"PillButton.blueForeground", mineral},
			{"PillButton.blueBorderColor", mineral},
			{"PillButton.blueHoverForeground", mineral},
			{"PillButton.blueHoverBackground", mineralWash},
			{"PillButton.blueHoverBorderColor", mineral},
		},
	},
	{
		comment: "Lists, trees, and tables",
		keys: []uiKey{
			{"List.background", surface},
			{"List.Tag.background", structure},
			{"List.Tag.foreground", muted},
			{"Tree.background", surface},
			{"Tree.foreground", ink},
			{"Tree.textBackground", surface},
			{"Tree.modifiedItemForeground", mineral},
			{"Tree.selectionBorderColor", line},
			{"Table.background", surface},
			{"Table.foreground", ink},
			{"Table.disabledForeground", muted},
			{"Table.alternativeRowBackground", structure},
			{"Table.hoverInactiveBackground", callout},
			{"Table.lightSelectionBackground", callout},
			{"Table.lightSelectionForeground", ink},
			{"Table.lightSelectionInactiveBackground", callout},
			{"TableHeader.background", surface},
			{"TableHeader.separatorColor", line},
			{"TableHeader.disabledForeground", muted},
			{"SpeedSearch.foreground", ink},
			{"Tag.background", structure},
		},
	},
	{
		comment: "Tabs",
		keys: []uiKey{
			{"EditorTabs.borderColor", line},
			{"EditorTabs.underlineColor", mineral},
			{"EditorTabs.inactiveUnderlineColor", line},
			{"EditorTabs.underlinedTabForeground", ink},
			{"DefaultTabs.background", surface},
			{"DefaultTabs.borderColor", line},
			{"DefaultTabs.underlineColor", mineral},
			{"DefaultTabs.inactiveUnderlineColor", line},
			{"DefaultTabs.underlinedTabForeground", ink},
			{"TabbedPane.underlineColor", mineral},
			{"TabbedPane.disabledUnderlineColor", line},
			{"TabbedPane.disabledForeground", muted},
			{"MainWindow.Tab.background", structure},
			{"MainWindow.Tab.borderColor", line},
			{"MainWindow.Tab.hoverForeground", ink},
			{"MainWindow.Tab.selectedForeground", ink},
			{"MainWindow.Tab.selectedInactiveBackground", callout},
		},
	},
	{
		comment: "Tool windows",
		keys: []uiKey{
			{"ToolWindow.Header.background", surface},
			{"ToolWindow.HeaderCloseButton.background", surface},
			{"ToolWindow.HeaderTab.selectedBackground", selection},
			{"ToolWindow.HeaderTab.selectedInactiveBackground", callout},
			{"ToolWindow.HeaderTab.underlineColor", mineral},
			{"ToolWindow.HeaderTab.inactiveUnderlineColor", line},
			{"ToolWindow.HeaderTab.underlinedTabForeground", ink},
			{"ToolWindow.HeaderTab.underlinedTabInactiveForeground", ink},
			{"ToolWindow.Button.hoverBackground", callout},
			{"ToolWindow.borderColor", line},
			{"ToolWindow.Stripe.borderColor", line},
			{"ToolWindow.Button.DragAndDrop.stripeBackground", mineralWash},
			{"ToolWindow.Button.DragAndDrop.buttonDropBackground", mineralWash},
			{"ToolWindow.DragAndDrop.areaBackground", mineralWash},
			{"DragAndDrop.areaBackground", mineralWash},
			{"DragAndDrop.areaBorderColor", mineral},
			{"DragAndDrop.areaForeground", ink},
			{"DragAndDrop.rowBackground", mineralWash},
		},
	},
	{
		comment: "Window frame, toolbar, and status bar",
		keys: []uiKey{
			{"MainToolbar.foreground", ink},
			{"MainToolbar.borderColor", line},
			{"MainToolbar.Icon.hoverBackground", callout},
			{"MainToolbar.Dropdown.background", structure},
			{"MainToolbar.Dropdown.hoverBackground", callout},
			{"MainWindow.FullScreeControl.Background", structure},
			{"RunWidget.foreground", ink},
			{"RunWidget.hoverBackground", callout},
			{"RunWidget.pressedBackground", selection},
			{"RunWidget.runIconColor", moss},
			{"StatusBar.borderColor", line},
			{"StatusBar.hoverBackground", callout},
			{"StatusBar.LightEditBackground", structure},
			{"StatusBar.Breadcrumbs.floatingBackground", surface},
			{"TitlePane.infoForeground", muted},
			{"TitlePane.inactiveInfoForeground", muted},
			{"NavBar.borderColor", line},
			{"MemoryIndicator.allocatedBackground", structure},
			{"MemoryIndicator.usedBackground", mineralWash},
			{"Panel.mouseShortcutBackground", structure},
			{"SplitPane.highlight", structure},
			{"ScrollBar.background", surface},
		},
	},
	{
		comment: "Menus, popups, and completion",
		keys: []uiKey{
			{"MainMenu.background", surface},
			{"MainMenu.foreground", ink},
			{"MenuBar.borderColor", line},
			{"MenuBar.disabledBackground", surface},
			{"Menu.borderColor", line},
			{"Menu.separatorColor", line},
			{"MenuItem.acceleratorForeground", muted},
			{"MenuItem.acceleratorSelectionForeground", ink},
			{"MenuItem.disabledForeground", muted},
			{"PopupMenu.translucentBackground", surface},
			{"Popup.separatorColor", line},
			{"Popup.separatorForeground", muted},
			{"Popup.inactiveBorderColor", line},
			{"Popup.Toolbar.borderColor", line},
			{"Separator.separatorColor", line},
			{"Group.separatorColor", line},
			{"Group.disabledSeparatorColor", line},
			{"CompletionPopup.background", surface},
			{"CompletionPopup.selectionBackground", selection},
			{"CompletionPopup.selectionInactiveBackground", callout},
			{"ParameterInfo.background", surface},
			{"ParameterInfo.foreground", ink},
			{"ParameterInfo.infoForeground", muted},
			{"ParameterInfo.disabledForeground", muted},
			{"ParameterInfo.borderColor", line},
			{"ParameterInfo.lineSeparatorColor", line},
			{"ParameterInfo.currentOverloadBackground", selection},
			{"ParameterInfo.currentParameterForeground", ink},
			{"SearchEverywhere.Header.background", surface},
			{"SearchEverywhere.List.separatorColor", line},
			{"SearchEverywhere.List.separatorForeground", muted},
			{"SearchEverywhere.SearchField.background", field},
			{"SearchEverywhere.SearchField.borderColor", line},
			{"SearchEverywhere.SearchField.infoForeground", muted},
		},
	},
	{
		comment: "Tooltips",
		keys: []uiKey{
			{"ToolTip.Actions.background", surface},
			{"ToolTip.Actions.infoForeground", muted},
			{"ToolTip.separatorColor", line},
			{"Tooltip.background", surface},
			{"Tooltip.foreground", ink},
			{"Tooltip.Actions.background", surface},
			{"Editor.ToolTip.foreground", ink},
			{"Editor.ToolTip.selectionBackground", selection},
			{"GutterTooltip.infoForeground", muted},
			{"GutterTooltip.lineSeparatorColor", line},
			{"GotItTooltip.animationBackground", structure},
			{"GotItTooltip.borderSimplifiedColor", line},
			{"Canvas.Tooltip.background", surface},
		},
	},
	{
		comment: "Links",
		keys: []uiKey{
			{"Hyperlink.linkColor", mineral},
			{"Link.Tag.background", structure},
			{"Link.Tag.foreground", muted},
		},
	},
	{
		comment: "Progress, badges, and notifications",
		keys: []uiKey{
			{"ProgressBar.foreground", muted},
			{"ProgressBar.progressCounterBackground", mineral},
			{"ProgressBar.progressCounterForeground", field},
			{"ProgressBar.indeterminateStartColor", mineral},
			{"ProgressBar.indeterminateEndColor", mineralWash},
			{"ProgressBar.passedColor", moss},
			{"ProgressBar.passedEndColor", moss},
			{"ProgressBar.failedColor", rust},
			{"ProgressBar.failedEndColor", rust},
			{"ProgressBar.passedendcolor", moss},
			{"ProgressBar.failedendcolor", rust},
			{"Badge.blueBackground", mineral},
			{"Badge.blueForeground", field},
			{"Badge.blueSecondaryBackground", mineralWash},
			{"Badge.blueSecondaryForeground", mineral},
			{"Badge.greenBackground", moss},
			{"Badge.greenForeground", field},
			{"Badge.greenSecondaryBackground", mossWash},
			{"Badge.greenSecondaryForeground", moss},
			{"Badge.purpleSecondaryBackground", mineralWash},
			{"Badge.purpleSecondaryForeground", mineral},
			{"Badge.graySecondaryBackground", structure},
			{"Badge.graySecondaryForeground", muted},
			{"Badge.disabledBackground", surface},
			{"Badge.disabledForeground", muted},
			{"Notification.errorBackground", rustWash},
			{"Notification.errorBorderColor", rust},
		},
	},
	{
		comment: "Bookmarks",
		keys: []uiKey{
			{"Bookmark.iconBackground", ochre},
			{"Bookmark.Mnemonic.iconBackground", ochreWash},
			{"Bookmark.Mnemonic.iconBorderColor", ochre},
			{"Bookmark.MnemonicCurrent.foreground", field},
			{"BookmarkMnemonicAssigned.background", ochreWash},
			{"BookmarkMnemonicAssigned.borderColor", ochre},
			{"BookmarkMnemonicAssigned.foreground", ink},
			{"BookmarkMnemonicAvailable.background", field},
			{"BookmarkMnemonicAvailable.borderColor", line},
			{"BookmarkMnemonicAvailable.foreground", ink},
			{"BookmarkMnemonicCurrent.background", mineral},
			{"BookmarkMnemonicCurrent.borderColor", mineral},
			{"BookmarkMnemonicCurrent.foreground", field},
		},
	},
	{
		comment: "Debugger, version control, and review",
		keys: []uiKey{
			{"Debugger.Variables.valueForeground", ink},
			{"Debugger.Variables.typeForeground", muted},
			{"Debugger.Variables.changedValueForeground", mineral},
			{"Debugger.Variables.modifyingValueForeground", mineral},
			{"Debugger.Variables.collectingDataForeground", muted},
			{"Debugger.Variables.evaluatingExpressionForeground", muted},
			{"Debugger.Variables.errorMessageForeground", rust},
			{"Debugger.Variables.exceptionForeground", rust},
			{"VersionControl.FileHistory.Commit.selectedBranchBackground", mineralWash},
			{"VersionControl.Log.Commit.currentBranchBackground", mineralWash},
			{"VersionControl.Log.Commit.hoveredBackground", callout},
			{"VersionControl.Log.Commit.unmatchedForeground", muted},
			{"VersionControl.Log.Commit.Reference.foreground", muted},
			{"VersionControl.RefLabel.foreground", muted},
			{"VersionControl.GitCommits.graphColor", muted},
			{"VersionControl.MarkerPopup.borderColor", line},
			{"VersionControl.Merge.Status.NoConflicts.foreground", moss},
			{"CombinedDiff.BlockBorder.selectedActiveColor", mineral},
			{"CombinedDiff.BlockBorder.selectedInactiveColor", line},
			{"Review.Branch.Background", structure},
			{"Review.Branch.Background.Hover", callout},
			{"Review.State.Background", structure},
			{"Review.State.Foreground", muted},
			{"Review.Timeline.Thread.Diff.AnchorLine", ochreWash},
		},
	},
	{
		comment: "Plugins and welcome screen",
		keys: []uiKey{
			{"Plugins.background", surface},
			{"Plugins.borderColor", line},
			{"Plugins.lightSelectionBackground", callout},
			{"Plugins.SearchField.background", field},
			{"Plugins.SectionHeader.background", structure},
			{"Plugins.SectionHeader.foreground", muted},
			{"Plugins.tagBackground", structure},
			{"Plugins.eapTagBackground", rustWash},
			{"Plugins.paidTagBackground", mineralWash},
			{"Plugins.trialTagBackground", mossWash},
			{"Plugins.Button.installBackground", field},
			{"Plugins.Button.installFocusedBackground", mossWash},
			{"WelcomeScreen.background", structure},
			{"WelcomeScreen.SidePanel.background", surface},
			{"WelcomeScreen.separatorColor", line},
			{"WelcomeScreen.Projects.actions.background", surface},
			{"WelcomeScreen.Projects.actions.selectionBackground", selection},
			{"WelcomeScreen.Projects.actions.selectionBorderColor", mineral},
			{"WelcomeScreen.Frame.DisclosureButton.defaultBackground", structure},
			{"SidePanel.background", surface},
		},
	},
	{
		comment: "Swing system colors, for components that use them directly",
		keys: []uiKey{
			{"control", surface},
			{"controlText", ink},
			{"text", field},
			{"textText", ink},
			{"textInactiveText", muted},
			{"window", field},
			{"infoText", ink},
			{"activeCaption", structure},
			{"inactiveCaption", structure},
			{"OptionPane.messageForeground", ink},
			{"TitledBorder.titleColor", ink},
		},
	},
}

// parentTheme returns the Islands theme quiet's theme inherits from.
func parentTheme(p core.Palette) (id string, dark bool, err error) {
	switch p.Mode {
	case core.ModeDay:
		return "Islands Light", false, nil

	case core.ModeNight:
		return "Islands Dark", true, nil

	default:
		return "", false, fmt.Errorf(
			"unsupported JetBrains palette mode %q",
			p.Mode,
		)
	}
}

// EncodeTheme writes the JetBrains UI theme for p.
//
// The theme is written directly rather than through encoding/json, so the
// colors stay in policy order, grouped as the policy groups them.
func EncodeTheme(
	w io.Writer,
	p core.Palette,
) error {
	resolved, err := resolve(p)
	if err != nil {
		return err
	}

	parent, dark, err := parentTheme(p)
	if err != nil {
		return err
	}

	names := make(map[role]string, len(roleNames))

	for _, n := range roleNames {
		names[n.role] = n.name
	}

	name := func(subject string, r role) (string, error) {
		n, ok := names[r]
		if !ok {
			return "", fmt.Errorf(
				"%s %s: role %d has no theme color name",
				p.Name,
				subject,
				r,
			)
		}

		return n, nil
	}

	var colors [][]entry

	var roles []entry

	for _, n := range roleNames {
		hex, err := resolved.hex(n.role)
		if err != nil {
			return fmt.Errorf(
				"%s %s: %w",
				p.Name,
				n.name,
				err,
			)
		}

		roles = append(roles, entry{n.name, "#" + strings.ToUpper(hex)})
	}

	colors = append(colors, roles)

	for _, s := range themeTokens {
		var group []entry

		for _, t := range s.tokens {
			r := t.day
			if dark {
				r = t.night
			}

			n, err := name(t.name, r)
			if err != nil {
				return err
			}

			group = append(group, entry{t.name, n})
		}

		colors = append(colors, group)
	}

	var ui [][]entry

	for _, s := range themeKeys {
		var group []entry

		for _, k := range s.keys {
			n, err := name(k.key, k.role)
			if err != nil {
				return err
			}

			group = append(group, entry{k.key, n})
		}

		ui = append(ui, group)
	}

	var b strings.Builder

	fmt.Fprintf(
		&b,
		"{\n"+
			"  \"name\": %s,\n"+
			"  \"dark\": %t,\n"+
			"  \"author\": \"Chris Mecca\",\n"+
			"  \"parentTheme\": %s,\n"+
			"  \"editorScheme\": %s,\n",
		quote(p.Name),
		dark,
		quote(parent),
		quote(p.Name),
	)

	writeObject(&b, "colors", colors)
	b.WriteString(",\n")
	writeObject(&b, "ui", ui)
	b.WriteString("\n}\n")

	if _, err := io.WriteString(
		w,
		b.String(),
	); err != nil {
		return fmt.Errorf(
			"write JetBrains theme: %w",
			err,
		)
	}

	return nil
}

type entry struct {
	key   string
	value string
}

// writeObject writes a JSON object of string values, with a blank line
// between groups.
func writeObject(
	b *strings.Builder,
	key string,
	groups [][]entry,
) {
	fmt.Fprintf(b, "  %s: {\n", quote(key))

	var lines []string

	for _, g := range groups {
		var group []string

		for _, e := range g {
			group = append(group, fmt.Sprintf(
				"    %s: %s",
				quote(e.key),
				quote(e.value),
			))
		}

		lines = append(lines, strings.Join(group, ",\n"))
	}

	b.WriteString(strings.Join(lines, ",\n\n"))
	b.WriteString("\n  }")
}

func quote(s string) string {
	b, _ := json.Marshal(s)

	return string(b)
}
