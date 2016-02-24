// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Version represents a semantic version.
type Version struct {
	Major, Minor, Patch int

	// If true, then this is the unstable version.
	Unstable bool
}

// String returns a string representation of this version, for example:
//
//  Version{Major=1, Minor=2, Patch=3}                -> "v1.2.3"
//  Version{Major=1, Minor=2, Patch=3, Unstable=true} -> "v1.2.3-unstable"
//
//  Version{Major=1, Minor=2, Patch=-1}                -> "v1.2"
//  Version{Major=1, Minor=2, Patch=-1, Unstable=true} -> "v1.2-unstable"
//
//  Version{Major=1, Minor=-1, Patch=-1}                -> "v1"
//  Version{Major=1, Minor=-1, Patch=-1, Unstable=true} -> "v1-unstable"
//
func (v Version) String() string {
	var s string
	if v.Major > 0 && v.Minor > 0 && v.Patch > 0 {
		s = fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
	} else if v.Major > 0 && v.Minor > 0 {
		s = fmt.Sprintf("v%d.%d", v.Major, v.Minor)
	} else if v.Major > 0 {
		s = fmt.Sprintf("v%d", v.Major)
	} else {
		return fmt.Sprintf("Version{Major=%d, Minor=%d, Patch=%d, Unstable=%t}", v.Major, v.Minor, v.Patch, v.Unstable)
	}
	if v.Unstable {
		return s + "-unstable"
	}
	return s
}

// Less tells if v is a lesser version than the other version.
//
// It follows semver specification (e.g. v1.200.300 is less than v2). A
// unstable version is *always* less than a stable version (e.g. v3-unstable is
// less than v2).
func (v Version) Less(other Version) bool {
	if v.Unstable && !other.Unstable {
		return true
	} else if other.Unstable && !v.Unstable {
		return false
	}

	if v.Major < other.Major {
		return true
	} else if v.Major > other.Major {
		return false
	}

	if v.Minor < other.Minor {
		return true
	} else if v.Minor > other.Minor {
		return false
	}

	if v.Patch < other.Patch {
		return true
	} else if v.Patch > other.Patch {
		return false
	}
	return false
}

// InvalidVersion represents a completely invalid version.
var InvalidVersion = Version{
	Major:    -1,
	Minor:    -1,
	Patch:    -1,
	Unstable: false,
}

// Matches strings like "1", "1.1", and "1.1.1".
var vsRegexp = regexp.MustCompile(`^([0-9]+)[\.]?([0-9]*)[\.]?([0-9]*)`)

// ParseVersion parses a version string in the form of:
//
//  "v1"
//  "v1.2"
//  "v1.2.1"
//  "v1-unstable"
//  "v1.2-unstable"
//  "v1.2.1-unstable"
//
// It returns InvalidVersion for strings not suffixed with "v", like:
//
//  "1"
//  "1.2-unstable"
//
func ParseVersion(vs string) Version {
	if vs[0] != 'v' {
		return InvalidVersion
	}
	vs = vs[1:] // Strip prefixed v

	// Split by the dash seperated suffix. We expect only one dash suffix, and
	// if present it must be "unstable".
	dashSplit := strings.Split(vs, "-")
	if len(dashSplit) > 2 || len(dashSplit) == 2 && dashSplit[1] != "unstable" {
		return InvalidVersion
	}

	// We now use regexp to match the last part of the version string, which
	// e.g. looks like "1", "1.1", or "1.1.1".
	var (
		m = vsRegexp.FindStringSubmatch(dashSplit[0])
		v = InvalidVersion
	)
	if len(m) > 1 && len(m[1]) > 0 {
		v.Major, _ = strconv.Atoi(m[1])
	}
	if len(m) > 2 && len(m[2]) > 0 {
		v.Minor, _ = strconv.Atoi(m[2])
	}
	if len(m) > 3 && len(m[3]) > 0 {
		v.Patch, _ = strconv.Atoi(m[3])
	}
	if v.Major != -1 && len(dashSplit) == 2 && dashSplit[1] == "unstable" {
		v.Unstable = true
	}
	return v
}
