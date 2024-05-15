// Package searcher provides a client for our just in time text searching
// service "searcher".
package searcher

import (
	"context"
	"io"
	"net/url"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	proto "github.com/sourcegraph/sourcegraph/internal/searcher/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func SearchCommits(
	ctx context.Context,
	searcherURLs *endpoint.Map,
	connectionCache *defaults.ConnectionCache,
	repo api.RepoName,
	repoID api.RepoID,
	req *protocol.CommitSearchRequest,
	// branch string,
	// commit api.CommitID,
	// indexed bool,
	// p *search.TextPatternInfo,
	// fetchTimeout time.Duration,
	// features search.Features,
	// contextLines int,
	onMatches func([]protocol.CommitMatch),
) (limitHit bool, err error) {
	// r := (&protocol.CommitSearchRequest{
	// 	Repo:   repo,
	// 	RepoID: repoID,
	// 	Commit: commit,
	// 	Branch: branch,
	// 	PatternInfo: protocol.PatternInfo{
	// 		Query:                        p.Query,
	// 		ExcludePaths:                 p.ExcludePaths,
	// 		IncludePaths:                 p.IncludePaths,
	// 		IncludeLangs:                 p.IncludeLangs,
	// 		ExcludeLangs:                 p.ExcludeLangs,
	// 		CombyRule:                    p.CombyRule,
	// 		Select:                       p.Select.Root(),
	// 		Limit:                        int(p.FileMatchLimit),
	// 		IsStructuralPat:              p.IsStructuralPat,
	// 		IsCaseSensitive:              p.IsCaseSensitive,
	// 		PathPatternsAreCaseSensitive: p.PathPatternsAreCaseSensitive,
	// 		PatternMatchesContent:        p.PatternMatchesContent,
	// 		PatternMatchesPath:           p.PatternMatchesPath,
	// 		Languages:                    p.Languages,
	// 	},
	// 	Indexed:         indexed,
	// 	FetchTimeout:    fetchTimeout,
	// 	NumContextLines: int32(contextLines),
	// }).ToProto()

	// Searcher caches the file contents for repo@commit since it is
	// relatively expensive to fetch from gitserver. So we use consistent
	// hashing to increase cache hits.
	consistentHashKey := string(repo) + "@" + string("RANDOM")

	nodes, err := searcherURLs.Endpoints()
	if err != nil {
		return false, err
	}

	// TODO: We don't care about which searcher this hits as there are no local
	// caches, but there's no "select a random one round robin" right now.
	urls, err := searcherURLs.GetN(consistentHashKey, len(nodes))
	if err != nil {
		return false, err
	}

	trySearch := func(attempt int) (bool, error) {
		parsed, err := url.Parse(urls[attempt%len(urls)])
		if err != nil {
			return false, errors.Wrap(err, "failed to parse URL")
		}

		conn, err := connectionCache.GetConnection(parsed.Host)
		if err != nil {
			return false, err
		}

		client := &automaticRetryClient{proto.NewSearcherServiceClient(conn)}
		resp, err := client.CommitSearch(ctx, req.ToProto())
		if err != nil {
			return false, err
		}

		for {
			msg, err := resp.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return limitHit, nil
				}
				return limitHit, err
			}

			switch m := msg.Message.(type) {
			case *proto.CommitSearchResponse_LimitHit:
				limitHit = limitHit || m.LimitHit
			case *proto.CommitSearchResponse_Match:
				onMatches([]protocol.CommitMatch{protocol.CommitMatchFromProto(m.Match)})
			default:
				return false, errors.Newf("unknown message type %T", m)
			}
		}
	}

	limitHit, err = trySearch(0)
	if err != nil && errcode.IsTemporary(err) {
		// Retry once if we get a temporary error back
		limitHit, err = trySearch(1)
	}
	return limitHit, err
}
