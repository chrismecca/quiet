// @(#)quiet / go / internal/port/jetbrains/ide_test.go / 2026-10-25

package jetbrains

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// The tests in this file read an installed JetBrains IDE, and skip when none
// is installed. QUIET_JETBRAINS_IDE names one explicitly: a macOS .app
// bundle or an installation directory containing lib/.

// platformJar holds the IntelliJ Platform's bundled themes and schemes.
const platformJar = "intellij.platform.ide.impl.jar"

// notMapped lists the parent theme's interface keys quiet leaves to
// JetBrains, by component or key prefix. Each is a categorical color that
// identifies something the user chose, a chart or data visualization, a
// product or onboarding surface, or legacy painting the Islands themes no
// longer use.
var notMapped = []string{
	// Categorical colors.
	"ProjectGradients.",
	"RecentProject.",
	"FileColor.",
	"CodeWithMe.",
	"VersionControl.GitLog.",
	"VersionControl.HgLog.",
	"VersionControl.RefLabel.backgroundBase",

	// Charts and data visualization.
	"Profiler.",
	"CompilationCharts.",
	"LineProfiler.",
	"FlameGraph.",
	"OpenTelemetry.",
	"DataSummary.",
	"AppInspector.",
	"ScreenView.",
	"UIDesigner.",

	// Product, licensing, and onboarding surfaces.
	"TrialWidget.",
	"LicenseDialog.",
	"ManagedIde",
	"Recap.",
	"Lesson.",
	"ToolTip.Learning.",
	"Tooltip.Learning.",
	"UnattendedHostStatus.",
	"PresentationAssistant.",
	"Space.",
	"NuGet.",
	"PackageSearch.",
	"AgentPicker.",
	"HelpBrowser.",

	// Legacy painting and debugging.
	"Button.darcula.",
	"RadioButton.darcula.",
	"Spinner.darcula.",
	"Link.background",
	"Link.foreground",
	"Focus.color",
}

func TestSchemeKeysExistInIDE(t *testing.T) {
	lib := ideLib(t)

	known := make(map[string]bool)

	editorColors := ideEntry(
		t,
		filepath.Join(lib, "intellij.platform.editor.ui.jar"),
		"com/intellij/openapi/editor/colors/EditorColors.class",
	)

	for _, data := range ideSchemes(t, lib) {
		var doc struct {
			Schemes []struct {
				Colors []struct {
					Name string `xml:"name,attr"`
				} `xml:"colors>option"`
				Attributes []struct {
					Name string `xml:"name,attr"`
				} `xml:"attributes>option"`
			} `xml:"scheme"`
			Colors []struct {
				Name string `xml:"name,attr"`
			} `xml:"colors>option"`
			Attributes []struct {
				Name string `xml:"name,attr"`
			} `xml:"attributes>option"`
		}

		if err := xml.Unmarshal(data, &doc); err != nil {
			continue
		}

		for _, o := range doc.Colors {
			known["color "+o.Name] = true
		}

		for _, o := range doc.Attributes {
			known["attributes "+o.Name] = true
		}

		for _, s := range doc.Schemes {
			for _, o := range s.Colors {
				known["color "+o.Name] = true
			}

			for _, o := range s.Attributes {
				known["attributes "+o.Name] = true
			}
		}
	}

	for _, s := range schemeColors {
		for _, o := range s.options {
			// Some color keys have defaults computed in code and appear in
			// no bundled scheme, only in EditorColors.
			if !known["color "+o.name] &&
				!bytes.Contains(editorColors, []byte(o.name)) {
				t.Errorf("color key %s is unknown to the IDE", o.name)
			}
		}
	}

	for _, s := range schemeAttributes {
		for _, o := range s.options {
			if !known["attributes "+o.name] {
				t.Errorf("attributes key %s is unknown to the IDE", o.name)
			}
		}
	}
}

