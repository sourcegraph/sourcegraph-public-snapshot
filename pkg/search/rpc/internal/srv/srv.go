package srv

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/search"
	"github.com/sourcegraph/sourcegraph/pkg/search/query"
)

type SearchArgs struct {
	Q    query.Q
	Opts *search.Options
}

type SearchReply struct {
	Result *search.Result
}

type Searcher struct {
	Searcher search.Searcher
}

func (s *Searcher) Search(ctx context.Context, args *SearchArgs, reply *SearchReply) error {
	timeout := 10 * time.Second
	if args.Opts.MaxWallTime > 0 {
		timeout = args.Opts.MaxWallTime + time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	r, err := s.Searcher.Search(ctx, args.Q, args.Opts)
	if err != nil {
		return err
	}
	reply.Result = r
	return nil
}
