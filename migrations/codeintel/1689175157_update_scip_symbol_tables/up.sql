CREATE TABLE IF NOT EXISTS codeintel_scip_symbols_lookup (
    upload_id       integer                  NOT NULL,
    segment_type    SymbolNameSegmentType    NOT NULL,
    segment_quality SymbolNameSegmentQuality,
    name            text                     NOT NULL,
    id              integer                  NOT NULL,
    parent_id       integer
);

-- Reconstruct names from leaf (descriptor) in tree
CREATE UNIQUE INDEX IF NOT EXISTS codeintel_scip_symbols_lookup_id ON codeintel_scip_symbols_lookup(upload_id, id);

-- Search by exact descriptor suffix
CREATE INDEX IF NOT EXISTS codeintel_scip_symbols_lookup_descriptor_suffix ON codeintel_scip_symbols_lookup(upload_id, name)
    WHERE segment_type = 'DESCRIPTOR_SUFFIX' AND segment_quality != 'FUZZY';

-- Search by fuzzily for descriptor suffix
CREATE INDEX IF NOT EXISTS codeintel_scip_symbols_lookup_fuzzy_descriptor_suffix ON codeintel_scip_symbols_lookup(upload_id, reverse(name) text_pattern_ops)
    WHERE segment_type = 'DESCRIPTOR_SUFFIX' AND segment_quality != 'PRECISE';

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
