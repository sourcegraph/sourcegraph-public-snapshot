CREATE INDEX CONCURRENTLY IF NOT EXISTS access_tokens_lookup_double_hash ON access_tokens USING HASH (digest(value_sha256, 'sha256'))
    WHERE
    deleted_at IS NULL;
