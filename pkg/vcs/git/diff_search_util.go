package git

import (
	regexpsyntax "regexp/syntax"
)

// regexpToGlobBestEffort performs a best-effort conversion of the regexp p to an equivalent glob
// pattern. The glob matches a superset of what the regexp matches. If equiv is true, then the glob
// is exactly equivalent to the pattern; otherwise it is a strict superset and post-filtering is
// necessary. The glob never matches a strict subset of p (that would make it possible to correctly
// post-filter).
//
// https://git-scm.com/docs/gitglossary#gitglossary-aiddefpathspecapathspec
func regexpToGlobBestEffort(p string) (glob string, equiv bool) {
	if p == "" {
		return "*", true
	}

	re, err := regexpsyntax.Parse(p, regexpsyntax.OneLine)
	if err != nil {
		return "", false
	}
	switch re.Op {
	case regexpsyntax.OpLiteral:
		return "*" + globQuoteMeta(re.Rune) + "*", true
	case regexpsyntax.OpConcat:
		if len(re.Sub) == 3 && re.Sub[0].Op == regexpsyntax.OpBeginText && re.Sub[1].Op == regexpsyntax.OpLiteral && re.Sub[2].Op == regexpsyntax.OpEndText {
			if len(re.Sub[1].Rune) > 0 && re.Sub[1].Rune[0] == ':' { // leading : has special meaning
				return "", false
			}
			return globQuoteMeta(re.Sub[1].Rune), true
		}
		if len(re.Sub) == 2 && re.Sub[0].Op == regexpsyntax.OpBeginText && re.Sub[1].Op == regexpsyntax.OpLiteral {
			if len(re.Sub[1].Rune) > 0 && re.Sub[1].Rune[0] == ':' { // leading : has special meaning
				return "", false
			}
			return globQuoteMeta(re.Sub[1].Rune) + "*", true
		}
		if len(re.Sub) == 2 && re.Sub[1].Op == regexpsyntax.OpEndText && re.Sub[0].Op == regexpsyntax.OpLiteral {
			return "*" + globQuoteMeta(re.Sub[0].Rune), true
		}
	}
	return "", false
}

func globQuoteMeta(s []rune) string {
	isSpecial := func(c rune) bool {
		switch c {
		case '*':
			return true
		case '?':
			return true
		case '[':
			return true
		case ']':
			return true
		case '\\':
			return true
		default:
			return false
		}
	}

	// Avoid extra work by counting additions. regexp.QuoteMeta does the same,
	// but is more efficient since it does it via bytes.
	count := 0
	for _, c := range s {
		if isSpecial(c) {
			count++
		}
	}
	if count == 0 {
		return string(s)
	}

	escaped := make([]rune, 0, len(s)+count)
	for _, c := range s {
		if isSpecial(c) {
			escaped = append(escaped, '\\')
		}
		escaped = append(escaped, c)
	}
	return string(escaped)
}
