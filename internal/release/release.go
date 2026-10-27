// @(#)quiet / go / internal/release/release.go / 2026-10-27

// Package release records the quiet release that the repository describes.
//
// Generated ports that carry a version, such as browser and IDE manifests,
// take it from Version. Authored manifests repeat it, and the package tests
// check that they agree.
package release

// Version is the current quiet release. It covers the whole repository:
// the specification, the canonical palettes, the reference implementation,
// and the first-party ports.
const Version = "0.6.0"