// Every named color and interface key quiet sets exists in the parent theme,
// so a renamed or misspelled key cannot silently do nothing.
func TestThemeKeysExistInIDE(t *testing.T) {
	lib := ideLib(t)

	colors := make(map[string]bool)
	ui := make(map[string]bool)

	for _, name := range canonical {
		parent := mergedParent(t, lib, theme(t, name))

		for k := range parent.colors {
			colors[k] = true
		}

		for k := range parent.ui {
			ui[k] = true
		}
	}

	for _, s := range themeTokens {
		for _, token := range s.tokens {
			if !colors[token.name] {
				t.Errorf("parent themes define no color %s", token.name)
			}
		}
	}

	for _, s := range themeKeys {
		for _, k := range s.keys {
			if !ui[k.key] {
				t.Errorf("parent themes define no key %s", k.key)
			}
		}
	}
}

// Every interface color resolves to a quiet role, unless the parent theme
// draws it translucent over quiet's colors or quiet leaves it to JetBrains
// by design.
func TestThemeCoversParentInterface(t *testing.T) {
	lib := ideLib(t)

	for _, name := range canonical {
		doc := theme(t, name)
		merged := mergedParent(t, lib, doc)

		for k, v := range doc.Colors {
			merged.colors[k] = v
		}

		for k, v := range doc.UI {
			merged.ui[k] = v
		}

		var leaks []string

		for key, value := range merged.ui {
			s, ok := value.(string)
			if !ok || leftToJetBrains(key) {
				continue
			}

			final, isColor := merged.resolve(s)
			if !isColor || strings.HasPrefix(final, "quiet-") {
				continue
			}

			if translucent(final) {
				continue
			}

			leaks = append(leaks, key+" = "+final)
		}

		sort.Strings(leaks)

		for _, leak := range leaks {
			t.Errorf("%s: %s", name, leak)
		}
	}
}

func leftToJetBrains(key string) bool {
	for _, prefix := range notMapped {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}

	return false
}

// translucent reports whether a #RRGGBBAA color has partial alpha.
func translucent(value string) bool {
	return len(value) == 9 &&
		strings.HasPrefix(value, "#") &&
		!strings.EqualFold(value[7:], "ff")
}

type mergedTheme struct {
	colors map[string]string
	ui     map[string]any
}

// resolve follows named colors to a quiet role name or a color literal.
func (m mergedTheme) resolve(value string) (string, bool) {
	for range len(m.colors) + 1 {
		if strings.HasPrefix(value, "quiet-") || strings.HasPrefix(value, "#") {
			return value, true
		}

		next, ok := m.colors[value]
		if !ok {
			return value, false
		}

		value = next
	}

	return value, false
}

// mergedParent loads the IDE theme doc inherits from and that theme's own
// ancestors, merged the way the IntelliJ Platform merges them: ancestors
// first, each descendant overriding its parent's colors and keys.
func mergedParent(t *testing.T, lib string, doc themeDocument) mergedTheme {
	t.Helper()

	jar := filepath.Join(lib, platformJar)
	paths := themePaths(t, jar)

	var chain []map[string]any

	for id := doc.ParentTheme; id != ""; {
		path, ok := paths[id]
		if !ok {
			t.Fatalf("IDE registers no theme %q", id)
		}

		var file map[string]any

		if err := json.Unmarshal(ideEntry(t, jar, path), &file); err != nil {
			t.Fatalf("%s: %v", path, err)
		}

		chain = append(chain, file)
		id = parentOf(id, file)
	}

	merged := mergedTheme{
		colors: make(map[string]string),
		ui:     make(map[string]any),
	}

	for i := len(chain) - 1; i >= 0; i-- {
		if colors, ok := chain[i]["colors"].(map[string]any); ok {
			for k, v := range colors {
				if s, ok := v.(string); ok {
					merged.colors[k] = s
				}
			}
		}

		if ui, ok := chain[i]["ui"].(map[string]any); ok {
			flattenUI("", ui, merged.ui)
		}
	}

	return merged
}

// parentOf returns the theme id inherits from, following
// UiThemeProviderListManager: IntelliJ inherits from Darcula, Darcula from
// nothing, and a theme without parentTheme from Darcula or IntelliJ.
func parentOf(id string, file map[string]any) string {
	if parent, ok := file["parentTheme"].(string); ok && parent != "" {
		return parent
	}

	switch id {
	case "Darcula":
		return ""

	case "IntelliJ":
		return "Darcula"
	}

	if dark, _ := file["dark"].(bool); dark {
		return "Darcula"
	}

	return "IntelliJ"
}

