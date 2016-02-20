package nosurf

import (
	"fmt"
	"net/http"
	pathModule "path"
	"reflect"
	"regexp"
)

// Checks if the given request is exempt from CSRF checks.
// It checks the ExemptFunc first, then the exact paths,
// then the globs and finally the regexps.
func (h *CSRFHandler) IsExempt(r *http.Request) bool {
	if h.exemptFunc != nil && h.exemptFunc(r) {
		return true
	}

	path := r.URL.Path
	if sContains(h.exemptPaths, path) {
		return true
	}

	// then the globs
	for _, glob := range h.exemptGlobs {
		matched, err := pathModule.Match(glob, path)
		if matched && err == nil {
			return true
		}
	}

	// finally, the regexps
	for _, re := range h.exemptRegexps {
		if re.MatchString(path) {
			return true
		}
	}

	return false
}

// Exempts an exact path from CSRF checks
// With this (and other Exempt* methods)
// you should take note that Go's paths
// include a leading slash.
func (h *CSRFHandler) ExemptPath(path string) {
	h.exemptPaths = append(h.exemptPaths, path)
}

// A variadic argument version of ExemptPath()
func (h *CSRFHandler) ExemptPaths(paths ...string) {
	for _, v := range paths {
		h.ExemptPath(v)
	}
}

// Exempts URLs that match the specified glob pattern
// (as used by filepath.Match()) from CSRF checks
//
// Note that ExemptGlob() is unable to detect syntax errors,
// because it doesn't have a path to check it against
// and filepath.Match() doesn't report an error
// if the path is empty.
// If we find a way to check the syntax, ExemptGlob
// MIGHT PANIC on a syntax error in the future.
// ALWAYS check your globs for syntax errors.
func (h *CSRFHandler) ExemptGlob(pattern string) {
	h.exemptGlobs = append(h.exemptGlobs, pattern)
}

// A variadic argument version of ExemptGlob()
func (h *CSRFHandler) ExemptGlobs(patterns ...string) {
	for _, v := range patterns {
		h.ExemptGlob(v)
	}
}

// Accepts a regular expression string or a compiled *regexp.Regexp
// and exempts URLs that match it from CSRF checks.
//
// If the given argument is neither of the accepted values,
// or the given string fails to compile, ExemptRegexp() panics.
func (h *CSRFHandler) ExemptRegexp(re interface{}) {
	var compiled *regexp.Regexp

	switch re.(type) {
	case string:
		compiled = regexp.MustCompile(re.(string))
	case *regexp.Regexp:
		compiled = re.(*regexp.Regexp)
	default:
		err := fmt.Sprintf("%v isn't a valid type for ExemptRegexp()", reflect.TypeOf(re))
		panic(err)
	}

	h.exemptRegexps = append(h.exemptRegexps, compiled)
}

// A variadic argument version of ExemptRegexp()
func (h *CSRFHandler) ExemptRegexps(res ...interface{}) {
	for _, v := range res {
		h.ExemptRegexp(v)
	}
}

func (h *CSRFHandler) ExemptFunc(fn func(r *http.Request) bool) {
	h.exemptFunc = fn
}
