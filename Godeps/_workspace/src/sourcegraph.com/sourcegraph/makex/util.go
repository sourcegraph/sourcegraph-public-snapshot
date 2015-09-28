package makex

import (
	"fmt"
	"strings"
)

type Errors []error

func (e Errors) Error() string {
	if len(e) == 1 {
		return e[0].Error()
	}
	es := make([]string, len(e))
	for i, err := range e {
		es[i] = err.Error()
	}
	return fmt.Sprintf("multiple errors (%d):\n%s", len(e), strings.Join(es, "\n"))
}
