package backend

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
	"sourcegraph.com/sqs/pbtypes"
)

var Async sourcegraph.AsyncServer = &async{}

type async struct{}

func (s *async) RefreshIndexes(ctx context.Context, op *sourcegraph.AsyncRefreshIndexesOp) (*pbtypes.Void, error) {
	// TODO(keegancsmith) perm check on ctx, since we want to use one not
	// tied to the gRPC request lifetime
	go func() {
		err := s.refreshIndexes(ctx, op)
		if err != nil {
			log15.Debug("Async.RefreshIndexes failed", "repo", op.Repo, "source", op.Source, "err", err)
		} else {
			log15.Debug("Async.RefreshIndexes success", "repo", op.Repo, "source", op.Source)
		}
	}()
	return &pbtypes.Void{}, nil
}

func (s *async) refreshIndexes(ctx context.Context, op *sourcegraph.AsyncRefreshIndexesOp) error {
	_, err := svc.Defs(ctx).RefreshIndex(ctx, &sourcegraph.DefsRefreshIndexOp{
		Repo:                op.Repo,
		RefreshRefLocations: true,
		Force:               op.Force,
	})
	if err != nil {
		return grpc.Errorf(grpc.Code(err), "Def.RefreshIndex failed on repo %s from source %s: %s", op.Repo, op.Source, err)
	}

	_, err = svc.Search(ctx).RefreshIndex(ctx, &sourcegraph.SearchRefreshIndexOp{
		Repos:         []int32{op.Repo},
		RefreshCounts: true,
		RefreshSearch: true,
	})
	if err != nil {
		return grpc.Errorf(grpc.Code(err), "Search.RefreshIndex failed on repo %s from source %s: %s", op.Repo, op.Source, err)
	}

	return nil
}
