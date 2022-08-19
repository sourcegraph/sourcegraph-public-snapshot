package querybuilder

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/compute"

	"github.com/grafana/regexp"

	searchquery "github.com/sourcegraph/sourcegraph/internal/search/query"
	searchrepos "github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// withDefaults builds a Sourcegraph query from a base input query setting default fields if they are not specified
// in the base query. For example an input query of `repo:myrepo test` might be provided a default `archived:no`,
// and the result would be generated as `repo:myrepo test archive:no`. This preserves the semantics of the original query
// by fully parsing and reconstructing the tree, and does **not** overwrite user supplied values for the default fields.
func withDefaults(inputQuery BasicQuery, defaults searchquery.Parameters) (BasicQuery, error) {
	plan, err := searchquery.Pipeline(searchquery.Init(string(inputQuery), searchquery.SearchTypeLiteral))
	if err != nil {
		return "", errors.Wrap(err, "Pipeline")
	}
	modified := make(searchquery.Plan, 0, len(plan))

	for _, basic := range plan {
		p := make(searchquery.Parameters, 0, len(basic.Parameters)+len(defaults))

		for _, defaultParam := range defaults {
			if !basic.Parameters.Exists(defaultParam.Field) {
				p = append(p, defaultParam)
			}
		}
		p = append(p, basic.Parameters...)
		modified = append(modified, basic.MapParameters(p))
	}

	return BasicQuery(searchquery.StringHuman(modified.ToQ())), nil
}

// CodeInsightsQueryDefaults returns the default query parameters for a Code Insights generated Sourcegraph query.
func CodeInsightsQueryDefaults(allReposInsight bool) searchquery.Parameters {
	forkArchiveValue := searchquery.No
	if !allReposInsight {
		forkArchiveValue = searchquery.Yes
	}
	return []searchquery.Parameter{
		{
			Field:      searchquery.FieldFork,
			Value:      string(forkArchiveValue),
			Negated:    false,
			Annotation: searchquery.Annotation{},
		},
		{
			Field:      searchquery.FieldArchived,
			Value:      string(forkArchiveValue),
			Negated:    false,
			Annotation: searchquery.Annotation{},
		},
		{
			Field:      searchquery.FieldPatternType,
			Value:      "literal",
			Negated:    false,
			Annotation: searchquery.Annotation{},
		},
	}
}

// withCountAll appends a count all argument to a query if one isn't already provided.
func withCountAll(s BasicQuery) BasicQuery {
	if strings.Contains(string(s), "count:") {
		return s
	}
	return s + " count:all"
}

// forRepoRevision appends the `repo@rev` target for a Code Insight query.
func forRepoRevision(query BasicQuery, repo, revision string) BasicQuery {
	return BasicQuery(fmt.Sprintf("%s repo:^%s$@%s", query, regexp.QuoteMeta(repo), revision))
}

// forRepos appends a single repo filter making an OR condition for all repos passed
func forRepos(query BasicQuery, repos []string) BasicQuery {
	escapedRepos := make([]string, len(repos))
	for i, repo := range repos {
		escapedRepos[i] = regexp.QuoteMeta(repo)
	}
	return BasicQuery(fmt.Sprintf("%s repo:^(%s)$", query, (strings.Join(escapedRepos, "|"))))
}

// SingleRepoQuery generates a Sourcegraph query with the provided default values given a user specified query and a repository / revision target. The repository string
// should be provided in plain text, and will be escaped for regexp before being added to the query.
func SingleRepoQuery(query BasicQuery, repo, revision string, defaultParams searchquery.Parameters) (BasicQuery, error) {
	modified := withCountAll(query)
	modified, err := withDefaults(modified, defaultParams)
	if err != nil {
		return "", errors.Wrap(err, "WithDefaults")
	}
	modified = forRepoRevision(modified, repo, revision)

	return modified, nil
}

// SingleRepoQueryIndexed generates a query against the current index for one repo
func SingleRepoQueryIndexed(query BasicQuery, repo string) BasicQuery {
	modified := withCountAll(query)
	modified = forRepos(modified, []string{repo})
	return modified
}

// GlobalQuery generates a Sourcegraph query with the provided default values given a user specified query. This query will be global (against all visible repositories).
func GlobalQuery(query BasicQuery, defaultParams searchquery.Parameters) (BasicQuery, error) {
	modified := withCountAll(query)
	modified, err := withDefaults(modified, defaultParams)
	if err != nil {
		return "", errors.Wrap(err, "WithDefaults")
	}
	return modified, nil
}

// MultiRepoQuery generates a Sourcegraph query with the provided default values given a user specified query and slice of repositories.
// Repositories should be provided in plain text, and will be escaped for regexp and OR'ed together before being added to the query.
func MultiRepoQuery(query BasicQuery, repos []string, defaultParams searchquery.Parameters) (BasicQuery, error) {
	modified := withCountAll(query)
	modified, err := withDefaults(modified, defaultParams)
	if err != nil {
		return "", errors.Wrap(err, "WithDefaults")
	}
	modified = forRepos(modified, repos)

	return modified, nil
}

