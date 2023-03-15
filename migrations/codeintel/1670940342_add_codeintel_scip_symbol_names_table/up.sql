CREATE TABLE IF NOT EXISTS codeintel_scip_symbol_names (
    id integer NOT NULL,
    upload_id integer NOT NULL,
    name_segment text NOT NULL,
    prefix_id integer,
    PRIMARY KEY (upload_id, id)
);

COMMENT ON TABLE codeintel_scip_symbol_names IS 'Stores a prefix tree of symbol names within a particular upload.';
COMMENT ON COLUMN codeintel_scip_symbol_names.id IS 'An identifier unique within the index for this symbol name segment.';
COMMENT ON COLUMN codeintel_scip_symbol_names.upload_id IS 'The identifier of the upload that provided this SCIP index.';
COMMENT ON COLUMN codeintel_scip_symbol_names.name_segment IS 'The portion of the symbol name that is unique to this symbol and its children.';
COMMENT ON COLUMN codeintel_scip_symbol_names.prefix_id IS 'The identifier of the segment that forms the prefix of this symbol, if any.';

ALTER TABLE codeintel_scip_symbols ADD COLUMN IF NOT EXISTS symbol_id integer NOT NULL;
COMMENT ON COLUMN codeintel_scip_symbols.symbol_id IS 'The identifier of the segment that terminates the name of this symbol. See the table [`codeintel_scip_symbol_names`](#table-publiccodeintel_scip_symbol_names) on how to reconstruct the full symbol name.';
ALTER TABLE codeintel_scip_symbols DROP CONSTRAINT IF EXISTS codeintel_scip_symbols_pkey;
ALTER TABLE codeintel_scip_symbols ADD PRIMARY KEY (upload_id, symbol_id, document_lookup_id);
ALTER TABLE codeintel_scip_symbols DROP COLUMN IF EXISTS symbol_name;
