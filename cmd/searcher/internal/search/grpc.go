package search

import (
	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/api"
	searcherapi "github.com/sourcegraph/sourcegraph/internal/searcher/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GRPCServer struct {
	searcherapi.UnimplementedSearcherServer
	Store *Store
}

var _ searcherapi.SearcherServer = (*GRPCServer)(nil)

func (s *GRPCServer) SearchStructuralUnindexed(req *searcherapi.SearchStructuralUnindexedRequest, stream searcherapi.Searcher_SearchStructuralUnindexedServer) error {
	ctx := stream.Context()

	getZf := func() (string, *zipFile, error) {
		path, err := s.Store.PrepareZip(ctx, api.RepoName(req.Repo), api.CommitID(req.Commit))
		if err != nil {
			return "", nil, err
		}
		zf, err := s.Store.zipCache.Get(path)
		return path, zf, err
	}

	zipPath, zf, err := getZipFileWithRetry(getZf)
	if err != nil {
		return errors.Wrap(err, "failed to get archive")
	}
	defer zf.Close()

	onMatches := func(match protocol.FileMatch) { stream.Send(matchToEvent(match)) }

	ctx, cancel, sender := newLimitedStream(ctx, int(req.Limit), onMatches)
	defer cancel()

	patternInfo := &protocol.PatternInfo{
		Pattern:                      req.PatternInfo.Pattern,
		IsNegated:                    false, // not supported by structural search
		IsRegExp:                     false, // always a structural pattern
		IsStructuralPat:              true,  // always a structural pattern
		IsWordMatch:                  false, // not supported by structural search
		IsCaseSensitive:              true,  // always true for structural search
		ExcludePattern:               req.PatternInfo.PathPatterns.Exclude,
		IncludePatterns:              req.PatternInfo.PathPatterns.Include,
		PathPatternsAreRegExps:       req.PatternInfo.PathPatterns.IsRegexp,
		PathPatternsAreCaseSensitive: req.PatternInfo.PathPatterns.IsCaseSensitive,
		Limit:                        int(req.Limit),
		PatternMatchesContent:        req.PatternInfo.PatternMatchesContent,
		PatternMatchesPath:           req.PatternInfo.PatternMatchesPath,
		Languages:                    req.PatternInfo.Languages,
		CombyRule:                    req.PatternInfo.CombyRule,
		Select:                       "", // not used for structural search
	}

	err = filteredStructuralSearch(ctx, zipPath, zf, patternInfo, api.RepoName(req.Repo), sender)
	if err != nil {
		return err
	}

	stream.Send(&searcherapi.SearchStructuralUnindexedResponse{
		Event: &searcherapi.SearchStructuralUnindexedResponse_Done{
			Done: &searcherapi.EventDone{LimitHit: sender.LimitHit()},
		},
	})

	return nil
}

func matchToEvent(match protocol.FileMatch) *searcherapi.SearchStructuralUnindexedResponse {
	return &searcherapi.SearchStructuralUnindexedResponse{
		Event: &searcherapi.SearchStructuralUnindexedResponse_Matches{
			Matches: &searcherapi.EventMatches{
				Matches: []*searcherapi.FileMatch{{
					Path:        match.Path,
					LineMatches: convertLineMatches(match.LineMatches),
					MatchCount:  int64(match.MatchCount),
					LimitHit:    match.LimitHit,
				}},
			},
		},
	}
}

func convertLineMatches(in []protocol.LineMatch) []*searcherapi.LineMatch {
	out := make([]*searcherapi.LineMatch, len(in))
	for i, lm := range in {
		out[i] = &searcherapi.LineMatch{
			Preview:          lm.Preview,
			LineNumber:       int64(lm.LineNumber),
			OffsetAndLengths: convertOffsetLengths(lm.OffsetAndLengths),
		}
	}
	return out
}

func convertOffsetLengths(in [][2]int) []*searcherapi.OffsetLength {
	out := make([]*searcherapi.OffsetLength, len(in))
	for i, ol := range in {
		out[i] = &searcherapi.OffsetLength{Offset: int64(ol[0]), Length: int64(ol[1])}
	}
	return out
}
