// Package protocol contains structures used by the searcher API.
package protocol

import (
	"fmt"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	proto "github.com/sourcegraph/sourcegraph/internal/searcher/v1"

	"google.golang.org/protobuf/types/known/durationpb"
)

// Request represents a request to searcher
type Request struct {
	// Repo is the name of the repository to search. eg "github.com/gorilla/mux"
	Repo api.RepoName

	// RepoID is the Sourcegraph repository id of the repo to search.
	RepoID api.RepoID

	// URL specifies the repository's Git remote URL (for gitserver). It is optional. See
	// (gitserver.ExecRequest).URL for documentation on what it is used for.
	URL string

	// Commit is which commit to search. It is required to be resolved,
	// not a ref like HEAD or master. eg
	// "599cba5e7b6137d46ddf58fb1765f5d928e69604"
	Commit api.CommitID

	// Branch is used for structural search as an alternative to Commit
	// because Zoekt only takes branch names
	Branch string

	PatternInfo

	// The amount of time to wait for a repo archive to fetch.
	// It is parsed with time.ParseDuration.
	//
	// This timeout should be low when searching across many repos
	// so that unfetched repos don't delay the search, and because we are likely
	// to get results from the repos that have already been fetched.
	//
	// This timeout should be high when searching across a single repo
	// because returning results slowly is better than returning no results at all.
	//
	// This only times out how long we wait for the fetch request;
	// the fetch will still happen in the background so future requests don't have to wait.
	FetchTimeout time.Duration

	// Whether the revision to be searched is indexed or unindexed. This matters for
	// structural search because it will query Zoekt for indexed structural search.
	Indexed bool

	NumContextLines int
}

// PatternInfo describes a search request on a repo. Most of the fields
// are based on PatternInfo used in vscode.
type PatternInfo struct {
	// Pattern is the search query. It is a regular expression if IsRegExp
	// is true, otherwise a fixed string. eg "route variable"
	Pattern string

	// IsNegated if true will invert the matching logic for regexp searches. IsNegated=true is
	// not supported for structural searches.
	IsNegated bool

	// IsRegExp if true will treat the Pattern as a regular expression.
	IsRegExp bool

	// IsStructuralPat if true will treat the pattern as a Comby structural search pattern.
	IsStructuralPat bool

	// IsWordMatch if true will only match the pattern at word boundaries.
	IsWordMatch bool

	// IsCaseSensitive if false will ignore the case of text and pattern
	// when finding matches.
	IsCaseSensitive bool

	// ExcludePattern is a pattern that may not match the returned files' paths.
	// eg '**/node_modules'
	ExcludePattern string

	// IncludePatterns is a list of patterns that must *all* match the returned
	// files' paths.
	// eg '**/node_modules'
	//
	// The patterns are ANDed together; a file's path must match all patterns
	// for it to be kept. That is also why it is a list (unlike the singular
	// ExcludePattern); it is not possible in general to construct a single
	// glob or Go regexp that represents multiple such patterns ANDed together.
	IncludePatterns []string

	// IncludeExcludePatternAreCaseSensitive indicates that ExcludePattern, IncludePattern,
	// and IncludePatterns are case sensitive.
	PathPatternsAreCaseSensitive bool

	// Limit is the cap on the total number of matches returned.
	// A match is either a path match, or a fragment of a line matched by the query.
	Limit int

	// PatternMatchesPath is whether the pattern should be matched against the content
	// of files.
	PatternMatchesContent bool

	// PatternMatchesPath is whether a file whose path matches Pattern (but whose contents don't) should be
	// considered a match.
	PatternMatchesPath bool

	// Languages is the languages passed via the lang filters (e.g., "lang:c")
	Languages []string

	// CombyRule is a rule that constrains matching for structural search.
	// It only applies when IsStructuralPat is true.
	// As a temporary measure, the expression `where "backcompat" == "backcompat"` acts as
	// a flag to activate the old structural search path, which queries zoekt for the
	// file list in the frontend and passes it to searcher.
	CombyRule string

	// Select is the value of the the select field in the query. It is not necessary to
	// use it since selection is done after the query completes, but exposing it can enable
	// optimizations.
	Select string
}