// flattenUI flattens nested interface keys the way UIThemeBean does,
// including per-OS values: the current OS's value, else os.default.
func flattenUI(prefix string, ui map[string]any, out map[string]any) {
	osKey := map[string]string{
		"darwin":  "os.mac",
		"windows": "os.windows",
	}[runtime.GOOS]

	if osKey == "" {
		osKey = "os.linux"
	}

	for k, v := range ui {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		nested, ok := v.(map[string]any)
		if !ok {
			out[key] = v

			continue
		}

		if !perOS(nested) {
			flattenUI(key, nested, out)

			continue
		}

		if value, ok := nested[osKey]; ok {
			out[key] = value
		} else if value, ok := nested["os.default"]; ok {
			out[key] = value
		}
	}
}

// perOS reports whether a nested value holds per-OS alternatives.
func perOS(nested map[string]any) bool {
	for k := range nested {
		if strings.HasPrefix(k, "os.") {
			return true
		}
	}

	return false
}

// themePaths maps the IDE's registered theme ids to their paths in jar.
func themePaths(t *testing.T, jar string) map[string]string {
	t.Helper()

	var doc struct {
		Extensions []struct {
			Themes []struct {
				ID   string `xml:"id,attr"`
				Path string `xml:"path,attr"`
			} `xml:"themeProvider"`
		} `xml:"extensions"`
	}

	if err := xml.Unmarshal(
		ideEntry(t, jar, "META-INF/PlatformExtensions.xml"),
		&doc,
	); err != nil {
		t.Fatal(err)
	}

	paths := make(map[string]string)

	for _, e := range doc.Extensions {
		for _, theme := range e.Themes {
			paths[theme.ID] = strings.TrimPrefix(theme.Path, "/")
		}
	}

	return paths
}

// ideSchemes returns every editor scheme file bundled with the IDE and its
// plugins.
func ideSchemes(t *testing.T, lib string) [][]byte {
	t.Helper()

	jars, _ := filepath.Glob(filepath.Join(lib, "*.jar"))
	plugins, _ := filepath.Glob(filepath.Join(lib, "..", "plugins", "*", "lib", "*.jar"))

	var schemes [][]byte

	for _, jar := range append(jars, plugins...) {
		r, err := zip.OpenReader(jar)
		if err != nil {
			continue
		}

		for _, f := range r.File {
			if !strings.HasSuffix(f.Name, ".xml") {
				continue
			}

			if !strings.Contains(f.Name, "colorSchemes/") &&
				!strings.Contains(f.Name, "themes/") &&
				f.Name != "DefaultColorSchemesManager.xml" {
				continue
			}

			schemes = append(schemes, readZipFile(t, f))
		}

		r.Close()
	}

	return schemes
}

func ideEntry(t *testing.T, jar string, name string) []byte {
	t.Helper()

	r, err := zip.OpenReader(jar)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == name {
			return readZipFile(t, f)
		}
	}

	t.Fatalf("%s has no %s", jar, name)

	return nil
}

// ideLib returns the lib directory of an installed JetBrains IDE, or skips.
func ideLib(t *testing.T) string {
	t.Helper()

	var candidates []string

	if dir := os.Getenv("QUIET_JETBRAINS_IDE"); dir != "" {
		candidates = append(candidates, dir)
	}

	if runtime.GOOS == "darwin" {
		home, _ := os.UserHomeDir()

		for _, apps := range []string{
			filepath.Join(home, "Applications"),
			"/Applications",
		} {
			bundles, _ := filepath.Glob(filepath.Join(apps, "*.app"))
			candidates = append(candidates, bundles...)
		}
	}

	for _, dir := range candidates {
		for _, lib := range []string{
			filepath.Join(dir, "Contents", "lib"),
			filepath.Join(dir, "lib"),
		} {
			if _, err := os.Stat(filepath.Join(lib, platformJar)); err == nil {
				return lib
			}
		}
	}

	t.Skip("no JetBrains IDE installed")

	return ""
}
