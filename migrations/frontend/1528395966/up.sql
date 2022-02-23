-- Note that we have to regenerate the reconciler_changesets view, as the SELECT
-- c.* in the view definition isn't refreshed when the fields change within the
-- changesets table.
DROP VIEW IF EXISTS
    reconciler_changesets;

ALTER TABLE
  changesets
ADD COLUMN IF NOT EXISTS
  external_fork_namespace CITEXT NULL;

CREATE VIEW reconciler_changesets AS
    SELECT c.* FROM changesets c
    INNER JOIN repo r on r.id = c.repo_id
    WHERE
        r.deleted_at IS NULL AND
        EXISTS (
            SELECT 1 FROM batch_changes
            LEFT JOIN users namespace_user ON batch_changes.namespace_user_id = namespace_user.id
            LEFT JOIN orgs namespace_org ON batch_changes.namespace_org_id = namespace_org.id
            WHERE
                c.batch_change_ids ? batch_changes.id::text AND
                namespace_user.deleted_at IS NULL AND
                namespace_org.deleted_at IS NULL
        )
;
