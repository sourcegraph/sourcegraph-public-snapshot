package campaigns

import (
	"context"

	"github.com/keegancsmith/sqlf"
)

type CodeHost struct {
	ExternalServiceType string
	ExternalServiceID   string
}

type GetCodeHostsOpts struct {
	Limit int64
}

func (s *Store) GetCodeHosts(ctx context.Context, opts GetCodeHostsOpts) ([]*CodeHost, error) {
	q := getCodeHostsQuery(opts)

	cs := make([]*CodeHost, 0)
	err := s.query(ctx, q, func(sc scanner) error {
		var c CodeHost
		if err := scanCodeHost(&c, sc); err != nil {
			return err
		}
		cs = append(cs, &c)
		return nil
	})

	return cs, err
}

var getCodeHostsQueryFmtstr = `
-- source: enterprise/internal/campaigns/store_codehost.go:GetCodeHosts
SELECT
	external_service_type, external_service_id
FROM repo
WHERE %s
GROUP BY external_service_type, external_service_id
%s
`

func getCodeHostsQuery(opts GetCodeHostsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		// Only show supported hosts.
		sqlf.Sprintf("external_service_type IN ('github','gitlab','bitbucketServer')"),
		sqlf.Sprintf("repo.deleted_at IS NULL"),
	}
	limitQuery := sqlf.Sprintf("")
	if opts.Limit > 0 {
		limitQuery = sqlf.Sprintf("LIMIT %s", opts.Limit)
	}
	return sqlf.Sprintf(getCodeHostsQueryFmtstr, sqlf.Join(preds, "AND"), limitQuery)
}

func scanCodeHost(c *CodeHost, sc scanner) error {
	return sc.Scan(
		&c.ExternalServiceType,
		&c.ExternalServiceID,
	)
}
