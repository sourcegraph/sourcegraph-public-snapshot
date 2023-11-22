package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/api"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type ListCodeHostsOpts struct {
	RepoIDs             []api.RepoID
	OnlyWithoutWebhooks bool
}

func (s *Store) ListCodeHosts(ctx context.Context, opts ListCodeHostsOpts) (cs []*btypes.CodeHost, err error) {
	ctx, _, endObservation := s.operations.listCodeHosts.With(ctx, &err, observation.Args{})
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
WITH
	-- esr_with_ssh includes all external_service_repos records where the
	-- external service is cloned over SSH.
    esr_with_ssh AS (
        SELECT
            *
        FROM
            external_service_repos
        WHERE
            -- Either start with ssh://
            clone_url ILIKE %s OR
            -- OR have no schema specified at all
            clone_url !~* '[[:alpha:]]+://'
    ),
	-- esr_with_webhooks includes all external_service_repos records where the
	-- external service has one or more webhooks configured.
    esr_with_webhooks AS (
        SELECT
            *
        FROM
            external_service_repos
        WHERE
            external_service_id IN (
                SELECT
                    id
                FROM
                    external_services
                WHERE
                    -- has_webhooks can be NULL if the OOB migration hasn't yet
                    -- calculated the field value. We'll fail open here: the
                    -- worst case is that we report that a repo has webhooks
                    -- when it doesn't, which is less disruptive than the
                    -- alternative.
                    has_webhooks IS NULL OR
                    has_webhooks = TRUE
            )
    ),
	aggregated_repos AS (
		SELECT
			repo.external_service_type,
			repo.external_service_id,
			COUNT(esr_with_ssh.external_service_id) AS ssh_required_count,
			COUNT(esr_with_webhooks.external_service_id) AS has_webhooks_count
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
			repo.external_service_type, repo.external_service_id
		ORDER BY
			repo.external_service_type ASC, repo.external_service_id ASC
	)
SELECT
	external_service_type,
	external_service_id,
	ssh_required_count > 0 AS ssh_required,
	has_webhooks_count > 0 AS has_webhooks
FROM
	aggregated_repos
WHERE
	%s
`

func listCodeHostsQuery(opts ListCodeHostsOpts) *sqlf.Query {
	repoPreds := []*sqlf.Query{
		// Only for those which have any enabled repositories.
		sqlf.Sprintf("repo.deleted_at IS NULL"),
	}

	// Only show supported hosts.
	supportedTypes := []*sqlf.Query{}
	for extSvcType := range btypes.GetSupportedExternalServices() {
		supportedTypes = append(supportedTypes, sqlf.Sprintf("%s", extSvcType))
	}
	repoPreds = append(repoPreds, sqlf.Sprintf("repo.external_service_type IN (%s)", sqlf.Join(supportedTypes, ", ")))

	if len(opts.RepoIDs) > 0 {
		repoPreds = append(repoPreds, sqlf.Sprintf("repo.id = ANY (%s)", pq.Array(opts.RepoIDs)))
	}

	var aggregatePreds []*sqlf.Query
	if opts.OnlyWithoutWebhooks {
		aggregatePreds = append(aggregatePreds, sqlf.Sprintf("has_webhooks_count = 0 AND external_service_id NOT IN (SELECT DISTINCT(code_host_urn) FROM webhooks)"))
	} else {
		aggregatePreds = append(aggregatePreds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		listCodeHostsQueryFmtstr,
		sqlf.Sprintf("%s", "ssh://%"),
		sqlf.Join(repoPreds, "AND"),
		sqlf.Join(aggregatePreds, "AND"),
	)
}

func scanCodeHost(c *btypes.CodeHost, sc dbutil.Scanner) error {
	return sc.Scan(
		&c.ExternalServiceType,
		&c.ExternalServiceID,
		&c.RequiresSSH,
		&c.HasWebhooks,
	)
}

type GetExternalServiceIDsOpts struct {
	ExternalServiceType string
	ExternalServiceID   string
}

func (s *Store) GetExternalServiceIDs(ctx context.Context, opts GetExternalServiceIDsOpts) (ids []int64, err error) {
	ctx, _, endObservation := s.operations.getExternalServiceIDs.With(ctx, &err, observation.Args{})
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
