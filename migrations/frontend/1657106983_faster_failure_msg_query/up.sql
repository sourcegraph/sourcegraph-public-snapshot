CREATE INDEX CONCURRENTLY IF NOT EXISTS external_service_sync_jobs_state_external_service_id ON external_service_sync_jobs (state, external_service_id) INCLUDE (finished_at);