type MapType string

const (
	Lang   MapType = "lang"
	Repo   MapType = "repo"
	Path   MapType = "path"
	Author MapType = "author"
	Date   MapType = "date"
)

// This is the compute command that corresponds to the execution for Code Insights.
const insightsComputeCommand = "output.extra"

// ComputeInsightCommandQuery will convert a standard Sourcegraph search query into a compute "map type" insight query. This command type will group by
// certain fields. The original search query semantic should be preserved, although any new limitations or restrictions in Compute will apply.
func ComputeInsightCommandQuery(query BasicQuery, mapType MapType) (ComputeInsightQuery, error) {
	q, err := compute.Parse(string(query))
	if err != nil {
		return "", err
	}
	pattern := q.Command.ToSearchPattern()
	return ComputeInsightQuery(searchquery.AddRegexpField(q.Parameters, searchquery.FieldContent, fmt.Sprintf("%s(%s -> $%s)", insightsComputeCommand, pattern, mapType))), nil
}

type BasicQuery string
type ComputeInsightQuery string

// These string functions just exist to provide a cleaner interface for clients
func (q BasicQuery) String() string {
	return string(q)
}

func (q ComputeInsightQuery) String() string {
	return string(q)
}

var QueryNotSupported = errors.New("query not supported")

// IsSingleRepoQuery - Returns a boolean indicating if the query provided targets only a single repo.
// At this time only queries with a single query plan step are supported.  Queries with multiple plan steps
// will error with `QueryNotSupported`
func IsSingleRepoQuery(query BasicQuery) (bool, error) {

	// because we are only attempting to understand if this query targets a single repo, the search type is not relevant
	planSteps, err := searchquery.Pipeline(searchquery.Init(string(query), searchquery.SearchTypeLiteral))
	if err != nil {
		return false, err
	}

	if len(planSteps) > 1 {
		return false, QueryNotSupported
	}

	for _, step := range planSteps {
		repoFilters, _ := step.Repositories()
		if !searchrepos.ExactlyOneRepo(repoFilters) {
			return false, nil
		}
	}

	return true, nil
}

func addAuthorFilter(query BasicQuery, author string) (BasicQuery, error) {
	plan, err := searchquery.Pipeline(searchquery.Init(string(query), searchquery.SearchTypeLiteral))
	if err != nil {
		return "", err
	}

	mutatedQuery := searchquery.MapPlan(plan, func(basic searchquery.Basic) searchquery.Basic {
		modified := make([]searchquery.Parameter, 0, len(basic.Parameters)+1)
		isCommitDiffType := false
		for _, parameter := range basic.Parameters {
			modified = append(modified, parameter)
			if parameter.Field == searchquery.FieldType && (parameter.Value == "commit" || parameter.Value == "diff") {
				isCommitDiffType = true
			}
		}
		if !isCommitDiffType {
			// we can't modify this plan to accept an author so return the original input
			return basic
		}
		modified = append(modified, searchquery.Parameter{
			Field:      searchquery.FieldAuthor,
			Value:      regexp.QuoteMeta(author),
			Negated:    false,
			Annotation: searchquery.Annotation{},
		})
		return basic.MapParameters(modified)
	})

	return BasicQuery(searchquery.StringHuman(mutatedQuery.ToQ())), nil
}

func addRepoFilter(query BasicQuery, repo string) (BasicQuery, error) {
	return addFilterSimple(query, searchquery.FieldRepo, repo)
}

func addFileFilter(query BasicQuery, file string) (BasicQuery, error) {
	return addFilterSimple(query, searchquery.FieldFile, file)
}

func addFilterSimple(query BasicQuery, field, value string) (BasicQuery, error) {
	plan, err := searchquery.Pipeline(searchquery.Init(string(query), searchquery.SearchTypeLiteral))
	if err != nil {
		return "", err
	}

	mutatedQuery := searchquery.MapPlan(plan, func(basic searchquery.Basic) searchquery.Basic {
		modified := make([]searchquery.Parameter, 0, len(basic.Parameters)+1)
		modified = append(modified, basic.Parameters...)
		modified = append(modified, searchquery.Parameter{
			Field:      field,
			Value:      regexp.QuoteMeta(value),
			Negated:    false,
			Annotation: searchquery.Annotation{},
		})
		return basic.MapParameters(modified)
	})
	return BasicQuery(searchquery.StringHuman(mutatedQuery.ToQ())), nil
}

type modificationFunc func(plan searchquery.Plan) (searchquery.Plan, error)

type builder struct {
	originalPlan  searchquery.Plan
	modifications []modificationFunc
}

