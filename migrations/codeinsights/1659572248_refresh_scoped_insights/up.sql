ALTER TABLE IF EXISTS insight_series
    ADD COLUMN IF NOT EXISTS needs_migration bool;


update insight_series
set needs_migration = true
where cardinality(repositories) > 0 AND generation_method in ('search', 'search-compute') AND needs_migration is NULL;
