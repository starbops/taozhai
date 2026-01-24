// Package version provides version information for the taozhai binary.
package version

import "fmt"

var (
	// Version is the semantic version, set via ldflags at build time.
	Version = "dev"
	// Commit is the git commit SHA, set via ldflags at build time.
	Commit = "unknown"
	// BuildDate is the build timestamp, set via ldflags at build time.
	BuildDate = "unknown"
)

// GetVersionString returns a formatted version string.
func GetVersionString() string {
	return fmt.Sprintf("taozhai %s (commit: %s, built: %s)", Version, Commit, BuildDate)
}
