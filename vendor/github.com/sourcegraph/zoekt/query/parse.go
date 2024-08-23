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
	"bytes"
	"fmt"
	"log"
	"regexp/syntax"

	"github.com/grafana/regexp"
	"github.com/sourcegraph/zoekt/internal/languages"
)

var _ = log.Printf

// parseStringLiteral parses a string literal, consumes the starting
// quote too.
func parseStringLiteral(in []byte) (lit []byte, n int, err error) {
	left := in[1:]
	found := false

loop:
	for len(left) > 0 {
		c := left[0]
		left = left[1:]
		switch c {
		case '"':
			found = true
			break loop
		case '\\':
			// TODO - other escape sequences.
			if len(left) == 0 {
				return nil, 0, fmt.Errorf("query: missing char after \\")
			}
			c = left[0]
			left = left[1:]

			lit = append(lit, c)
		default:
			lit = append(lit, c)
		}
	}
	if !found {
		return nil, 0, fmt.Errorf("query: unterminated quoted string")
	}
	return lit, len(in) - len(left), nil
}

// orOperator is a placeholder intermediate so we can represent [A,
// or, B] before we convert it to Or{A, B}
type orOperator struct{}

func (o *orOperator) String() string {
	return "orOp"
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t'
}

// Parse parses a string into a query.
func Parse(qStr string) (Q, error) {
	b := []byte(qStr)

	qs, _, err := parseExprList(b)
	if err != nil {
		return nil, err
	}

	q, err := parseOperators(qs)
	if err != nil {
		return nil, err
	}

	return Simplify(q), nil
}

// parseExpr parses a single expression, returning the result, and the
// number of bytes consumed.
func parseExpr(in []byte) (Q, int, error) {
	b := in[:]
	var expr Q
	for len(b) > 0 && isSpace(b[0]) {
		b = b[1:]
	}

	tok, err := nextToken(b)
	if err != nil {
		return nil, 0, err
	}
	if tok == nil {
		return nil, 0, nil
	}
	b = b[len(tok.Input):]

	text := string(tok.Text)
	switch tok.Type {
	case tokCase:
		switch text {
		case "yes":
		case "no":
		case "auto":
		default:
			return nil, 0, fmt.Errorf("query: unknown case argument %q, want {yes,no,auto}", text)
		}
		expr = &caseQ{text}
	case tokRepo:
		r, err := regexp.Compile(text)
		if err != nil {
			return nil, 0, err
		}

		expr = &Repo{r}
	case tokArchived:
		switch text {
		case "yes":
			expr = RawConfig(RcOnlyArchived)
		case "no":
			expr = RawConfig(RcNoArchived)
		default:
			return nil, 0, fmt.Errorf("query: unknown archived argument %q, want {yes,no}", text)
		}
	case tokFork:
		switch text {
		case "yes":
			expr = RawConfig(RcOnlyForks)
		case "no":
			expr = RawConfig(RcNoForks)
		default:
			return nil, 0, fmt.Errorf("query: unknown fork argument %q, want {yes,no}", text)
		}
	case tokPublic:
		switch text {
		case "yes":
			expr = RawConfig(RcOnlyPublic)
		case "no":
			expr = RawConfig(RcOnlyPrivate)
		default:
			return nil, 0, fmt.Errorf("query: unknown public argument %q, want {yes,no}", text)
		}
	case tokBranch:
		expr = &Branch{Pattern: text}
	case tokText, tokRegex:
		q, err := RegexpQuery(text, false, false)
		if err != nil {
			return nil, 0, err
		}
		expr = q
	case tokFile:
		q, err := RegexpQuery(text, false, true)
		if err != nil {
			return nil, 0, err
		}
		expr = q
	case tokContent:
		q, err := RegexpQuery(text, true, false)
		if err != nil {
			return nil, 0, err
		}
		expr = q
	case tokLang:
		canonical, ok := languages.GetLanguageByAlias(text)
		if !ok {
			expr = &Const{false}
		} else {
			expr = &Language{Language: canonical}
		}

	case tokSym:
		if text == "" {
			return nil, 0, fmt.Errorf("the sym: atom must have an argument")
		}

		q, err := RegexpQuery(text, false, false)
		if err != nil {
			return nil, 0, err
		}

		expr = &Symbol{q}
	case tokParenClose:
		// Caller must consume paren.
		expr = nil

	case tokParenOpen:
		qs, n, err := parseExprList(b)
		b = b[n:]
		if err != nil {
			return nil, 0, err
		}

		pTok, err := nextToken(b)
		if err != nil {
			return nil, 0, err
		}
		if pTok == nil || pTok.Type != tokParenClose {
			return nil, 0, fmt.Errorf("query: missing close paren, got token %v", pTok)
		}

		b = b[len(pTok.Input):]
		expr, err = parseOperators(qs)
		if err != nil {
			return nil, 0, err
		}
	case tokNegate:
		subQ, n, err := parseExpr(b)
		if err != nil {
			return nil, 0, err
		}
		if subQ == nil {
			return nil, 0, fmt.Errorf("query: '-' operator needs an argument")
		}
		b = b[n:]
		expr = &Not{subQ}

	case tokType:
		var t uint8
		switch text {
		case "filematch":
			t = TypeFileMatch
		case "filename", "file":
			t = TypeFileName
		case "repo":
			t = TypeRepo
		default:
			return nil, 0, fmt.Errorf("query: unknown type argument %q, want {filematch,filename,repo}", text)
		}
		// Later we will lift this into a root, like we do for caseQ
		expr = &Type{Type: t, Child: nil}
	}

	return expr, len(in) - len(b), nil
}

