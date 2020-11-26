BEGIN;

DROP VIEW lsif_indexes_with_repository_name;

-- Put it back how it was :(
CREATE VIEW lsif_indexes_with_repository_name AS SELECT u.id,
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
    r.name AS repository_name
FROM lsif_indexes u
JOIN repo r ON r.id = u.repository_id
WHERE r.deleted_at IS NULL;

COMMIT;
