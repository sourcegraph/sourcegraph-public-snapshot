-- In the previous migration (1659085788_add_repo_stats_table) the triggers
-- adding data to the stats table and the initial state of the stats table were
-- setup in the wrong order:
--
--   1. Create repo_statistics/gitserver_repos_statistics tables
--   2. Insert initial state into tables
--   3. Setup database triggers that insert/update these tables
--
-- That's wrong. Instead we should've :
--
--   1. Create repo_statistics/gitserver_repos_statistics tables
--   2. Setup database triggers that insert/update these tables
--   3. Update tables to have correct initial state
--
--
-- What this migration does is correct the initial wrong states by

-- 1. Lock repo and gitserver_repos tables, so that no inserts/updates/deletes happen while we compute new total state.
--    EXCLUSIVE is the mode that says "no one else can write, only read".

LOCK repo IN EXCLUSIVE MODE;            -- first lock repo, since we have triggers that write `repo` and cause updates on `gitserver_repos` but not the other way around
LOCK gitserver_repos IN EXCLUSIVE MODE; -- then lock `gitserver_repos`

--- 2. Delete old state in `repo_statistics` (we can't update the state, since this is an append-only table).
DELETE FROM repo_statistics;

--- 3. Insert new total counts in `repo_statistics`
INSERT INTO repo_statistics (total, soft_deleted, not_cloned, cloning, cloned, failed_fetch)
SELECT
  COUNT(*) AS total,
  (SELECT COUNT(*) FROM repo WHERE deleted_at is NOT NULL AND blocked IS NULL) AS soft_deleted,
  COUNT(*) FILTER(WHERE gitserver_repos.clone_status = 'not_cloned') AS not_cloned,
  COUNT(*) FILTER(WHERE gitserver_repos.clone_status = 'cloning') AS cloning,
  COUNT(*) FILTER(WHERE gitserver_repos.clone_status = 'cloned') AS cloned,
  COUNT(*) FILTER(WHERE gitserver_repos.last_error IS NOT NULL) AS failed_fetch
FROM repo
JOIN gitserver_repos ON gitserver_repos.repo_id = repo.id
WHERE
  repo.deleted_at is NULL AND repo.blocked IS NULL;

--- 4. Insert/update `gitserver_repos_statistics` by updating/insert the recalculated counts per shard.
INSERT INTO
  gitserver_repos_statistics (shard_id, total, not_cloned, cloning, cloned, failed_fetch)
SELECT
  shard_id,
  COUNT(*) AS total,
  COUNT(*) FILTER(WHERE clone_status = 'not_cloned') AS not_cloned,
  COUNT(*) FILTER(WHERE clone_status = 'cloning') AS cloning,
  COUNT(*) FILTER(WHERE clone_status = 'cloned') AS cloned,
  COUNT(*) FILTER(WHERE last_error IS NOT NULL) AS failed_fetch
FROM
  gitserver_repos
GROUP BY shard_id
ON CONFLICT(shard_id)
DO UPDATE
SET
  total        = excluded.total,
  not_cloned   = excluded.not_cloned,
  cloning      = excluded.cloning,
  cloned       = excluded.cloned,
  failed_fetch = excluded.failed_fetch
;