const regexpFlags syntax.Flags = syntax.ClassNL | syntax.PerlX | syntax.UnicodeGroups

// RegexpQuery parses an atom into either a regular expression, or a
// simple substring atom.
func RegexpQuery(text string, content, file bool) (Q, error) {
	var expr Q

	r, err := syntax.Parse(text, regexpFlags)
	if err != nil {
		return nil, err
	}

	r = OptimizeRegexp(r, regexpFlags)

	if r.Op == syntax.OpLiteral {
		expr = &Substring{
			Pattern:  string(r.Rune),
			FileName: file,
			Content:  content,
		}
	} else {
		expr = &Regexp{
			Regexp:   r,
			FileName: file,
			Content:  content,
		}
	}

	return expr, nil
}

// parseOperators interprets the orOperator in a list of queries.
func parseOperators(in []Q) (Q, error) {
	top := &Or{}
	cur := &And{}

	seenOr := false
	for _, q := range in {
		if _, ok := q.(*orOperator); ok {
			seenOr = true
			if len(cur.Children) == 0 {
				return nil, fmt.Errorf("query: OR operator should have operand")
			}
			top.Children = append(top.Children, cur)
			cur = &And{}
		} else {
			cur.Children = append(cur.Children, q)
		}
	}

	if seenOr && len(cur.Children) == 0 {
		return nil, fmt.Errorf("query: OR operator should have operand")
	}
	top.Children = append(top.Children, cur)
	return top, nil
}

// parseExprList parses a list of query expressions. It is the
// workhorse of the Parse function.
func parseExprList(in []byte) ([]Q, int, error) {
	b := in[:]
	var qs []Q
	for len(b) > 0 {
		for len(b) > 0 && isSpace(b[0]) {
			b = b[1:]
		}
		tok, _ := nextToken(b)
		if tok != nil && tok.Type == tokParenClose {
			break
		} else if tok != nil && tok.Type == tokOr {
			qs = append(qs, &orOperator{})
			b = b[len(tok.Input):]
			continue
		}

		q, n, err := parseExpr(b)
		if err != nil {
			return nil, 0, err
		}

		if q == nil {
			// eof or a ')'
			break
		}
		qs = append(qs, q)
		b = b[n:]
	}

	setCase := "auto"
	newQS := qs[:0]
	typeT := uint8(100)
	for _, q := range qs {
		switch s := q.(type) {
		case *caseQ:
			setCase = s.Flavor
		case *Type:
			if s.Type < typeT {
				typeT = s.Type
			}
		default:
			newQS = append(newQS, q)
		}
	}
	qs = mapQueryList(newQS, func(q Q) Q {
		if sc, ok := q.(setCaser); ok {
			sc.setCase(setCase)
		}
		return q
	})
	if typeT != 100 {
		qs = []Q{&Type{Type: typeT, Child: NewAnd(qs...)}}
	}
	return qs, len(in) - len(b), nil
}

type token struct {
	Type int
	// The value of the token
	Text []byte

	// The input that we consumed to form the token.
	Input []byte
}

