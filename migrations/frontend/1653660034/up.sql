ALTER TABLE
    lsif_configuration_policies
ADD
    COLUMN IF NOT EXISTS tags text [];

-- Migrate boolean to tag, but keep boolean column around for backwards
-- compatibility.
UPDATE
    lsif_configuration_policies
SET
    tags = '{"indexing"}'
WHERE
    indexing_enabled;

UPDATE
    lsif_configuration_policies
SET
    tags = CASE
        WHEN (
            indexing_enabled
            AND for_retention
        ) THEN '{"indexing", "retention"}'
        WHEN indexing_enabled '{"indexing"}'
        WHEN for_retention '{"retention"}'
        ELSE '{}'
    END;
