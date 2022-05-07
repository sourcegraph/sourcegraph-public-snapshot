package jobutil

import (
	"strings"
	"time"

	"github.com/grafana/regexp"
	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/commit"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/limits"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	searchrepos "github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/searchcontexts"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/search/structural"
	"github.com/sourcegraph/sourcegraph/internal/search/symbol"
	"github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewFlatJob converts a query parse tree to the _internal_ representation
// needed to run a search routine. To understand why this conversion matters, think
// about the fact that the query parse tree doesn't know anything about our
// backends or architecture. It doesn't decide certain defaults, like whether we
// should return multiple result types (pattern matches content, or a file name,
// or a repo name). If we want to optimize a Sourcegraph query parse tree for a
// particular backend (e.g., skip repository resolution and just run a Zoekt
// query on all indexed repositories) then we need to convert our tree to
// Zoekt's internal inputs and representation. These concerns are all handled by
// NewFlatJob.
func NewFlatJob(searchInputs *run.SearchInputs, f query.Flat) (job.Job, error) {
	maxResults := f.MaxResults(searchInputs.DefaultLimit())
	types, _ := f.IncludeExcludeValues(query.FieldType)
	resultTypes := computeResultTypes(types, f.ToBasic(), searchInputs.PatternType)
	patternInfo := toTextPatternInfo(f.ToBasic(), resultTypes, searchInputs.Protocol)

	// searcher to use full deadline if timeout: set or we are streaming.
	useFullDeadline := f.GetTimeout() != nil || f.Count() != nil || searchInputs.Protocol == search.Streaming

	fileMatchLimit := int32(computeFileMatchLimit(f.ToBasic(), searchInputs.Protocol))
	selector, _ := filter.SelectPathFromString(f.FindValue(query.FieldSelect)) // Invariant: select is validated

	features := toFeatures(searchInputs.Features)
	repoOptions := toRepoOptions(f.ToBasic(), searchInputs.UserSettings)

	repoUniverseSearch, skipRepoSubsetSearch, _ := jobMode(f.ToBasic(), resultTypes, searchInputs.PatternType, searchInputs.OnSourcegraphDotCom)

	var allJobs []job.Job
	addJob := func(job job.Job) {
		allJobs = append(allJobs, job)
	}

	{
		// This code block creates search jobs under specific
		// conditions, and depending on generic process of `args` above.
		// It which specializes search logic in doResults. In time, all
		// of the above logic should be used to create search jobs
		// across all of Sourcegraph.

		// Create Text Search Jobs
		if resultTypes.Has(result.TypeFile | result.TypePath) {
			// Create Text Search jobs over repo set.
			if !skipRepoSubsetSearch {
				searcherJob := &searcher.SearcherJob{
					PatternInfo:     patternInfo,
					Indexed:         false,
					UseFullDeadline: useFullDeadline,
				}

				addJob(&repoPagerJob{
					child:            searcherJob,
					repoOptions:      repoOptions,
					useIndex:         f.Index(),
					containsRefGlobs: query.ContainsRefGlobs(f.ToBasic().ToParseTree()),
				})
			}
		}

		// Create Symbol Search Jobs
		if resultTypes.Has(result.TypeSymbol) {
			// Create Symbol Search jobs over repo set.
			if !skipRepoSubsetSearch {
				symbolSearchJob := &searcher.SymbolSearcherJob{
					PatternInfo: patternInfo,
					Limit:       maxResults,
				}

				addJob(&repoPagerJob{
					child:            symbolSearchJob,
					repoOptions:      repoOptions,
					useIndex:         f.Index(),
					containsRefGlobs: query.ContainsRefGlobs(f.ToBasic().ToParseTree()),
				})
			}
		}

		if resultTypes.Has(result.TypeStructural) {
			typ := search.TextRequest
			zoektQuery, err := zoekt.QueryToZoektQuery(f.ToBasic(), resultTypes, &features, typ)
			if err != nil {
				return nil, err
			}
			zoektArgs := &search.ZoektParameters{
				Query:          zoektQuery,
				Typ:            typ,
				FileMatchLimit: fileMatchLimit,
				Select:         selector,
			}

			searcherArgs := &search.SearcherParameters{
				PatternInfo:     patternInfo,
				UseFullDeadline: useFullDeadline,
			}

			addJob(&structural.StructuralSearchJob{
				ZoektArgs:        zoektArgs,
				SearcherArgs:     searcherArgs,
				UseIndex:         f.Index(),
				ContainsRefGlobs: query.ContainsRefGlobs(f.ToBasic().ToParseTree()),
				RepoOpts:         repoOptions,
			})
		}

		if resultTypes.Has(result.TypeRepo) {
			valid := func() bool {
				fieldAllowlist := map[string]struct{}{
					query.FieldRepo:               {},
					query.FieldContext:            {},
					query.FieldType:               {},
					query.FieldDefault:            {},
					query.FieldIndex:              {},
					query.FieldCount:              {},
					query.FieldTimeout:            {},
					query.FieldFork:               {},
					query.FieldArchived:           {},
					query.FieldVisibility:         {},
					query.FieldCase:               {},
					query.FieldRepoHasFile:        {},
					query.FieldRepoHasCommitAfter: {},
					query.FieldPatternType:        {},
					query.FieldSelect:             {},
				}

				// Don't run a repo search if the search contains fields that aren't on the allowlist.
				exists := true
				query.VisitParameter(f.ToBasic().ToParseTree(), func(field, _ string, _ bool, _ query.Annotation) {
					if _, ok := fieldAllowlist[field]; !ok {
						exists = false
					}
				})
				return exists
			}

			// returns an updated RepoOptions if the pattern part of a query can be used to
			// search repos. A problematic case we check for is when the pattern contains `@`,
			// which may confuse downstream logic to interpret it as part of `repo@rev` syntax.
			addPatternAsRepoFilter := func(pattern string, opts search.RepoOptions) (search.RepoOptions, bool) {
				if pattern == "" {
					return opts, true
				}

				opts.RepoFilters = append(make([]string, 0, len(opts.RepoFilters)), opts.RepoFilters...)
				opts.CaseSensitiveRepoFilters = f.IsCaseSensitive()

				patternPrefix := strings.SplitN(pattern, "@", 2)
				if len(patternPrefix) == 1 {
					// No "@" in pattern? We're good.
					opts.RepoFilters = append(opts.RepoFilters, pattern)
					return opts, true
				}

				if patternPrefix[0] != "" {
					// Extend the repo search using the pattern value, but
					// since the pattern contains @, only search the part
					// prefixed by the first @. This because downstream
					// logic will get confused by the presence of @ and try
					// to resolve repo revisions. See #27816.
					if _, err := regexp.Compile(patternPrefix[0]); err != nil {
						// Prefix is not valid regexp, so just reject it. This can happen for patterns where we've automatically added `(...).*?(...)`
						// such as `foo @bar` which becomes `(foo).*?(@bar)`, which when stripped becomes `(foo).*?(` which is unbalanced and invalid.
						// Why is this a mess? Because validation for everything, including repo values, should be done up front so far possible, not downtsream
						// after possible modifications. By the time we reach this code, the pattern should already have been considered valid to continue with
						// a search. But fixing the order of concerns for repo code is not something @rvantonder is doing today.
						return search.RepoOptions{}, false
					}
					opts.RepoFilters = append(opts.RepoFilters, patternPrefix[0])
					return opts, true
				}

				// This pattern starts with @, of the form "@thing". We can't
				// consistently handle search repos of this form, because
				// downstream logic will attempt to interpret "thing" as a repo
				// revision, may fail, and cause us to raise an alert for any
				// non `type:repo` search. Better to not attempt a repo search.
				return search.RepoOptions{}, false
			}

			if valid() {
				if repoOptions, ok := addPatternAsRepoFilter(f.ToBasic().PatternString(), repoOptions); ok {
					var mode search.GlobalSearchMode
					if repoUniverseSearch {
						mode = search.ZoektGlobalSearch
					}
					if skipRepoSubsetSearch {
						mode = search.SkipUnindexed
					}
					addJob(&run.RepoSearchJob{
						RepoOpts:                     repoOptions,
						Features:                     features,
						FilePatternsReposMustInclude: patternInfo.FilePatternsReposMustInclude,
						FilePatternsReposMustExclude: patternInfo.FilePatternsReposMustExclude,
						Mode:                         mode,
					})
				}
			}
		}
	}

	return NewParallelJob(allJobs...), nil
}

func mapSlice(values []string, f func(string) string) []string {
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = f(v)
	}
	return result
}

