// @(#)quiet / go / internal/port/jetbrains/plugin.go / 2026-10-25

package jetbrains

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/chrismecca/quiet/internal/core"
	"github.com/chrismecca/quiet/internal/release"
)

// pluginID identifies the plugin to JetBrains IDEs and the JetBrains
// Marketplace. It must not change once the plugin is published.
const pluginID = "com.chrismecca.quiet"

// sinceBuild is the IntelliJ Platform release the themes are verified
// against: 2026.2, which ships the Islands themes they inherit from.
const sinceBuild = "262"

// archiveTime is the modification time of every plugin entry, so the archive
// is reproducible.
var archiveTime = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)

const pluginXML = `<idea-plugin>
  <id>` + pluginID + `</id>
  <name>quiet</name>
  <version>` + release.Version + `</version>
  <vendor email="cm@chrismecca.com" url="https://github.com/chrismecca/quiet">Chris Mecca</vendor>
  <description>A restrained color system for long-duration work. Provides the quiet-day and quiet-night interface themes and editor color schemes.</description>
  <idea-version since-build="` + sinceBuild + `"/>
  <depends>com.intellij.modules.platform</depends>

  <extensions defaultExtensionNs="com.intellij">
    <themeProvider id="quiet-day" path="/themes/quiet-day.theme.json" targetUi="islands"/>
    <themeProvider id="quiet-night" path="/themes/quiet-night.theme.json" targetUi="islands"/>
    <bundledColorScheme id="quiet-day" path="/colors/quiet-day"/>
    <bundledColorScheme id="quiet-night" path="/colors/quiet-night"/>
  </extensions>
</idea-plugin>
`

type archiveEntry struct {
	name   string
	encode func(io.Writer) error
}

// EncodePlugin writes the JetBrains plugin archive containing the quiet-day
// and quiet-night UI themes and editor color schemes.
//
// Entries are stored uncompressed with a fixed modification time, so the
// archive bytes depend only on its contents.
func EncodePlugin(
	w io.Writer,
	day core.Palette,
	night core.Palette,
) error {
	if day.Mode != core.ModeDay || night.Mode != core.ModeNight {
		return fmt.Errorf(
			"JetBrains palettes are %q and %q, want %q and %q",
			day.Mode,
			night.Mode,
			core.ModeDay,
			core.ModeNight,
		)
	}

	entries := []archiveEntry{
		{
			"META-INF/plugin.xml",
			func(w io.Writer) error {
				_, err := io.WriteString(w, pluginXML)

				return err
			},
		},
	}

	for _, p := range []core.Palette{day, night} {
		entries = append(
			entries,
			archiveEntry{
				"themes/" + p.Name + ".theme.json",
				func(w io.Writer) error {
					return EncodeTheme(w, p)
				},
			},
			archiveEntry{
				"colors/" + p.Name + ".xml",
				func(w io.Writer) error {
					return EncodeScheme(w, p)
				},
			},
		)
	}

	var archive bytes.Buffer

	zw := zip.NewWriter(&archive)

	for _, e := range entries {
		f, err := zw.CreateHeader(&zip.FileHeader{
			Name:     e.name,
			Method:   zip.Store,
			Modified: archiveTime,
		})
		if err != nil {
			return fmt.Errorf(
				"JetBrains plugin %s: %w",
				e.name,
				err,
			)
		}

		if err := e.encode(f); err != nil {
			return fmt.Errorf(
				"JetBrains plugin %s: %w",
				e.name,
				err,
			)
		}
	}

	if err := zw.Close(); err != nil {
		return fmt.Errorf(
			"close JetBrains plugin: %w",
			err,
		)
	}

	if _, err := w.Write(archive.Bytes()); err != nil {
		return fmt.Errorf(
			"write JetBrains plugin: %w",
			err,
		)
	}

	return nil
}
