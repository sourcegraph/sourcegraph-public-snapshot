ALTER TABLE exhaustive_search_jobs ADD COLUMN IF NOT EXISTS is_aggregated boolean NOT NULL DEFAULT false;
