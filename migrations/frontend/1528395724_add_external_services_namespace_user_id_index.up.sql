-- Note: CREATE INDEX CONCURRENTLY cannot run inside a transaction block

CREATE INDEX CONCURRENTLY external_services_namespace_user_id_idx ON external_services (namespace_user_id);

