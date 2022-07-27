package oobmigration

import (
	"fmt"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

type Version struct {
	Major int
	Minor int
}

func NewVersion(major, minor int) Version {
	return Version{
		Major: major,
		Minor: minor,
	}
}

var versionPattern = lazyregexp.New(`^v?(\d+)\.(\d+)(?:\.\d+)?$`)

func NewVersionFromString(v string) (Version, bool) {
	if matches := versionPattern.FindStringSubmatch(v); len(matches) >= 3 {
		major, _ := strconv.Atoi(matches[1])
		minor, _ := strconv.Atoi(matches[2])

		return NewVersion(major, minor), true
	}

	return Version{}, false
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

type VersionOrder int

const (
	VersionOrderBefore VersionOrder = iota
	VersionOrderEqual
	VersionOrderAfter
)

// compareVersions returns the relationship between `a (op) b`.
func compareVersions(a, b Version) VersionOrder {
	for _, pair := range [][2]int{
		{a.Major, b.Major},
		{a.Minor, b.Minor},
	} {
		if pair[0] < pair[1] {
			return VersionOrderBefore
		}
		if pair[0] > pair[1] {
			return VersionOrderAfter
		}
	}

	return VersionOrderEqual
}

// pointIntersectsInterval returns true if point falls within the interval [lower, upper].
func pointIntersectsInterval(lower, upper, point Version) bool {
	return compareVersions(point, lower) != VersionOrderBefore && compareVersions(upper, point) != VersionOrderBefore
}
