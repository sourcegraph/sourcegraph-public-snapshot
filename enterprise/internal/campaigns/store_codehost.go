package campaigns

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
)

type ListCodeHostsOpts struct {
	RepoIDs []api.RepoID
}

func (s *Store) ListCodeHosts(ctx context.Context, opts ListCodeHostsOpts) ([]*campaigns.CodeHost, error) {
	q := listCodeHostsQuery(opts)

	cs := make([]*campaigns.CodeHost, 0)
	err := s.query(ctx, q, func(sc scanner) error {
		var c campaigns.CodeHost
		if err := scanCodeHost(&c, sc); err != nil {
			return err
		}
		cs = append(cs, &c)
		return nil
	})

	return cs, err
}

var listCodeHostsQueryFmtstr = `
-- source: enterprise/internal/campaigns/store_codehost.go:ListCodeHosts
SELECT
	external_service_type, external_service_id
FROM repo
WHERE %s
GROUP BY external_service_type, external_service_id
ORDER BY external_service_type ASC, external_service_id ASC
`

func listCodeHostsQuery(opts ListCodeHostsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		// Only for those which have any enabled repositories.
		sqlf.Sprintf("repo.deleted_at IS NULL"),
	}

	// Only show supported hosts.
	supportedTypes := []*sqlf.Query{}
	for extSvcType := range campaigns.SupportedExternalServices {
		supportedTypes = append(supportedTypes, sqlf.Sprintf("%s", extSvcType))
	}
	preds = append(preds, sqlf.Sprintf("external_service_type IN (%s)", sqlf.Join(supportedTypes, ", ")))

	if len(opts.RepoIDs) > 0 {
		repoIDs := make([]*sqlf.Query, len(opts.RepoIDs))
		for i, id := range opts.RepoIDs {
			repoIDs[i] = sqlf.Sprintf("%s", id)
		}
		preds = append(preds, sqlf.Sprintf("repo.id IN (%s)", sqlf.Join(repoIDs, ",")))
	}

	return sqlf.Sprintf(listCodeHostsQueryFmtstr, sqlf.Join(preds, "AND"))
}

func scanCodeHost(c *campaigns.CodeHost, sc scanner) error {
	return sc.Scan(
		&c.ExternalServiceType,
		&c.ExternalServiceID,
	)
}

type GetExternalServiceIDOpts struct {
	ExternalServiceType string
	ExternalServiceID   string
}

func (s *Store) GetExternalServiceID(ctx context.Context, opts GetExternalServiceIDOpts) (int64, error) {
	q := getExternalServiceIDQuery(opts)

	id, ok, err := basestore.ScanFirstInt(s.Query(ctx, q))
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, ErrNoResults
	}
	return int64(id), nil
}

const getExternalServiceIDQueryFmtstr = `
-- source: enterprise/internal/campaigns/store_codehost.go:GetExternalServiceID
SELECT
	external_services.id
FROM external_services
JOIN external_service_repos esr ON esr.external_service_id = external_services.id
JOIN repo ON esr.repo_id = repo.id
WHERE %s
ORDER BY external_services.id ASC
LIMIT 1
`

func getExternalServiceIDQuery(opts GetExternalServiceIDOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.external_service_type = %s", opts.ExternalServiceType),
		sqlf.Sprintf("repo.external_service_id = %s", opts.ExternalServiceID),
		sqlf.Sprintf("external_services.deleted_at IS NULL"),
		sqlf.Sprintf("repo.deleted_at IS NULL"),
	}
	return sqlf.Sprintf(getExternalServiceIDQueryFmtstr, sqlf.Join(preds, "AND"))
}
