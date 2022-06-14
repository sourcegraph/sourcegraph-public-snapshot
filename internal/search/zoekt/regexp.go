package zoekt

import (
	"regexp/syntax"
)

// TODO(jac) This is copy pasta to fix tests. It accidently implements the
// optimizations jac wants to share from zoekt, but by copying instead of
// using zoekt - keegancsmith

// optimizeRegexp converts capturing groups to non-capturing groups.
// Returns original input if an error is encountered
func optimizeRegexp(re *syntax.Regexp) *syntax.Regexp {
	r := convertCapture(re)
	return r.Simplify()
}

func convertCapture(re *syntax.Regexp) *syntax.Regexp {
	if !hasCapture(re) {
		return re
	}

	// Make a copy so in unlikely event of an error the original can be used as a fallback
	r, err := syntax.Parse(re.String(), regexpFlags)
	if err != nil {
		return re
	}

	r = uncapture(r)

	// Parse again for new structure to take effect
	r, err = syntax.Parse(r.String(), regexpFlags)
	if err != nil {
		return re
	}

	return r
}

func hasCapture(r *syntax.Regexp) bool {
	if r.Op == syntax.OpCapture {
		return true
	}

	for _, s := range r.Sub {
		if hasCapture(s) {
			return true
		}
	}

	return false
}

func uncapture(r *syntax.Regexp) *syntax.Regexp {
	if r.Op == syntax.OpCapture {
		// Captures only have one subexpression
		r.Op = syntax.OpConcat
		r.Cap = 0
		r.Name = ""
	}

	for i, s := range r.Sub {
		r.Sub[i] = uncapture(s)
	}

	return r
}