func (p *PatternInfo) String() string {
	args := []string{fmt.Sprintf("%q", p.Pattern)}
	if p.IsRegExp {
		args = append(args, "re")
	}
	if p.IsStructuralPat {
		if p.CombyRule != "" {
			args = append(args, fmt.Sprintf("comby:%s", p.CombyRule))
		} else {
			args = append(args, "comby")
		}
	}
	if p.IsWordMatch {
		args = append(args, "word")
	}
	if p.IsCaseSensitive {
		args = append(args, "case")
	}
	if !p.PatternMatchesContent {
		args = append(args, "nocontent")
	}
	if !p.PatternMatchesPath {
		args = append(args, "nopath")
	}
	if p.Limit > 0 {
		args = append(args, fmt.Sprintf("limit:%d", p.Limit))
	}
	for _, lang := range p.Languages {
		args = append(args, fmt.Sprintf("lang:%s", lang))
	}
	if p.Select != "" {
		args = append(args, fmt.Sprintf("select:%s", p.Select))
	}

	path := "f"
	if p.PathPatternsAreCaseSensitive {
		path = "F"
	}
	if p.ExcludePattern != "" {
		args = append(args, fmt.Sprintf("-%s:%q", path, p.ExcludePattern))
	}
	for _, inc := range p.IncludePatterns {
		args = append(args, fmt.Sprintf("%s:%q", path, inc))
	}

	return fmt.Sprintf("PatternInfo{%s}", strings.Join(args, ","))
}

func (r *Request) ToProto() *proto.SearchRequest {
	return &proto.SearchRequest{
		Repo:      string(r.Repo),
		RepoId:    uint32(r.RepoID),
		CommitOid: string(r.Commit),
		Branch:    r.Branch,
		Indexed:   r.Indexed,
		Url:       r.URL,
		PatternInfo: &proto.PatternInfo{
			Pattern:                      r.PatternInfo.Pattern,
			IsNegated:                    r.PatternInfo.IsNegated,
			IsRegexp:                     r.PatternInfo.IsRegExp,
			IsStructural:                 r.PatternInfo.IsStructuralPat,
			IsWordMatch:                  r.PatternInfo.IsWordMatch,
			IsCaseSensitive:              r.PatternInfo.IsCaseSensitive,
			ExcludePattern:               r.PatternInfo.ExcludePattern,
			IncludePatterns:              r.PatternInfo.IncludePatterns,
			PathPatternsAreCaseSensitive: r.PatternInfo.PathPatternsAreCaseSensitive,
			Limit:                        int64(r.PatternInfo.Limit),
			PatternMatchesContent:        r.PatternInfo.PatternMatchesContent,
			PatternMatchesPath:           r.PatternInfo.PatternMatchesPath,
			CombyRule:                    r.PatternInfo.CombyRule,
			Languages:                    r.PatternInfo.Languages,
			Select:                       r.PatternInfo.Select,
		},
		FetchTimeout:    durationpb.New(r.FetchTimeout),
		NumContextLines: int32(r.NumContextLines),
	}
}

func (r *Request) FromProto(req *proto.SearchRequest) {
	*r = Request{
		Repo:   api.RepoName(req.Repo),
		RepoID: api.RepoID(req.RepoId),
		URL:    req.Url,
		Commit: api.CommitID(req.CommitOid),
		Branch: req.Branch,
		PatternInfo: PatternInfo{
			Pattern:                      req.PatternInfo.Pattern,
			IsNegated:                    req.PatternInfo.IsNegated,
			IsRegExp:                     req.PatternInfo.IsRegexp,
			IsStructuralPat:              req.PatternInfo.IsStructural,
			IsWordMatch:                  req.PatternInfo.IsWordMatch,
			IsCaseSensitive:              req.PatternInfo.IsCaseSensitive,
			ExcludePattern:               req.PatternInfo.ExcludePattern,
			IncludePatterns:              req.PatternInfo.IncludePatterns,
			PathPatternsAreCaseSensitive: req.PatternInfo.PathPatternsAreCaseSensitive,
			Limit:                        int(req.PatternInfo.Limit),
			PatternMatchesContent:        req.PatternInfo.PatternMatchesContent,
			PatternMatchesPath:           req.PatternInfo.PatternMatchesPath,
			Languages:                    req.PatternInfo.Languages,
			CombyRule:                    req.PatternInfo.CombyRule,
			Select:                       req.PatternInfo.Select,
		},
		FetchTimeout:    req.FetchTimeout.AsDuration(),
		Indexed:         req.Indexed,
		NumContextLines: int(req.NumContextLines),
	}
}

// Response represents the response from a Search request.
type Response struct {
	Matches []FileMatch

	// LimitHit is true if Matches may not include all FileMatches because a match limit was hit.
	LimitHit bool

	// DeadlineHit is true if Matches may not include all FileMatches because a deadline was hit.
	DeadlineHit bool
}

