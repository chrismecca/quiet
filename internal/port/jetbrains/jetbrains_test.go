// @(#)quiet / go / internal/port/jetbrains/jetbrains_test.go / 2026-10-25

package jetbrains

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chrismecca/quiet/internal/color"
	"github.com/chrismecca/quiet/internal/core"
)

var canonical = []string{
	"quiet-day",
	"quiet-night",
}

func TestCanonicalPalettesMatchCommittedArtifacts(t *testing.T) {
	for _, name := range canonical {
		p := loadPalette(t, name)

		t.Run(name+".icls", func(t *testing.T) {
			assertCommitted(t, encodeWith(t, EncodeScheme, p), name+".icls")
		})

		t.Run(name+".theme.json", func(t *testing.T) {
			assertCommitted(t, encodeWith(t, EncodeTheme, p), name+".theme.json")
		})
	}

	t.Run("quiet.jar", func(t *testing.T) {
		assertCommitted(t, plugin(t), "quiet.jar")
	})
}

type schemeDocument struct {
	XMLName xml.Name `xml:"scheme"`
	Name    string   `xml:"name,attr"`
	Version string   `xml:"version,attr"`
	Parent  string   `xml:"parent_scheme,attr"`
	Colors  []struct {
		Name string `xml:"name,attr"`
	} `xml:"colors>option"`
	Attributes []struct {
		Name string `xml:"name,attr"`
	} `xml:"attributes>option"`
}

func TestSchemeIsWellFormedAndUnique(t *testing.T) {
	for _, name := range canonical {
		var doc schemeDocument

		if err := xml.Unmarshal(
			encodeWith(t, EncodeScheme, loadPalette(t, name)),
			&doc,
		); err != nil {
			t.Fatalf("%s: %v", name, err)
		}

		if doc.Name != name {
			t.Errorf("%s: scheme name %q", name, doc.Name)
		}

		seen := make(map[string]bool)

		for _, o := range doc.Colors {
			if seen["color "+o.Name] {
				t.Errorf("%s: duplicate color %s", name, o.Name)
			}

			seen["color "+o.Name] = true
		}

		for _, o := range doc.Attributes {
			if seen["attributes "+o.Name] {
				t.Errorf("%s: duplicate attributes %s", name, o.Name)
			}

			seen["attributes "+o.Name] = true
		}
	}
}

// Source text is drawn on Field and on the backgrounds the scheme lays under
// it, such as search results and diff lines. Every source foreground must
// reach 4.5:1 on each of them.
//
// The caret row is excluded: it uses Callout, where quiet-core guarantees
// Muted, as in every quiet editor port.
func TestSourceTextIsReadable(t *testing.T) {
	foregrounds := make(map[role]string)
	backgrounds := map[role]string{field: "TEXT"}

	var pairs []attributeOption

	for _, s := range schemeAttributes {
		for _, o := range s.options {
			a := o.attributes

			switch {
			// The console's ANSI keys use one color for both ANSI
			// foreground and background sequences.
			case a.fg != none && a.fg == a.bg:

			// Coverage keys color gutter markers, not text.
			case strings.HasPrefix(o.name, "LINE_"):

			case a.fg != none && a.bg != none:
				pairs = append(pairs, o)

			case a.fg != none:
				foregrounds[a.fg] = o.name

			case a.bg != none:
				backgrounds[a.bg] = o.name
			}
		}
	}

	for _, name := range canonical {
		resolved := resolved(t, name)

		for fg, fgKey := range foregrounds {
			for bg, bgKey := range backgrounds {
				ratio := color.Contrast(resolved[fg], resolved[bg])

				if ratio < 4.5 {
					t.Errorf(
						"%s: %s on %s = %.2f, want at least 4.5",
						name,
						fgKey,
						bgKey,
						ratio,
					)
				}
			}
		}

		for _, o := range pairs {
			ratio := color.Contrast(
				resolved[o.attributes.fg],
				resolved[o.attributes.bg],
			)

			if ratio < 4.5 {
				t.Errorf(
					"%s: %s = %.2f, want at least 4.5",
					name,
					o.name,
					ratio,
				)
			}
		}
	}
}

