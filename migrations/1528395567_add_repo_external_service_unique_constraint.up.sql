ALTER TABLE repo ADD CONSTRAINT repo_external_service_unique UNIQUE (external_id, external_service_type, external_service_id);
