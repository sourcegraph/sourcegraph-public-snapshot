ALTER TABLE codeintel_scip_symbols ADD COLUMN IF NOT EXISTS symbol_name text NOT NULL;
ALTER TABLE codeintel_scip_symbols DROP CONSTRAINT IF EXISTS codeintel_scip_symbols_pkey;
ALTER TABLE codeintel_scip_symbols ADD PRIMARY KEY (upload_id, symbol_name, document_lookup_id);
ALTER TABLE codeintel_scip_symbols DROP COLUMN IF EXISTS symbol_id;

DROP TABLE IF EXISTS codeintel_scip_symbol_names;
