package graphqlbackend

import (
	"context"
	"regexp"
	"regexp/syntax"
	"strings"
	"time"
	"unicode/utf8"

	zoektquery "github.com/google/zoekt/query"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/search"
)

var matchHoleRegexp = lazyregexp.New(splitOnHolesPattern())

func splitOnHolesPattern() string {
	word := `\w+`
	whitespaceAndOptionalWord := `[ ]+(` + word + `)?`
	holeAnything := `:\[` + word + `\]`
	holeAlphanum := `:\[\[` + word + `\]\]`
	holeWithPunctuation := `:\[` + word + `\.\]`
	holeWithNewline := `:\[` + word + `\\n\]`
	holeWhitespace := `:\[` + whitespaceAndOptionalWord + `\]`
	return strings.Join([]string{
		holeAnything,
		holeAlphanum,
		holeWithPunctuation,
		holeWithNewline,
		holeWhitespace,
	}, "|")
}

var matchRegexpPattern = lazyregexp.New(`:\[(\w+)?~(.*)\]`)

type Term interface {
	term()
	String() string
}

type Literal string
type Hole string

func (Literal) term() {}
func (t Literal) String() string {
	return string(t)
}

func (Hole) term() {}
func (t Hole) String() string {
	return string(t)
}

// parseTemplate parses a comby pattern to a list of Terms where a Term is
// either a literal or hole metasyntax.
func parseTemplate(buf []byte) []Term {
	// Track context of whether we are inside an opening hole, e.g., after
	// ':['. Value is greater than 1 when inside.
	var open int
	// Track whether we are balanced inside a regular expression character
	// set like '[a]' inside an open hole, e.g., :[foo~[a]]. Value is greater
	// than 1 when inside.
	var inside int

	var start int
	var r rune
	var token []rune
	var result []Term

	next := func() rune {
		r, start := utf8.DecodeRune(buf)
		buf = buf[start:]
		return r
	}

	appendTerm := func(term Term) {
		result = append(result, term)
		token = []rune{}
	}

	for len(buf) > 0 {
		r = next()
		switch r {
		case ':':
			if open > 0 {
				// ':' inside a hole, likely part of a regexp pattern.
				token = append(token, ':')
				continue
			}
			if len(buf[start:]) > 0 {
				// Look ahead and see if this is the start of a hole.
				if r, _ := utf8.DecodeRune(buf); r == '[' {
					// It is the start of a hole, consume the '['.
					r = next()
					open++
					appendTerm(Literal(token))
					// Persist the literal token scanned up to this point.
					token = append(token, ':', '[')
					continue
				}
				// Something else, push the ':' we saw and continue.
				token = append(token, ':')
				continue
			}
			// Trailing ':'
			token = append(token, ':')
		case '\\':
			if len(buf[start:]) > 0 && open > 0 {
				// Assume this is an escape sequence for a regexp hole.
				r = next()
				token = append(token, '\\', r)
				continue
			}
			token = append(token, '\\')
		case '[':
			if open > 0 {
				// Assume this is a character set inside a regexp hole.
				inside++
				token = append(token, '[')
				continue
			}
			token = append(token, '[')
		case ']':
			if open > 0 && inside > 0 {
				// This ']' closes a regular expression inside a hole.
				inside--
				token = append(token, ']')
				continue
			}
			if open > 0 {
				// This ']' closes a hole.
				open--
				token = append(token, ']')
				appendTerm(Hole(token))
				continue
			}
			token = append(token, r)
		default:
			token = append(token, r)
		}
	}
	if len(token) > 0 {
		result = append(result, Literal(token))
	}
	return result
}

var onMatchWhitespace = lazyregexp.New(`[\s]+`)

