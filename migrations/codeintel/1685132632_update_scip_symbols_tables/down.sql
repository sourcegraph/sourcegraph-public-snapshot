ALTER TABLE codeintel_scip_symbols DROP COLUMN IF EXISTS scheme_id;
ALTER TABLE codeintel_scip_symbols DROP COLUMN IF EXISTS package_manager_id;
ALTER TABLE codeintel_scip_symbols DROP COLUMN IF EXISTS package_name_id;
ALTER TABLE codeintel_scip_symbols DROP COLUMN IF EXISTS package_version_id;
ALTER TABLE codeintel_scip_symbols DROP COLUMN IF EXISTS descriptor_id;
ALTER TABLE codeintel_scip_symbols DROP COLUMN IF EXISTS descriptor_no_suffix_id;

DROP INDEX IF EXISTS codeintel_scip_symbols_fuzzy_selector;
DROP INDEX IF EXISTS codeintel_scip_symbols_precise_selector;

DROP TABLE IF EXISTS codeintel_scip_symbols_lookup;

DROP INDEX IF EXISTS codeintel_scip_symbols_lookup_unique_fuzzy;
DROP INDEX IF EXISTS codeintel_scip_symbols_lookup_unique_precise;
