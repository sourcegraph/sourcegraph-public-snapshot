CREATE TABLE IF NOT EXISTS codeintel_scip_symbols_lookup (
    upload_id     integer               NOT NULL,
    segment_type  SymbolNameSegmentType NOT NULL,
    name          text                  NOT NULL,
    id            integer               NOT NULL,
    parent_id     integer
);

CREATE UNIQUE INDEX IF NOT EXISTS codeintel_scip_symbols_lookup_id ON codeintel_scip_symbols_lookup(upload_id, id);
CREATE INDEX IF NOT EXISTS codeintel_scip_symbols_lookup_search ON codeintel_scip_symbols_lookup(upload_id, segment_type, name)
    WHERE
        segment_type = 'DESCRIPTOR_SUFFIX' OR
        segment_type = 'DESCRIPTOR_SUFFIX_FUZZY';

CREATE TABLE IF NOT EXISTS codeintel_scip_symbols_lookup_leaves (
    upload_id                   integer NOT NULL,
    symbol_id                   integer NOT NULL,
    descriptor_suffix_id        integer NOT NULL,
    fuzzy_descriptor_suffix_id  integer NOT NULL
);

CREATE INDEX IF NOT EXISTS codeintel_scip_symbols_lookup_leaves_descriptor_suffix_id ON codeintel_scip_symbols_lookup_leaves(upload_id, descriptor_suffix_id);
CREATE INDEX IF NOT EXISTS codeintel_scip_symbols_lookup_leaves_fuzzy_descriptor_suffix_id ON codeintel_scip_symbols_lookup_leaves(upload_id, fuzzy_descriptor_suffix_id);

CREATE TABLE IF NOT EXISTS codeintel_scip_symbols_migration_progress (
    upload_id  integer NOT NULL PRIMARY KEY,
    symbol_id  integer NOT NULL
);