func count(q query.Basic, p search.Protocol) int {
	if count := q.Count(); count != nil {
		return *count
	}

	if q.IsStructural() {
		return limits.DefaultMaxSearchResults
	}

	switch p {
	case search.Batch:
		return limits.DefaultMaxSearchResults
	case search.Streaming:
		return limits.DefaultMaxSearchResultsStreaming
	}
	panic("unreachable")
}

// toTextPatternInfo converts a an atomic query to internal values that drive
// text search. An atomic query is a Basic query where the Pattern is either
// nil, or comprises only one Pattern node (hence, an atom, and not an
// expression). See TextPatternInfo for the values it computes and populates.
func toTextPatternInfo(q query.Basic, resultTypes result.Types, p search.Protocol) *search.TextPatternInfo {
	// Handle file: and -file: filters.
	filesInclude, filesExclude := q.IncludeExcludeValues(query.FieldFile)
	// Handle lang: and -lang: filters.
	langInclude, langExclude := q.IncludeExcludeValues(query.FieldLang)
	filesInclude = append(filesInclude, mapSlice(langInclude, query.LangToFileRegexp)...)
	filesExclude = append(filesExclude, mapSlice(langExclude, query.LangToFileRegexp)...)
	filesReposMustInclude, filesReposMustExclude := q.IncludeExcludeValues(query.FieldRepoHasFile)
	selector, _ := filter.SelectPathFromString(q.FindValue(query.FieldSelect)) // Invariant: select is validated
	count := count(q, p)

	// Ugly assumption: for a literal search, the IsRegexp member of
	// TextPatternInfo must be set true. The logic assumes that a literal
	// pattern is an escaped regular expression.
	isRegexp := q.IsLiteral() || q.IsRegexp()

	if q.Pattern == nil {
		// For compatibility: A nil pattern implies isRegexp is set to
		// true. This has no effect on search logic.
		isRegexp = true
	}

	negated := false
	if p, ok := q.Pattern.(query.Pattern); ok {
		negated = p.Negated
	}

	return &search.TextPatternInfo{
		// Values dependent on pattern atom.
		IsRegExp:        isRegexp,
		IsStructuralPat: q.IsStructural(),
		IsCaseSensitive: q.IsCaseSensitive(),
		FileMatchLimit:  int32(count),
		Pattern:         q.PatternString(),
		IsNegated:       negated,

		// Values dependent on parameters.
		IncludePatterns:              filesInclude,
		ExcludePattern:               query.UnionRegExps(filesExclude),
		FilePatternsReposMustInclude: filesReposMustInclude,
		FilePatternsReposMustExclude: filesReposMustExclude,
		PatternMatchesPath:           resultTypes.Has(result.TypePath),
		PatternMatchesContent:        resultTypes.Has(result.TypeFile),
		Languages:                    langInclude,
		PathPatternsAreCaseSensitive: q.IsCaseSensitive(),
		CombyRule:                    q.FindValue(query.FieldCombyRule),
		Index:                        q.Index(),
		Select:                       selector,
	}
}

