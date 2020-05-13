package campaigns

import (
	"context"

	"github.com/pkg/errors"
)

// MigratePatchesWithoutDiffStats loads all Patches from the database where
// the DiffStat* fields are not set. It then computes the diff for each Patch
// and updates the Patch in the database.
// It is a blocking operation that does all work in a database transaction and
// fails if setting the diff stat on a single patch failed.
func MigratePatchesWithoutDiffStats(ctx context.Context, s *Store) (err error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return errors.Wrap(err, "starting transaction")
	}
	defer tx.Done(&err)

	opts := ListPatchesOpts{OnlyWithoutDiffStats: true, Limit: -1}
	patches, _, err := s.ListPatches(ctx, opts)

	for _, p := range patches {
		err = p.ComputeDiffStat()
		if err != nil {
			return errors.Wrapf(err, "failed to compute diff stat for patch %d", p.ID)
		}

		err = s.UpdatePatch(ctx, p)
		if err != nil {
			return errors.Wrapf(err, "failed to update patch %d", p.ID)
		}
	}

	return nil
}
