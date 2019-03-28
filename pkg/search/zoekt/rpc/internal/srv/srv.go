package srv

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/search/zoekt"
	"github.com/sourcegraph/sourcegraph/pkg/search/zoekt/query"
)

type SearchArgs struct {
	Q    query.Q
	Opts *zoekt.Options
}

type SearchReply struct {
	Result *zoekt.Result
}

type Searcher struct {
	Searcher zoekt.Searcher
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
