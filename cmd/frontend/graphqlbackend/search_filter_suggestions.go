package graphqlbackend

import (
	"context"
)

// Search provides search results and suggestions.
func (r *schemaResolver) SearchFilterSuggestions(ctx context.Context) (*searchFilterSuggestions, error) {
	groupsByName, err := resolveRepoGroups(ctx)
	if err != nil {
		return nil, err
	}
	repogroups := make([]string, 0, len(groupsByName))
	for name := range groupsByName {
		repogroups = append(repogroups, name)
	}

	repoRevs, _, _, err := resolveRepositories(ctx, resolveRepoOp{})
	if err != nil {
		return nil, err
	}
	const maxRepoSuggestions = 10
	repos := make([]string, 0, maxRepoSuggestions)
	for _, rev := range repoRevs {
		repos = append(repos, string(rev.Repo.Name))
		if len(repos) >= maxRepoSuggestions {
			break
		}
	}

	return &searchFilterSuggestions{
		repogroup: repogroups,
		repo:      repos,
	}, nil
}

type searchFilterSuggestions struct {
	repogroup []string
	repo      []string
}

func (s *searchFilterSuggestions) Filters() []descriptiveValue {
	return []descriptiveValue{
		{
			value:       "repo",
			description: "regex-pattern (include results whose repository path matches)",
		},
		{
			value:       "-repo",
			description: "regex-pattern (exclude results whose repository path matches)",
		},
		{
			value:       "repogroup",
			description: "group-name (include results from the named group)",
		},
		{
			value:       "repohasfile",
			description: "regex-pattern (include results from repos that contain a matching file)",
		},
		{
			value:       "repohascommitafter",
			description: `"string specifying time frame" (filter out stale repositories without recent commits)`,
		},
		{
			value:       "file",
			description: "regex-pattern (include results whose file path matches)",
		},
		{
			value:       "-file",
			description: "regex-pattern (exclude results whose file path matches)",
		},
		{
			value:       "type",
			description: "code | diff | commit | symbol",
		},
		{
			value:       "case",
			description: "yes | no (default)",
		},
		{
			value:       "lang",
			description: "lang-name (include results from the named language)",
		},
		{
			value:       "-lang",
			description: "lang-name (exclude results from the named language)",
		},
		{
			value:       "fork",
			description: "no | only | yes (default)",
		},
		{
			value:       "archived",
			description: "no | only | yes (default)",
		},
		{
			value:       "count",
			description: "integer (number of results to fetch)",
		},
		{
			value:       "timeout",
			description: `"string specifying time duration" (duration before timeout)`,
		},
	}
}

func (s *searchFilterSuggestions) Type() searchFilterDiscreteValues {
	return searchFilterDiscreteValues{
		def:    "code",
		values: []string{"code", "diff", "commit", "symbol"},
	}
}

func (s *searchFilterSuggestions) Case() searchFilterDiscreteValues {
	return searchFilterDiscreteValues{
		def:    string(No),
		values: []string{string(Yes), string(No)},
	}
}

func (s *searchFilterSuggestions) Fork() searchFilterDiscreteValues {
	return searchFilterDiscreteValues{
		def:    string(Yes),
		values: []string{string(No), string(Only), string(Yes)},
	}
}

func (s *searchFilterSuggestions) Archived() searchFilterDiscreteValues {
	return searchFilterDiscreteValues{
		def:    string(Yes),
		values: []string{string(No), string(Only), string(Yes)},
	}
}

func (s *searchFilterSuggestions) File() []descriptiveValue {
	return []descriptiveValue{
		{
			value:       `(test|spec)`,
			description: "Test files",
		},
		{
			value:       `\.json$`,
			description: "JSON files",
		},
		{
			value:       `(vendor|node_modules)/`,
			description: "Vendored code",
		},
		{
			value:       `\.md$`,
			description: "Markdown files",
		},
		{
			value:       `\.(txt|md)$`,
			description: "Text documents",
		},
	}
}

func (s *searchFilterSuggestions) Lang() []string      { return []string{"javascript", "go", "markdown"} }
func (s *searchFilterSuggestions) Repogroup() []string { return s.repogroup }
func (s *searchFilterSuggestions) Repo() []string      { return s.repo }

func (s *searchFilterSuggestions) Repohasfile() []string {
	return []string{"go.mod", "package.json", "Gemfile"}
}

func (s *searchFilterSuggestions) Repohascommitafter() []string {
	return []string{"1 week ago", "1 month ago"}
}

func (s *searchFilterSuggestions) Count() []int32    { return []int32{100, 1000} }
func (s *searchFilterSuggestions) Timeout() []string { return []string{"10s", "30s"} }

type descriptiveValue struct {
	value       string
	description string
}

func (v descriptiveValue) Value() string       { return v.value }
func (v descriptiveValue) Description() string { return v.description }

type searchFilterDiscreteValues struct {
	def    string
	values []string
}

func (v searchFilterDiscreteValues) Default() string  { return v.def }
func (v searchFilterDiscreteValues) Values() []string { return v.values }
