ALTER TABLE lsif_uploads ADD COLUMN IF NOT EXISTS sanitized_indexer TEXT GENERATED ALWAYS AS (
    split_part(
        split_part(
            CASE
                -- Strip sourcegraph/ prefix if it exists
                WHEN strpos(indexer, 'sourcegraph/') = 1 THEN substr(indexer, length('sourcegraph/') + 1)
                ELSE indexer
            END,
        '@', 1), -- strip off @sha256:...
    ':', 1) -- strip off tag
) STORED;

ALTER TABLE lsif_indexes ADD COLUMN IF NOT EXISTS sanitized_indexer TEXT GENERATED ALWAYS AS (
    split_part(
        split_part(
            CASE
                -- Strip sourcegraph/ prefix if it exists
                WHEN strpos(indexer, 'sourcegraph/') = 1 THEN substr(indexer, length('sourcegraph/') + 1)
                ELSE indexer
            END,
        '@', 1), -- strip off @sha256:...
    ':', 1) -- strip off tag
) STORED;
