BEGIN;

DROP INDEX IF EXISTS cert_cache_key_idx;
DROP TABLE IF EXISTS cert_cache;

COMMIT;
