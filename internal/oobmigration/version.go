package oobmigration

import "fmt"

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
