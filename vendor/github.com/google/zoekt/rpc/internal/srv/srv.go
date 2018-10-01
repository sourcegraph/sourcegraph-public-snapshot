package srv

import (
	"context"
	"log"
	"time"

	"github.com/google/zoekt"
	"github.com/google/zoekt/query"
)

type SearchArgs struct {
	Q    query.Q
	Opts *zoekt.SearchOptions
}

type SearchReply struct {
	Result *zoekt.SearchResult
}

type ListArgs struct {
	Q query.Q
}

type ListReply struct {
	List *zoekt.RepoList
}

type Searcher struct {
	Searcher zoekt.Searcher
}

func (s *Searcher) Search(ctx context.Context, args *SearchArgs, reply *SearchReply) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	log.Printf("got rpc query %q", args.Q)
	r, err := s.Searcher.Search(ctx, args.Q, args.Opts)
	if err != nil {
		return err
	}
	reply.Result = r
	return nil
}

func (s *Searcher) List(ctx context.Context, args *ListArgs, reply *ListReply) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	r, err := s.Searcher.List(ctx, args.Q)
	if err != nil {
		return err
	}
	reply.List = r
	return nil
}
