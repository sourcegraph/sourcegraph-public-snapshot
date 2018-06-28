package proxy

import (
	"context"
	"fmt"
	"sync"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
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

// getInitializationOptions returns the initializationOptions value to use in an LSP
// initialize request.
func getInitializationOptions(ctx context.Context, lang string) (map[string]interface{}, error) {
	for _, ls := range conf.EnabledLangservers() {
		if ls.Language == lang {
			return ls.InitializationOptions, nil
		}
	}
	// TODO(extensions): Clean this up: remove these 2 similar funcs
	// getInitializationOptions{,ForExtension}.
	return getInitializationOptionsForExtension(ctx, lang)
}
