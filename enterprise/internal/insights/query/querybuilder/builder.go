package querybuilder

import (
	"fmt"
	"strings"

	"github.com/inconshreveable/log15"

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

type PatternReplacer interface {
	Replace(replacement string) (BasicQuery, error)
}

type baseReplacer struct {
	original searchquery.Plan
	pattern  string
}

type literalReplacer struct {
	baseReplacer
}

func newLiteralReplacer(plan searchquery.Plan) (*literalReplacer, error) {
	return &literalReplacer{baseReplacer{original: plan}}, nil
}

func (r *regexpReplacer) replaceContent(replacement string) (BasicQuery, error) {
	modified := searchquery.MapPattern(r.original.ToQ(), func(patternValue string, negated bool, annotation searchquery.Annotation) searchquery.Node {
		return searchquery.Pattern{
			Value:      replacement,
			Negated:    false,
			Annotation: annotation,
		}
	})

	return BasicQuery(searchquery.StringHuman(modified)), nil
}

type regexpReplacer struct {
	baseReplacer
	groups []group
}

func (r *regexpReplacer) Replace(replacement string) (BasicQuery, error) {
	log15.Info("replacer", "groups", r.groups)
	if len(r.groups) == 0 {
		// replace the entire content field
		return r.replaceContent(replacement)
	}
	var matches [][]string
	matches = append(matches, make([]string, 0, 2))
	matches[0] = append(matches[0], "") // empty value for "total match"
	matches[0] = append(matches[0], replacement)

	return BasicQuery(replaceCaptureGroupsWithString(r.pattern, r.groups, matches, 1)), nil
}

func NewPatternReplacer(query BasicQuery, searchType searchquery.SearchType) (PatternReplacer, error) {
	plan, err := searchquery.Pipeline(searchquery.Init(string(query), searchType))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse search query")
	}
	var patterns []searchquery.Pattern
	searchquery.VisitPattern(plan.ToQ(), func(value string, negated bool, annotation searchquery.Annotation) {
		patterns = append(patterns, searchquery.Pattern{
			Value:      value,
			Negated:    negated,
			Annotation: annotation,
		})
	})
	if len(patterns) > 1 {
		return nil, errors.New("pattern replacement does not support queries with multiple patterns")
	}

	pattern := patterns[0]
	br := baseReplacer{
		original: plan,
		pattern:  pattern.Value,
	}
	log15.Info("annotations:", "labels", pattern.Annotation.Labels.String())
	if !pattern.Annotation.Labels.IsSet(searchquery.Regexp) {
		return nil, errors.New("pattern replacement is only supported for regexp patterns")
	}

	regexpGroups := findGroups(pattern.Value)

	return &regexpReplacer{baseReplacer: br, groups: regexpGroups}, nil
}