// computeResultTypes returns result types based three inputs: `type:...` in the query,
// the `pattern`, and top-level `searchType` (coming from a GQL value).
func computeResultTypes(types []string, b query.Basic, searchType query.SearchType) result.Types {
	var rts result.Types
	if searchType == query.SearchTypeStructural && !b.IsEmptyPattern() {
		rts = result.TypeStructural
	} else {
		if len(types) == 0 {
			rts = result.TypeFile | result.TypePath | result.TypeRepo
		} else {
			for _, t := range types {
				rts = rts.With(result.TypeFromString[t])
			}
		}
	}
	return rts
}

func toRepoOptions(b query.Basic, userSettings *schema.Settings) search.RepoOptions {
	repoFilters, minusRepoFilters := b.Repositories()

	var settingForks, settingArchived bool
	if v := userSettings.SearchIncludeForks; v != nil {
		settingForks = *v
	}
	if v := userSettings.SearchIncludeArchived; v != nil {
		settingArchived = *v
	}

	fork := query.No
	if searchrepos.ExactlyOneRepo(repoFilters) || settingForks {
		// fork defaults to No unless either of:
		// (1) exactly one repo is being searched, or
		// (2) user/org/global setting includes forks
		fork = query.Yes
	}
	if setFork := b.Fork(); setFork != nil {
		fork = *setFork
	}

	archived := query.No
	if searchrepos.ExactlyOneRepo(repoFilters) || settingArchived {
		// archived defaults to No unless either of:
		// (1) exactly one repo is being searched, or
		// (2) user/org/global setting includes archives in all searches
		archived = query.Yes
	}
	if setArchived := b.Archived(); setArchived != nil {
		archived = *setArchived
	}

	visibility := b.Visibility()
	commitAfter := b.FindValue(query.FieldRepoHasCommitAfter)
	searchContextSpec := b.FindValue(query.FieldContext)

	return search.RepoOptions{
		RepoFilters:       repoFilters,
		MinusRepoFilters:  minusRepoFilters,
		Dependencies:      b.Dependencies(),
		SearchContextSpec: searchContextSpec,
		ForkSet:           b.Fork() != nil,
		OnlyForks:         fork == query.Only,
		NoForks:           fork == query.No,
		ArchivedSet:       b.Archived() != nil,
		OnlyArchived:      archived == query.Only,
		NoArchived:        archived == query.No,
		Visibility:        visibility,
		CommitAfter:       commitAfter,
	}
}

