package store

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type ListExternalServicesForBatchChangeOpts struct {
	LimitOpts
	Cursor        int64
	BatchChangeID int64
}

// ListExternalServicesForBatchChange lists the external services that the given
// batch change has changesets published on.
//
// 🚨 SECURITY: Although this is filtered using authz, only site admins are
// permitted access to the configuration within external services, since it may
// include secrets. The raw results of this method MUST NOT be used unless a
// site admin check has already occurred before invoking this method.
func (s *Store) ListExternalServicesForBatchChange(ctx context.Context, opts ListExternalServicesForBatchChangeOpts) (es []*types.ExternalService, next int64, err error) {
	ctx, endObservation := s.operations.listExternalServicesForBatchChange.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	es, next, err = s.listExternalServicesForBatchChange(ctx, &opts)
	return
}

func (s *Store) listExternalServicesForBatchChange(ctx context.Context, opts *ListExternalServicesForBatchChangeOpts) ([]*types.ExternalService, int64, error) {
	// We'll retrieve the external service IDs for the given batch change first,
	// then call ExternalServiceStore.List to actually hydrate the list of
	// returned external services.
	//
	// This involves two SELECTs, which would normally be a bit silly, but
	// there's a good reason for it here: we need to apply the authz query
	// conditions at this level, since we have the repos here, and then
	// ExternalServiceStore.List has logic to handle decrypting the
	// configuration, which we can't do in the database.
	//
	// The actual cost of this, in practice, is generally trivial: most batch
	// changes are going to ultimately only touch one external service (since
	// most customers only have one external service). We'll spend far more time
	// doing the joins and grouping required for the first query (to get the
	// IDs) -- which we'd have to do anyway, even if we replicated the
	// decryption logic here -- that we will making a SELECT of a single record
	// by its ID.

	repoAuthzConds, err := database.AuthzQueryConds(ctx, s.Handle().DB())
	if err != nil {
		return nil, 0, errors.Wrap(err, "ListExternalServices generating authz query conds")
	}
	q := listExternalServicesForBatchChangeQuery(opts, repoAuthzConds)

	// Let's go get some external service IDs.
	ids := make([]int64, 0, opts.DBLimit())
	if err := s.query(ctx, q, func(sc dbutil.Scanner) error {
		var id int64
		if err := sc.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
		return nil
	}); err != nil {
		return nil, 0, errors.Wrap(err, "ListExternalServices querying external service IDs")
	}

	// ExternalServiceStore.List will treat an empty ID list as being a query to
	// retrieve _all_ external services, so we need to short circuit that here.
	if len(ids) == 0 {
		return []*types.ExternalService{}, 0, nil
	}

	// Calculate the next cursor, if any.
	var next int64
	if opts.Limit != 0 && len(ids) == opts.DBLimit() {
		next = ids[len(ids)-1]
		ids = ids[:len(ids)-1]
	}

	// Now we'll go retrieve the real ExternalService objects.
	es, err := database.ExternalServicesWith(s.Store).List(ctx, database.ExternalServicesListOptions{
		IDs:              ids,
		OrderByDirection: "ASC",
	})
	if err != nil {
		return nil, 0, errors.Wrap(err, "ListExternalServices querying external services")
	}

	return es, next, nil
}

const listExternalServicesForBatchChangeQueryFmtstr = `
-- source: enterprise/internal/batches/store/external_services.go:ListExternalServicesForBatchChange
SELECT
	DISTINCT external_service_repos.external_service_id
FROM
	external_service_repos
INNER JOIN
	repo ON external_service_repos.repo_id = repo.id
INNER JOIN
	changesets ON repo.id = changesets.repo_id
WHERE
	changesets.batch_change_ids ? %s AND
	repo.deleted_at IS NULL AND
	%s AND -- authz conditions
	%s -- cursor, if given
ORDER BY
	external_service_repos.external_service_id ASC
`

func listExternalServicesForBatchChangeQuery(opts *ListExternalServicesForBatchChangeOpts, repoAuthzConds *sqlf.Query) *sqlf.Query {
	var cursor *sqlf.Query
	if opts.Cursor != 0 {
		cursor = sqlf.Sprintf("external_service_repos.external_service_id >= %s", opts.Cursor)
	} else {
		cursor = sqlf.Sprintf("TRUE")
	}

	return sqlf.Sprintf(
		listExternalServicesForBatchChangeQueryFmtstr+opts.LimitOpts.ToDB(),
		fmt.Sprint(opts.BatchChangeID),
		repoAuthzConds,
		cursor,
	)
}

// CountExternalServicesForBatchChange returns the number of external services
// that the given batch change has changesets on.
func (s *Store) CountExternalServicesForBatchChange(ctx context.Context, batchChangeID int64) (count int64, err error) {
	ctx, endObservation := s.operations.countExternalServicesForBatchChange.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	count, err = s.countExternalServicesForBatchChange(ctx, batchChangeID)
	return
}

func (s *Store) countExternalServicesForBatchChange(ctx context.Context, batchChangeID int64) (int64, error) {
	repoAuthzConds, err := database.AuthzQueryConds(ctx, s.Handle().DB())
	if err != nil {
		return 0, errors.Wrap(err, "CountExternalServices generating authz query conds")
	}
	q := countExternalServicesQuery(batchChangeID, repoAuthzConds)

	row := s.QueryRow(ctx, q)
	if row == nil {
		return 0, ErrNoResults
	}

	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, errors.Wrap(err, "CountExternalServices executing query")
	}

	return count, nil
}

const countExternalServicesForBatchChangeQueryFmtstr = `
-- source: enterprise/internal/batches/store/external_services.go:CountExternalServicesForBatchChange
SELECT
	COUNT(DISTINCT external_service_repos.external_service_id)
FROM
	external_service_repos
INNER JOIN
	repo ON external_service_repos.repo_id = repo.id
INNER JOIN
	changesets ON repo.id = changesets.repo_id
WHERE
	changesets.batch_change_ids ? %s AND
	repo.deleted_at IS NULL AND
	%s -- authz conditions
`

func countExternalServicesQuery(batchChangeID int64, repoAuthzConds *sqlf.Query) *sqlf.Query {
	return sqlf.Sprintf(
		countExternalServicesForBatchChangeQueryFmtstr,
		fmt.Sprint(batchChangeID),
		repoAuthzConds,
	)
}