func (t *token) String() string {
	return fmt.Sprintf("%s:%q", tokNames[t.Type], t.Text)
}

// token types.
const (
	tokText       = 0
	tokFile       = 1
	tokRepo       = 2
	tokCase       = 3
	tokBranch     = 4
	tokParenOpen  = 5
	tokParenClose = 6
	tokError      = 7
	tokNegate     = 8
	tokRegex      = 9
	tokOr         = 10
	tokContent    = 11
	tokLang       = 12
	tokSym        = 13
	tokType       = 14
	tokArchived   = 15
	tokPublic     = 16
	tokFork       = 17
)

var tokNames = map[int]string{
	tokArchived:   "Archived",
	tokBranch:     "Branch",
	tokCase:       "Case",
	tokError:      "Error",
	tokFile:       "File",
	tokFork:       "Fork",
	tokNegate:     "Negate",
	tokOr:         "Or",
	tokParenClose: "ParenClose",
	tokParenOpen:  "ParenOpen",
	tokPublic:     "Public",
	tokRegex:      "Regex",
	tokRepo:       "Repo",
	tokText:       "Text",
	tokLang:       "Language",
	tokSym:        "Symbol",
	tokType:       "Type",
}

var prefixes = map[string]int{
	"archived:": tokArchived,
	"b:":        tokBranch,
	"branch:":   tokBranch,
	"c:":        tokContent,
	"case:":     tokCase,
	"content:":  tokContent,
	"f:":        tokFile,
	"file:":     tokFile,
	"fork:":     tokFork,
	"public:":   tokPublic,
	"r:":        tokRepo,
	"regex:":    tokRegex,
	"repo:":     tokRepo,
	"lang:":     tokLang,
	"sym:":      tokSym,
	"t:":        tokType,
	"type:":     tokType,
}

var reservedWords = map[string]int{
	"or": tokOr,
}

func (t *token) setType() {
	// After we consumed the input, we have to interpret some of the text,
	// eg. to distinguish between ")" the text and ) the query grouping
	// parenthesis.
	if len(t.Text) == 1 && t.Text[0] == '(' {
		t.Type = tokParenOpen
	}
	if len(t.Text) == 1 && t.Text[0] == ')' {
		t.Type = tokParenClose
	}

	for w, typ := range reservedWords {
		if string(t.Text) == w && string(t.Input) == w {
			t.Type = typ
			break
		}
	}

	for pref, typ := range prefixes {
		if !bytes.HasPrefix(t.Input, []byte(pref)) {
			continue
		}

		t.Text = t.Text[len(pref):]
		t.Type = typ
		break
	}
}

// nextToken returns the next token from the given input.
func nextToken(in []byte) (*token, error) {
	left := in[:]
	parenCount := 0
	var cur token
	if len(left) == 0 {
		return nil, nil
	}

	if left[0] == '-' {
		return &token{
			Type:  tokNegate,
			Text:  []byte{'-'},
			Input: in[:1],
		}, nil
	}

	foundSpace := false

loop:
	for len(left) > 0 {
		c := left[0]
		switch c {
		case '(':
			parenCount++
			cur.Text = append(cur.Text, c)
			left = left[1:]
		case ')':
			if parenCount == 0 {
				if len(cur.Text) == 0 {
					cur.Text = []byte{')'}
					left = left[1:]
				}
				break loop
			}

			cur.Text = append(cur.Text, c)
			left = left[1:]
			parenCount--

		case '"':
			t, n, err := parseStringLiteral(left)
			if err != nil {
				return nil, err
			}
			cur.Text = append(cur.Text, t...)
			left = left[n:]
		case '\\':
			left = left[1:]
			if len(left) == 0 {
				return nil, fmt.Errorf("query: lone \\ at end")
			}
			c := left[0]
			cur.Text = append(cur.Text, '\\', c)
			left = left[1:]

		case ' ', '\n', '\t':
			if parenCount > 0 {
				foundSpace = true
			}
			break loop
		default:
			cur.Text = append(cur.Text, c)
			left = left[1:]
		}
	}

	if len(cur.Text) == 0 {
		return nil, nil
	}

	if foundSpace && cur.Text[0] == '(' {
		cur.Text = cur.Text[:1]
		cur.Input = in[:1]
	} else {
		cur.Input = in[:len(in)-len(left)]
	}
	cur.setType()
	return &cur, nil
}
