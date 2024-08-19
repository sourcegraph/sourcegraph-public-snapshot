package web

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// stringPattern is a struct describing
type stringPattern struct {
	raw      string
	pats     []string
	breaks   []byte
	literals []string
	wildcard bool
}

func (s stringPattern) Prefix() string {
	return s.literals[0]
}
func (s stringPattern) Match(r *http.Request, c *C) bool {
	return s.match(r, c, true)
}
func (s stringPattern) Run(r *http.Request, c *C) {
	s.match(r, c, false)
}
func (s stringPattern) match(r *http.Request, c *C, dryrun bool) bool {
	path := r.URL.Path
	var matches map[string]string
	if !dryrun {
		if s.wildcard {
			matches = make(map[string]string, len(s.pats)+1)
		} else if len(s.pats) != 0 {
			matches = make(map[string]string, len(s.pats))
		}
	}
	for i, pat := range s.pats {
		sli := s.literals[i]
		if !strings.HasPrefix(path, sli) {
			return false
		}
		path = path[len(sli):]

		m := 0
		bc := s.breaks[i]
		for ; m < len(path); m++ {
			if path[m] == bc || path[m] == '/' {
				break
			}
		}
		if m == 0 {
			// Empty strings are not matches, otherwise routes like
			// "/:foo" would match the path "/"
			return false
		}
		if !dryrun {
			matches[pat] = path[:m]
		}
		path = path[m:]
	}
	// There's exactly one more literal than pat.
	tail := s.literals[len(s.pats)]
	if s.wildcard {
		if !strings.HasPrefix(path, tail) {
			return false
		}
		if !dryrun {
			matches["*"] = path[len(tail)-1:]
		}
	} else if path != tail {
		return false
	}

	if c == nil || dryrun {
		return true
	}

	if c.URLParams == nil {
		c.URLParams = matches
	} else {
		for k, v := range matches {
			c.URLParams[k] = v
		}
	}
	return true
}

func (s stringPattern) String() string {
	return fmt.Sprintf("stringPattern(%q)", s.raw)
}

func (s stringPattern) Raw() string {
	return s.raw
}

// "Break characters" are characters that can end patterns. They are not allowed
// to appear in pattern names. "/" was chosen because it is the standard path
// separator, and "." was chosen because it often delimits file extensions. ";"
// and "," were chosen because Section 3.3 of RFC 3986 suggests their use.
const bc = "/.;,"

var patternRe = regexp.MustCompile(`[` + bc + `]:([^` + bc + `]+)`)

func parseStringPattern(s string) stringPattern {
	raw := s
	var wildcard bool
	if strings.HasSuffix(s, "/*") {
		s = s[:len(s)-1]
		wildcard = true
	}

	matches := patternRe.FindAllStringSubmatchIndex(s, -1)
	pats := make([]string, len(matches))
	breaks := make([]byte, len(matches))
	literals := make([]string, len(matches)+1)
	n := 0
	for i, match := range matches {
		a, b := match[2], match[3]
		literals[i] = s[n : a-1] // Need to leave off the colon
		pats[i] = s[a:b]
		if b == len(s) {
			breaks[i] = '/'
		} else {
			breaks[i] = s[b]
		}
		n = b
	}
	literals[len(matches)] = s[n:]
	return stringPattern{
		raw:      raw,
		pats:     pats,
		breaks:   breaks,
		literals: literals,
		wildcard: wildcard,
	}
}
