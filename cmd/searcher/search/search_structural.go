package search

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/store"
)

// The Sourcegraph frontend and interface only allow LineMatches (matches on a
// single line) and it isn't possible to specify a line and column range
// spanning multiple lines for highlighting. This function chops up potentially
// multiline matches into multiple LineMatches.
func highlightMultipleLines(r *comby.Match) (matches []protocol.LineMatch) {
	lineSpan := r.Range.End.Line - r.Range.Start.Line + 1
	if lineSpan == 1 {
		return []protocol.LineMatch{
			{
				LineNumber: r.Range.Start.Line - 1,
				OffsetAndLengths: [][2]int{
					{
						r.Range.Start.Column - 1,
						r.Range.End.Column - r.Range.Start.Column,
					},
				},
				Preview: r.Matched,
			},
		}
	}

	contentLines := strings.Split(r.Matched, "\n")
	for i, line := range contentLines {
		var columnStart, columnEnd int
		if i == 0 {
			// First line.
			columnStart = r.Range.Start.Column - 1
			columnEnd = len(line)
		} else if i == (lineSpan - 1) {
			// Last line.
			columnStart = 0
			columnEnd = r.Range.End.Column - 1 // don't include trailing newline
		} else {
			// In between line.
			columnStart = 0
			columnEnd = len(line)
		}

		matches = append(matches, protocol.LineMatch{
			LineNumber: r.Range.Start.Line + i - 1,
			OffsetAndLengths: [][2]int{
				{
					columnStart,
					columnEnd,
				},
			},
			Preview: line,
		})
	}
	return matches
}

func ToFileMatch(combyMatches []comby.FileMatch) (matches []protocol.FileMatch) {
	for _, m := range combyMatches {
		var lineMatches []protocol.LineMatch
		for _, r := range m.Matches {
			lineMatches = append(lineMatches, highlightMultipleLines(&r)...)
		}
		matches = append(matches,
			protocol.FileMatch{
				Path:        m.URI,
				LineMatches: lineMatches,
				MatchCount:  len(m.Matches),
				LimitHit:    false,
			})
	}
	return matches
}

var isValidMatcher = lazyregexp.New(`\.(s|sh|bib|c|cs|css|dart|clj|elm|erl|ex|f|fsx|go|html|hs|java|js|json|jl|kt|tex|lisp|nim|md|ml|org|pas|php|py|re|rb|rs|rst|scala|sql|swift|tex|txt|ts)$`)

func extensionToMatcher(extension string) string {
	if isValidMatcher.MatchString(extension) {
		return extension
	}
	return ".generic"
}

// lookupMatcher looks up a key for specifying -matcher in comby. Comby accepts
// a representative file extension to set a language, so this lookup does not
// need to consider all possible file extensions for a language. There is a generic
// fallback language, so this lookup does not need to be exhaustive either.
func lookupMatcher(language string) string {
	switch strings.ToLower(language) {
	case "assembly", "asm":
		return ".s"
	case "bash":
		return ".sh"
	case "c":
		return ".c"
	case "c#, csharp":
		return ".cs"
	case "css":
		return ".css"
	case "dart":
		return ".dart"
	case "clojure":
		return ".clj"
	case "elm":
		return ".elm"
	case "erlang":
		return ".erl"
	case "elixir":
		return ".ex"
	case "fortran":
		return ".f"
	case "f#", "fsharp":
		return ".fsx"
	case "go":
		return ".go"
	case "html":
		return ".html"
	case "haskell":
		return ".hs"
	case "java":
		return ".java"
	case "javascript":
		return ".js"
	case "json":
		return ".json"
	case "julia":
		return ".jl"
	case "kotlin":
		return ".kt"
	case "laTeX":
		return ".tex"
	case "lisp":
		return ".lisp"
	case "nim":
		return ".nim"
	case "ocaml":
		return ".ml"
	case "pascal":
		return ".pas"
	case "php":
		return ".php"
	case "python":
		return ".py"
	case "reason":
		return ".re"
	case "ruby":
		return ".rb"
	case "rust":
		return ".rs"
	case "scala":
		return ".scala"
	case "sql":
		return ".sql"
	case "swift":
		return ".swift"
	case "text":
		return ".txt"
	case "typescript", "ts":
		return ".ts"
	case "xml":
		return ".xml"
	}
	return ".generic"
}

// filteredStructuralSearch filters the list of files with a regex search before passing the zip to comby
func filteredStructuralSearch(ctx context.Context, zipPath string, zipFile *store.ZipFile, p *protocol.PatternInfo, repo api.RepoName) (matches []protocol.FileMatch, limitHit bool, err error) {
	// Make a copy of the pattern info to modify it to work for a regex search
	rp := *p
	rp.Pattern = comby.StructuralPatToRegexpQuery(p.Pattern, false)
	rp.IsStructuralPat = false
	rp.IsRegExp = true
	rg, err := compile(&rp)
	if err != nil {
		return nil, false, err
	}

	fileMatches, _, err := regexSearch(ctx, rg, zipFile, p.FileMatchLimit, true, false, false)
	if err != nil {
		return nil, false, err
	}

	matchedPaths := make([]string, 0, len(fileMatches))
	for _, fm := range fileMatches {
		matchedPaths = append(matchedPaths, fm.Path)
	}

	var extensionHint string
	if len(matchedPaths) > 0 {
		extensionHint = filepath.Ext(matchedPaths[0])
	}

	return structuralSearch(ctx, zipPath, Subset(matchedPaths), extensionHint, p.Pattern, p.CombyRule, p.Languages, repo)
}

