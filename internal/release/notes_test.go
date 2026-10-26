// @(#)quiet / go / internal/release/notes_test.go / 2026-10-26

package release

import (
	"os"
	"strings"
	"testing"
)

const changelog = `= quiet changelog

Preamble.

== Unreleased

* Pending.

== 0.2.0 (2026-11-01)

Summary.

=== Ports

* Added a port,
  over two lines.

==== Detail

Text.

== 0.1.0 (2026-10-25)

* First.
`

func TestNotesConvertsSection(t *testing.T) {
	got, err := Notes(changelog, "0.2.0")
	if err != nil {
		t.Fatal(err)
	}

	want := `Summary.

### Ports

* Added a port,
  over two lines.

#### Detail

Text.
`

	if got != want {
		t.Errorf("Notes = %q, want %q", got, want)
	}
}

func TestNotesReadsLastSection(t *testing.T) {
	got, err := Notes(changelog, "0.1.0")
	if err != nil {
		t.Fatal(err)
	}

	if got != "* First.\n" {
		t.Errorf("Notes = %q, want %q", got, "* First.\n")
	}
}

func TestNotesRejectsMissingVersion(t *testing.T) {
	// A version prefix of a released section is not that section.
	for _, version := range []string{"0.3.0", "0.2", "Unreleased"} {
		if _, err := Notes(changelog, version); err == nil {
			t.Errorf("Notes accepted %q", version)
		}
	}
}

func TestNotesForVersion(t *testing.T) {
	data, err := os.ReadFile(repoPath("CHANGELOG.adoc"))
	if err != nil {
		t.Fatal(err)
	}

	notes, err := Notes(string(data), Version)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(notes, "\n=") {
		t.Errorf("notes for %s keep an AsciiDoc heading:\n%s", Version, notes)
	}
}
