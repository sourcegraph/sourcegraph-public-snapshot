package search

import (
	"fmt"
	"io"
	"path/filepath"
	"unicode/utf8"

	"strings"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// Parse tokenizes q and returns a query State object for determining
// the currently active token.
func Parse(q sourcegraph.RawQuery) ([]sourcegraph.Token, State, error) {
	toks, tpos, err := tokenize(q.Text)
	st := State{tpos: tpos, ins: int(q.InsertionPoint)}
	return toks, st, err
}

// tokenize returns tokens in query.
func tokenize(q string) ([]sourcegraph.Token, tokenPos, error) {
	rtoks, tpos, err := scan(q)
	if err != nil {
		return nil, tpos, err
	}
	var toks []sourcegraph.Token
	for _, rtok := range rtoks {
		tok, err := tokenizeOne(rtok)
		if err != nil {
			return nil, tpos, err
		}
		toks = append(toks, tok)
	}
	return toks, tpos, nil
}

// tokenizeOne returns a single token from tok.
func tokenizeOne(tok string) (sourcegraph.Token, error) {
	if tok == "" {
		return nil, &TokenizeError{Reason: "blank token"}
	}
	switch tok[0] {
	case '@':
		return sourcegraph.UserToken{Login: tok[1:]}, nil
	case ':':
		return sourcegraph.RevToken{Rev: tok[1:]}, nil
	case '/':
		return sourcegraph.FileToken{Path: filepath.Clean(tok[1:])}, nil
	case '~':
		t := sourcegraph.UnitToken{Name: tok[1:]}
		if sep := strings.Index(t.Name, "@"); sep != -1 {
			origName := t.Name
			t.Name = origName[:sep]
			t.UnitType = origName[sep+1:]
		}
		return t, nil
	}
	for i, r := range tok {
		if r == '/' && i > 0 {
			return sourcegraph.RepoToken{URI: tok}, nil
		}
	}
	return sourcegraph.AnyToken(tok), nil
}

// A TokenizeError occurs when query tokenization fails.
type TokenizeError struct {
	TokenType string
	Reason    string
}

func (e *TokenizeError) Error() string {
	return fmt.Sprintf("query: tokenize %s failed: %s", e.TokenType, e.Reason)
}

// scan reads whitespace-delimited tokens from query and returns the
// tokens and their character positions.
func scan(query string) ([]string, tokenPos, error) {
	var toks []string
	tpos := tokenPos{}
	var s scanner
	s.Init(query)
	for {
		tok, err := s.Scan()
		if err == io.EOF {
			return toks, tpos, nil
		}
		if err != nil {
			return toks, tpos, err
		}
		tpos = append(tpos, [2]int{s.lastTokStart, s.lastTokStart + utf8.RuneCountInString(tok)})
		toks = append(toks, tok)
	}
}

// tokenPos holds the character range (in the original query) of
// tokens.
type tokenPos [][2]int

// State tracks the query's insertion point and the character range
// (in the original query) of each token.
type State struct {
	tpos tokenPos // token char ranges
	ins  int      // insertion point
}

// IsActive returns true if the i'th token is active (the insertion
// point is inside its character range).
func (s *State) IsActive(i int) bool {
	if i < 0 || i >= len(s.tpos) {
		panic("token i out of bounds")
	}
	return s.ins >= s.tpos[i][0] && s.ins <= s.tpos[i][1]
}

type scanner struct {
	src string

	srcPos int // character position

	lastTokStart int
}

func (s *scanner) Init(src string) {
	s.src = src
	s.srcPos = 0
}

func (s *scanner) Scan() (string, error) {
	var rs []rune
loop:
	for _, r := range s.src[s.srcPos:] {
		switch r {
		case ' ':
			if rs != nil {
				break loop
			}
		default:
			if rs == nil {
				s.lastTokStart = s.srcPos
			}
			rs = append(rs, r)
		}
		s.srcPos += utf8.RuneLen(r)
	}
	if rs == nil {
		return "", io.EOF
	}
	return string(rs), nil
}

// A ScanError occurs when query scanning fails.
type ScanError struct {
	TokenType string
	Reason    string
}

func (e *ScanError) Error() string {
	return fmt.Sprintf("query: scan %s failed: %s", e.TokenType, e.Reason)
}
