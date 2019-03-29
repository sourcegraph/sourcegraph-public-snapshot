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

// RegexpToQuery tries to distill a substring search query that
// matches a superset of the regexp.
func RegexpToQuery(r *syntax.Regexp, minTextSize int) Q {
	q := regexpToQueryRecursive(r, minTextSize)
	q = Simplify(q)
	return q
}

func regexpToQueryRecursive(r *syntax.Regexp, minTextSize int) Q {
	// TODO - we could perhaps transform Begin/EndText in '\n'?
	// TODO - we could perhaps transform CharClass in (OrQuery )
	// if there are just a few runes, and part of a OpConcat?
	switch r.Op {
	case syntax.OpLiteral:
		s := string(r.Rune)
		if len(s) >= minTextSize {
			return &Substring{Pattern: s}
		}
	case syntax.OpCapture:
		return regexpToQueryRecursive(r.Sub[0], minTextSize)

	case syntax.OpPlus:
		return regexpToQueryRecursive(r.Sub[0], minTextSize)

	case syntax.OpRepeat:
		if r.Min >= 1 {
			return regexpToQueryRecursive(r.Sub[0], minTextSize)
		}

	case syntax.OpConcat, syntax.OpAlternate:
		var qs []Q
		for _, sr := range r.Sub {
			if sq := regexpToQueryRecursive(sr, minTextSize); sq != nil {
				qs = append(qs, sq)
			}
		}
		if r.Op == syntax.OpConcat {
			return &And{qs}
		}
		return &Or{qs}
	}
	return &Const{true}
}