var failedToParseQueryErr = errors.New("failed to parse search query")

func FromQuery(query string, searchType searchquery.SearchType) (*builder, error) {
	plan, err := searchquery.Pipeline(searchquery.Init(query, searchType))
	if err != nil {
		return nil, failedToParseQueryErr
	}

	return &builder{originalPlan: plan}, nil
}

func (b *builder) Build() (string, error) {
	result, err := b.doModifications()
	if err != nil {
		return "", errors.Wrap(err, "failed to apply search query modifications")
	}
	return searchquery.StringHuman(result.ToQ()), nil
}

func (b *builder) WithRepo(repo string) *builder {
	mod := func(plan searchquery.Plan) (searchquery.Plan, error) {
		return escapedField(plan, searchquery.FieldRepo, repo)
	}
	b.modifications = append(b.modifications, mod)
	return b
}

func (b *builder) WithRepoPattern(repoPattern string) *builder {
	mod := func(plan searchquery.Plan) (searchquery.Plan, error) {
		return newField(plan, searchquery.FieldRepo, repoPattern)
	}
	b.modifications = append(b.modifications, mod)
	return b
}

func (b *builder) WithRepoRevision(repo, revision string) *builder {
	mod := func(plan searchquery.Plan) (searchquery.Plan, error) {
		return escapedField(plan, searchquery.FieldRepo, fmt.Sprintf("%s@%s", repo, revision))
	}
	b.modifications = append(b.modifications, mod)
	return b
}

func (b *builder) WithFile(file string) *builder {
	mod := func(plan searchquery.Plan) (searchquery.Plan, error) {
		return escapedField(plan, searchquery.FieldFile, file)
	}
	b.modifications = append(b.modifications, mod)
	return b
}

func (b *builder) WithFilePattern(filePattern string) *builder {
	mod := func(plan searchquery.Plan) (searchquery.Plan, error) {
		return newField(plan, searchquery.FieldFile, filePattern)
	}
	b.modifications = append(b.modifications, mod)
	return b
}

func (b *builder) WithCountAll() *builder {
	mod := func(plan searchquery.Plan) (searchquery.Plan, error) {
		return searchquery.MapPlan(plan, func(basic searchquery.Basic) searchquery.Basic {
			params := make([]searchquery.Parameter, 0, len(basic.Parameters))
			for _, parameter := range basic.Parameters {
				if parameter.Field == searchquery.FieldCount {
					continue
				}
				params = append(params, parameter)
			}
			params = append(params, searchquery.Parameter{
				Field:      searchquery.FieldCount,
				Value:      "all",
				Negated:    false,
				Annotation: searchquery.Annotation{},
			})
			return basic.MapParameters(params)
		}), nil
	}
	b.modifications = append(b.modifications, mod)
	return b
}

func (b *builder) WithAuthor(author string) *builder {
	mod := func(plan searchquery.Plan) (searchquery.Plan, error) {
		mutatedQuery := searchquery.MapPlan(plan, func(basic searchquery.Basic) searchquery.Basic {
			modified := make([]searchquery.Parameter, 0, len(basic.Parameters)+1)
			isCommitDiffType := false
			for _, parameter := range basic.Parameters {
				modified = append(modified, parameter)
				if parameter.Field == searchquery.FieldType && (parameter.Value == "commit" || parameter.Value == "diff") {
					isCommitDiffType = true
				}
			}
			if !isCommitDiffType {
				// we can't modify this query to accept an author so return the original input
				return basic
			}
			modified = append(modified, searchquery.Parameter{
				Field:      searchquery.FieldAuthor,
				Value:      regexp.QuoteMeta(author),
				Negated:    false,
				Annotation: searchquery.Annotation{},
			})
			return basic.MapParameters(modified)
		})
		return mutatedQuery, nil
	}
	b.modifications = append(b.modifications, mod)
	return b
}

func escapedField(plan searchquery.Plan, field, value string) (searchquery.Plan, error) {
	return newField(plan, field, regexp.QuoteMeta(value))
}

func newField(plan searchquery.Plan, field, value string) (searchquery.Plan, error) {
	mutated := searchquery.MapPlan(plan, func(basic searchquery.Basic) searchquery.Basic {
		modified := make([]searchquery.Parameter, 0, len(basic.Parameters)+1)
		modified = append(modified, basic.Parameters...)
		modified = append(modified, searchquery.Parameter{
			Field:      field,
			Value:      value,
			Negated:    false,
			Annotation: searchquery.Annotation{},
		})
		return basic.MapParameters(modified)
	})

	return mutated, nil
}

func (b *builder) doModifications() (searchquery.Plan, error) {
	current := b.originalPlan
	for _, modification := range b.modifications {
		next, err := modification(current)
		if err != nil {
			return nil, err
		}
		current = next
	}
	return current, nil
}
