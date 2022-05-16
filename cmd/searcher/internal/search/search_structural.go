package search

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/RoaringBitmap/roaring"
	zoektquery "github.com/google/zoekt/query"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
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

func toFileMatch(combyMatch *comby.FileMatch) protocol.FileMatch {
	var lineMatches []protocol.LineMatch
	for _, r := range combyMatch.Matches {
		lineMatches = append(lineMatches, highlightMultipleLines(&r)...)
	}
	return protocol.FileMatch{
		Path:        combyMatch.URI,
		LineMatches: lineMatches,
		MatchCount:  len(combyMatch.Matches),
		LimitHit:    false,
	}
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
func filteredStructuralSearch(ctx context.Context, zipPath string, zf *zipFile, p *protocol.PatternInfo, repo api.RepoName, sender matchSender) error {
	// Make a copy of the pattern info to modify it to work for a regex search
	rp := *p
	rp.Pattern = comby.StructuralPatToRegexpQuery(p.Pattern, false)
	rp.IsStructuralPat = false
	rp.IsRegExp = true
	rg, err := compile(&rp)
	if err != nil {
		return err
	}

	fileMatches, _, err := regexSearchBatch(ctx, rg, zf, p.Limit, true, false, false)
	if err != nil {
		return err
	}

	matchedPaths := make([]string, 0, len(fileMatches))
	for _, fm := range fileMatches {
		matchedPaths = append(matchedPaths, fm.Path)
	}

	var extensionHint string
	if len(matchedPaths) > 0 {
		extensionHint = filepath.Ext(matchedPaths[0])
	}

	return structuralSearch(ctx, zipPath, subset(matchedPaths), extensionHint, p.Pattern, p.CombyRule, p.Languages, repo, sender)
}

// toMatcher returns the matcher that parameterizes structural search. It
// derives either from an explicit language, or an inferred extension hint.
func toMatcher(languages []string, extensionHint string) string {
	if len(languages) > 0 {
		// Pick the first language, there is no support for applying
		// multiple language matchers in a single search query.
		matcher := lookupMatcher(languages[0])
		requestTotalStructuralSearch.WithLabelValues(matcher).Inc()
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
// (type universalSet), or just a subset (type Subset).
type filePatterns interface {
	Value()
}

func (universalSet) Value() {}
func (subset) Value()       {}

type universalSet struct{}
type subset []string

var all universalSet = struct{}{}

func structuralSearch(ctx context.Context, zipPath string, paths filePatterns, extensionHint, pattern, rule string, languages []string, repo api.RepoName, sender matchSender) (err error) {
	span, ctx := ot.StartSpanFromContext(ctx, "StructuralSearch")
	span.SetTag("repo", repo)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	// Cap the number of forked processes to limit the size of zip contents being mapped to memory. Resolving #7133 could help to lift this restriction.
	numWorkers := 4

	matcher := toMatcher(languages, extensionHint)

	var filePatterns []string
	if v, ok := paths.(subset); ok {
		filePatterns = []string(v)
	}
	span.LogFields(otlog.Int("paths", len(filePatterns)))

	args := comby.Args{
		Input:         comby.ZipPath(zipPath),
		Matcher:       matcher,
		MatchTemplate: pattern,
		ResultKind:    comby.MatchOnly,
		FilePatterns:  filePatterns,
		Rule:          rule,
		NumWorkers:    numWorkers,
	}

	combyMatches, err := comby.Matches(ctx, args)
	if err != nil {
		return err
	}

	for _, combyMatch := range combyMatches {
		if ctx.Err() != nil {
			return nil
		}
		sender.Send(toFileMatch(combyMatch))
	}
	return nil
}

func structuralSearchWithZoekt(ctx context.Context, p *protocol.Request, sender matchSender) (err error) {
	patternInfo := &search.TextPatternInfo{
		Pattern:                      p.Pattern,
		IsNegated:                    p.IsNegated,
		IsRegExp:                     p.IsRegExp,
		IsStructuralPat:              p.IsStructuralPat,
		CombyRule:                    p.CombyRule,
		IsWordMatch:                  p.IsWordMatch,
		IsCaseSensitive:              p.IsCaseSensitive,
		FileMatchLimit:               int32(p.Limit),
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
	branchRepos := []zoektquery.BranchRepos{{Branch: p.Branch, Repos: roaring.BitmapOf(uint32(p.RepoID))}}
	zoektMatches, _, _, err := zoektSearch(ctx, patternInfo, branchRepos, time.Since, p.IndexerEndpoints, nil)
	if err != nil {
		return err
	}

	if len(zoektMatches) == 0 {
		return nil
	}

	zipFile, err := os.CreateTemp("", "*.zip")
	if err != nil {
		return err
	}
	defer zipFile.Close()
	defer os.Remove(zipFile.Name())

	if err = writeZip(ctx, zipFile, zoektMatches); err != nil {
		return err
	}

	var extensionHint string
	if len(zoektMatches) > 0 {
		filename := zoektMatches[0].FileName
		extensionHint = filepath.Ext(filename)
	}

	return structuralSearch(ctx, zipFile.Name(), all, extensionHint, p.Pattern, p.CombyRule, p.Languages, p.Repo, sender)
}

var requestTotalStructuralSearch = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "searcher_service_request_total_structural_search",
	Help: "Number of returned structural search requests.",
}, []string{"language"})
