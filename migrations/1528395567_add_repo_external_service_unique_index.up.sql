CREATE UNIQUE INDEX repo_external_service_unique_idx
ON repo (external_service_type, external_service_id, external_id)
WHERE external_service_type IS NOT NULL
AND external_service_id IS NOT NULL
AND external_id IS NOT NULL;
