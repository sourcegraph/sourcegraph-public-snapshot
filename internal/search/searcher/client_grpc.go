// Package searcher provides a client for our just in time text searching
// service "searcher".
package searcher

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/proto"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Search searches repo@commit with p.
func SearchGRPC(
	ctx context.Context,
	searcherURLs *endpoint.Map,
	repo api.RepoName,
	repoID api.RepoID,
	branch string,
	commit api.CommitID,
	indexed bool,
	p *search.TextPatternInfo,
	fetchTimeout time.Duration,
	features search.Features,
	onMatch func(*proto.FileMatch),
) (limitHit bool, err error) {
	r := proto.SearchRequest{
		Repo:      string(repo),
		RepoId:    int32(repoID),
		CommitOid: string(commit),
		Branch:    branch,
		Indexed:   indexed,
		PatternInfo: &proto.PatternInfo{
			Pattern:   p.Pattern,
			IsNegated: p.IsNegated, IsRegexp: p.IsRegExp,
			IsStructural:                 p.IsStructuralPat,
			IsWordMatch:                  p.IsWordMatch,
			IsCaseSensitive:              p.IsCaseSensitive,
			ExcludePattern:               p.ExcludePattern,
			IncludePatterns:              p.IncludePatterns,
			PathPatternsAreCaseSensitive: p.PathPatternsAreCaseSensitive,
			Limit:                        p.FileMatchLimit,
			PatternMatchesContent:        p.PatternMatchesContent,
			PatternMatchesPath:           p.PatternMatchesPath,
			CombyRule:                    p.CombyRule,
			Languages:                    p.Languages,
			Select:                       p.Select.Root(),
		},
		FetchTimeout: fetchTimeout.String(),
		FeatHybrid:   features.HybridSearch,
	}

	// Searcher caches the file contents for repo@commit since it is
	// relatively expensive to fetch from gitserver. So we use consistent
	// hashing to increase cache hits.
	consistentHashKey := string(repo) + "@" + string(commit)

	nodes, err := searcherURLs.Endpoints()
	if err != nil {
		return false, err
	}

	urls, err := searcherURLs.GetN(consistentHashKey, len(nodes))
	if err != nil {
		return false, err
	}

	for attempt := 0; attempt < 2; attempt++ {
		parsed, err := url.Parse(urls[attempt%len(urls)])
		if err != nil {
			return false, err
		}
		clientConn, err := grpc.Dial(parsed.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return false, err
		}
		client := proto.NewSearcherClient(clientConn)

		resp, err := client.Search(ctx, &r)
		if err != nil {
			return false, err
		}

		// TODO proper retries
		for {
			msg, err := resp.Recv()
			if errors.Is(err, io.EOF) {
				return limitHit, nil
			} else if err != nil {
				return false, err
			}

			onMatch(msg.FileMatch)
		}
	}

	return false, err
}
