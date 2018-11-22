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
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	r, err := s.Searcher.Search(ctx, args.Q, args.Opts)
	if err != nil {
		return err
	}
	reply.Result = r
	return nil
}
