-- This reverts a previous migration, so this is literally just 1669576792_make_batch_spec_of_batch_change_nullable/down.sql
WITH reconstructed_batch_specs AS (
    INSERT INTO batch_specs
        (batch_change_id, user_id, namespace_user_id, namespace_org_id, rand_id, raw_spec, spec, created_from_raw)
    SELECT
        id, creator_id, namespace_user_id, namespace_org_id, md5(CONCAT(id, name)::bytea), CONCAT('name: ', name), json_build_object('name', name), TRUE
    FROM
        batch_changes
    WHERE
        batch_spec_id IS NULL
    RETURNING
	    batch_change_id, id
)
UPDATE
    batch_changes
SET batch_spec_id = (SELECT id FROM reconstructed_batch_specs WHERE batch_change_id = batch_changes.id)
WHERE id IN (SELECT batch_change_id FROM reconstructed_batch_specs);

ALTER TABLE batch_changes ALTER COLUMN batch_spec_id SET NOT NULL;
