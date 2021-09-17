package background

import (
	"context"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/service"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

// batchSpecWorkspaceCreator takes in BatchSpecs, resolves them into
// RepoWorkspaces and then persists those as pending BatchSpecWorkspaces.
type batchSpecWorkspaceCreator struct {
	store *store.Store
}

// HandlerFunc returns a workeruitl.HandlerFunc that can be passed to a
// workerutil.Worker to process queued changesets.
func (e *batchSpecWorkspaceCreator) HandlerFunc() workerutil.HandlerFunc {
	return func(ctx context.Context, record workerutil.Record) (err error) {
		tx, err := e.store.Transact(ctx)
		if err != nil {
			return err
		}
		defer func() { err = tx.Done(err) }()

		return e.process(ctx, tx, service.NewWorkspaceResolver, record.(*btypes.BatchSpecResolutionJob))
	}
}

func (r *batchSpecWorkspaceCreator) process(
	ctx context.Context,
	tx *store.Store,
	newResolver service.WorkspaceResolverBuilder,
	job *btypes.BatchSpecResolutionJob,
) error {
	spec, err := tx.GetBatchSpec(ctx, store.GetBatchSpecOpts{ID: job.BatchSpecID})
	if err != nil {
		return err
	}

	evaluatableSpec, err := batcheslib.ParseBatchSpec([]byte(spec.RawSpec), batcheslib.ParseBatchSpecOptions{
		AllowArrayEnvironments: true,
		AllowTransformChanges:  true,
		AllowConditionalExec:   true,
	})
	if err != nil {
		return err
	}

	resolver := newResolver(tx)
	workspaces, unsupported, ignored, err := resolver.ResolveWorkspacesForBatchSpec(ctx, evaluatableSpec, service.ResolveWorkspacesForBatchSpecOpts{
		AllowUnsupported: job.AllowUnsupported,
		AllowIgnored:     job.AllowIgnored,
	})
	if err != nil {
		return err
	}

	log15.Info("resolved workspaces for batch spec", "job", job.ID, "spec", spec.ID, "workspaces", len(workspaces), "unsupported", len(unsupported), "ignored", len(ignored))

	var ws []*btypes.BatchSpecWorkspace
	for _, w := range workspaces {
		ws = append(ws, &btypes.BatchSpecWorkspace{
			BatchSpecID:      spec.ID,
			ChangesetSpecIDs: []int64{},

			RepoID:             w.Repo.ID,
			Branch:             w.Branch,
			Commit:             string(w.Commit),
			Path:               w.Path,
			FileMatches:        w.FileMatches,
			OnlyFetchWorkspace: w.OnlyFetchWorkspace,
			Steps:              w.Steps,
		})
	}

	return tx.CreateBatchSpecWorkspace(ctx, ws...)
}
