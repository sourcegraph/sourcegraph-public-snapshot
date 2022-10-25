package streaming

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/inventory"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

// SearchFilters computes the filters to show a user based on results.
//
// Note: it currently live in graphqlbackend. However, once we have a non
// resolver based SearchResult type it can be extracted. It lives in its own
// file to make that more obvious. We already have the filter type extracted
// (Filter).
type SearchFilters struct {
	filters filters
}

// commonFileFilters are common filters used. It is used by SearchFilters to
// propose them if they match shown results.
var commonFileFilters = []struct {
	label       string
	regexp      *lazyregexp.Regexp
	regexFilter string
	globFilter  string
}{
	{
		label:       "Exclude Go tests",
		regexp:      lazyregexp.New(`_test\.go$`),
		regexFilter: `-file:_test\.go$`,
		globFilter:  `-file:**_test.go`,
	},
	{
		label:       "Exclude Go vendor",
		regexp:      lazyregexp.New(`(^|/)vendor/`),
		regexFilter: `-file:(^|/)vendor/`,
		globFilter:  `-file:vendor/** -file:**/vendor/**`,
	},
	{
		label:       "Exclude node_modules",
		regexp:      lazyregexp.New(`(^|/)node_modules/`),
		regexFilter: `-file:(^|/)node_modules/`,
		globFilter:  `-file:node_modules/** -file:**/node_modules/**`,
	},
	{
		label:       "Exclude minified JavaScript",
		regexp:      lazyregexp.New(`\.min\.js$`),
		regexFilter: `-file:\.min\.js$`,
		globFilter:  `-file:**.min.js`,
	},
	{
		label:       "Exclude JavaScript maps",
		regexp:      lazyregexp.New(`\.js\.map$`),
		regexFilter: `-file:\.js\.map$`,
		globFilter:  `-file:**.js.map`,
	},
}

// Update internal state for the results in event.
func (s *SearchFilters) Update(event SearchEvent) {
	// Initialize state on first call.
	if s.filters == nil {
		s.filters = make(filters)
	}

	addRepoFilter := func(repoName api.RepoName, repoID api.RepoID, rev string, lineMatchCount int32) {
		filter := fmt.Sprintf(`repo:^%s$`, regexp.QuoteMeta(string(repoName)))
		if rev != "" {
			// We don't need to quote rev. The only special characters we interpret
			// are @ and :, both of which are disallowed in git refs
			filter = filter + fmt.Sprintf(`@%s`, rev)
		}
		limitHit := event.Stats.Status.Get(repoID)&search.RepoStatusLimitHit != 0
		s.filters.Add(filter, string(repoName), lineMatchCount, limitHit, "repo")
	}

	addFileFilter := func(fileMatchPath string, lineMatchCount int32, limitHit bool) {
		for _, ff := range commonFileFilters {
			// use regexp to match file paths unconditionally, whether globbing is enabled or not,
			// since we have no native library call to match `**` for globs.
			if ff.regexp.MatchString(fileMatchPath) {
				s.filters.Add(ff.regexFilter, ff.label, lineMatchCount, limitHit, "file")
			}
		}
	}

	addLangFilter := func(fileMatchPath string, lineMatchCount int32, limitHit bool) {
		if ext := path.Ext(fileMatchPath); ext != "" {
			rawLanguage, _ := inventory.GetLanguageByFilename(fileMatchPath)
			language := strings.ToLower(rawLanguage)
			if language != "" {
				if strings.Contains(language, " ") {
					language = strconv.Quote(language)
				}
				value := fmt.Sprintf(`lang:%s`, language)
				s.filters.Add(value, rawLanguage, lineMatchCount, limitHit, "lang")
			}
		}
	}

	if event.Stats.ExcludedForks > 0 {
		s.filters.Add("fork:yes", "Include forked repos", int32(event.Stats.ExcludedForks), event.Stats.IsLimitHit, "utility")
		s.filters.MarkImportant("fork:yes")
	}
	if event.Stats.ExcludedArchived > 0 {
		s.filters.Add("archived:yes", "Include archived repos", int32(event.Stats.ExcludedArchived), event.Stats.IsLimitHit, "utility")
		s.filters.MarkImportant("archived:yes")
	}

	for _, match := range event.Results {
		switch v := match.(type) {
		case *result.FileMatch:
			rev := ""
			if v.InputRev != nil {
				rev = *v.InputRev
			}
			lines := int32(v.ResultCount())
			addRepoFilter(v.Repo.Name, v.Repo.ID, rev, lines)
			addLangFilter(v.Path, lines, v.LimitHit)
			addFileFilter(v.Path, lines, v.LimitHit)
		case *result.RepoMatch:
			// It should be fine to leave this blank since revision specifiers
			// can only be used with the 'repo:' scope. In that case,
			// we shouldn't be getting any repositoy name matches back.
			addRepoFilter(v.Name, v.ID, "", 1)
		case *result.CommitMatch:
			// We leave "rev" empty, instead of using "CommitMatch.Commit.ID". This way we
			// get 1 filter per repo instead of 1 filter per sha in the side-bar.
			addRepoFilter(v.Repo.Name, v.Repo.ID, "", int32(v.ResultCount()))
		}
	}
}

// Compute returns an ordered slice of Filters to present to the user based on
// events passed to Next.
func (s *SearchFilters) Compute() []*Filter {
	return s.filters.Compute(computeOpts{
		MaxRepos: 40,
		MaxOther: 40,
	})
}
