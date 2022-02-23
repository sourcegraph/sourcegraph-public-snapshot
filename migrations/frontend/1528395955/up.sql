-- Bust the cache, since we can't recreate the user_id for existing cache entries.
DELETE FROM batch_spec_execution_cache_entries;

ALTER TABLE batch_spec_execution_cache_entries
  ADD COLUMN IF NOT EXISTS user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE DEFERRABLE,
  DROP CONSTRAINT IF EXISTS batch_spec_execution_cache_entries_key_unique,
    DROP CONSTRAINT IF EXISTS batch_spec_execution_cache_entries_user_id_key_unique,
  ADD CONSTRAINT batch_spec_execution_cache_entries_user_id_key_unique UNIQUE (user_id, key);
