BEGIN;

--
-- Public

-- Change comment
COMMENT ON COLUMN lsif_data_docs_search_current_public.dump_id IS 'The associated dump identifier.';

-- Create new created_at column to decide a leader
ALTER TABLE lsif_data_docs_search_current_public ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
COMMENT ON COLUMN lsif_data_docs_search_current_public.created_at IS 'The time this record was inserted. The records with the latest created_at value for the same repository, root, and language is the only visible one and others will be deleted asynchronously.';

-- Add default to last_cleanup_scan_at column
ALTER TABLE lsif_data_docs_search_current_public ALTER COLUMN last_cleanup_scan_at SET DEFAULT NOW();
COMMENT ON COLUMN lsif_data_docs_search_current_public.last_cleanup_scan_at IS 'The last time this record was checked as part of a data retention scan.';

-- Create new indexes
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_current_public_lookup
ON lsif_data_docs_search_current_public(repo_id, dump_root, lang_name_id, created_at)
INCLUDE (dump_id);

CREATE INDEX IF NOT EXISTS lsif_data_docs_search_current_public_last_cleanup_scan_at ON lsif_data_docs_search_current_public(last_cleanup_scan_at);

-- Drop existing primary key
ALTER TABLE lsif_data_docs_search_current_public DROP CONSTRAINT IF EXISTS lsif_data_docs_search_current_public_pkey;

-- Create new serial primary key
ALTER TABLE lsif_data_docs_search_current_public ADD COLUMN IF NOT EXISTS id SERIAL PRIMARY KEY;

--
-- Private

-- Change comment
COMMENT ON COLUMN lsif_data_docs_search_current_private.dump_id IS 'The associated dump identifier.';

-- Create new created_at column to decide a leader
ALTER TABLE lsif_data_docs_search_current_private ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
COMMENT ON COLUMN lsif_data_docs_search_current_private.created_at IS 'The time this record was inserted. The records with the latest created_at value for the same repository, root, and language is the only visible one and others will be deleted asynchronously.';

-- Add default to last_cleanup_scan_at column
ALTER TABLE lsif_data_docs_search_current_private ALTER COLUMN last_cleanup_scan_at SET DEFAULT NOW();
COMMENT ON COLUMN lsif_data_docs_search_current_private.last_cleanup_scan_at IS 'The last time this record was checked as part of a data retention scan.';

-- Add index to last_cleanup_scan_at


-- Create new indexes
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_current_private_lookup
ON lsif_data_docs_search_current_private(repo_id, dump_root, lang_name_id, created_at)
INCLUDE (dump_id);

CREATE INDEX IF NOT EXISTS lsif_data_docs_search_current_private_last_cleanup_scan_at ON lsif_data_docs_search_current_private(last_cleanup_scan_at);

-- Drop existing primary key
ALTER TABLE lsif_data_docs_search_current_private DROP CONSTRAINT IF EXISTS lsif_data_docs_search_current_private_pkey;

-- Create new serial primary key
ALTER TABLE lsif_data_docs_search_current_private ADD COLUMN IF NOT EXISTS id SERIAL PRIMARY KEY;

COMMIT;
