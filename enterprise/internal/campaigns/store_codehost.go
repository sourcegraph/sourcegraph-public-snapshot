package campaigns

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

func (s *Store) GetCodeHosts(ctx context.Context) ([]*campaigns.CodeHost, error) {
	q := getCodeHostsQuery()

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

var getCodeHostsQueryFmtstr = `
-- source: enterprise/internal/campaigns/store_codehost.go:GetCodeHosts
SELECT
	external_service_type, external_service_id
FROM repo
WHERE %s
GROUP BY external_service_type, external_service_id
ORDER BY external_service_type ASC, external_service_id ASC
`

func getCodeHostsQuery() *sqlf.Query {
	preds := []*sqlf.Query{
		// Only show supported hosts.
		sqlf.Sprintf("external_service_type IN (%s, %s, %s)", extsvc.TypeGitHub, extsvc.TypeGitLab, extsvc.TypeBitbucketServer),
		// And only for those which have any enabled repositories.
		sqlf.Sprintf("repo.deleted_at IS NULL"),
	}
	return sqlf.Sprintf(getCodeHostsQueryFmtstr, sqlf.Join(preds, "AND"))
}

func scanCodeHost(c *campaigns.CodeHost, sc scanner) error {
	return sc.Scan(
		&c.ExternalServiceType,
		&c.ExternalServiceID,
	)
}
