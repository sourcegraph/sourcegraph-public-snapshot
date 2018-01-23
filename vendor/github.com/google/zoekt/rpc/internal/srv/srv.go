package srv

import (
	"context"
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
	zoekt.Searcher
}

func (s *Searcher) Search(args *SearchArgs, reply *SearchReply) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r, err := s.Searcher.Search(ctx, args.Q, args.Opts)
	if err != nil {
		return err
	}
	reply.Result = r
	return nil
}

func (s *Searcher) List(args *ListArgs, reply *ListReply) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r, err := s.Searcher.List(ctx, args.Q)
	if err != nil {
		return err
	}
	reply.List = r
	return nil
}
