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
		repogroups: repogroups,
		repos:      repos,
	}, nil
}

// Static values of search filter suggestions.
var (
	searchFilterSuggestionsFilters = []descriptiveValue{
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
			description: "lang-name (include results from the named language, e.g. go)",
		},
		{
			value:       "-lang",
			description: "lang-name (exclude results from the named language, e.g. go)",
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
			description: "integer (number of results to fetch, e.g. 1000)",
		},
		{
			value:       "timeout",
			description: `"string specifying time duration" (duration before timeout, e.g. 30s)`,
		},
	}
	searchFilterSuggestionsType = searchFilterDiscreteValues{
		def:    "code",
		values: []string{"code", "diff", "commit", "symbol"},
	}
	searchFilterSuggestionsCase = searchFilterDiscreteValues{
		def:    string(No),
		values: []string{string(Yes), string(No)},
	}
	searchFilterSuggestionsFork = searchFilterDiscreteValues{
		def:    string(Yes),
		values: []string{string(No), string(Only), string(Yes)},
	}
	searchFilterSuggestionsArchived = searchFilterDiscreteValues{
		def:    string(Yes),
		values: []string{string(No), string(Only), string(Yes)},
	}
	searchFilterSuggestionsFile = []descriptiveValue{
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
	searchFilterSuggestionsLang               = []string{"javascript", "go", "markdown"}
	searchFilterSuggestionsRepohasfile        = []string{"go.mod", "package.json", "Dockerfile"}
	searchFilterSuggestionsRepohascommitafter = []string{`"1 week ago"`, `"last Thursday"`, `"June 25 2017"`}
	searchFilterSuggestionsCount              = []int32{10, 100, 1000}
	searchFilterSuggestionsTimeout            = []string{"10s", "30s"}
)

// searchFilterSuggestions holds suggestions of search filters and their default values.
type searchFilterSuggestions struct {
	repogroups []string
	repos      []string
}

// Filters returns a list of filters and their description.
func (s *searchFilterSuggestions) Filters() []descriptiveValue {
	return searchFilterSuggestionsFilters
}

// Type returns discrete and default values of search filter "type:".
func (s *searchFilterSuggestions) Type() searchFilterDiscreteValues {
	return searchFilterSuggestionsType
}

// Case returns discrete and default values of search filter "case:".
func (s *searchFilterSuggestions) Case() searchFilterDiscreteValues {
	return searchFilterSuggestionsCase
}

// Fork returns discrete and default values of search filter "fork:".
func (s *searchFilterSuggestions) Fork() searchFilterDiscreteValues {
	return searchFilterSuggestionsFork
}

// Archived returns discrete and default values of search filter "archived:".
func (s *searchFilterSuggestions) Archived() searchFilterDiscreteValues {
	return searchFilterSuggestionsArchived
}

// File returns example values of search filter "file:".
func (s *searchFilterSuggestions) File() []descriptiveValue {
	return searchFilterSuggestionsFile
}

// Lang returns example values of search filter "lang:".
func (s *searchFilterSuggestions) Lang() []string {
	return searchFilterSuggestionsLang
}

// Repogroup returns all repository groups defined in the settings.
func (s *searchFilterSuggestions) Repogroup() []string {
	return s.repogroups
}

// Repo returns a list of repositories as the default value for suggestion.
func (s *searchFilterSuggestions) Repo() []string {
	return s.repos
}

// Repohasfile returns example values of search filter "repohasfile:".
func (s *searchFilterSuggestions) Repohasfile() []string {
	return searchFilterSuggestionsRepohasfile
}

// Repohascommitafter returns example values of search filter "repohascommitafter:".
func (s *searchFilterSuggestions) Repohascommitafter() []string {
	return searchFilterSuggestionsRepohascommitafter
}

// Repohascommitafter returns example values of search filter "repohascommitafter:".
func (s *searchFilterSuggestions) Count() []int32 {
	return searchFilterSuggestionsCount
}

// Timeout returns example values of search filter "timeout:".
func (s *searchFilterSuggestions) Timeout() []string {
	return searchFilterSuggestionsTimeout
}

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
