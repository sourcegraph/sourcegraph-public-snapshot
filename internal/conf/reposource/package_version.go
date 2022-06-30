package reposource

import (
	"strconv"
	"unicode"
)

// PackageVersion is a Package that additionally includes a concrete version.
// The version must be a concrete version, it cannot be a version range.
type PackageVersion interface {
	Package

	// The version of the package.
	PackageVersion() string

	// Returns the git tag associated with the given dependency version, used
	// rev: or repo:foo@rev
	GitTagFromVersion() string

	// PackageVersionSyntax is the string-formatted encoding of this PackageVersion.
	PackageVersionSyntax() string

	// Less implements a comparison method with another PackageVersion for sorting.
	Less(PackageVersion) bool
}

var (
	_ PackageVersion = (*MavenPackageVersion)(nil)
	_ PackageVersion = (*NpmPackageVersion)(nil)
	_ PackageVersion = (*GoPackageVersion)(nil)
	_ PackageVersion = (*PythonPackageVersion)(nil)
	_ PackageVersion = (*RustPackageVersion)(nil)
)

// versionGreaterThan is a generalized version of comparing two strings
// using semantic versioning that allows for non-numeric characters.
// When a non-numeric character is encountered, the comparison switches to
// lexicographic.
//
// For example, 11.0x > 11.0a > 11.0 > 8.0.
func versionGreaterThan(version1, version2 string) bool {
	index := 0
	end := len(version1)
	if len(version2) < end {
		end = len(version2)
	}
	for index < end {
		rune1 := rune(version1[index])
		rune2 := rune(version2[index])
		if unicode.IsDigit(rune1) && unicode.IsDigit(rune2) {
			int1 := versionParseInt(index, version1)
			int2 := versionParseInt(index, version2)
			if int1 == int2 {
				index = versionNextNonDigitOffset(index, version1)
			} else {
				return int1 > int2
			}
		} else {
			if rune1 == rune2 {
				index += 1
			} else {
				return rune1 > rune2
			}
		}
	}
	return len(version1) < len(version2)
}

// versionParseInt returns the integer value of the number that appears at given
// index of the given string.
func versionParseInt(index int, a string) int {
	end := versionNextNonDigitOffset(index, a)
	value, _ := strconv.Atoi(a[index:end])
	return value
}

// versionNextNonDigitOffset returns the offset of the next non-digit character
// of the given string starting at the given index.
func versionNextNonDigitOffset(index int, b string) int {
	offset := index
	for offset < len(b) && unicode.IsDigit(rune(b[offset])) {
		offset += 1
	}
	return offset
}
