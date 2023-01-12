package casetransform

import (
	"regexp/syntax" //nolint:depguard // using the grafana fork of regexp clashes with zoekt, which uses the std regexp/syntax.
	"unicode"
	"unicode/utf8"
)

// LowerRegexpASCII lowers rune literals and expands char classes to include
// lowercase. It does it inplace. We can't just use strings.ToLower since it
// will change the meaning of regex shorthands like \S or \B.
func LowerRegexpASCII(re *syntax.Regexp) {
	for _, c := range re.Sub {
		if c != nil {
			LowerRegexpASCII(c)
		}
	}
	switch re.Op {
	case syntax.OpLiteral:
		// For literal strings we can simplify lower each character.
		for i := range re.Rune {
			re.Rune[i] = unicode.ToLower(re.Rune[i])
		}
	case syntax.OpCharClass:
		l := len(re.Rune)

		// An exclusion class is something like [^A-Z]. We need to specially
		// handle it since the user intention of [^A-Z] should map to
		// [^a-z]. If we use the normal mapping logic, we will do nothing
		// since [a-z] is in [^A-Z]. We assume we have an exclusion class if
		// our inclusive range starts at 0 and ends at the end of the unicode
		// range. Note this means we don't support unusual ranges like
		// [^\x00-B] or [^B-\x{10ffff}].
		isExclusion := l >= 4 && re.Rune[0] == 0 && re.Rune[l-1] == utf8.MaxRune
		if isExclusion {
			// Algorithm:
			// Assume re.Rune is sorted (it is!)
			// 1. Build a list of inclusive ranges in a-z that are excluded in A-Z (excluded)
			// 2. Copy across classes, ensuring all ranges are outside of ranges in excluded.
			//
			// In our comments we use the mathematical notation [x, y] and (a,
			// b). [ and ] are range inclusive, ( and ) are range
			// exclusive. So x is in [x, y], but not in (x, y).

			// excluded is a list of _exclusive_ ranges in ['a', 'z'] that need
			// to be removed.
			excluded := []rune{}

			// Note i starts at 1, so we are inspecting the gaps between
			// ranges. So [re.Rune[0], re.Rune[1]] and [re.Rune[2],
			// re.Rune[3]] impiles we have an excluded range of (re.Rune[1],
			// re.Rune[2]).
			for i := 1; i < l-1; i += 2 {
				// (a, b) is a range that is excluded
				a, b := re.Rune[i], re.Rune[i+1]
				// This range doesn't exclude [A-Z], so skip (does not
				// intersect with ['A', 'Z']).
				if a > 'Z' || b < 'A' {
					continue
				}
				// We know (a, b) intersects with ['A', 'Z']. So clamp such
				// that we have the intersection (a, b) ^ [A, Z]
				if a < 'A' {
					a = 'A' - 1
				}
				if b > 'Z' {
					b = 'Z' + 1
				}
				// (a, b) is now a range contained in ['A', 'Z'] that needs to
				// be excluded. So we map it to the lower case version and add
				// it to the excluded list.
				excluded = append(excluded, a+'a'-'A', b+'b'-'B')
			}

			// Adjust re.Rune to exclude excluded. This may require shrinking
			// or growing the list, so we do it to a copy.
			copy := make([]rune, 0, len(re.Rune))
			for i := 0; i < l; i += 2 {
				// [a, b] is a range that is included
				a, b := re.Rune[i], re.Rune[i+1]

				// Remove exclusions ranges that occur before a. They would of
				// been previously processed.
				for len(excluded) > 0 && a >= excluded[1] {
					excluded = excluded[2:]
				}

				// If our exclusion range happens after b, that means we
				// should only consider it later.
				if len(excluded) == 0 || b <= excluded[0] {
					copy = append(copy, a, b)
					continue
				}

				// We now know that the current exclusion range intersects
				// with [a, b]. Break it into two parts, the range before a
				// and the range after b.
				if a <= excluded[0] {
					copy = append(copy, a, excluded[0])
				}
				if b >= excluded[1] {
					copy = append(copy, excluded[1], b)
				}
			}
			re.Rune = copy
		} else {
			for i := 0; i < l; i += 2 {
				// We found a char class that includes a-z. No need to
				// modify.
				if re.Rune[i] <= 'a' && re.Rune[i+1] >= 'z' {
					return
				}
			}
			for i := 0; i < l; i += 2 {
				a, b := re.Rune[i], re.Rune[i+1]
				// This range doesn't include A-Z, so skip
				if a > 'Z' || b < 'A' {
					continue
				}
				simple := true
				if a < 'A' {
					simple = false
					a = 'A'
				}
				if b > 'Z' {
					simple = false
					b = 'Z'
				}
				a, b = unicode.ToLower(a), unicode.ToLower(b)
				if simple {
					// The char range is within A-Z, so we can
					// just modify it to be the equivalent in a-z.
					re.Rune[i], re.Rune[i+1] = a, b
				} else {
					// The char range includes characters outside
					// of A-Z. To be safe we just append a new
					// lowered range which is the intersection
					// with A-Z.
					re.Rune = append(re.Rune, a, b)
				}
			}
		}
	default:
		return
	}
	// Copy to small storage if necessary
	for i := 0; i < 2 && i < len(re.Rune); i++ {
		re.Rune0[i] = re.Rune[i]
	}
}
