package backend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
	"sourcegraph.com/sourcegraph/srclib/store/pb"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// NOTE: Only the importer part of the graph service is available over
// gRPC for now, because it is more necessary and has a simpler
// interface to convert to protobufs.

var Graph = &graph_{}

// The "_" differentiates it from package graph, which other files in
// this package import.
type graph_ struct{}

func (s *graph_) Import(ctx context.Context, op *pb.ImportOp) error {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Graph.Import", op.Repo); err != nil {
		return err
	}

	// Update global deps
	dstore := localstore.GlobalDeps
	resolution := &unit.Resolution{
		Resolved: unit.Key{
			Repo:     op.Repo,
			CommitID: op.CommitID,
			Name:     op.Unit.Name,
			Type:     op.Unit.Type,
		},
		Raw: unit.Key{
			Repo:    unit.UnitRepoUnresolved,
			Version: op.Unit.Version,
			Name:    op.Unit.Name,
			Type:    op.Unit.Type,
		},
	}
	if err := dstore.Upsert(ctx, []*unit.Resolution{resolution}); err != nil {
		return err
	}

	// Resolve ref dependency info
	resolveCache := make(map[unit.Key]*unit.Key)
	for _, ref := range op.Data.Refs {
		if ref.DefRepo == unit.UnitRepoUnresolved {
			raw := unit.Key{Type: ref.DefUnitType, Name: ref.DefUnit}
			if _, isResolved := resolveCache[raw]; !isResolved {
				resolved, err := dstore.Resolve(ctx, &raw)
				if err != nil {
					return err
				}
				if len(resolved) >= 1 {
					r := resolved[0]
					resolveCache[raw] = &r
				} else {
					resolveCache[raw] = nil
				}
			}
			resolved := resolveCache[raw]
			if resolved != nil {
				ref.DefRepo, ref.DefUnit, ref.DefUnitType = resolved.Repo, resolved.Name, resolved.Type
			}
		}
	}

	gstore := localstore.Graph
	if _, err := pb.Server(gstore).Import(ctx, op); err != nil {
		return err
	}

	return nil
}

func (s *graph_) Index(ctx context.Context, op *pb.IndexOp) error {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Graph.Index", op.Repo); err != nil {
		return err
	}
	_, err := pb.Server(localstore.Graph).Index(ctx, op)
	return err
}

func (s *graph_) CreateVersion(ctx context.Context, op *pb.CreateVersionOp) error {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Graph.CreateVersion", op.Repo); err != nil {
		return err
	}
	_, err := pb.Server(localstore.Graph).CreateVersion(ctx, op)
	return err
}