// FileMatch is the struct used by vscode to receive search results
type FileMatch struct {
	Path string

	ChunkMatches []ChunkMatch

	// LimitHit is true if LineMatches may not include all LineMatches.
	LimitHit bool
}

func (fm *FileMatch) ToProto() *proto.FileMatch {
	chunkMatches := make([]*proto.ChunkMatch, len(fm.ChunkMatches))
	for i, cm := range fm.ChunkMatches {
		chunkMatches[i] = cm.ToProto()
	}
	return &proto.FileMatch{
		Path:         []byte(fm.Path),
		ChunkMatches: chunkMatches,
		LimitHit:     fm.LimitHit,
	}
}

func (fm *FileMatch) FromProto(pm *proto.FileMatch) {
	chunkMatches := make([]ChunkMatch, len(pm.GetChunkMatches()))
	for i, cm := range pm.GetChunkMatches() {
		chunkMatches[i].FromProto(cm)
	}
	*fm = FileMatch{
		Path:         string(pm.GetPath()), // WARNING: It is not safe to assume that Path is utf-8 encoded.
		ChunkMatches: chunkMatches,
		LimitHit:     pm.GetLimitHit(),
	}
}

func (fm FileMatch) MatchCount() int {
	if len(fm.ChunkMatches) == 0 {
		return 1 // path match is still one match
	}
	count := 0
	for _, cm := range fm.ChunkMatches {
		count += len(cm.Ranges)
	}
	return count
}

func (fm *FileMatch) Limit(limit int) {
	for i := range fm.ChunkMatches {
		l := len(fm.ChunkMatches[i].Ranges)
		if l <= limit {
			limit -= l
			continue
		}

		// invariant: limit < l
		fm.ChunkMatches[i].Ranges = fm.ChunkMatches[i].Ranges[:limit]
		if limit > 0 {
			fm.ChunkMatches = fm.ChunkMatches[:i+1]
		} else {
			fm.ChunkMatches = fm.ChunkMatches[:i]
		}
		fm.LimitHit = true
		return
	}
}

type ChunkMatch struct {
	Content      string // Warning: It is not safe to assume that Content is utf-8 encoded.
	ContentStart Location
	Ranges       []Range
}

func (cm ChunkMatch) MatchedContent() []string {
	res := make([]string, 0, len(cm.Ranges))
	for _, rr := range cm.Ranges {
		res = append(res, cm.Content[rr.Start.Offset-cm.ContentStart.Offset:rr.End.Offset-cm.ContentStart.Offset])
	}
	return res
}

func (cm *ChunkMatch) ToProto() *proto.ChunkMatch {
	ranges := make([]*proto.Range, len(cm.Ranges))
	for i, r := range cm.Ranges {
		ranges[i] = r.ToProto()
	}
	return &proto.ChunkMatch{
		Content:      []byte(cm.Content),
		ContentStart: cm.ContentStart.ToProto(),
		Ranges:       ranges,
	}
}

func (cm *ChunkMatch) FromProto(pm *proto.ChunkMatch) {
	var contentStart Location
	contentStart.FromProto(pm.GetContentStart())

	ranges := make([]Range, len(pm.GetRanges()))
	for i, r := range pm.GetRanges() {
		ranges[i].FromProto(r)
	}

	*cm = ChunkMatch{
		Content:      string(pm.GetContent()), // WARNING: It is not safe to assume that the chunk match content is utf-8 encoded.
		ContentStart: contentStart,
		Ranges:       ranges,
	}
}

type Range struct {
	Start Location
	End   Location
}

func (r *Range) ToProto() *proto.Range {
	return &proto.Range{
		Start: r.Start.ToProto(),
		End:   r.End.ToProto(),
	}
}

func (r *Range) FromProto(pr *proto.Range) {
	r.Start.FromProto(pr.GetStart())
	r.End.FromProto(pr.GetEnd())
}

type Location struct {
	// The byte offset from the beginning of the file.
	Offset int32

	// Line is the count of newlines before the offset in the file.
	// Line is 0-based.
	Line int32

	// Column is the rune offset from the beginning of the last line.
	Column int32
}

func (l *Location) ToProto() *proto.Location {
	return &proto.Location{
		Offset: l.Offset,
		Line:   l.Line,
		Column: l.Column,
	}
}

func (l *Location) FromProto(pl *proto.Location) {
	*l = Location{
		Offset: pl.GetOffset(),
		Line:   pl.GetLine(),
		Column: pl.GetColumn(),
	}
}
