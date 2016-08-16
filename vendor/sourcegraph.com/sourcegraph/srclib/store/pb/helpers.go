package pb

import (
	"context"

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
	_, err := c.u.Import(c.ctx, &ImportOp{
		Repo:     repo,
		CommitID: commitID,
		Unit:     u,
		Data:     &data,
	})
	return err
}

func (c *client) CreateVersion(repo, commitID string) error {
	_, err := c.u.CreateVersion(c.ctx, &CreateVersionOp{
		Repo:     repo,
		CommitID: commitID,
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
	if op.Data == nil {
		op.Data = &graph.Output{}
	}
	if err := s.u.Import(op.Repo, op.CommitID, op.Unit, *op.Data); err != nil {
		return nil, err
	}
	return &pbtypes.Void{}, nil
}

func (s *server) CreateVersion(ctx context.Context, op *CreateVersionOp) (*pbtypes.Void, error) {
	if err := s.u.CreateVersion(op.Repo, op.CommitID); err != nil {
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