// jobBuilder represents computed static values that are backend agnostic: we
// generally need to compute these values before we're able to create (or build)
// multiple specific jobs. If you want to add new fields or state to run a
// search, ask yourself: is this value specific to a backend like Zoekt,
// searcher, or gitserver, or a new backend? If yes, then that new field does
// not belong in this builder type, and your new field should probably be
// computed either using values in this builder, or obtained from the outside
// world where you construct your specific search job.
//
// If you _may_ need the value available to start a search across differnt
// backends, then this builder type _may_ be the right place for it to live.
// If in doubt, ask the search team.
type jobBuilder struct {
	query          query.Basic
	resultTypes    result.Types
	repoOptions    search.RepoOptions
	features       *search.Features
	fileMatchLimit int32
	selector       filter.SelectPath
}

func (b *jobBuilder) newZoektGlobalSearch(typ search.IndexedRequestType) (job.Job, error) {
	zoektQuery, err := zoekt.QueryToZoektQuery(b.query, b.resultTypes, b.features, typ)
	if err != nil {
		return nil, err
	}

	defaultScope, err := zoekt.DefaultGlobalQueryScope(b.repoOptions)
	if err != nil {
		return nil, err
	}

	includePrivate := b.repoOptions.Visibility == query.Private || b.repoOptions.Visibility == query.Any
	globalZoektQuery := zoekt.NewGlobalZoektQuery(zoektQuery, defaultScope, includePrivate)

	zoektArgs := &search.ZoektParameters{
		// TODO(rvantonder): the Query value is set when the global zoekt query is
		// enriched with private repository data in the search job's Run method, and
		// is therefore set to `nil` below.
		// Ideally, The ZoektParameters type should not expose this field for Universe text
		// searches at all, and will be removed once jobs are fully migrated.
		Query:          nil,
		Typ:            typ,
		FileMatchLimit: b.fileMatchLimit,
		Select:         b.selector,
	}

	switch typ {
	case search.SymbolRequest:
		return &symbol.RepoUniverseSymbolSearchJob{
			GlobalZoektQuery: globalZoektQuery,
			ZoektArgs:        zoektArgs,
			RepoOptions:      b.repoOptions,
		}, nil
	case search.TextRequest:
		return &zoekt.ZoektGlobalSearchJob{
			GlobalZoektQuery: globalZoektQuery,
			ZoektArgs:        zoektArgs,
			RepoOptions:      b.repoOptions,
		}, nil
	}
	return nil, errors.Errorf("attempt to create unrecognized zoekt global search with value %v", typ)
}

