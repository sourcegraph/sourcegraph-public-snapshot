BEGIN;

DELETE FROM changeset_events WHERE kind LIKE 'gitlab:%';

-- Update all gitlab changesets to appear as being synced at least 8 hours ago.
-- That will increase their priority in the syncer and make sure those are updated more quickly,
-- so the state lost before is restored rather fast.
UPDATE changesets SET updated_at = updated_at - '8 hours'::interval WHERE external_service_type = 'gitlab';

COMMIT;
