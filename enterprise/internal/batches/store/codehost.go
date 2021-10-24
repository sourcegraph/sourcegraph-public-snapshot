package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type ListCodeHostsOpts struct {
	RepoIDs []api.RepoID
}

func (s *Store) ListCodeHosts(ctx context.Context, opts ListCodeHostsOpts) (cs []*btypes.CodeHost, err error) {
	ctx, endObservation := s.operations.listCodeHosts.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := listCodeHostsQuery(opts)

	cs = make([]*btypes.CodeHost, 0)
	err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		var c btypes.CodeHost
		if err := scanCodeHost(&c, sc); err != nil {
			return err
		}
		cs = append(cs, &c)
		return nil
	})

	return cs, err
}

var listCodeHostsQueryFmtstr = `
-- source: enterprise/internal/batches/store/codehost.go:ListCodeHosts
SELECT
	repo.external_service_type, repo.external_service_id, COUNT(esr.external_service_id) > 0 AS ssh_required
FROM repo
LEFT JOIN external_service_repos esr
	ON
		esr.repo_id = repo.id AND
		(
			-- Either start with ssh://
			esr.clone_url ILIKE %s OR
			-- OR have no scheme specified at all
			esr.clone_url !~* '[[:alpha:]]+://'
		)
WHERE %s
GROUP BY repo.external_service_type, repo.external_service_id
ORDER BY repo.external_service_type ASC, repo.external_service_id ASC
`

func listCodeHostsQuery(opts ListCodeHostsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		// Only for those which have any enabled repositories.
		sqlf.Sprintf("repo.deleted_at IS NULL"),
	}

	// Only show supported hosts.
	supportedTypes := []*sqlf.Query{}
	for extSvcType := range btypes.SupportedExternalServices {
		supportedTypes = append(supportedTypes, sqlf.Sprintf("%s", extSvcType))
	}
	preds = append(preds, sqlf.Sprintf("repo.external_service_type IN (%s)", sqlf.Join(supportedTypes, ", ")))

	if len(opts.RepoIDs) > 0 {
		preds = append(preds, sqlf.Sprintf("repo.id = ANY (%s)", pq.Array(opts.RepoIDs)))
	}

	return sqlf.Sprintf(listCodeHostsQueryFmtstr, sqlf.Sprintf("%s", "ssh://%"), sqlf.Join(preds, "AND"))
}

func scanCodeHost(c *btypes.CodeHost, sc dbutil.Scanner) error {
	return sc.Scan(
		&c.ExternalServiceType,
		&c.ExternalServiceID,
		&c.RequiresSSH,
	)
}

type GetExternalServiceIDsOpts struct {
	ExternalServiceType string
	ExternalServiceID   string
}

func (s *Store) GetExternalServiceIDs(ctx context.Context, opts GetExternalServiceIDsOpts) (ids []int64, err error) {
	ctx, endObservation := s.operations.getExternalServiceIDs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := getExternalServiceIDsQuery(opts)

	err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		var id int64
		sc.Scan(&id)
		if err := sc.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return ids, ErrNoResults
	}

	return ids, nil
}

const getExternalServiceIDsQueryFmtstr = `
-- source: enterprise/internal/batches/store/codehost.go:GetExternalServiceIDs
SELECT
	external_services.id
FROM external_services
JOIN external_service_repos esr ON esr.external_service_id = external_services.id
JOIN repo ON esr.repo_id = repo.id
WHERE %s
ORDER BY external_services.id ASC
`

func getExternalServiceIDsQuery(opts GetExternalServiceIDsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.external_service_type = %s", opts.ExternalServiceType),
		sqlf.Sprintf("repo.external_service_id = %s", opts.ExternalServiceID),
		sqlf.Sprintf("external_services.deleted_at IS NULL"),
		sqlf.Sprintf("repo.deleted_at IS NULL"),
	}
	return sqlf.Sprintf(getExternalServiceIDsQueryFmtstr, sqlf.Join(preds, "AND"))
}
