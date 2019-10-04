package graphqlbackend

import (
	"context"
)

// SearchFilterSuggestions provides search filter and default value suggestions.
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

// searchFilterSuggestions holds suggestions of search filters and their default values.
type searchFilterSuggestions struct {
	repogroup []string
	repo      []string
}

// Filters returns a list of filters and their description.
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

// Type returns discrete and default values of search filter "type:".
func (s *searchFilterSuggestions) Type() searchFilterDiscreteValues {
	return searchFilterDiscreteValues{
		def:    "code",
		values: []string{"code", "diff", "commit", "symbol"},
	}
}

// Case returns discrete and default values of search filter "case:".
func (s *searchFilterSuggestions) Case() searchFilterDiscreteValues {
	return searchFilterDiscreteValues{
		def:    string(No),
		values: []string{string(Yes), string(No)},
	}
}

// Fork returns discrete and default values of search filter "fork:".
func (s *searchFilterSuggestions) Fork() searchFilterDiscreteValues {
	return searchFilterDiscreteValues{
		def:    string(Yes),
		values: []string{string(No), string(Only), string(Yes)},
	}
}

// Archived returns discrete and default values of search filter "archived:".
func (s *searchFilterSuggestions) Archived() searchFilterDiscreteValues {
	return searchFilterDiscreteValues{
		def:    string(Yes),
		values: []string{string(No), string(Only), string(Yes)},
	}
}

// File returns example values of search filter "file:".
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

// Lang returns example values of search filter "lang:".
func (s *searchFilterSuggestions) Lang() []string { return []string{"javascript", "go", "markdown"} }

// Repogroup returns all repository groups defined in the settings.
func (s *searchFilterSuggestions) Repogroup() []string { return s.repogroup }

// Repo returns a list of repositories as the default value for suggestion.
func (s *searchFilterSuggestions) Repo() []string { return s.repo }

// Repohasfile returns example values of search filter "repohasfile:".
func (s *searchFilterSuggestions) Repohasfile() []string {
	return []string{"go.mod", "package.json", "Dockerfile"}
}

// Repohascommitafter returns example values of search filter "repohascommitafter:".
func (s *searchFilterSuggestions) Repohascommitafter() []string {
	return []string{`"1 week ago"`, `"last Thursday"`, `"June 25 2017"`}
}

// Repohascommitafter returns example values of search filter "repohascommitafter:".
func (s *searchFilterSuggestions) Count() []int32 { return []int32{100, 1000} }

// Timeout returns example values of search filter "timeout:".
func (s *searchFilterSuggestions) Timeout() []string { return []string{"10s", "30s"} }

// descriptiveValue contains a value with its description.
type descriptiveValue struct {
	value       string
	description string
}

func (v descriptiveValue) Value() string       { return v.value }
func (v descriptiveValue) Description() string { return v.description }

// searchFilterDiscreteValues contains a list of discrete values and the default choice.
type searchFilterDiscreteValues struct {
	def    string
	values []string
}

func (v searchFilterDiscreteValues) Default() string  { return v.def }
func (v searchFilterDiscreteValues) Values() []string { return v.values }
