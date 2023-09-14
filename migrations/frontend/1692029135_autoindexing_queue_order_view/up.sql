CREATE OR REPLACE FUNCTION lsif_indexes_enqueue_candidates(lookback_window interval, cooldown interval)
RETURNS TABLE(
    id bigint ,
    commit text ,
    queued_at timestamptz ,
    state text ,
    failure_message text,
    started_at timestamptz,
    finished_at timestamptz,
    repository_id integer ,
    process_after timestamptz,
    num_resets integer ,
    num_failures integer ,
    docker_steps jsonb[] ,
    root text ,
    indexer text ,
    indexer_args text[] ,
    outfile text ,
    log_contents text,
    execution_logs json[],
    local_steps text[] ,
    should_reindex boolean ,
    requested_envvars text[],
    repository_name text ,
    enqueuer_user_id integer
)
AS $$
    WITH newest_queued AS MATERIALIZED (
        -- used to calculate top of window
        SELECT MAX(queued_at) AS newest
        FROM lsif_indexes
        WHERE state IN ('queued', 'errored')
        LIMIT 1
    ),
    newest_in_window AS NOT MATERIALIZED (
        -- select the newest index per repo within the window
        SELECT DISTINCT ON (repository_id) *
        FROM lsif_indexes
        WHERE
            state IN ('queued', 'errored')
            AND enqueuer_user_id = 0
            AND queued_at BETWEEN
                (SELECT newest - lookback_window FROM newest_queued) AND
                (SELECT newest FROM newest_queued)
        ORDER BY repository_id, queued_at DESC, id
    ),
    potentially_starving AS NOT MATERIALIZED (
        -- select the newest index per repo outside the window without a completed index
        -- wwithin a grace period
        SELECT DISTINCT ON (repository_id) *
        FROM lsif_indexes l1
        WHERE
            state IN ('queued', 'errored')
            AND enqueuer_user_id = 0
            AND queued_at < (SELECT newest - lookback_window FROM newest_queued)
            AND NOT EXISTS (
                SELECT 1
                FROM lsif_indexes
                WHERE
                    l1.repository_id = repository_id
                    AND state = 'completed'
                    AND finished_at >= ((SELECT newest - lookback_window FROM newest_queued) - cooldown)
            )
        ORDER BY repository_id, queued_at DESC, id
    ),
    manually_enqueued AS MATERIALIZED (
        -- select indexes that were manually enqueued that should be at the
        -- top of the queue
        SELECT *
        FROM lsif_indexes
        WHERE
            enqueuer_user_id > 0
            AND state IN ('queued', 'errored')
    ),
    final_candidates AS NOT MATERIALIZED (
        -- merge them all together, with a priority marker for ordering
        SELECT *, 0 AS priority
        FROM manually_enqueued
        UNION ALL
        (
            SELECT *, 1 AS priority
            FROM potentially_starving
            WHERE id NOT IN (
                SELECT id
                FROM manually_enqueued
            )
            UNION ALL
            (
                SELECT *, 2 AS priority
                FROM newest_in_window
                WHERE id NOT IN (
                    SELECT id
                    FROM manually_enqueued
                )
            )
        )
    )
    SELECT u.id,
        u.commit,
        u.queued_at,
        u.state,
        u.failure_message,
        u.started_at,
        u.finished_at,
        u.repository_id,
        u.process_after,
        u.num_resets,
        u.num_failures,
        u.docker_steps,
        u.root,
        u.indexer,
        u.indexer_args,
        u.outfile,
        u.log_contents,
        u.execution_logs,
        u.local_steps,
        u.should_reindex,
        u.requested_envvars,
        r.name AS repository_name,
        u.enqueuer_user_id
    FROM final_candidates u
    JOIN repo r
    ON r.id = u.repository_id
    WHERE (r.deleted_at IS NULL)
    ORDER BY u.priority ASC
$$ LANGUAGE SQL STABLE STRICT;
