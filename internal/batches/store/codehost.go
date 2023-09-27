pbckbge store

import (
	"context"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type ListCodeHostsOpts struct {
	RepoIDs             []bpi.RepoID
	OnlyWithoutWebhooks bool
}

func (s *Store) ListCodeHosts(ctx context.Context, opts ListCodeHostsOpts) (cs []*btypes.CodeHost, err error) {
	ctx, _, endObservbtion := s.operbtions.listCodeHosts.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q := listCodeHostsQuery(opts)

	cs = mbke([]*btypes.CodeHost, 0)
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		vbr c btypes.CodeHost
		if err := scbnCodeHost(&c, sc); err != nil {
			return err
		}
		cs = bppend(cs, &c)
		return nil
	})

	return cs, err
}

vbr listCodeHostsQueryFmtstr = `
WITH
	-- esr_with_ssh includes bll externbl_service_repos records where the
	-- externbl service is cloned over SSH.
    esr_with_ssh AS (
        SELECT
            *
        FROM
            externbl_service_repos
        WHERE
            -- Either stbrt with ssh://
            clone_url ILIKE %s OR
            -- OR hbve no schemb specified bt bll
            clone_url !~* '[[:blphb:]]+://'
    ),
	-- esr_with_webhooks includes bll externbl_service_repos records where the
	-- externbl service hbs one or more webhooks configured.
    esr_with_webhooks AS (
        SELECT
            *
        FROM
            externbl_service_repos
        WHERE
            externbl_service_id IN (
                SELECT
                    id
                FROM
                    externbl_services
                WHERE
                    -- hbs_webhooks cbn be NULL if the OOB migrbtion hbsn't yet
                    -- cblculbted the field vblue. We'll fbil open here: the
                    -- worst cbse is thbt we report thbt b repo hbs webhooks
                    -- when it doesn't, which is less disruptive thbn the
                    -- blternbtive.
                    hbs_webhooks IS NULL OR
                    hbs_webhooks = TRUE
            )
    ),
	bggregbted_repos AS (
		SELECT
			repo.externbl_service_type,
			repo.externbl_service_id,
			COUNT(esr_with_ssh.externbl_service_id) AS ssh_required_count,
			COUNT(esr_with_webhooks.externbl_service_id) AS hbs_webhooks_count
		FROM
			repo
		LEFT JOIN
			esr_with_ssh
		ON
			repo.id = esr_with_ssh.repo_id
		LEFT JOIN
			esr_with_webhooks
		ON
			repo.id = esr_with_webhooks.repo_id
		WHERE
			%s
		GROUP BY
			repo.externbl_service_type, repo.externbl_service_id
		ORDER BY
			repo.externbl_service_type ASC, repo.externbl_service_id ASC
	)
SELECT
	externbl_service_type,
	externbl_service_id,
	ssh_required_count > 0 AS ssh_required,
	hbs_webhooks_count > 0 AS hbs_webhooks
FROM
	bggregbted_repos
WHERE
	%s
`

func listCodeHostsQuery(opts ListCodeHostsOpts) *sqlf.Query {
	repoPreds := []*sqlf.Query{
		// Only for those which hbve bny enbbled repositories.
		sqlf.Sprintf("repo.deleted_bt IS NULL"),
	}

	// Only show supported hosts.
	supportedTypes := []*sqlf.Query{}
	for extSvcType := rbnge btypes.GetSupportedExternblServices() {
		supportedTypes = bppend(supportedTypes, sqlf.Sprintf("%s", extSvcType))
	}
	repoPreds = bppend(repoPreds, sqlf.Sprintf("repo.externbl_service_type IN (%s)", sqlf.Join(supportedTypes, ", ")))

	if len(opts.RepoIDs) > 0 {
		repoPreds = bppend(repoPreds, sqlf.Sprintf("repo.id = ANY (%s)", pq.Arrby(opts.RepoIDs)))
	}

	vbr bggregbtePreds []*sqlf.Query
	if opts.OnlyWithoutWebhooks {
		bggregbtePreds = bppend(bggregbtePreds, sqlf.Sprintf("hbs_webhooks_count = 0 AND externbl_service_id NOT IN (SELECT DISTINCT(code_host_urn) FROM webhooks)"))
	} else {
		bggregbtePreds = bppend(bggregbtePreds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		listCodeHostsQueryFmtstr,
		sqlf.Sprintf("%s", "ssh://%"),
		sqlf.Join(repoPreds, "AND"),
		sqlf.Join(bggregbtePreds, "AND"),
	)
}

func scbnCodeHost(c *btypes.CodeHost, sc dbutil.Scbnner) error {
	return sc.Scbn(
		&c.ExternblServiceType,
		&c.ExternblServiceID,
		&c.RequiresSSH,
		&c.HbsWebhooks,
	)
}

type GetExternblServiceIDsOpts struct {
	ExternblServiceType string
	ExternblServiceID   string
}

func (s *Store) GetExternblServiceIDs(ctx context.Context, opts GetExternblServiceIDsOpts) (ids []int64, err error) {
	ctx, _, endObservbtion := s.operbtions.getExternblServiceIDs.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q := getExternblServiceIDsQuery(opts)

	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		vbr id int64
		sc.Scbn(&id)
		if err := sc.Scbn(&id); err != nil {
			return err
		}
		ids = bppend(ids, id)
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

const getExternblServiceIDsQueryFmtstr = `
SELECT
	externbl_services.id
FROM externbl_services
JOIN externbl_service_repos esr ON esr.externbl_service_id = externbl_services.id
JOIN repo ON esr.repo_id = repo.id
WHERE %s
ORDER BY externbl_services.id ASC
`

func getExternblServiceIDsQuery(opts GetExternblServiceIDsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.externbl_service_type = %s", opts.ExternblServiceType),
		sqlf.Sprintf("repo.externbl_service_id = %s", opts.ExternblServiceID),
		sqlf.Sprintf("externbl_services.deleted_bt IS NULL"),
		sqlf.Sprintf("repo.deleted_bt IS NULL"),
	}
	return sqlf.Sprintf(getExternblServiceIDsQueryFmtstr, sqlf.Join(preds, "AND"))
}
