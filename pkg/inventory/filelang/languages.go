// Package filelang detects the programming language of files by
// their filenames and (probabilistically/heuristically) their
// contents.
package filelang

import (
	"path"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

//go:generate env GO111MODULE=on go run gen/generate.go

// A Language represents a programming or markup language.
//
// See
// https://github.com/github/linguist/blob/master/lib/linguist/languages.yml
// for a description of each field.
type Language struct {
	Name string `yaml:"-"`

	Type         string   `yaml:"type,omitempty"`
	Aliases      []string `yaml:"aliases,omitempty"`
	Wrap         bool     `yaml:"wrap,omitempty"`
	Extensions   []string `yaml:"extensions,omitempty"`
	Interpreters []string `yaml:"interpreters,omitempty"`
	Searchable   bool     `yaml:"searchable,omitempty"`
	Color        string   `yaml:"color,omitempty"`
	Group        string   `yaml:"group,omitempty"`
	Filenames    []string `yaml:"filenames,omitempty"`

	// Omitted fields:
	//  - ace_mode
	//  - search_term
	//  - tm_scope
}

// MatchFilename returns true if the language is associated with files
// with the given name.
func (l *Language) MatchFilename(name string) bool {
	// If you adjust this implementation, remember to update CompileByFilename
	for _, n := range l.Filenames {
		if name == n {
			return true
		}
	}
	if ext := path.Ext(name); ext != "" {
		for _, x := range l.Extensions {
			if strings.EqualFold(ext, x) {
				return true
			}
		}
	}
	return false
}

// primaryMatchFilename returns true if name's extension equals l's
// primary extension or if name matches one of l's filenames.
func (l *Language) primaryMatchFilename(name string) bool {
	for _, n := range l.Filenames {
		if name == n {
			return true
		}
	}
	if ext := path.Ext(name); ext != "" && len(l.Extensions) > 0 {
		return strings.EqualFold(l.Extensions[0], ext)
	}
	return false
}

// The ConfigName()s of these must match the keys of `StaticInfo` in
// cmd/frontend/internal/pkg/langservers/langservers.go
// TODO see about adding a test for this
var builtInLanguages = map[string]struct{}{
	"Shell":      struct{}{},
	"Clojure":    struct{}{},
	"C++":        struct{}{},
	"C#":         struct{}{},
	"CSS":        struct{}{},
	"Dockerfile": struct{}{},
	"Go":         struct{}{},
	"Elixir":     struct{}{},
	"HTML":       struct{}{},
	"Java":       struct{}{},
	"JavaScript": struct{}{},
	"Lua":        struct{}{},
	"OCaml":      struct{}{},
	"PHP":        struct{}{},
	"Python":     struct{}{},
	"R":          struct{}{},
	"Ruby":       struct{}{},
	"Rust":       struct{}{},
	"TypeScript": struct{}{},
	"Haskell":    struct{}{},
}

// IsBuiltIn means the language is statically known to have a
// sourcegraph/codeintel-* language server. Currently, this is only used to
// favor languages with language servers during language detection.
func (l *Language) IsBuiltIn() bool {
	_, ok := builtInLanguages[l.Name]
	return ok
}

// Languages is a dataset of languages.
type Languages []*Language

func (ls Languages) MarshalYAML() (interface{}, error) {
	m := make(yaml.MapSlice, len(ls))
	for i, l := range ls {
		m[i] = yaml.MapItem{
			Key:   l.Name,
			Value: l,
		}
	}
	return m, nil
}

func (ls *Languages) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Note: does not preserve order.

	// var m map[string]Language
	// if err := unmarshal(&m); err != nil {
	// 	return err
	// }
	// *ls = make(Languages, 0, len(m))
	// for name, lang := range m {
	// 	lang.Name = name
	// 	*ls = append(*ls, lang)
	// }
	// return nil

	// Preserves order.

	var m yaml.MapSlice
	if err := unmarshal(&m); err != nil {
		return err
	}
	*ls = make(Languages, len(m))
	for i, mi := range m {
		b, err := yaml.Marshal(mi.Value)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(b, &(*ls)[i]); err != nil {
			return err
		}
		(*ls)[i].Name = mi.Key.(string)
	}
	return nil
}

func (ls Languages) CompileByFilename() func(string) []*Language {
	byFilename := map[string][]*Language{}
	byExt := map[string][]*Language{}
	for _, l := range ls {
		for _, n := range l.Filenames {
			byFilename[n] = append(byFilename[n], l)
		}
		for _, x := range l.Extensions {
			x = strings.ToLower(x)
			byExt[x] = append(byExt[x], l)
		}
	}
	return func(name string) []*Language {
		var matches []*Language
		matches = append(matches, byFilename[name]...)
		for _, l := range byExt[strings.ToLower(path.Ext(name))] {
			contains := false
			for _, l2 := range matches {
				if l2.Name == l.Name {
					contains = true
					break
				}
			}
			if !contains {
				matches = append(matches, l)
			}
		}
		sort.Sort(&sortByPrimaryMatch{name, matches})
		return matches
	}
}

// ByFilename returns a list of languages associated with the given
// filename or its extension.
func (ls Languages) ByFilename(name string) []*Language {
	var matches []*Language
	for _, l := range ls {
		if l.MatchFilename(name) {
			matches = append(matches, l)
		}
	}
	sort.Sort(&sortByPrimaryMatch{name, matches})
	return matches
}

// sortByPrimaryMatch sorts matches first by whether the language is built-in,
// then by whether the filename is a primary match (of the language's filenames
// or the language's primary extension), and then alphabetically by language
// name.
type sortByPrimaryMatch struct {
	name  string
	langs []*Language
}

func (v *sortByPrimaryMatch) Len() int { return len(v.langs) }

func (v *sortByPrimaryMatch) Less(i, j int) bool {
	if _, ok := builtInLanguages[v.langs[i].Name]; ok {
		return true
	}

	if _, ok := builtInLanguages[v.langs[j].Name]; ok {
		return false
	}

	ipm := v.langs[i].primaryMatchFilename(v.name)
	jpm := v.langs[j].primaryMatchFilename(v.name)
	if ipm == jpm {
		return v.langs[i].Name < v.langs[j].Name
	}
	return ipm
}

func (v *sortByPrimaryMatch) Swap(i, j int) { v.langs[i], v.langs[j] = v.langs[j], v.langs[i] }
