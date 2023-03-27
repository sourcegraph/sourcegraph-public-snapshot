package repos

import (
	"reflect"
	"strings"

	"github.com/grafana/regexp"
)

// excludeFunc takes a string and returns true if it should be excluded. In
// the case of repo sourcing it will take a repository name or ID as input.
type excludeFunc func(interface{}) bool

// excludeBuilder builds an excludeFunc.
type excludeBuilder struct {
	exact      map[string]struct{}
	patterns   []*regexp.Regexp
	fieldsTrue []string

	err error
}

// Exact will case-insensitively exclude the string name.
func (e *excludeBuilder) Exact(name string) {
	if e.exact == nil {
		e.exact = map[string]struct{}{}
	}
	if name == "" {
		return
	}
	e.exact[strings.ToLower(name)] = struct{}{}
}

// Pattern will exclude strings matching the regex pattern.
func (e *excludeBuilder) Pattern(pattern string) {
	if pattern == "" {
		return
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		e.err = err
		return
	}
	e.patterns = append(e.patterns, re)
}

// FieldTrue will exclude the object if the specified field name is set to true.
func (e *excludeBuilder) FieldTrue(filedName string) {
	if e.fieldsTrue == nil {
		e.fieldsTrue = []string{}
	}
	if filedName == "" {
		return
	}
	e.fieldsTrue = append(e.fieldsTrue, filedName)
}

// Build will return an excludeFunc based on the previous calls to Exact, FieldTrue, and
// Pattern.
func (e *excludeBuilder) Build() (excludeFunc, error) {
	return func(input interface{}) bool {
		if name, ok := input.(string); ok {
			if _, ok := e.exact[strings.ToLower(name)]; ok {
				return true
			}

			for _, re := range e.patterns {
				if re.MatchString(name) {
					return true
				}
			}
		} else {
			v := reflect.ValueOf(input)
			if v.Kind() == reflect.Struct {
				for _, fieldName := range e.fieldsTrue {
					fieldValue := v.FieldByName(fieldName)
					if fieldValue.Kind() == reflect.Bool && fieldValue.Bool() {
						return true
					}
				}
			}
		}

		return false
	}, e.err
}