// Selected text is forced to Ink, because quiet-core guarantees only Ink on
// Selection.
func TestSelectedTextIsInk(t *testing.T) {
	for _, s := range schemeColors {
		for _, o := range s.options {
			if o.name == "SELECTION_FOREGROUND" && o.role != ink {
				t.Fatalf("SELECTION_FOREGROUND is role %d, want Ink", o.role)
			}
		}
	}
}

type themeDocument struct {
	Name         string            `json:"name"`
	Dark         bool              `json:"dark"`
	ParentTheme  string            `json:"parentTheme"`
	EditorScheme string            `json:"editorScheme"`
	Colors       map[string]string `json:"colors"`
	UI           map[string]string `json:"ui"`
}

func TestThemeNamesQuietScheme(t *testing.T) {
	for _, name := range canonical {
		doc := theme(t, name)

		if doc.Name != name || doc.EditorScheme != name {
			t.Errorf(
				"%s: name %q, editor scheme %q",
				name,
				doc.Name,
				doc.EditorScheme,
			)
		}

		if doc.Dark != (name == "quiet-night") {
			t.Errorf("%s: dark = %t", name, doc.Dark)
		}
	}
}

// Every color the theme assigns is one of quiet's named roles.
func TestThemeUsesOnlyQuietRoles(t *testing.T) {
	for _, name := range canonical {
		doc := theme(t, name)

		for key, value := range doc.Colors {
			if strings.HasPrefix(key, "quiet-") {
				continue
			}

			if !strings.HasPrefix(value, "quiet-") {
				t.Errorf("%s: color %s = %q", name, key, value)
			}
		}

		for key, value := range doc.UI {
			if !strings.HasPrefix(value, "quiet-") {
				t.Errorf("%s: ui %s = %q", name, key, value)
			}
		}
	}
}

// Interface text must reach 4.5:1 on the surfaces the Islands themes draw
// it on.
func TestThemeTextIsReadable(t *testing.T) {
	texts := []string{
		"text-default",
		"text-muted",
		"text-secondary",
		"text-link",
		"text-error",
		"text-warning",
		"text-success",
	}

	surfaces := []string{
		"editor-bg",
		"tool-window-bg",
		"main-window-bg",
		"popup-bg",
		"dialog-bg",
		"control-bg",
	}

	pairs := [][2]string{
		{"text-over-accent", "accent-brand-bg"},
		{"text-default", "selection-bg-active"},
		{"text-default", "selection-bg-inactive"},
		{"text-default", "tab-selected-bg-active"},
		{"text-default", "feedback-brand-bg"},
		{"text-default", "feedback-success-bg"},
		{"text-default", "feedback-warning-bg"},
		{"text-default", "feedback-error-bg"},
	}

	for _, text := range texts {
		for _, surface := range surfaces {
			pairs = append(pairs, [2]string{text, surface})
		}
	}

	for _, name := range canonical {
		doc := theme(t, name)

		for _, pair := range pairs {
			ratio := color.Contrast(
				themeColor(t, doc, pair[0]),
				themeColor(t, doc, pair[1]),
			)

			if ratio < 4.5 {
				t.Errorf(
					"%s: %s on %s = %.2f, want at least 4.5",
					name,
					pair[0],
					pair[1],
					ratio,
				)
			}
		}
	}
}

type pluginDocument struct {
	ID         string `xml:"id"`
	Extensions struct {
		Themes []struct {
			ID   string `xml:"id,attr"`
			Path string `xml:"path,attr"`
		} `xml:"themeProvider"`
		Schemes []struct {
			ID   string `xml:"id,attr"`
			Path string `xml:"path,attr"`
		} `xml:"bundledColorScheme"`
	} `xml:"extensions"`
}

