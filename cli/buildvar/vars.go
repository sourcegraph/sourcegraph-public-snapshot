// Package buildvar contains variables that are set at build time by
// `-ldflags -X ...` flags.
package buildvar

var (
	// Version of the current release. For built releases, this is set
	// at build time by sgtool. Otherwise it is "dev".
	//
	// Version is the only build variable that is exported directly
	// (all others are only accessible via All). This is because it is
	// intended to be part of the published API usable by other
	// packages, unlike the other vars.
	Version = "dev"
)
