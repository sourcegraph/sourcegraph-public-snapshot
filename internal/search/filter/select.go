package filter

import (
	"fmt"
	"strings"
)

type SelectType string

const (
	Commit     SelectType = "commit"
	Content    SelectType = "content"
	File       SelectType = "file"
	Repository SelectType = "repo"
	Symbol     SelectType = "symbol"
)

// SelectPath represents a parsed and validated select value and fields.
type SelectPath struct {
	Type   SelectType
	Fields []string
}

func (sp SelectPath) String() string {
	return string(sp.Type)
}

type Object map[string]interface{}

var empty = struct{}{}

var validSelectors = map[SelectType]Object{
	Commit: {
		"diff": Object{
			"added":   empty,
			"removed": empty,
		},
	},
	Content:    {},
	File:       {},
	Repository: {},
	Symbol: {
		/* cf. SymbolKind https://microsoft.github.io/language-server-protocol/specification */
		"file":           empty,
		"module":         empty,
		"namespace":      empty,
		"package":        empty,
		"class":          empty,
		"method":         empty,
		"property":       empty,
		"field":          empty,
		"constructor":    empty,
		"enum":           empty,
		"interface":      empty,
		"function":       empty,
		"variable":       empty,
		"constant":       empty,
		"string":         empty,
		"number":         empty,
		"boolean":        empty,
		"array":          empty,
		"object":         empty,
		"key":            empty,
		"null":           empty,
		"enum-member":    empty,
		"struct":         empty,
		"event":          empty,
		"operator":       empty,
		"type-parameter": empty,
	},
}

func splitFields(s string) (string, []string) {
	v := strings.Split(s, ".")
	return v[0], v[1:]
}

func SelectPathFromString(s string) (SelectPath, error) {
	selector, fields := splitFields(s)
	if _, ok := validSelectors[SelectType(selector)]; !ok {
		return SelectPath{}, fmt.Errorf("invalid select type '%s'", s)
	}
	if len(fields) > 0 {
		if _, ok := validSelectors[SelectType(selector)][fields[0]]; !ok {
			return SelectPath{}, fmt.Errorf("invalid field '%s' on select type '%s'", fields[0], selector)
		}
	}
	return SelectPath{Type: SelectType(selector), Fields: fields}, nil
}
