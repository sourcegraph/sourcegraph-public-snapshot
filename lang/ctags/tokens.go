package ctags

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

// getTokenFromFile extracts the token at the given position (line, col) in the file
func getTokenFromFile(filename string, l int, c int) (tokenInfo, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return tokenInfo{}, err
	}
	lines := strings.Split(string(b), "\n")
	if l >= len(lines) {
		return tokenInfo{}, fmt.Errorf("line index exceeds number of lines in file")
	}
	line := lines[l]

	return getToken(line, c)
}

// getToken extracts the token that overlaps index i from string s in a language-agnostic fashion
func getToken(s string, i int) (tokenInfo, error) {
	toks, err := tokenize(s)
	if err != nil {
		return tokenInfo{}, err
	}
	before := 0
	for _, tok := range toks {
		if before <= i && i < before+len(tok.token) {
			return tok, nil
		}
		before += len(tok.token)
	}
	return tokenInfo{}, fmt.Errorf("could not return token at index out of bounds %d, length of string was %d", i, len(s))
}

type tokenInfo struct {
	token string
	kind  string
}

func tokenize(s string) ([]tokenInfo, error) {
	var tokens []tokenInfo

	for s != "" {
		match := tokenizer.FindStringSubmatch(s)
		if len(match) == 0 {
			return nil, fmt.Errorf("failed to tokenize string `%s`", s)
		}
		toks := match[1:]

		if len(toks) != numTokenKinds {
			return nil, fmt.Errorf("unexpected number of matching groups, expected %d, found %d", numTokenKinds, len(toks))
		}

		var newToken tokenInfo
		for j := 0; j < numTokenKinds; j++ {
			if tok := toks[j]; tok != "" {
				newToken.kind = tokenKinds[j]
				newToken.token = tok
			}
		}
		tokens = append(tokens, newToken)
		s = s[len(newToken.token):]
	}
	return tokens, nil
}

var (
	tokenizer     = regexp.MustCompile(fmt.Sprintf(`^(%s)|(%s)|(%s)|(%s)|(%s)|(%s)|$`, whitespace, name, other, string1, string2, comment))
	tokenKinds    = []string{tokWhitespace, tokName, tokOther, tokString1, tokString2, tokComment}
	numTokenKinds = len(tokenKinds)
)

const (
	whitespace = `\s+`
	name       = `[A-Za-z0-9_\?\$\@\pL\pN]+`
	string1    = `'(?:\\'|[^'])*'`
	string2    = `"(?:\\"|[^"])*"`
	comment    = `//[\S\s]*`
	other      = `[^A-Za-z0-9_\?\$\@\pL\pN\s]+`

	tokWhitespace = "whitespace"
	tokName       = "name"
	tokOther      = "other"
	tokString1    = "string1"
	tokString2    = "string2"
	tokComment    = "comment"
)
