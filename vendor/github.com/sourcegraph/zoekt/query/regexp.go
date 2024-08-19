// Copyright 2016 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package query

import (
	"log"
	"regexp/syntax"

	"github.com/sourcegraph/zoekt/internal/syntaxutil"
)

var _ = log.Println

func LowerRegexp(r *syntax.Regexp) *syntax.Regexp {
	newRE := *r
	switch r.Op {
	case syntax.OpLiteral, syntax.OpCharClass:
		newRE.Rune = make([]rune, len(r.Rune))
		for i, c := range r.Rune {
			if c >= 'A' && c <= 'Z' {
				newRE.Rune[i] = c + 'a' - 'A'
			} else {
				newRE.Rune[i] = c
			}
		}
	default:
		newRE.Sub = make([]*syntax.Regexp, len(newRE.Sub))
		for i, s := range r.Sub {
			newRE.Sub[i] = LowerRegexp(s)
		}
	}

	return &newRE
}

// OptimizeRegexp converts capturing groups to non-capturing groups.
// Returns original input if an error is encountered
func OptimizeRegexp(re *syntax.Regexp, flags syntax.Flags) *syntax.Regexp {
	r := convertCapture(re, flags)
	return r.Simplify()
}

func convertCapture(re *syntax.Regexp, flags syntax.Flags) *syntax.Regexp {
	if !hasCapture(re) {
		return re
	}

	// Make a copy so in unlikely event of an error the original can be used as a fallback
	r, err := syntax.Parse(syntaxutil.RegexpString(re), flags)
	if err != nil {
		log.Printf("failed to copy regexp `%s`: %v", re, err)
		return re
	}

	r = uncapture(r)

	// Parse again for new structure to take effect
	r, err = syntax.Parse(syntaxutil.RegexpString(r), flags)
	if err != nil {
		log.Printf("failed to parse regexp after uncapture `%s`: %v", r, err)
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
