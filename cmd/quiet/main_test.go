// @(#)quiet / go / cmd/quiet/main_test.go / 2026-10-26

package main

import (
	"bytes"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/chrismecca/quiet/internal/release"
)

// execute runs the command with args and returns its exit status and output.
func execute(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	status := run(args, &stdout, &stderr)
	return status, stdout.String(), stderr.String()
}

func TestVersion(t *testing.T) {
	status, stdout, _ := execute("version")
	if status != exitOK {
		t.Fatalf("status = %d, want %d", status, exitOK)
	}

	if stdout != release.Version+"\n" {
		t.Errorf("version = %q, want %q", stdout, release.Version+"\n")
	}
}

func TestReleaseNotes(t *testing.T) {
	changelog := filepath.Join("..", "..", "CHANGELOG.adoc")

	status, stdout, stderr := execute("release-notes", changelog, release.Version)
	if status != exitOK {
		t.Fatalf("status = %d, want %d: %s", status, exitOK, stderr)
	}

	if stdout == "" {
		t.Error("release notes are empty")
	}

	status, _, _ = execute("release-notes", changelog, "0.0.0")
	if status != exitError {
		t.Errorf("status for a missing version = %d, want %d", status, exitError)
	}
}

func TestUsageErrors(t *testing.T) {
	for _, args := range [][]string{
		nil,
		{"unknown"},
		{"version", "extra"},
		{"release-notes", "CHANGELOG.adoc"},
		{"port"},
		{"port", "unknown"},
		{"port", "zed", "one-palette.json"},
	} {
		if status, _, _ := execute(args...); status != exitUsage {
			t.Errorf("%q: status = %d, want %d", args, status, exitUsage)
		}
	}
}

func TestPortCommandsAreAlphabetic(t *testing.T) {
	names := make([]string, 0, len(portCommands))
	for _, command := range portCommands {
		names = append(names, command.name)
	}

	if !slices.IsSorted(names) {
		t.Errorf("port commands are not alphabetic: %s", strings.Join(names, " "))
	}
}
