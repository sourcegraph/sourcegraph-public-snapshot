BEGIN;
-- When row-level security is enabled, the table immediately "fails closed"
-- for all roles other than the table owner. The "sg_repo_access_policy" then
-- dictates the filtering mechanism used to decide what rows can be seen or
-- updated by the specified role(s).
--
-- The USING clause requires four LOCAL variables to be set:
--     1. rls.bypass: whether the policy should be skipped (such as for local admins)
--     2. rls.use_permissions_user_mapping: the switch to turn on permissions uesr mapping
--     3. rls.user_id: the effective ID for the user making the request
--     4. rls.permission: the permission that the user needs (e.g. "read")
ALTER POLICY sg_repo_access_policy ON repo TO sg_service USING (
  (
    current_setting('rls.bypass')::BOOLEAN -- Permit complete access for local admins
  )
  OR
  (
    NOT current_setting(
      'rls.use_permissions_user_mapping'
    )::BOOLEAN -- Disregard unrestricted state when permissions user mapping is enabled
    AND (
      repo.private IS false -- Happy path of non-private repositories
      OR EXISTS (
        -- Each external service defines if repositories are unrestricted
        SELECT
        FROM
          external_services AS es
          JOIN external_service_repos AS esr ON (
            esr.external_service_id = es.id
            AND esr.repo_id = repo.id
            AND es.unrestricted = TRUE
            AND es.deleted_at IS NULL
          )
        LIMIT
          1
      )
    )
  ) OR EXISTS (
    -- We assume that all repos added by the authenticated user should be shown
    SELECT
      1
    FROM
      external_service_repos
    WHERE
      repo_id = repo.id
      AND user_id = current_setting('rls.user_id')::INTEGER
  )
  OR (
    -- Restricted repositories require checking permissions
    SELECT
      object_ids_ints @> INTSET(repo.id)
    FROM
      user_permissions
    WHERE
      user_id = current_setting('rls.user_id')::INTEGER
      AND permission = current_setting('rls.permission')::TEXT
      AND object_type = 'repos'
  )
);
COMMIT;
