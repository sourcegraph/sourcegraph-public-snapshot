// Package searcher provides a client for our just in time text searching
// service "searcher".
package searcher

import (
	"context"
	"io"
	"net/url"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/search"
	proto "github.com/sourcegraph/sourcegraph/internal/searcher/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	// TODO: Not used.
	MockSearch func(ctx context.Context, repo api.RepoName, repoID api.RepoID, commit api.CommitID, p *search.TextPatternInfo, fetchTimeout time.Duration, onMatches func([]*protocol.FileMatch)) (limitHit bool, err error)
)

// Search searches repo@commit with p.
func Search(
	ctx context.Context,
	searcherURLs *endpoint.Map,
	connectionCache *defaults.ConnectionCache,
	repo api.RepoName,
	repoID api.RepoID,
	branch string,
	commit api.CommitID,
	indexed bool,
	p *search.TextPatternInfo,
	fetchTimeout time.Duration,
	features search.Features,
	contextLines int,
	onMatch func(*proto.FileMatch),
) (limitHit bool, err error) {
	r := (&protocol.Request{
		Repo:   repo,
		RepoID: repoID,
		Commit: commit,
		Branch: branch,
		PatternInfo: protocol.PatternInfo{
			Pattern:                      p.Pattern,
			ExcludePattern:               p.ExcludePattern,
			IncludePatterns:              p.IncludePatterns,
			Languages:                    p.Languages,
			CombyRule:                    p.CombyRule,
			Select:                       p.Select.Root(),
			Limit:                        int(p.FileMatchLimit),
			IsRegExp:                     p.IsRegExp,
			IsStructuralPat:              p.IsStructuralPat,
			IsWordMatch:                  p.IsWordMatch,
			IsCaseSensitive:              p.IsCaseSensitive,
			PathPatternsAreCaseSensitive: p.PathPatternsAreCaseSensitive,
			IsNegated:                    p.IsNegated,
			PatternMatchesContent:        p.PatternMatchesContent,
			PatternMatchesPath:           p.PatternMatchesPath,
		},
		Indexed:         indexed,
		FetchTimeout:    fetchTimeout,
		NumContextLines: int32(contextLines),
	}).ToProto()

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

	trySearch := func(attempt int) (bool, error) {
		parsed, err := url.Parse(urls[attempt%len(urls)])
		if err != nil {
			return false, errors.Wrap(err, "failed to parse URL")
		}

		conn, err := connectionCache.GetConnection(parsed.Host)
		if err != nil {
			return false, err
		}

		client := proto.NewSearcherServiceClient(conn)
		resp, err := client.Search(ctx, r)
		if err != nil {
			return false, err
		}

		for {
			msg, err := resp.Recv()
			if errors.Is(err, io.EOF) {
				return false, nil
			} else if status.Code(err) == codes.Canceled {
				return false, context.Canceled
			} else if err != nil {
				return false, err
			}

			switch v := msg.Message.(type) {
			case *proto.SearchResponse_FileMatch:
				onMatch(v.FileMatch)
			case *proto.SearchResponse_DoneMessage:
				return v.DoneMessage.LimitHit, nil
			default:
				return false, errors.Newf("unknown SearchResponse message %T", v)
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
