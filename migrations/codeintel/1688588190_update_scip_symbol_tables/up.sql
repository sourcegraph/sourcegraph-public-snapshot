ALTER TABLE codeintel_scip_symbols ADD COLUMN IF NOT EXISTS descriptor_id integer;
ALTER TABLE codeintel_scip_symbols ADD COLUMN IF NOT EXISTS descriptor_no_suffix_id integer;

CREATE INDEX IF NOT EXISTS codeintel_scip_symbols_precise_selector ON codeintel_scip_symbols(upload_id, descriptor_id);
CREATE INDEX IF NOT EXISTS codeintel_scip_symbols_fuzzy_selector ON codeintel_scip_symbols(upload_id, descriptor_no_suffix_id);

CREATE TABLE IF NOT EXISTS codeintel_scip_symbols_lookup (
    id bigint NOT NULL,
    upload_id integer NOT NULL,
    name text NOT NULL,
    scip_name_type text NOT NULL,
    parent_id integer
);

CREATE INDEX IF NOT EXISTS codeintel_scip_symbols_lookup_unique_fuzzy ON codeintel_scip_symbols_lookup(upload_id, scip_name_type, name);
CREATE UNIQUE INDEX IF NOT EXISTS codeintel_scip_symbols_lookup_unique_precise ON codeintel_scip_symbols_lookup(upload_id, id);

CREATE TABLE IF NOT EXISTS codeintel_scip_symbols_migration_progress (
    upload_id integer NOT NULL PRIMARY KEY,
    symbol_id integer NOT NULL
);
