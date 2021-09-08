package background

import (
	"context"

	"github.com/hashicorp/go-multierror"
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
		defer func() {
			doneErr := tx.Done(nil)
			if err != nil && doneErr != nil {
				err = multierror.Append(err, doneErr)
			}
			if doneErr != nil {
				err = doneErr
			}
		}()

		return e.process(ctx, tx, record.(*btypes.BatchSpec))
	}
}

func (r *batchSpecWorkspaceCreator) process(ctx context.Context, tx *store.Store, spec *btypes.BatchSpec) error {
	evaluatableSpec, err := batcheslib.ParseBatchSpec([]byte(spec.RawSpec), batcheslib.ParseBatchSpecOptions{
		AllowArrayEnvironments: true,
		AllowTransformChanges:  true,
		AllowConditionalExec:   true,
	})
	if err != nil {
		return err
	}

	workspaces, unsupported, ignored, err := service.New(tx).ResolveWorkspacesForBatchSpec(ctx, evaluatableSpec, service.ResolveWorkspacesForBatchSpecOpts{
		// TODO: Persist these also on batch_spec
		AllowIgnored:     true,
		AllowUnsupported: true,
	})
	if err != nil {
		return err
	}

	log15.Info("resolved workspaces for batch spec", "spec", spec.ID, "workspaces", len(workspaces), "unsupported", len(unsupported), "ignored", len(ignored))

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

			State: btypes.BatchSpecWorkspaceStatePending,
		})
	}

	return tx.CreateBatchSpecWorkspace(ctx, ws...)
}
