-- +++
-- parent: 1528395896
-- +++

CREATE INDEX CONCURRENTLY IF NOT EXISTS registry_extension_releases_registry_extension_id_created_at ON registry_extension_releases(registry_extension_id, created_at) WHERE deleted_at IS NULL;
