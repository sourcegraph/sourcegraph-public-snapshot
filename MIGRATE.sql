-- 4.22.2016
CREATE TABLE global_defs (
    repo text NOT NULL,
    commit_id text NOT NULL,
    unit_type text NOT NULL,
    unit text NOT NULL,
    path text NOT NULL,
    name text,
    kind text,
    file text,
    ref_ct integer DEFAULT 0,
    updated_at timestamp with time zone,
    data bytea,
    bow text,
    doc text
);

ALTER TABLE global_defs ALTER COLUMN updated_at TYPE timestamp with time zone USING updated_at::timestamp with time zone;
ALTER TABLE global_defs ALTER COLUMN ref_ct SET DEFAULT 0;
CREATE INDEX bow_idx ON global_defs USING gin(to_tsvector('english', bow));
CREATE INDEX doc_idx ON global_defs USING gin(to_tsvector('english', doc));
