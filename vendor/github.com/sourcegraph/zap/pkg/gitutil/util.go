package gitutil

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
)

// splitLines is like strings.Split(s, "\n"), but if s is empty, it returns nil
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func splitNulls(s string) []string {
	if s == "" {
		return nil
	}
	if s[len(s)-1] == '\x00' {
		s = s[:len(s)-1]
	}
	if s == "" {
		return nil
	}
	return strings.Split(s, "\x00")
}

func splitNullsBytes(s []byte) [][]byte {
	if len(s) == 0 {
		return nil
	}
	if s[len(s)-1] == '\x00' {
		s = s[:len(s)-1]
	}
	if len(s) == 0 {
		return nil
	}
	return bytes.Split(s, []byte("\x00"))
}

func splitNullsBytesToStrings(s []byte) []string {
	items := splitNullsBytes(s)
	strings := make([]string, len(items))
	for i, item := range items {
		strings[i] = string(item)
	}
	return strings
}

type errorList struct {
	mu     sync.Mutex
	errors []error
}

// add adds err to the list of errors. It is safe to call it from
// concurrent goroutines.
func (e *errorList) add(err error) {
	e.mu.Lock()
	e.errors = append(e.errors, err)
	e.mu.Unlock()
}

func (e *errorList) error() error {
	switch len(e.errors) {
	case 0:
		return nil
	case 1:
		return e.errors[0]
	default:
		return fmt.Errorf("%s [and %d more errors]", e.errors[0], len(e.errors)-1)
	}
}
