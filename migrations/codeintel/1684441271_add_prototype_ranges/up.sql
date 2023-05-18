ALTER TABLE codeintel_scip_symbols ADD COLUMN IF NOT EXISTS prototype_ranges bytea;

COMMENT ON COLUMN codeintel_scip_symbols.prototype_ranges IS  'An encoded set of ranges within the associated document that have a **prototype** relationship to the associated symbol.'