// StructuralPatToRegexpQuery converts a comby pattern to an approximate regular
// expression query. It converts whitespace in the pattern so that content
// across newlines can be matched in the index. As an incomplete approximation,
// we use the regex pattern .*? to scan ahead. A shortcircuit option returns a
// regexp query that may find true matches faster, but may miss all possible
// matches.
//
// Example:
// "ParseInt(:[args]) if err != nil" -> "ParseInt(.*)\s+if\s+err!=\s+nil"
func StructuralPatToRegexpQuery(pattern string, shortcircuit bool) string {
	var pieces []string

	terms := parseTemplate([]byte(pattern))
	for _, term := range terms {
		if term.String() == "" {
			continue
		}
		switch v := term.(type) {
		case Literal:
			piece := regexp.QuoteMeta(v.String())
			piece = onMatchWhitespace.ReplaceAllLiteralString(piece, `[\s]+`)
			pieces = append(pieces, piece)
		case Hole:
			if matchRegexpPattern.MatchString(v.String()) {
				extractedRegexp := matchRegexpPattern.ReplaceAllString(v.String(), `$2`)
				pieces = append(pieces, extractedRegexp)
			}
		default:
			panic("Unreachable")
		}
	}

	if len(pieces) == 0 {
		// Match anything.
		return "(.|\\s)*?"
	}

	if shortcircuit {
		// As a shortcircuit, do not match across newlines of structural search pieces.
		return "(" + strings.Join(pieces, ").*?(") + ")"
	}
	return "(" + strings.Join(pieces, ")(.|\\s)*?(") + ")"
}

func HandleFilePathPatterns(query *search.TextPatternInfo) (zoektquery.Q, error) {
	var and []zoektquery.Q

	// Zoekt uses regular expressions for file paths.
	// Unhandled cases: PathPatternsAreCaseSensitive and whitespace in file path patterns.
	for _, p := range query.IncludePatterns {
		q, err := fileRe(p, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, q)
	}
	if query.ExcludePattern != "" {
		q, err := fileRe(query.ExcludePattern, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, &zoektquery.Not{Child: q})
	}

	// For conditionals that happen on a repo we can use type:repo queries. eg
	// (type:repo file:foo) (type:repo file:bar) will match all repos which
	// contain a filename matching "foo" and a filename matchinb "bar".
	//
	// Note: (type:repo file:foo file:bar) will only find repos with a
	// filename containing both "foo" and "bar".
	for _, p := range query.FilePatternsReposMustInclude {
		q, err := fileRe(p, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, &zoektquery.Type{Type: zoektquery.TypeRepo, Child: q})
	}
	for _, p := range query.FilePatternsReposMustExclude {
		q, err := fileRe(p, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, &zoektquery.Not{Child: &zoektquery.Type{Type: zoektquery.TypeRepo, Child: q}})
	}

	return zoektquery.NewAnd(and...), nil
}

func buildQuery(args *search.TextParameters, repos *indexedRepoRevs, filePathPatterns zoektquery.Q, shortcircuit bool) (zoektquery.Q, error) {
	regexString := StructuralPatToRegexpQuery(args.PatternInfo.Pattern, shortcircuit)
	if len(regexString) == 0 {
		return &zoektquery.Const{Value: true}, nil
	}
	re, err := syntax.Parse(regexString, syntax.ClassNL|syntax.PerlX|syntax.UnicodeGroups)
	if err != nil {
		return nil, err
	}
	return zoektquery.NewAnd(
		&zoektquery.RepoBranches{Set: repos.repoBranches},
		filePathPatterns,
		&zoektquery.Regexp{
			Regexp:        re,
			CaseSensitive: true,
			Content:       true,
		},
	), nil
}