func (b *jobBuilder) newZoektSearch(typ search.IndexedRequestType) (job.Job, error) {
	zoektQuery, err := zoekt.QueryToZoektQuery(b.query, b.resultTypes, b.features, typ)
	if err != nil {
		return nil, err
	}

	switch typ {
	case search.SymbolRequest:
		return &zoekt.ZoektSymbolSearchJob{
			Query:          zoektQuery,
			FileMatchLimit: b.fileMatchLimit,
			Select:         b.selector,
		}, nil
	case search.TextRequest:
		return &zoekt.ZoektRepoSubsetSearchJob{
			Query:          zoektQuery,
			Typ:            typ,
			FileMatchLimit: b.fileMatchLimit,
			Select:         b.selector,
		}, nil
	}
	return nil, errors.Errorf("attempt to create unrecognized zoekt search with value %v", typ)
}

func jobMode(b query.Basic, resultTypes result.Types, st query.SearchType, onSourcegraphDotCom bool) (repoUniverseSearch, skipRepoSubsetSearch, runZoektOverRepos bool) {
	isGlobalSearch := func() bool {
		if st == query.SearchTypeStructural {
			return false
		}

		return query.ForAll(b.ToParseTree(), func(node query.Node) bool {
			n, ok := node.(query.Parameter)
			if !ok {
				return true
			}
			switch n.Field {
			case query.FieldContext:
				return searchcontexts.IsGlobalSearchContextSpec(n.Value)
			case query.FieldRepo:
				// We allow -repo: in global search.
				return n.Negated
			case query.FieldRepoHasFile:
				return false
			default:
				return true
			}
		})
	}

	hasGlobalSearchResultType := resultTypes.Has(result.TypeFile | result.TypePath | result.TypeSymbol)
	isIndexedSearch := b.Index() != query.No
	noPattern := b.IsEmptyPattern()
	noFile := !b.Exists(query.FieldFile)
	noLang := !b.Exists(query.FieldLang)
	isEmpty := noPattern && noFile && noLang

	repoUniverseSearch = isGlobalSearch() && isIndexedSearch && hasGlobalSearchResultType && !isEmpty
	// skipRepoSubsetSearch is a value that controls whether to
	// run unindexed search in a specific scenario of queries that
	// contain no repo-affecting filters (global mode). When on
	// sourcegraph.com, we resolve only a subset of all indexed
	// repos to search. This control flow implies len(searcherRepos)
	// is always 0, meaning that we should not create jobs to run
	// unindexed searcher.
	skipRepoSubsetSearch = isEmpty || (repoUniverseSearch && onSourcegraphDotCom)

	// runZoektOverRepos controls whether we run Zoekt over a set of
	// resolved repositories. Because Zoekt can run natively run over all
	// repositories (AKA global search), we can sometimes skip searching
	// over resolved repos.
	//
	// The decision to run over a set of repos is as follows:
	// (1) When we don't run global search, run Zoekt over repositories (we have to, otherwise
	// we'd be skipping indexed search entirely).
	// (2) If on Sourcegraph.com, resolve repos unconditionally (we run both global search
	// and search over resolved repos, and return results from either job).
	runZoektOverRepos = !repoUniverseSearch || onSourcegraphDotCom

	return repoUniverseSearch, skipRepoSubsetSearch, runZoektOverRepos
}

func toFeatures(flags featureflag.FlagSet) search.Features {
	if flags == nil {
		flags = featureflag.FlagSet{}
		metricFeatureFlagUnavailable.Inc()
		log15.Warn("search feature flags are not available")
	}

	return search.Features{
		ContentBasedLangFilters: flags.GetBoolOr("search-content-based-lang-detection", false),
	}
}

// toAndJob creates a new job from a basic query whose pattern is an And operator at the root.
func toAndJob(inputs *run.SearchInputs, q query.Basic) (job.Job, error) {
	// Invariant: this function is only reachable from callers that
	// guarantee a root node with one or more queryOperands.
	queryOperands := q.Pattern.(query.Operator).Operands

	// Limit the number of results from each child to avoid a huge amount of memory bloat.
	// With streaming, we should re-evaluate this number.
	//
	// NOTE: It may be possible to page over repos so that each intersection is only over
	// a small set of repos, limiting massive number of results that would need to be
	// kept in memory otherwise.
	maxTryCount := 40000

	operands := make([]job.Job, 0, len(queryOperands))
	for _, queryOperand := range queryOperands {
		operand, err := toPatternExpressionJob(inputs, q.MapPattern(queryOperand))
		if err != nil {
			return nil, err
		}
		operands = append(operands, NewLimitJob(maxTryCount, operand))
	}

	return NewAndJob(operands...), nil
}

