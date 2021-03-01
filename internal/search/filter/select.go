package filter

import "fmt"

type SelectType string

const (
	Commit     SelectType = "commit"
	Content    SelectType = "content"
	File       SelectType = "file"
	Repository SelectType = "repo"
	Symbol     SelectType = "symbol"
)

// SelectPath is the parsed representation of the select field's value.
type SelectPath struct {
	Type SelectType
}

func (sp SelectPath) String() string {
	return string(sp.Type)
}

var validSelectors = map[SelectType]struct{}{
	Commit:     {},
	Content:    {},
	File:       {},
	Repository: {},
	Symbol:     {},
}

func SelectPathFromString(s string) (SelectPath, error) {
	if _, ok := validSelectors[SelectType(s)]; !ok {
		return SelectPath{}, fmt.Errorf("invalid select type '%s'", s)
	}
	return SelectPath{SelectType(s)}, nil
}