// zoektSearchHEADOnlyFiles searches repositories using zoekt, returning only the file paths containing
// content matching the given pattern.
//
// Timeouts are reported through the context, and as a special case errNoResultsInTimeout
// is returned if no results are found in the given timeout (instead of the more common
// case of finding partial or full results in the given timeout).
func zoektSearchHEADOnlyFiles(ctx context.Context, args *search.TextParameters, repos *indexedRepoRevs, since func(t time.Time) time.Duration) (fm []*FileMatchResolver, limitHit bool, reposLimitHit map[string]struct{}, err error) {
	if len(repos.repoRevs) == 0 {
		return nil, false, nil, nil
	}

	k := zoektResultCountFactor(len(repos.repoBranches), args.PatternInfo.FileMatchLimit, args.Mode == search.ZoektGlobalSearch)
	searchOpts := zoektSearchOpts(ctx, k, args.PatternInfo)

	if args.UseFullDeadline {
		// If the user manually specified a timeout, allow zoekt to use all of the remaining timeout.
		deadline, _ := ctx.Deadline()
		searchOpts.MaxWallTime = time.Until(deadline)

		// We don't want our context's deadline to cut off zoekt so that we can get the results
		// found before the deadline.
		//
		// We'll create a new context that gets cancelled if the other context is cancelled for any
		// reason other than the deadline being exceeded. This essentially means the deadline for the new context
		// will be `deadline + time for zoekt to cancel + network latency`.
		var cancel context.CancelFunc
		ctx, cancel = contextWithoutDeadline(ctx)
		defer cancel()
	}

	filePathPatterns, err := HandleFilePathPatterns(args.PatternInfo)
	if err != nil {
		return nil, false, nil, err
	}

	t0 := time.Now()
	q, err := buildQuery(args, repos, filePathPatterns, true)
	if err != nil {
		return nil, false, nil, err
	}
	resp, err := args.Zoekt.Client.Search(ctx, q, &searchOpts)
	if err != nil {
		return nil, false, nil, err
	}
	if since(t0) >= searchOpts.MaxWallTime {
		return nil, false, nil, errNoResultsInTimeout
	}

	// We always return approximate results (limitHit true) unless we run the branch to perform a more complete search.
	limitHit = true
	// If the previous indexed search did not return a substantial number of matching file candidates or count was
	// manually specified, run a more complete and expensive search.
	if resp.FileCount < 10 || args.PatternInfo.FileMatchLimit != defaultMaxSearchResults {
		q, err = buildQuery(args, repos, filePathPatterns, false)
		if err != nil {
			return nil, false, nil, err
		}
		resp, err = args.Zoekt.Client.Search(ctx, q, &searchOpts)
		if err != nil {
			return nil, false, nil, err
		}
		if since(t0) >= searchOpts.MaxWallTime {
			return nil, false, nil, errNoResultsInTimeout
		}
		// This is the only place limitHit can be set false, meaning we covered everything.
		limitHit = resp.FilesSkipped+resp.ShardsSkipped > 0
	}

	if len(resp.Files) == 0 {
		return nil, false, nil, nil
	}

	// Zoekt did not evaluate some files in repositories or ignored some repositories. Record skipped repos.
	reposLimitHit = make(map[string]struct{})
	if limitHit {
		for _, file := range resp.Files {
			if _, ok := reposLimitHit[file.Repository]; !ok {
				reposLimitHit[file.Repository] = struct{}{}
			}
		}
	}

	if fileMatchLimit := int(args.PatternInfo.FileMatchLimit); len(resp.Files) > fileMatchLimit {
		// Trim files based on count.
		fileMatchesInSkippedRepos := resp.Files[fileMatchLimit:]
		resp.Files = resp.Files[:fileMatchLimit]

		if !limitHit {
			// Record skipped repos with trimmed files.
			for _, file := range fileMatchesInSkippedRepos {
				if _, ok := reposLimitHit[file.Repository]; !ok {
					reposLimitHit[file.Repository] = struct{}{}
				}
			}
		}
		limitHit = true
	}

	maxLineMatches := 25 + k
	matches := make([]*FileMatchResolver, len(resp.Files))
	repoResolvers := make(RepositoryResolverCache)
	for i, file := range resp.Files {
		fileLimitHit := false
		if len(file.LineMatches) > maxLineMatches {
			file.LineMatches = file.LineMatches[:maxLineMatches]
			fileLimitHit = true
			limitHit = true
		}
		repoRev := repos.repoRevs[file.Repository]
		if repoResolvers[repoRev.Repo.Name] == nil {
			repoResolvers[repoRev.Repo.Name] = &RepositoryResolver{repo: repoRev.Repo}
		}
		matches[i] = &FileMatchResolver{
			JPath:     file.FileName,
			JLimitHit: fileLimitHit,
			uri:       fileMatchURI(repoRev.Repo.Name, "", file.FileName),
			Repo:      repoResolvers[repoRev.Repo.Name],
			CommitID:  api.CommitID(file.Version),
		}
	}

	return matches, limitHit, reposLimitHit, nil
}
