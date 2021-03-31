package predicate

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

// RepoContains represents the `repo:contains()` predicate,
// which filters to repos that contain either a file or content
type RepoContains struct {
	File    string
	Content string
}

func (f *RepoContains) ParseParams(params string) error {
	nodes, err := query.Parse(params, query.SearchTypeRegex)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		switch v := node.(type) {
		case query.Parameter:
			switch strings.ToLower(v.Field) {
			case "file":
				if f.File != "" {
					return errors.New("cannot specify file multiple times")
				}
				f.File = v.Value
			case "content":
				if f.Content != "" {
					return errors.New("cannot specify content multiple times")
				}
				f.Content = v.Value
			default:
				return fmt.Errorf("unsupported option %q", v.Field)
			}
		case query.Pattern:
			if f.Content != "" {
				return errors.New("cannot specify content multiple times")
			}
			f.Content = v.Value
		default:
			return fmt.Errorf("unsupported node type %T", node)
		}
	}

	if f.File == "" && f.Content == "" {
		return errors.New("one of file or content must be set")
	}

	return nil
}

func (f *RepoContains) Field() string { return query.FieldRepo }
func (f *RepoContains) Name() string  { return "contains" }
func (f *RepoContains) Plan(parent query.Basic) (query.Plan, error) {
	nodes := make([]query.Node, 0, 3)
	nodes = append(nodes, query.Parameter{
		Field: query.FieldSelect,
		Value: "repo",
	}, query.Parameter{
		Field: query.FieldCount,
		Value: "99999",
	})

	if f.File != "" {
		nodes = append(nodes, query.Parameter{
			Field: query.FieldFile,
			Value: f.File,
		})
	}

	if f.Content != "" {
		nodes = append(nodes, query.Pattern{
			Value: f.Content,
		})
	}

	nodes = append(nodes, nonPredicateRepos(parent)...)
	return query.ToPlan(query.Dnf(nodes))
}

// nonPredicateRepos returns the repo nodes in a query that aren't predicates
func nonPredicateRepos(q query.Basic) []query.Node {
	var res []query.Node
	query.VisitField(q.ToParseTree(), query.FieldRepo, func(value string, negated bool, ann query.Annotation) {
		if _, _, err := ParseAsPredicate(value); err == nil {
			// Skip predicates
			return
		}

		res = append(res, query.Parameter{
			Field:      query.FieldRepo,
			Value:      value,
			Negated:    negated,
			Annotation: ann,
		})
	})
	return res
}
