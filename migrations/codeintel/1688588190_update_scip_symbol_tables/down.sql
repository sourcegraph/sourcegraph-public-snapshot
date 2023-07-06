DROP INDEX IF EXISTS codeintel_scip_symbols_fuzzy_selector;
DROP INDEX IF EXISTS codeintel_scip_symbols_precise_selector;

ALTER TABLE codeintel_scip_symbols DROP COLUMN IF EXISTS descriptor_id;
ALTER TABLE codeintel_scip_symbols DROP COLUMN IF EXISTS descriptor_no_suffix_id;

DROP TABLE IF EXISTS codeintel_scip_symbols_lookup;
DROP TABLE IF EXISTS codeintel_scip_symbols_migration_progress;
