package querybuilder

import (
	"fmt"
	"strings"
	"time"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
	searchquery "github.com/sourcegraph/sourcegraph/internal/search/query"
	searchrepos "github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// withDefaults builds a Sourcegraph query from a base input query setting default fields if they are not specified
// in the base query. For example an input query of `repo:myrepo test` might be provided a default `archived:no`,
// and the result would be generated as `repo:myrepo test archive:no`. This preserves the semantics of the original query
// by fully parsing and reconstructing the tree, and does **not** overwrite user supplied values for the default fields.
// This also converts count:all to count:99999999.
func withDefaults(inputQuery BasicQuery, defaults searchquery.Parameters) (BasicQuery, error) {
	if len(defaults) == 0 {
		return inputQuery, nil
	}

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

// AggregationQuery takes an existing query and adds a count:all and timeout:[timeoutSeconds]s
// If a count or timeout parameter already exist in the query they will be updated.
func AggregationQuery(inputQuery BasicQuery, timeoutSeconds int, count string) (BasicQuery, error) {
	upsertParams := searchquery.Parameters{
		{
			Field:      searchquery.FieldCount,
			Value:      count,
			Negated:    false,
			Annotation: searchquery.Annotation{},
		},
		{
			Field:      searchquery.FieldTimeout,
			Value:      fmt.Sprintf("%ds", timeoutSeconds),
			Negated:    false,
			Annotation: searchquery.Annotation{},
		},
	}

	plan, err := searchquery.Pipeline(searchquery.Init(string(inputQuery), searchquery.SearchTypeLiteral))
	if err != nil {
		return "", errors.Wrap(err, "Pipeline")
	}
	modified := make(searchquery.Plan, 0, len(plan))

	for _, basic := range plan {
		p := make(searchquery.Parameters, 0, len(basic.Parameters)+len(upsertParams))

		for _, param := range basic.Parameters {
			if upsertParams.Exists(param.Field) {
				continue
			}
			p = append(p, param)
		}

		p = append(p, upsertParams...)
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
	return BasicQuery(fmt.Sprintf("%s repo:^(%s)$", query, strings.Join(escapedRepos, "|")))
}

type PointDiffQueryOpts struct {
	Before             time.Time
	After              *time.Time
	FilterRepoIncludes []string // Includes repos included from a selected context
	FilterRepoExcludes []string // includes repos excluded from a selected context
	RepoList           []string
	RepoSearch         *string
	SearchQuery        BasicQuery
}

func PointDiffQuery(diffInfo PointDiffQueryOpts) (BasicQuery, error) {
	// Build up a list of parameters that should be added to the original query
	newFilters := []searchquery.Parameter{}

	if len(diffInfo.FilterRepoIncludes) > 0 {
		newFilters = append(newFilters, searchquery.Parameter{
			Field:   searchquery.FieldRepo,
			Value:   strings.Join(diffInfo.FilterRepoIncludes, "|"),
			Negated: false,
		})
	}

	if len(diffInfo.FilterRepoExcludes) > 0 {
		newFilters = append(newFilters, searchquery.Parameter{
			Field:   searchquery.FieldRepo,
			Value:   strings.Join(diffInfo.FilterRepoExcludes, "|"),
			Negated: true,
		})
	}

	if len(diffInfo.RepoList) > 0 {
		escapedRepos := make([]string, len(diffInfo.RepoList))
		for i, repo := range diffInfo.RepoList {
			escapedRepos[i] = regexp.QuoteMeta(repo)
		}
		newFilters = append(newFilters, searchquery.Parameter{
			Field:   searchquery.FieldRepo,
			Value:   fmt.Sprintf("^(%s)$", strings.Join(escapedRepos, "|")),
			Negated: false,
		})

	}
	if diffInfo.After != nil {
		newFilters = append(newFilters, searchquery.Parameter{
			Field:   searchquery.FieldAfter,
			Value:   diffInfo.After.UTC().Format(time.RFC3339),
			Negated: false,
		})
	}
	newFilters = append(newFilters, searchquery.Parameter{
		Field:   searchquery.FieldBefore,
		Value:   diffInfo.Before.UTC().Format(time.RFC3339),
		Negated: false,
	})
	newFilters = append(newFilters, searchquery.Parameter{
		Field:   searchquery.FieldType,
		Value:   "diff",
		Negated: false,
	})

	queryPlan, err := ParseQuery(diffInfo.SearchQuery.String(), "literal")
	if err != nil {
		return "", err
	}
	modifiedPlan := make(searchquery.Plan, 0, len(queryPlan))
	for _, step := range queryPlan {
		s := make(searchquery.Parameters, 0, len(step.Parameters)+len(newFilters))
		for _, filter := range newFilters {
			s = append(s, filter)
		}
		s = append(s, step.Parameters...)
		modifiedPlan = append(modifiedPlan, step.MapParameters(s))
	}
	query := searchquery.StringHuman(modifiedPlan.ToQ())

	// If a repo search was provided treat it like its own query and combine to preserve proper groupings in compound query cases
	if diffInfo.RepoSearch != nil {
		queryWithRepo, err := MakeQueryWithRepoFilters(*diffInfo.RepoSearch, BasicQuery(query), false)
		if err != nil {
			return "", err
		}
		query = queryWithRepo.String()
	}

	return BasicQuery(query), nil
}

func PointInTimeQuery(diffInfo PointDiffQueryOpts) (BasicQuery, error) {
	// Build up a list of parameters that should be added to the original query
	newFilters := []searchquery.Parameter{}

	if len(diffInfo.FilterRepoIncludes) > 0 {
		newFilters = append(newFilters, searchquery.Parameter{
			Field:   searchquery.FieldRepo,
			Value:   strings.Join(diffInfo.FilterRepoIncludes, "|"),
			Negated: false,
		})
	}

	if len(diffInfo.FilterRepoExcludes) > 0 {
		newFilters = append(newFilters, searchquery.Parameter{
			Field:   searchquery.FieldRepo,
			Value:   strings.Join(diffInfo.FilterRepoExcludes, "|"),
			Negated: true,
		})
	}

	if len(diffInfo.RepoList) > 0 {
		escapedRepos := make([]string, len(diffInfo.RepoList))
		for i, repo := range diffInfo.RepoList {
			escapedRepos[i] = regexp.QuoteMeta(repo)
		}
		newFilters = append(newFilters, searchquery.Parameter{
			Field: searchquery.FieldRepo,
			Value: fmt.Sprintf("^(%s)$", strings.Join(escapedRepos, "|")),
		})

	}
	newFilters = append(newFilters, searchquery.Parameter{
		Field: searchquery.FieldRev,
		Value: fmt.Sprintf("at.time(%s)", diffInfo.Before.UTC().Format(time.RFC3339)),
	})

	queryPlan, err := ParseQuery(diffInfo.SearchQuery.String(), "literal")
	if err != nil {
		return "", err
	}
	modifiedPlan := make(searchquery.Plan, 0, len(queryPlan))
	for _, step := range queryPlan {
		s := make(searchquery.Parameters, 0, len(step.Parameters)+len(newFilters))
		for _, filter := range newFilters {
			s = append(s, filter)
		}
		s = append(s, step.Parameters...)
		modifiedPlan = append(modifiedPlan, step.MapParameters(s))
	}
	query := searchquery.StringHuman(modifiedPlan.ToQ())

	// If a repo search was provided treat it like its own query and combine to preserve proper groupings in compound query cases
	if diffInfo.RepoSearch != nil {
		queryWithRepo, err := MakeQueryWithRepoFilters(*diffInfo.RepoSearch, BasicQuery(query), false)
		if err != nil {
			return "", err
		}
		query = queryWithRepo.String()
	}

	return BasicQuery(query), nil
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
func ComputeInsightCommandQuery(query BasicQuery, mapType MapType, gitserverClient gitserver.Client) (ComputeInsightQuery, error) {
	q, err := ParseComputeQuery(string(query), gitserverClient)
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

// WithCount adds or updates a count paramerter for an existing query
func (q BasicQuery) WithCount(count string) (BasicQuery, error) {
	upsertParams := searchquery.Parameters{
		{
			Field:      searchquery.FieldCount,
			Value:      count,
			Negated:    false,
			Annotation: searchquery.Annotation{},
		},
	}

	plan, err := searchquery.Pipeline(searchquery.Init(string(q), searchquery.SearchTypeLiteral))
	if err != nil {
		return "", errors.Wrap(err, "Pipeline")
	}
	modified := make(searchquery.Plan, 0, len(plan))

	for _, basic := range plan {
		p := make(searchquery.Parameters, 0, len(basic.Parameters)+len(upsertParams))

		for _, param := range basic.Parameters {
			if upsertParams.Exists(param.Field) {
				continue
			}
			p = append(p, param)
		}

		p = append(p, upsertParams...)
		modified = append(modified, basic.MapParameters(p))
	}

	return BasicQuery(searchquery.StringHuman(modified.ToQ())), nil
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

func AddAuthorFilter(query BasicQuery, author string) (BasicQuery, error) {
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
			Value:      buildFilterText(author),
			Negated:    false,
			Annotation: searchquery.Annotation{},
		})
		return basic.MapParameters(modified)
	})

	return BasicQuery(searchquery.StringHuman(mutatedQuery.ToQ())), nil
}

func AddRepoFilter(query BasicQuery, repo string) (BasicQuery, error) {
	return addFilterSimple(query, searchquery.FieldRepo, repo)
}

func AddFileFilter(query BasicQuery, file string) (BasicQuery, error) {
	return addFilterSimple(query, searchquery.FieldFile, file)
}

func AddRepoMetadataFilter(query BasicQuery, repoMeta string) (BasicQuery, error) {
	if repoMeta == types.NO_REPO_METADATA_TEXT {
		return query, errors.New("Can't search for no metadata key")
	}
	plan, err := searchquery.Pipeline(searchquery.Init(string(query), searchquery.SearchTypeLiteral))
	if err != nil {
		return "", err
	}

	mutatedQuery := searchquery.MapPlan(plan, func(basic searchquery.Basic) searchquery.Basic {
		modified := make([]searchquery.Parameter, 0, len(basic.Parameters)+1)
		modified = append(modified, basic.Parameters...)
		fValue := fmt.Sprint("has.meta(", repoMeta, ")")
		meta := strings.Split(repoMeta, ":")
		if len(meta) == 2 {
			key := meta[0]
			value := meta[1]
			fValue = fmt.Sprint("has.meta(", key, ":", value, ")")
		}
		modified = append(modified, searchquery.Parameter{
			Field:      searchquery.FieldRepo,
			Value:      fValue,
			Negated:    false,
			Annotation: searchquery.Annotation{},
		})
		return basic.MapParameters(modified)
	})

	return BasicQuery(searchquery.StringHuman(mutatedQuery.ToQ())), nil
}

func buildFilterText(raw string) string {
	quoted := regexp.QuoteMeta(raw)
	if strings.Contains(raw, " ") {
		return fmt.Sprintf("(^%s$)", quoted)
	}
	return fmt.Sprintf("^%s$", quoted)
}

func AddFilter(query BasicQuery, field, value string, negated bool) (BasicQuery, error) {
	plan, err := searchquery.Pipeline(searchquery.Init(string(query), searchquery.SearchTypeLiteral))
	if err != nil {
		return "", err
	}

	mutatedQuery := searchquery.MapPlan(plan, func(basic searchquery.Basic) searchquery.Basic {
		modified := make([]searchquery.Parameter, 0, len(basic.Parameters)+1)
		modified = append(modified, basic.Parameters...)
		modified = append(modified, searchquery.Parameter{
			Field:      field,
			Value:      buildFilterText(value),
			Negated:    negated,
			Annotation: searchquery.Annotation{},
		})
		return basic.MapParameters(modified)
	})
	return BasicQuery(searchquery.StringHuman(mutatedQuery.ToQ())), nil
}

func addFilterSimple(query BasicQuery, field, value string) (BasicQuery, error) {
	return AddFilter(query, field, value, false)
}

func SetCaseSensitivity(query BasicQuery, sensitive bool) (BasicQuery, error) {
	plan, err := searchquery.Pipeline(searchquery.Init(string(query), searchquery.SearchTypeLiteral))
	if err != nil {
		return "", err
	}

	mutatedQuery := searchquery.MapPlan(plan, func(basic searchquery.Basic) searchquery.Basic {
		params := make([]searchquery.Parameter, 0, len(basic.Parameters))
		for _, parameter := range basic.Parameters {
			if parameter.Field == searchquery.FieldCase {
				continue
			}
			params = append(params, parameter)
		}

		value := "yes"
		if !sensitive {
			value = "no"
		}
		params = append(params, searchquery.Parameter{
			Field:      searchquery.FieldCase,
			Value:      value,
			Negated:    false,
			Annotation: searchquery.Annotation{},
		})

		return basic.MapParameters(params)
	})
	return BasicQuery(searchquery.StringHuman(mutatedQuery.ToQ())), nil
}

// RepositoryScopeQuery adds fork:yes archived:yes count:all to a user inputted query.
// It overwrites any input such as fork:no archived:no.
func RepositoryScopeQuery(query string) (BasicQuery, error) {
	repositoryScopeParameters := searchquery.Parameters{
		{
			Field:      searchquery.FieldFork,
			Value:      string(searchquery.Yes),
			Negated:    false,
			Annotation: searchquery.Annotation{},
		},
		{
			Field:      searchquery.FieldArchived,
			Value:      string(searchquery.Yes),
			Negated:    false,
			Annotation: searchquery.Annotation{},
		},
		{
			Field:      searchquery.FieldCount,
			Value:      "all",
			Negated:    false,
			Annotation: searchquery.Annotation{},
		},
	}
	plan, err := searchquery.Pipeline(searchquery.Init(query, searchquery.SearchTypeLiteral))
	if err != nil {
		return "", errors.Wrap(err, "Pipeline")
	}

	modified := make(searchquery.Plan, 0, len(plan))
	for _, basic := range plan {
		p := repositoryScopeParameters
		for _, param := range basic.Parameters {
			if !repositoryScopeParameters.Exists(param.Field) {
				p = append(p, param)
			}
		}
		modified = append(modified, basic.MapParameters(p))
	}
	return BasicQuery(searchquery.StringHuman(modified.ToQ())), nil
}

func MakeQueryWithRepoFilters(repositoryCriteria string, query BasicQuery, countAll bool, defaults ...searchquery.Parameter) (BasicQuery, error) {
	if countAll {
		query = withCountAll(query)
	}
	modifiedQuery, err := withDefaults(query, defaults)
	if err != nil {
		return "", errors.Wrap(err, "error parsing search query")
	}
	repositoryPlan, err := ParseQuery(repositoryCriteria, "literal")
	if err != nil {
		return "", errors.Wrap(err, "error parsing repository filters")
	}
	return BasicQuery(searchquery.StringHuman(repositoryPlan.ToQ()) + " " + modifiedQuery.String()), nil
}
