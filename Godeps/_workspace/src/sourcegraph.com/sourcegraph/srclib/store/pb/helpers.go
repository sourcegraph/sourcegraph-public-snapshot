package pb

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/store"
	"sourcegraph.com/sourcegraph/srclib/unit"
	"sourcegraph.com/sqs/pbtypes"
)

// A MultiRepoImporterIndexer implements both store.MultiRepoImporter
// and store.MultiRepoIndexer.
type MultiRepoImporterIndexer interface {
	store.MultiRepoImporter
	store.MultiRepoIndexer
}

// Client wraps a gRPC MultiRepoImporterClient and makes it implement
// store.MultiRepoImporter. Clients should not be long-lived (because
// they must hold the ctx).
func Client(ctx context.Context, c MultiRepoImporterClient) MultiRepoImporterIndexer {
	return &client{ctx, c}
}

type client struct {
	ctx context.Context
	u   MultiRepoImporterClient
}

func (c *client) Import(repo, commitID string, u *unit.SourceUnit, data graph.Output) error {
	rsUnit, err := unit.NewRepoSourceUnit(u)
	if err != nil {
		return err
	}
	_, err = c.u.Import(c.ctx, &ImportOp{
		Repo:     repo,
		CommitID: commitID,
		Unit:     rsUnit,
		Data:     &data,
	})
	return err
}

func (c *client) Index(repo, commitID string) error {
	_, err := c.u.Index(c.ctx, &IndexOp{Repo: repo, CommitID: commitID})
	return err
}

// Server wraps a store.MultiRepoImporter and makes it implement
// MultiRepoImporterServer.
func Server(s MultiRepoImporterIndexer) MultiRepoImporterServer { return &server{s} }

type server struct{ u MultiRepoImporterIndexer }

func (s *server) Import(ctx context.Context, op *ImportOp) (*pbtypes.Void, error) {
	unit, err := op.Unit.SourceUnit()
	if err != nil {
		return nil, err
	}
	if op.Data == nil {
		op.Data = &graph.Output{}
	}
	if err := s.u.Import(op.Repo, op.CommitID, unit, *op.Data); err != nil {
		return nil, err
	}
	return &pbtypes.Void{}, nil
}

func (s *server) Index(ctx context.Context, op *IndexOp) (*pbtypes.Void, error) {
	if err := s.u.Index(op.Repo, op.CommitID); err != nil {
		return nil, err
	}
	return &pbtypes.Void{}, nil
}