// The plugin registers each theme and scheme at a path the archive contains,
// with the same content as the standalone files.
func TestPluginRegistersPackagedFiles(t *testing.T) {
	data := plugin(t)

	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}

	files := make(map[string][]byte)

	for _, f := range archive.File {
		files[f.Name] = readZipFile(t, f)
	}

	var doc pluginDocument

	if err := xml.Unmarshal(files["META-INF/plugin.xml"], &doc); err != nil {
		t.Fatalf("plugin.xml: %v", err)
	}

	if doc.ID != pluginID {
		t.Errorf("plugin id %q, want %q", doc.ID, pluginID)
	}

	want := make(map[string][]byte)

	for _, name := range canonical {
		p := loadPalette(t, name)
		want["themes/"+name+".theme.json"] = encodeWith(t, EncodeTheme, p)
		want["colors/"+name+".xml"] = encodeWith(t, EncodeScheme, p)
	}

	registered := 0

	for _, theme := range doc.Extensions.Themes {
		registered++
		assertPackaged(t, files, want, strings.TrimPrefix(theme.Path, "/"))
	}

	for _, scheme := range doc.Extensions.Schemes {
		registered++
		assertPackaged(t, files, want, strings.TrimPrefix(scheme.Path, "/")+".xml")
	}

	if registered != len(want) {
		t.Errorf("plugin registers %d files, want %d", registered, len(want))
	}
}

func assertPackaged(
	t *testing.T,
	files map[string][]byte,
	want map[string][]byte,
	path string,
) {
	t.Helper()

	got, ok := files[path]
	if !ok {
		t.Errorf("plugin registers %s, which the archive lacks", path)

		return
	}

	if !bytes.Equal(got, want[path]) {
		t.Errorf("packaged %s differs from generator output", path)
	}
}

func readZipFile(t *testing.T, f *zip.File) []byte {
	t.Helper()

	r, err := f.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	return data
}

func loadPalette(t *testing.T, name string) core.Palette {
	t.Helper()

	p, err := core.LoadPaletteFile(
		repoPath(
			t,
			"spec",
			name+".json",
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	return p
}

func resolved(t *testing.T, name string) palette {
	t.Helper()

	r, err := resolve(loadPalette(t, name))
	if err != nil {
		t.Fatal(err)
	}

	return r
}

func encodeWith(
	t *testing.T,
	encode func(io.Writer, core.Palette) error,
	p core.Palette,
) []byte {
	t.Helper()

	var b bytes.Buffer

	if err := encode(&b, p); err != nil {
		t.Fatal(err)
	}

	return b.Bytes()
}

func plugin(t *testing.T) []byte {
	t.Helper()

	var b bytes.Buffer

	if err := EncodePlugin(
		&b,
		loadPalette(t, "quiet-day"),
		loadPalette(t, "quiet-night"),
	); err != nil {
		t.Fatal(err)
	}

	return b.Bytes()
}

func theme(t *testing.T, name string) themeDocument {
	t.Helper()

	var doc themeDocument

	if err := json.Unmarshal(
		encodeWith(t, EncodeTheme, loadPalette(t, name)),
		&doc,
	); err != nil {
		t.Fatalf("%s: theme is not valid JSON: %v", name, err)
	}

	return doc
}

// themeColor resolves a named color through the theme's own color map.
func themeColor(t *testing.T, doc themeDocument, name string) color.SRGB {
	t.Helper()

	value := name

	for range len(doc.Colors) {
		next, ok := doc.Colors[value]
		if !ok {
			break
		}

		value = next
	}

	c, err := color.ParseHex(value)
	if err != nil {
		t.Fatalf("%s: %s resolves to %q: %v", doc.Name, name, value, err)
	}

	return c
}

func assertCommitted(t *testing.T, got []byte, file string) {
	t.Helper()

	want, err := os.ReadFile(
		repoPath(
			t,
			"ports",
			"jetbrains",
			file,
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, want) {
		t.Fatalf("generated output differs from %s", file)
	}
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
