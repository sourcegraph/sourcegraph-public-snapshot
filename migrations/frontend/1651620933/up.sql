CREATE UNIQUE INDEX IF NOT EXISTS external_services_unique_kind_org_id ON external_services (kind, namespace_org_id) WHERE (deleted_at IS NULL AND namespace_org_id IS NOT NULL);
