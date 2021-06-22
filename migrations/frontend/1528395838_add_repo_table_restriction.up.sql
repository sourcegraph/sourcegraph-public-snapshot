BEGIN;

-- The "sgservice" role is one that the frontend and other services
-- will assume on startup/init. It lowers the privilege of those services
-- such that we can apply security policies to the role and let Postgres
-- manage things that previously would need to be done in app-level code.
CREATE ROLE sgservice;
GRANT USAGE ON SCHEMA public TO sgservice;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO sgservice;

-- When row-level security is enabled, the table immediately "fails closed"
-- for all roles other than the table owner. The "restricted_repo_policy"
-- dictates the filtering mechanism used to decide what rows can be seen or
-- updated by the specified role(s).
--
-- The USING clause requires two LOCAL variables to be set:
--   rls.user_id: the effective ID for the user making the request
--   rls.permission: the permission that the user needs (e.g. "write")
ALTER TABLE repo ENABLE ROW LEVEL SECURITY;
CREATE POLICY restricted_repo_policy
    ON repo
    FOR ALL
    TO sgservice
    USING (
        repo.private IS false     -- Happy path of non-private repositories
        OR  EXISTS (              -- Each external service defines if repositories are unrestricted
            SELECT
            FROM external_services AS es
            JOIN external_service_repos AS esr ON (
                    esr.external_service_id = es.id
                AND esr.repo_id = repo.id
                AND es.unrestricted = TRUE
                AND es.deleted_at IS NULL
            )
            LIMIT 1
        )
        OR  EXISTS (              -- We assume that all repos added by the authenticated user should be shown
			SELECT 1
			FROM external_service_repos
			WHERE repo_id = repo.id
			AND user_id = current_setting('rls.user_id')::INTEGER
		)
        OR (                      -- Restricted repositories require checking permissions
			SELECT object_ids_ints @> INTSET(repo.id)
			FROM user_permissions
			WHERE user_id = current_setting('rls.user_id')::INTEGER
			AND permission = current_setting('rls.permission')::TEXT
			AND object_type = 'repos'
		)
    );

COMMIT;
