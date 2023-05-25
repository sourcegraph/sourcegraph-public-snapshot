DROP VIEW IF EXISTS codeintel_configuration_policies;
DROP VIEW IF EXISTS codeintel_configuration_policies_repository_pattern_lookup;

ALTER TABLE lsif_configuration_policies DROP COLUMN IF EXISTS embeddings_enabled;
ALTER TABLE lsif_configuration_policies ADD COLUMN IF NOT EXISTS lockfile_indexing_enabled BOOLEAN NOT NULL DEFAULT false;
