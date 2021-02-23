BEGIN;

CREATE VIEW reconciler_changesets AS
    SELECT c.* FROM changesets c
    INNER JOIN repo r on r.id = c.repo_id
    WHERE
        r.deleted_at IS NULL AND
        EXISTS (
            SELECT 1 FROM campaigns
            LEFT JOIN users namespace_user ON campaigns.namespace_user_id = namespace_user.id
            LEFT JOIN orgs namespace_org ON campaigns.namespace_org_id = namespace_org.id
            WHERE
                c.campaign_ids ? campaigns.id::text AND
                namespace_user.deleted_at IS NULL AND
                namespace_org.deleted_at IS NULL
        )
;

COMMIT;
