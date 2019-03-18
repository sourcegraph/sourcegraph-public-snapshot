package graphqlfile

import (
	"bufio"
	"bytes"
	"regexp"
)

// StripInternalComments removes lines starting with #! (e.g. internal
// comments in schema.graphql).
func StripInternalComments(schema []byte) ([]byte, error) {
	var (
		scanner = bufio.NewScanner(bytes.NewReader(schema))
		out     []byte
		re      = regexp.MustCompile("^ *#!")
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
