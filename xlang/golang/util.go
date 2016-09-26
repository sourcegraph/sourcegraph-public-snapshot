package golang

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

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

// errors returns the list of errors as a single error. It is NOT safe
// to call from concurrent goroutines.
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

func pathHasPrefix(s, prefix string) bool {
	var prefixSlash string
	if prefix != "" && !strings.HasSuffix(prefix, string(os.PathSeparator)) {
		prefixSlash = prefix + string(os.PathSeparator)
	}
	return s == prefix || strings.HasPrefix(s, prefixSlash)
}

func pathTrimPrefix(s, prefix string) string {
	if s == prefix {
		return ""
	}
	if !strings.HasSuffix(prefix, string(os.PathSeparator)) {
		prefix += string(os.PathSeparator)
	}
	return strings.TrimPrefix(s, prefix)
}