// toOrJob creates a new job from a basic query whose pattern is an Or operator at the top level
func toOrJob(inputs *run.SearchInputs, q query.Basic) (job.Job, error) {
	// Invariant: this function is only reachable from callers that
	// guarantee a root node with one or more queryOperands.
	queryOperands := q.Pattern.(query.Operator).Operands

	operands := make([]job.Job, 0, len(queryOperands))
	for _, term := range queryOperands {
		operand, err := toPatternExpressionJob(inputs, q.MapPattern(term))
		if err != nil {
			return nil, err
		}
		operands = append(operands, operand)
	}
	return NewOrJob(operands...), nil
}

func toPatternExpressionJob(inputs *run.SearchInputs, q query.Basic) (job.Job, error) {
	switch term := q.Pattern.(type) {
	case query.Operator:
		if len(term.Operands) == 0 {
			return NewNoopJob(), nil
		}

		switch term.Kind {
		case query.And:
			return toAndJob(inputs, q)
		case query.Or:
			return toOrJob(inputs, q)
		}
	case query.Pattern:
		return NewFlatJob(inputs, query.Flat{Parameters: q.Parameters, Pattern: &term})
	case query.Parameter:
		// evaluatePatternExpression does not process Parameter nodes.
		return NewNoopJob(), nil
	}
	// Unreachable.
	return nil, errors.Errorf("unrecognized type %T in evaluatePatternExpression", q.Pattern)
}

func toFlatJobs(inputs *run.SearchInputs, q query.Basic) (job.Job, error) {
	if q.Pattern == nil {
		return NewFlatJob(inputs, query.Flat{Parameters: q.Parameters, Pattern: nil})
	} else {
		return toPatternExpressionJob(inputs, q)
	}
}

func NewJob(inputs *run.SearchInputs, plan query.Plan) (job.Job, error) {
	children := make([]job.Job, 0, len(plan))
	for _, q := range plan {
		child, err := NewBasicJob(inputs, q)
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}
	return NewAlertJob(inputs, NewOrJob(children...)), nil
}

