package streaming

import (
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/regexp"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/inventory"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

// SearchFilters computes the filters to show a user based on results.
//
// Note: it currently lives in graphqlbackend. However, once we have a non
// resolver based SearchResult type it can be extracted. It lives in its own
// file to make that more obvious. We already have the filter type extracted
// (Filter).
type SearchFilters struct {
	filters filters

	// Dirty is true if SearchFilters has changed since the last call to Compute()
	Dirty bool
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

const (
	// After/Before Filters
	AFTER  = "after"
	BEFORE = "before"

	// After/Before Values
	YESTERDAY     = "yesterday"
	ONE_WEEK_AGO  = "1 week ago"
	ONE_MONTH_AGO = "1 month ago"
)

type dateFilterInfo struct {
	Timeframe string
	Value     string
	Label     string
}

func determineTimeframe(date time.Time) dateFilterInfo {
	now := time.Now()

	switch {
	case date.After(now.Add(-25 * time.Hour)):
		return dateFilterInfo{
			Timeframe: AFTER,
			Value:     YESTERDAY,
			Label:     "Last 24 hours",
		}
	case date.After(now.Add(-8 * 24 * time.Hour)):
		return dateFilterInfo{
			Timeframe: BEFORE,
			Value:     ONE_WEEK_AGO,
			Label:     "Last week",
		}
	case date.After(now.Add(-31 * 24 * time.Hour)):
		return dateFilterInfo{
			Timeframe: BEFORE,
			Value:     ONE_MONTH_AGO,
			Label:     "Last month",
		}

	default:
		return dateFilterInfo{
			Timeframe: BEFORE,
			Value:     "2 months ago",
			Label:     "Older than 2 months",
		}
	}
}

// Update internal state for the results in event.
func (s *SearchFilters) Update(event SearchEvent) {
	// Initialize state on first call.
	if s.filters == nil {
		s.filters = make(filters)
	}

	addRepoFilter := func(repoName api.RepoName, rev string, lineMatchCount int32) {
		filter := fmt.Sprintf(`repo:^%s$`, regexp.QuoteMeta(string(repoName)))
		if rev != "" {
			// We don't need to quote rev. The only special characters we interpret
			// are @ and :, both of which are disallowed in git refs
			filter = filter + fmt.Sprintf(`@%s`, rev)
		}
		s.filters.Add(filter, string(repoName), lineMatchCount, "repo")
	}

	addFileFilter := func(fileMatchPath string, lineMatchCount int32) {
		for _, ff := range commonFileFilters {
			// use regexp to match file paths unconditionally, whether globbing is enabled or not,
			// since we have no native library call to match `**` for globs.
			if ff.regexp.MatchString(fileMatchPath) {
				s.filters.Add(ff.regexFilter, ff.label, lineMatchCount, "file")
			}
		}
	}

	addLangFilter := func(fileMatchPath string, lineMatchCount int32) {
		if ext := path.Ext(fileMatchPath); ext != "" {
			rawLanguage, _ := inventory.GetLanguageByFilename(fileMatchPath)
			language := strings.ToLower(rawLanguage)
			if language != "" {
				if strings.Contains(language, " ") {
					language = strconv.Quote(language)
				}
				value := fmt.Sprintf(`lang:%s`, language)
				s.filters.Add(value, rawLanguage, lineMatchCount, "lang")
			}
		}
	}

	addSymbolFilter := func(symbols []*result.SymbolMatch) {
		for _, sym := range symbols {
			selectKind := result.ToSelectKind[strings.ToLower(sym.Symbol.Kind)]
			filter := fmt.Sprintf(`select:symbol.%s`, selectKind)
			s.filters.Add(filter, selectKind, 1, "symbol type")
		}
	}

	addCommitAuthorFilter := func(commit gitdomain.Commit) {
		filter := fmt.Sprintf(`author:%s`, regexp.QuoteMeta(commit.Author.Email))
		s.filters.Add(filter, commit.Author.Name, 1, "author")
	}

	addCommitDateFilter := func(commit gitdomain.Commit) {
		var cd time.Time
		if commit.Committer != nil {
			cd = commit.Committer.Date
		} else {
			cd = commit.Author.Date
		}

		df := determineTimeframe(cd)
		filter := fmt.Sprintf("%s:%s", df.Timeframe, df.Value)
		s.filters.Add(filter, df.Label, 1, "commit date")
	}

	if event.Stats.ExcludedForks > 0 {
		s.filters.Add("fork:yes", "Include forked repos", int32(event.Stats.ExcludedForks), "utility")
		s.filters.MarkImportant("fork:yes")
		s.Dirty = true
	}
	if event.Stats.ExcludedArchived > 0 {
		s.filters.Add("archived:yes", "Include archived repos", int32(event.Stats.ExcludedArchived), "utility")
		s.filters.MarkImportant("archived:yes")
		s.Dirty = true
	}

	for _, match := range event.Results {
		switch v := match.(type) {
		case *result.FileMatch:
			rev := ""
			if v.InputRev != nil {
				rev = *v.InputRev
			}
			lines := int32(v.ResultCount())

			addRepoFilter(v.Repo.Name, rev, lines)
			addLangFilter(v.Path, lines)
			addFileFilter(v.Path, lines)
			addSymbolFilter(v.Symbols)
			s.Dirty = true
		case *result.RepoMatch:
			// It should be fine to leave this blank since revision specifiers
			// can only be used with the 'repo:' scope. In that case,
			// we shouldn't be getting any repositoy name matches back.
			addRepoFilter(v.Name, "", 1)
			s.Dirty = true
		case *result.CommitMatch:
			// We leave "rev" empty, instead of using "CommitMatch.Commit.ID". This way we
			// get 1 filter per repo instead of 1 filter per sha in the side-bar.
			addRepoFilter(v.Repo.Name, "", int32(v.ResultCount()))
			addCommitAuthorFilter(v.Commit)
			addCommitDateFilter(v.Commit)
			s.Dirty = true

			// =========== TODO: Jason Repo Metadata filters ============
			// file paths are in v.ModifiedFiles which is a []string
		}
	}
}

// Compute returns an ordered slice of Filters to present to the user based on
// events passed to Next.
func (s *SearchFilters) Compute() []*Filter {
	s.Dirty = false
	return s.filters.Compute(computeOpts{
		MaxRepos: 40,
		MaxOther: 40,
	})
}
