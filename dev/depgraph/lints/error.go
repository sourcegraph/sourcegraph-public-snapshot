package lints

import (
	"fmt"
	"strings"
)

type lintError struct {
	name        string
	pkg         string
	description string
}

func (e lintError) Error() string {
	desc := e.description
	if desc != "" {
		desc = ": " + desc
	}

	pkg := e.pkg
	if pkg == "" {
		pkg = "<root>"
	}

	return fmt.Sprintf("(%s) %s%s", e.name, pkg, desc)
}

type lintErrors []lintError

func (es lintErrors) Error() string {
	errors := make([]string, 0, len(es))
	for _, e := range es {
		errors = append(errors, e.Error())
	}

	return fmt.Sprintf("%d errors:\n%s", len(es), strings.Join(errors, "\n"))
}

func multi(es []lintError) lintErrors {
	if len(es) == 0 {
		return nil
	}

	return lintErrors(es)
}
