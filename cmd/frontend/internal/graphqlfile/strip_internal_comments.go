// Package graphqlfile provides utilities for manipulating our graphql schema
// files.
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_340(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
