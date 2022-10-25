package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func main() {
	if err := anonymiseProtection(os.Stdin, os.Stdout); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// anonymiseProtection anonymises a protection table by consistently renaming
// all tokens to random values.
func anonymiseProtection(r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		// Drop comments
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "//COMMENT") {
			continue
		}

		// Anonymise path patterns
		if strings.Contains(line, "//") {
			line = anonymiseLine(line)
		}

		if _, err := fmt.Fprintln(w, line); err != nil {
			return errors.Wrapf(err, "writing line")
		}
	}
	if err := scanner.Err(); err != nil {
		return errors.Wrap(err, "scanning input")
	}
	return nil
}

var h = sha256.New()
var groupRegexp = regexp.MustCompile("group ([\\w]*)")
var userRegexp = regexp.MustCompile("user (\\w)*")

func anonymiseLine(line string) string {
	start := strings.Index(line, "//")
	if start == -1 {
		return line
	}

	// Parts separated by slashes
	path := line[start+2:] // +2 to drop //
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "..." {
			continue
		}
		if strings.Contains(part, "*") {
			// TODO: We need to anonymise text around the star
			parts[i] = randomiseWithWildcards(part)
			continue
		}

		parts[i] = randomise(part)
	}
	line = line[:start+2]
	line = line + strings.Join(parts, "/")

	// We also want to replace user and group names, which is any string starting
	// with "group " or "user "
	line = replaceGroupOrUser(line, groupRegexp, "group ")
	line = replaceGroupOrUser(line, userRegexp, "user ")

	return line
}

func replaceGroupOrUser(line string, r *regexp.Regexp, prefix string) string {
	return r.ReplaceAllStringFunc(line, func(s string) string {
		if s == prefix {
			return s
		}
		return prefix + randomise(s[len(prefix):])
	})
}

var nonWildcardRegexp = regexp.MustCompile("([^\\*]*)")

// asterisk(s) can appear anywhere in the string and should
// remain there. Everything around them should be randomised
func randomiseWithWildcards(input string) string {
	return nonWildcardRegexp.ReplaceAllStringFunc(input, func(s string) string {
		if s == "" {
			return s
		}
		return randomise(s)
	})
}

func randomise(s string) string {
	const desiredLength = 6

	h.Reset()
	h.Write([]byte(s))
	b := h.Sum(nil)
	s = hex.EncodeToString(b)
	return s[:desiredLength]
}
