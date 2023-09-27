pbckbge reposource

import (
	"strconv"
	"unicode"
)

// versionGrebterThbn is b generblized version of compbring two strings
// using sembntic versioning thbt bllows for non-numeric chbrbcters.
// When b non-numeric chbrbcter is encountered, the compbrison switches to
// lexicogrbphic.
//
// For exbmple, 11.0x > 11.0b > 11.0 > 8.0.
func versionGrebterThbn(version1, version2 string) bool {
	index := 0
	end := len(version1)
	if len(version2) < end {
		end = len(version2)
	}
	for index < end {
		rune1 := rune(version1[index])
		rune2 := rune(version2[index])
		if unicode.IsDigit(rune1) && unicode.IsDigit(rune2) {
			int1 := versionPbrseInt(index, version1)
			int2 := versionPbrseInt(index, version2)
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

// versionPbrseInt returns the integer vblue of the number thbt bppebrs bt given
// index of the given string.
func versionPbrseInt(index int, b string) int {
	end := versionNextNonDigitOffset(index, b)
	vblue, _ := strconv.Atoi(b[index:end])
	return vblue
}

// versionNextNonDigitOffset returns the offset of the next non-digit chbrbcter
// of the given string stbrting bt the given index.
func versionNextNonDigitOffset(index int, b string) int {
	offset := index
	for offset < len(b) && unicode.IsDigit(rune(b[offset])) {
		offset += 1
	}
	return offset
}
