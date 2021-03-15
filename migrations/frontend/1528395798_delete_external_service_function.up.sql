BEGIN;
  DROP TRIGGER IF EXISTS trig_delete_external_service_ref_on_external_service_repos ON external_services;
  DROP FUNCTION IF EXISTS delete_external_service_ref_on_external_service_repos;

  CREATE OR REPLACE FUNCTION delete_external_service(_id bigint) RETURNS boolean AS $$
  DECLARE repo_ids int[]; orphan_repo_ids int[]; existing boolean;
  BEGIN
    -- Does this external service exist?
    SELECT true INTO existing FROM external_services WHERE id = _id AND deleted_at IS NULL;

    IF NOT COALESCE(existing, false) THEN
      RETURN false;
    END IF;

    -- We begin by soft deleting the external service row itself.
    UPDATE external_services SET deleted_at = now() WHERE id = _id;

    -- Which repos were associated with this external service?
    SELECT array_agg(repo_id) INTO repo_ids
    FROM external_service_repos WHERE external_service_id = _id;

    -- Disable triggers for this transaction only for the next statements.
    -- We don't want trig_soft_delete_orphan_repo_by_external_service_repo nor
    -- trig_delete_repo_ref_on_external_service_repos to run because
    -- we handle that efficiently below.
    SET LOCAL session_replication_role = replica;

    -- Delete those associations, but disable all triggers
    DELETE FROM external_service_repos
    WHERE external_service_id = _id AND repo_id = ANY(repo_ids);

    -- Which of those repos are no longer associated with any external service?
    SELECT array_agg(id) INTO orphan_repo_ids
    FROM repo WHERE id = ANY(repo_ids) AND NOT EXISTS (
      SELECT FROM external_service_repos WHERE repo_id = ANY(repo_ids)
    );

    -- Delete those repos.
    UPDATE repo SET (name, deleted_at) =
      (soft_deleted_repository_name(name), transaction_timestamp())
    WHERE id = ANY(orphan_repo_ids);

    -- Return true for success.
    RETURN true;
  END;
  $$ LANGUAGE plpgsql;
COMMIT;
