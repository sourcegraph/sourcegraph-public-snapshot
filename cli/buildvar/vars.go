// Package buildvar contains variables that are set at build time by
// `-ldflags -X ...` flags.
package buildvar

import "time"

var (
	// Version of the current release. For built releases, this is set
	// at build time by sgtool. Otherwise it is "dev".
	//
	// Version is the only build variable that is exported directly
	// (all others are only accessible via All). This is because it is
	// intended to be part of the published API usable by other
	// packages, unlike the other vars.
	Version = "dev"

	// dateStr is the date when this release was built (or the zero
	// time.Time for development builds).
	dateStr string

	// commitID is the commit ID that this release was built from, or
	// "" for development builds.
	commitID string

	// branch is the branch that this release was built from, or
	// "" for development builds.
	branch string

	// dirtyStr is a nonempty string (typically "true") if the git
	// index or working tree has uncommitted files (i.e., if `git
	// status --porcelain` is nonempty), when the release was
	// built. It indicates that the build's CommitID might not fully
	// reflect the files from which this release was built.
	dirtyStr string

	// host is the `uname -a` output of the host machine that compiled
	// this release, or "" for development builds.
	host string

	// user is the login of the user who build this release, or "" for
	// development builds.
	user string
)

// Vars holds the parsed build variables (set at build time as
// strings).
//
// These values are currently used for debugging purposes only and are
// not intended to be part of the published API; therefore, they are
// not documented. You should not rely on any of these values. For
// more information about each field, view the source code for the
// buildvar package.
type Vars struct {
	Version  string    `json:",omitempty"`
	Date     time.Time `json:",omitempty"`
	CommitID string    `json:",omitempty"`
	Branch   string    `json:",omitempty"`
	Dirty    bool      `json:",omitempty"`
	Host     string    `json:",omitempty"`
	User     string    `json:",omitempty"`
}

// All holds the parsed values of all build variables. Some will be
// empty for development builds.
var All = parse()

// Public holds the parsed values of the useful build variables which security
// researches won't mistake for something important.
var Public = sanitize(All)

func parse() Vars {
	date, _ := time.Parse(time.UnixDate, dateStr)
	return Vars{
		Version:  Version,
		Date:     date,
		CommitID: commitID,
		Branch:   branch,
		Dirty:    dirtyStr != "",
		Host:     host,
		User:     user,
	}
}

func sanitize(v Vars) Vars {
	return Vars{
		Version:  v.Version,
		Date:     v.Date,
		CommitID: v.CommitID,
		Branch:   v.Branch,
	}
}
