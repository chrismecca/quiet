// @(#)quiet / go / internal/release/notes.go / 2026-10-26

package release

import (
	"fmt"
	"strings"
)

// Notes returns the CHANGELOG.adoc section for version as the Markdown that
// GitHub release pages render.
//
// A section runs from its "== VERSION (DATE)" heading to the next level-two
// heading. The changelog's prose, lists, and inline code read the same in
// both languages, so only the section's own headings are converted.
func Notes(changelog, version string) (string, error) {
	var (
		body  []string
		found bool
	)

	for line := range strings.SplitSeq(changelog, "\n") {
		if strings.HasPrefix(line, "== ") {
			if found {
				break
			}

			found = strings.HasPrefix(line, "== "+version+" ")
			continue
		}

		if !found {
			continue
		}

		switch {
		case strings.HasPrefix(line, "==== "):
			line = "#### " + strings.TrimPrefix(line, "==== ")
		case strings.HasPrefix(line, "=== "):
			line = "### " + strings.TrimPrefix(line, "=== ")
		}

		body = append(body, line)
	}

	if !found {
		return "", fmt.Errorf("changelog has no section for %s", version)
	}

	notes := strings.Trim(strings.Join(body, "\n"), "\n")
	if notes == "" {
		return "", fmt.Errorf("changelog section for %s is empty", version)
	}

	return notes + "\n", nil
}
