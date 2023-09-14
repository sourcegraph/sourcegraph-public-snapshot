WITH
progress AS (
    SELECT graph_key
    FROM codeintel_ranking_progress
    ORDER BY mappers_started_at DESC
    LIMIT 1
),
del_references AS (
    DELETE FROM codeintel_ranking_references_processed
    WHERE graph_key NOT IN (SELECT graph_key FROM progress)
    RETURNING 1
),
del_paths AS (
    DELETE FROM codeintel_initial_path_ranks_processed
    WHERE graph_key NOT IN (SELECT graph_key FROM progress)
    RETURNING 1
)
SELECT
    (SELECT COUNT(*) FROM del_references),
    (SELECT COUNT(*) FROM del_paths);