// toMatcher returns the matcher that parameterizes structural search. It
// derives either from an explicit language, or an inferred extension hint.
func toMatcher(languages []string, extensionHint string) string {
	if len(languages) > 0 {
		// Pick the first language, there is no support for applying
		// multiple language matchers in a single search query.
		matcher := lookupMatcher(languages[0])
		requestTotalStructuralSearch.WithLabelValues(matcher).Inc()
		log15.Debug("structural search", "language", languages[0], "matcher", matcher)
		return matcher
	}

	if extensionHint != "" {
		extension := extensionToMatcher(extensionHint)
		requestTotalStructuralSearch.WithLabelValues("inferred:" + extension).Inc()
		return extension
	}
	requestTotalStructuralSearch.WithLabelValues("inferred:.generic").Inc()
	return ".generic"
}

// A variant type that represents whether to search all files in a Zip file
// (type UniversalSet), or just a subset (type Subset).
type filePatterns interface {
	Value()
}

func (UniversalSet) Value() {}
func (Subset) Value()       {}

type UniversalSet struct{}
type Subset []string

var All UniversalSet = struct{}{}

func structuralSearch(ctx context.Context, zipPath string, paths filePatterns, extensionHint, pattern, rule string, languages []string, repo api.RepoName) (matches []protocol.FileMatch, limitHit bool, err error) {
	log15.Info("structural search", "repo", string(repo))

	// Cap the number of forked processes to limit the size of zip contents being mapped to memory. Resolving #7133 could help to lift this restriction.
	numWorkers := 4

	matcher := toMatcher(languages, extensionHint)

	var filePatterns []string
	if v, ok := paths.(Subset); ok {
		filePatterns = []string(v)
	}

	args := comby.Args{
		Input:         comby.ZipPath(zipPath),
		Matcher:       matcher,
		MatchTemplate: pattern,
		MatchOnly:     true,
		FilePatterns:  filePatterns,
		Rule:          rule,
		NumWorkers:    numWorkers,
	}

	combyMatches, err := comby.Matches(ctx, args)
	if err != nil {
		return nil, false, err
	}

	matches = ToFileMatch(combyMatches)
	if err != nil {
		return nil, false, err
	}
	return matches, false, err
}

func structuralSearchWithZoekt(ctx context.Context, p *protocol.Request) (matches []protocol.FileMatch, limitHit, deadlineHit bool, err error) {
	// Since we are returning file content, limit the number of file matches
	// until streaming from Zoekt is implemented
	fileMatchLimit := p.FileMatchLimit
	if fileMatchLimit > maxFileMatchLimit {
		fileMatchLimit = maxFileMatchLimit
	}

	patternInfo :=
		&search.TextPatternInfo{
			Pattern:                      p.Pattern,
			IsNegated:                    p.IsNegated,
			IsRegExp:                     p.IsRegExp,
			IsStructuralPat:              p.IsStructuralPat,
			CombyRule:                    p.CombyRule,
			IsWordMatch:                  p.IsWordMatch,
			IsCaseSensitive:              p.IsCaseSensitive,
			FileMatchLimit:               int32(fileMatchLimit),
			IncludePatterns:              p.IncludePatterns,
			ExcludePattern:               p.ExcludePattern,
			PathPatternsAreCaseSensitive: p.PathPatternsAreCaseSensitive,
			PatternMatchesContent:        p.PatternMatchesContent,
			PatternMatchesPath:           p.PatternMatchesPath,
			Languages:                    p.Languages,
		}

	if p.Branch == "" {
		p.Branch = "HEAD"
	}
	repoBranches := map[string][]string{string(p.Repo): {p.Branch}}
	useFullDeadline := false
	zoektMatches, limitHit, _, err := zoektSearch(ctx, patternInfo, repoBranches, time.Since, p.IndexerEndpoints, useFullDeadline, nil)
	if err != nil {
		return nil, false, false, err
	}

	if len(zoektMatches) == 0 {
		return nil, false, false, nil
	}

	zipFile, err := os.CreateTemp("", "*.zip")
	if err != nil {
		return nil, false, false, err
	}
	defer zipFile.Close()
	defer os.Remove(zipFile.Name())

	if err = writeZip(ctx, zipFile, zoektMatches); err != nil {
		return nil, false, false, err
	}

	var extensionHint string
	if len(zoektMatches) > 0 {
		filename := zoektMatches[0].FileName
		extensionHint = filepath.Ext(filename)
	}

	matches, limitHit, err = structuralSearch(ctx, zipFile.Name(), All, extensionHint, p.Pattern, p.CombyRule, p.Languages, p.Repo)
	return matches, limitHit, false, err
}

var requestTotalStructuralSearch = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "searcher_service_request_total_structural_search",
	Help: "Number of returned structural search requests.",
}, []string{"language"})
