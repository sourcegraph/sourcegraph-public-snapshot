-- Should run as single non-transaction block
CREATE INDEX IF NOT EXISTS repo_hashed_name_idx ON repo USING BTREE (encode(sha256(lower(name)::bytea), 'hex'));
