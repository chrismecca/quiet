// @(#)quiet / go / internal/port/dircolors/dircolors_test.go / 2026-10-18

package dircolors

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestCanonicalDircolorsMatchesCommittedArtifact(t *testing.T) {
	var got bytes.Buffer

	if err := Encode(
		&got,
	); err != nil {
		t.Fatalf(
			"Encode: %v",
			err,
		)
	}

	want, err := os.ReadFile(
		repoPath(
			t,
			"ports",
			"dircolors",
			"quiet.dircolors",
		),
	)
	if err != nil {
		t.Fatalf(
			"ReadFile: %v",
			err,
		)
	}

	if !bytes.Equal(
		got.Bytes(),
		want,
	) {
		t.Fatal(
			"generated dircolors differs from ports/dircolors/quiet.dircolors",
		)
	}
}

func TestSemanticSlotAssignments(t *testing.T) {
	want := map[string]string{
		"RESET":                 "0",
		"NORMAL":                "0",
		"FILE":                  "0",
		"MULTIHARDLINK":         "0",
		"DIR":                   "34",
		"LINK":                  "36",
		"EXEC":                  "32",
		"FIFO":                  "33",
		"SOCK":                  "35",
		"DOOR":                  "35",
		"BLK":                   "33",
		"CHR":                   "33",
		"ORPHAN":                "31",
		"MISSING":               "31",
		"SETUID":                "01;31",
		"SETGID":                "01;31",
		"CAPABILITY":            "01;31",
		"STICKY_OTHER_WRITABLE": "01;33",
		"OTHER_WRITABLE":        "01;31",
		"STICKY":                "33",
	}

	for _, group := range groups {
		for _, entry := range group.entries {
			expected, ok := want[entry.keyword]
			if !ok {
				t.Errorf(
					"unexpected dircolors keyword %q",
					entry.keyword,
				)
				continue
			}

			if entry.sgr != expected {
				t.Errorf(
					"%s = %q, want %q",
					entry.keyword,
					entry.sgr,
					expected,
				)
			}

			delete(
				want,
				entry.keyword,
			)
		}
	}

	for keyword := range want {
		t.Errorf(
			"missing dircolors keyword %q",
			keyword,
		)
	}
}

func TestOrdinaryFilesRemainUncolored(t *testing.T) {
	for _, keyword := range []string{
		"NORMAL",
		"FILE",
		"MULTIHARDLINK",
	} {
		sgr, ok := lookup(keyword)
		if !ok {
			t.Fatalf(
				"%s is missing",
				keyword,
			)
		}

		if sgr != "0" {
			t.Errorf(
				"%s = %q, want terminal default",
				keyword,
				sgr,
			)
		}
	}
}

func TestNoExtensionRules(t *testing.T) {
	for _, group := range groups {
		for _, entry := range group.entries {
			if strings.HasPrefix(
				entry.keyword,
				".",
			) || strings.HasPrefix(
				entry.keyword,
				"*",
			) {
				t.Errorf(
					"dircolors must not classify filename extension %q",
					entry.keyword,
				)
			}
		}
	}
}

func TestNoBackgroundColors(t *testing.T) {
	for _, group := range groups {
		for _, entry := range group.entries {
			for _, field := range strings.Split(
				entry.sgr,
				";",
			) {
				value, err := strconv.Atoi(field)
				if err != nil {
					t.Fatalf(
						"%s contains invalid SGR component %q",
						entry.keyword,
						field,
					)
				}

				if value >= 40 && value <= 47 {
					t.Errorf(
						"%s uses background SGR %d",
						entry.keyword,
						value,
					)
				}

				if value >= 100 && value <= 107 {
					t.Errorf(
						"%s uses bright background SGR %d",
						entry.keyword,
						value,
					)
				}
			}
		}
	}
}

func lookup(
	keyword string,
) (string, bool) {
	for _, group := range groups {
		for _, entry := range group.entries {
			if entry.keyword == keyword {
				return entry.sgr, true
			}
		}
	}

	return "", false
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

	parts := append(
		[]string{root},
		elem...,
	)

	return filepath.Join(parts...)
}

func TestPermissionStatesAvoidColoredBackgrounds(t *testing.T) {
	data, err := os.ReadFile(
		repoPath(
			t,
			"ports",
			"dircolors",
			"quiet.dircolors",
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	got := string(data)

	want := []string{
		"SETUID 01;31",
		"SETGID 01;31",
		"CAPABILITY 01;31",
		"STICKY_OTHER_WRITABLE 01;33",
		"OTHER_WRITABLE 01;31",
		"STICKY 33",
	}

	for _, line := range want {
		if !strings.Contains(
			got,
			line+"\n",
		) {
			t.Errorf(
				"dircolors permission-state mapping missing %q",
				line,
			)
		}
	}
}

func TestPermissionStatesContainNoBackgroundSGR(t *testing.T) {
	data, err := os.ReadFile(
		repoPath(
			t,
			"ports",
			"dircolors",
			"quiet.dircolors",
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	sensitive := map[string]bool{
		"SETUID":                true,
		"SETGID":                true,
		"CAPABILITY":            true,
		"STICKY_OTHER_WRITABLE": true,
		"OTHER_WRITABLE":        true,
		"STICKY":                true,
	}

	for _, line := range strings.Split(
		string(data),
		"\n",
	) {
		fields := strings.Fields(line)

		if len(fields) != 2 ||
			!sensitive[fields[0]] {
			continue
		}

		for _, sgr := range strings.Split(
			fields[1],
			";",
		) {
			switch sgr {
			case "40", "41", "42", "43",
				"44", "45", "46", "47",
				"100", "101", "102", "103",
				"104", "105", "106", "107":
				t.Errorf(
					"%s uses forbidden ANSI background SGR %s in %q",
					fields[0],
					sgr,
					line,
				)
			}
		}
	}
}
