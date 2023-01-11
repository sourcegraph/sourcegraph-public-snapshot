ALTER TABLE
    lsif_references
ADD
    COLUMN IF NOT EXISTS filter bytea;