func NewBasicJob(inputs *run.SearchInputs, q query.Basic) (job.Job, error) {
	var children []job.Job
	addJob := func(j job.Job) {
		children = append(children, j)
	}

	{
		// This block generates jobs that can be built directly from
		// a basic query rather than first being expanded into
		// flat queries.
		types, _ := q.IncludeExcludeValues(query.FieldType)
		resultTypes := computeResultTypes(types, q, inputs.PatternType)
		fileMatchLimit := int32(computeFileMatchLimit(q, inputs.Protocol))
		selector, _ := filter.SelectPathFromString(q.FindValue(query.FieldSelect)) // Invariant: select is validated
		features := toFeatures(inputs.Features)
		repoOptions := toRepoOptions(q, inputs.UserSettings)
		repoUniverseSearch, skipRepoSubsetSearch, runZoektOverRepos := jobMode(q, resultTypes, inputs.PatternType, inputs.OnSourcegraphDotCom)

		builder := &jobBuilder{
			query:          q,
			resultTypes:    resultTypes,
			repoOptions:    repoOptions,
			features:       &features,
			fileMatchLimit: fileMatchLimit,
			selector:       selector,
		}

		if resultTypes.Has(result.TypeFile | result.TypePath) {
			// Create Global Text Search jobs.
			if repoUniverseSearch {
				job, err := builder.newZoektGlobalSearch(search.TextRequest)
				if err != nil {
					return nil, err
				}
				addJob(job)
			}

			if !skipRepoSubsetSearch && runZoektOverRepos {
				job, err := builder.newZoektSearch(search.TextRequest)
				if err != nil {
					return nil, err
				}
				addJob(&repoPagerJob{
					child:            job,
					repoOptions:      repoOptions,
					useIndex:         q.Index(),
					containsRefGlobs: query.ContainsRefGlobs(q.ToParseTree()),
				})
			}
		}

		if resultTypes.Has(result.TypeSymbol) {
			// Create Global Symbol Search jobs.
			if repoUniverseSearch {
				job, err := builder.newZoektGlobalSearch(search.SymbolRequest)
				if err != nil {
					return nil, err
				}
				addJob(job)
			}

			if !skipRepoSubsetSearch && runZoektOverRepos {
				job, err := builder.newZoektSearch(search.SymbolRequest)
				if err != nil {
					return nil, err
				}
				addJob(&repoPagerJob{
					child:            job,
					repoOptions:      repoOptions,
					useIndex:         q.Index(),
					containsRefGlobs: query.ContainsRefGlobs(q.ToParseTree()),
				})
			}
		}

		if resultTypes.Has(result.TypeCommit) || resultTypes.Has(result.TypeDiff) {
			diff := resultTypes.Has(result.TypeDiff)
			addJob(&commit.CommitSearchJob{
				Query:                commit.QueryToGitQuery(q, diff),
				RepoOpts:             repoOptions,
				Diff:                 diff,
				HasTimeFilter:        q.Exists("after") || q.Exists("before"),
				Limit:                int(fileMatchLimit),
				IncludeModifiedFiles: authz.SubRepoEnabled(authz.DefaultSubRepoPermsChecker),
			})
		}

		addJob(&searchrepos.ComputeExcludedReposJob{
			RepoOpts: repoOptions,
		})
	}

	{
		// This block generates jobs using the expansion of a basic query into
		// flat queries
		flatJob, err := toFlatJobs(inputs, q)
		if err != nil {
			return nil, err
		}
		addJob(flatJob)
	}

	basicJob := NewParallelJob(children...)

	{ // Apply selectors
		if v, _ := q.ToParseTree().StringValue(query.FieldSelect); v != "" {
			sp, _ := filter.SelectPathFromString(v) // Invariant: select already validated
			basicJob = NewSelectJob(sp, basicJob)
		}
	}

	{ // Apply subrepo permissions checks
		checker := authz.DefaultSubRepoPermsChecker
		if authz.SubRepoEnabled(checker) {
			basicJob = NewFilterJob(basicJob)
		}
	}

	{ // Apply limit
		maxResults := q.ToParseTree().MaxResults(inputs.DefaultLimit())
		basicJob = NewLimitJob(maxResults, basicJob)
	}

	{ // Apply timeout
		timeout := TimeoutDuration(q)
		basicJob = NewTimeoutJob(timeout, basicJob)
	}

	return basicJob, nil
}

// FromExpandedPlan takes a query plan that has had all predicates expanded,
// and converts it to a job.
func FromExpandedPlan(inputs *run.SearchInputs, plan query.Plan) (job.Job, error) {
	return NewJob(inputs, plan)
}

var metricFeatureFlagUnavailable = promauto.NewCounter(prometheus.CounterOpts{
	Name: "src_search_featureflag_unavailable",
	Help: "temporary counter to check if we have feature flag available in practice.",
})

func computeFileMatchLimit(q query.Basic, p search.Protocol) int {
	if count := q.Count(); count != nil {
		return *count
	}

	if q.IsStructural() {
		return limits.DefaultMaxSearchResults
	}

	switch p {
	case search.Batch:
		return limits.DefaultMaxSearchResults
	case search.Streaming:
		return limits.DefaultMaxSearchResultsStreaming
	}
	panic("unreachable")
}

func TimeoutDuration(b query.Basic) time.Duration {
	d := limits.DefaultTimeout
	maxTimeout := time.Duration(limits.SearchLimits(conf.Get()).MaxTimeoutSeconds) * time.Second
	timeout := b.GetTimeout()
	if timeout != nil {
		d = *timeout
	} else if b.Count() != nil {
		// If `count:` is set but `timeout:` is not explicitly set, use the max timeout
		d = maxTimeout
	}
	if d > maxTimeout {
		d = maxTimeout
	}
	return d
}
