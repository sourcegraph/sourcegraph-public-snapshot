-- Bust the cache, since we might run into unique key constraint errors.
DELETE FROM batch_spec_execution_cache_entries;

ALTER TABLE batch_spec_execution_cache_entries
  DROP CONSTRAINT IF EXISTS batch_spec_execution_cache_entries_user_id_key_unique,
  DROP COLUMN IF EXISTS user_id,
  DROP CONSTRAINT IF EXISTS batch_spec_execution_cache_entries_key_unique,
  ADD CONSTRAINT batch_spec_execution_cache_entries_key_unique UNIQUE (key);
