// Package graphqlfile provides utilities for manipulating our graphql schema
// files.
package graphqlfile

import (
	"bufio"
	"bytes"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

// StripInternalComments removes lines starting with #! (e.g. internal
// comments in schema.graphql).
func StripInternalComments(schema []byte) ([]byte, error) {
	var (
		scanner = bufio.NewScanner(bytes.NewReader(schema))
		out     []byte
		re      = lazyregexp.New("^ *#!")
	)
	for scanner.Scan() {
		line := scanner.Text()
		if !re.MatchString(line) {
			out = append(out, []byte(line+"\n")...)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
