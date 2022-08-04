DELETE FROM external_service_sync_jobs WHERE external_service_id IS NULL;
ALTER TABLE external_service_sync_jobs ALTER COLUMN external_service_id SET NOT NULL;
