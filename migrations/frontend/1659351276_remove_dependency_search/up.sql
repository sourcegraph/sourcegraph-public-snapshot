ALTER TABLE lsif_configuration_policies
  DROP COLUMN IF EXISTS lockfile_indexing_enabled;

DROP TABLE IF EXISTS codeintel_lockfiles;
DROP TABLE IF EXISTS codeintel_lockfile_references;
