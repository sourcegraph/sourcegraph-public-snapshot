package compute

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	searchresult "github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/languages"
)

// Template is just a list of Atom, where an Atom is either a Variable or a Constant string.
type Template []Atom

type Atom interface {
	atom()
	String() string
}

type Attribute string

const (
	LengthAttr Attribute = "length"
	RangeAttr  Attribute = "range"
)

// Variable represents a variable in the template that may be substituted for. A
// variable is optionally qualified by an attribute, which is data associated
// with a variable (e.g., length, range). Attributes are currently unused, and
// exist for future expansion.
type Variable struct {
	Name      string
	Attribute Attribute
}

type Constant string

func (Variable) atom() {}
func (Constant) atom() {}

func (v Variable) String() string {
	if v.Attribute != "" {
		return v.Name + "." + string(v.Attribute)
	}
	return v.Name
}
func (c Constant) String() string { return string(c) }

const varAllowed = "abcdefghijklmnopqrstuvwxyzABCEDEFGHIJKLMNOPQRSTUVWXYZ1234567890_."

// scanTemplate scans an input string to produce a Template. Recognized
// metavariable syntax is `$(varAllowed+)`.
func scanTemplate(buf []byte) *Template {
	// Tracks whether the current token is a variable.
	var isVariable bool

	var start int
	var r rune
	var token []rune
	var result []Atom

	next := func() rune {
		r, start := utf8.DecodeRune(buf)
		buf = buf[start:]
		return r
	}

	appendAtom := func(atom Atom) {
		if a, ok := atom.(Constant); ok && len(a) == 0 {
			return
		}
		if a, ok := atom.(Variable); ok && len(a.Name) == 0 {
			return
		}
		result = append(result, atom)
		// Reset token, but reuse the backing memory.
		token = token[:0]
	}

	for len(buf) > 0 {
		r = next()
		switch r {
		case '$':
			if len(buf[start:]) > 0 {
				if r, _ = utf8.DecodeRune(buf); strings.ContainsRune(varAllowed, r) {
					// Start of a recognized variable.
					if isVariable {
						// We were busy scanning a variable.
						appendAtom(Variable{Name: string(token)}) // Push variable.
					} else {
						// We were busy scanning a constant.
						appendAtom(Constant(token))
					}
					token = append(token, '$')
					isVariable = true
					continue
				}
				// Something else, push the '$' we saw and continue.
				token = append(token, '$')
				isVariable = false
				continue
			}
			// Trailing '$'
			if isVariable {
				appendAtom(Variable{Name: string(token)}) // Push variable.
				isVariable = false
			} else {
				appendAtom(Constant(token))
			}
			token = append(token, '$')
		case '\\':
			if isVariable {
				// We were busy scanning a variable. A '\' always terminates it.
				appendAtom(Variable{Name: string(token)}) // Push variable.
				isVariable = false
			}
			if len(buf[start:]) > 0 {
				r = next()
				switch r {
				case 'n':
					token = append(token, '\n')
				case 'r':
					token = append(token, '\r')
				case 't':
					token = append(token, '\t')
				case '\\', '$', ' ', '.':
					token = append(token, r)
				default:
					token = append(token, '\\', r)
				}
				continue
			}
			// Trailing '\'
			token = append(token, '\\')
		default:
			if isVariable && !strings.ContainsRune(varAllowed, r) {
				appendAtom(Variable{Name: string(token)}) // Push variable.
				isVariable = false
			}
			token = append(token, r)
		}
	}
	if len(token) > 0 {
		if isVariable {
			appendAtom(Variable{Name: string(token)})
		} else {
			appendAtom(Constant(token))
		}
	}
	t := Template(result)
	return &t
}

func toJSON(atom Atom) any {
	switch a := atom.(type) {
	case Constant:
		return struct {
			Value string `json:"constant"`
		}{
			Value: string(a),
		}
	case Variable:
		return struct {
			Name      string `json:"variable"`
			Attribute string `json:"attribute,omitempty"`
		}{
			Name:      a.Name,
			Attribute: string(a.Attribute),
		}
	}
	panic("unreachable")
}

func toJSONString(template *Template) string {
	var jsons []any
	for _, atom := range *template {
		jsons = append(jsons, toJSON(atom))
	}
	j, _ := json.Marshal(jsons)
	return string(j)
}

