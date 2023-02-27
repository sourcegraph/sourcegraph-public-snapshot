-- Force recalculation of reference counts
UPDATE
    lsif_uploads
SET
    reference_count = NULL;

-- Remove new table
DROP TABLE IF EXISTS lsif_uploads_reference_counts;
