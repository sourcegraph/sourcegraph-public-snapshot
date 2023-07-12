CREATE TABLE IF NOT EXISTS codeintel_scip_symbols_lookup (
    upload_id integer NOT NULL,
    scip_name_type text NOT NULL,
    name text NOT NULL,
    id integer NOT NULL,
    parent_id integer
);

CREATE UNIQUE INDEX IF NOT EXISTS codeintel_scip_symbols_lookup_id ON codeintel_scip_symbols_lookup(upload_id, id);
CREATE INDEX IF NOT EXISTS codeintel_scip_symbols_lookup_search ON codeintel_scip_symbols_lookup(upload_id, scip_name_type, name) WHERE scip_name_type = 'DESCRIPTOR' OR scip_name_type = 'DESCRIPTOR_NO_SUFFIX';

CREATE TABLE IF NOT EXISTS codeintel_scip_symbols_lookup_leaves (
    upload_id integer NOT NULL,
    symbol_id integer NOT NULL,
    descriptor_id integer NOT NULL,
    descriptor_no_suffix_id integer NOT NULL
);

CREATE INDEX IF NOT EXISTS codeintel_scip_symbols_lookup_leaves_descriptor_id ON codeintel_scip_symbols_lookup_leaves(upload_id, descriptor_id);
CREATE INDEX IF NOT EXISTS codeintel_scip_symbols_lookup_leaves_descriptor_no_suffix_id ON codeintel_scip_symbols_lookup_leaves(upload_id, descriptor_no_suffix_id);

CREATE TABLE IF NOT EXISTS codeintel_scip_symbols_migration_progress (
    upload_id integer NOT NULL PRIMARY KEY,
    symbol_id integer NOT NULL
);