type MetaEnvironment struct {
	Repo    string
	Path    string
	Content string
	Commit  string
	Author  string
	Date    time.Time
	Email   string
	Lang    string
	Owner   string
}

var empty = struct{}{}

var builtinVariables = map[string]struct{}{
	"repo":            empty,
	"path":            empty,
	"content":         empty,
	"commit":          empty,
	"author":          empty,
	"date":            empty,
	"date.day":        empty,
	"date.month":      empty,
	"date.month.name": empty,
	"date.year":       empty,
	"email":           empty,
	"lang":            empty,
}

func templatize(pattern string, env *MetaEnvironment) string {
	t := scanTemplate([]byte(pattern))
	var templatized []string
	for _, atom := range *t {
		switch a := atom.(type) {
		case Constant:
			templatized = append(templatized, string(a))
		case Variable:
			if _, ok := builtinVariables[a.Name[1:]]; ok {
				switch a.Name[1:] {
				case "date.year":
					templatized = append(templatized, strconv.Itoa(env.Date.Year()))
				case "date.month.name":
					templatized = append(templatized, env.Date.Month().String())
				case "date.month":
					templatized = append(templatized, fmt.Sprintf("%02d", int(env.Date.Month())))
				case "date.day":
					templatized = append(templatized, strconv.Itoa(env.Date.Day()))
				case "date":
					templatized = append(templatized, env.Date.Format("2006-01-02"))
				default:
					templateVar := cases.Title(language.English).String(a.Name[1:])
					templatized = append(templatized, `{{.`+templateVar+`}}`)
				}
				continue
			}
			// Leave alone other variables that don't correspond to
			// builtins (e.g., regex capture groups)
			templatized = append(templatized, a.Name)
		}
	}
	return strings.Join(templatized, "")
}

func substituteMetaVariables(pattern string, env *MetaEnvironment) (string, error) {
	templated := templatize(pattern, env)
	t, err := template.New("").Parse(templated)
	if err != nil {
		return "", err
	}
	var result strings.Builder
	if err := t.Execute(&result, env); err != nil {
		return "", err
	}
	return result.String(), nil
}

// NewMetaEnvironment maps results to a metavariable:value environment where
// metavariables can be referenced and substituted for in an output template.
func NewMetaEnvironment(r searchresult.Match, content string) *MetaEnvironment {
	switch m := r.(type) {
	case *searchresult.RepoMatch:
		return &MetaEnvironment{
			Repo:    string(m.Name),
			Content: string(m.Name),
		}
	case *searchresult.FileMatch:
		// FIXME(id: language-detection-failure-handling):
		// Handle failure in language detection as well as ambiguity
		langs, _ := languages.GetLanguages(m.Path, func() ([]byte, error) {
			return []byte(content), nil
		})
		lang := ""
		if len(langs) > 0 {
			lang = langs[0]
		}
		return &MetaEnvironment{
			Repo:    string(m.Repo.Name),
			Path:    m.Path,
			Commit:  string(m.CommitID),
			Content: content,
			Lang:    lang,
		}
	case *searchresult.CommitMatch:
		return &MetaEnvironment{
			Repo:    string(m.Repo.Name),
			Commit:  string(m.Commit.ID),
			Author:  m.Commit.Author.Name,
			Date:    m.Commit.Committer.Date,
			Email:   m.Commit.Author.Email,
			Content: content,
		}
	case *searchresult.CommitDiffMatch:
		path := m.Path()
		// FIXME(id: language-detection-failure-handling):
		// Handle failure in language detection as well as ambiguity
		langs, _ := languages.GetLanguages(path, func() ([]byte, error) {
			return []byte(content), nil
		})
		lang := ""
		if len(langs) > 0 {
			lang = langs[0]
		}
		return &MetaEnvironment{
			Repo:    string(m.Repo.Name),
			Commit:  string(m.Commit.ID),
			Author:  m.Commit.Author.Name,
			Date:    m.Commit.Committer.Date,
			Email:   m.Commit.Author.Email,
			Path:    path,
			Lang:    lang,
			Content: content,
		}
	case *searchresult.OwnerMatch:
		return &MetaEnvironment{
			Repo:    string(m.Repo.Name),
			Commit:  string(m.CommitID),
			Owner:   m.ResolvedOwner.Identifier(),
			Content: content,
		}
	}
	return &MetaEnvironment{}
}
