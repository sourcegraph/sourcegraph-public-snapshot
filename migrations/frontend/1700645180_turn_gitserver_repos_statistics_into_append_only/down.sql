-- Drop
DROP TABLE IF EXISTS gitserver_repos_statistics;

-- Recreate the table
CREATE TABLE IF NOT EXISTS gitserver_repos_statistics (
  shard_id text PRIMARY KEY,

  total        BIGINT NOT NULL DEFAULT 0,
  not_cloned   BIGINT NOT NULL DEFAULT 0,
  cloning      BIGINT NOT NULL DEFAULT 0,
  cloned       BIGINT NOT NULL DEFAULT 0,
  failed_fetch BIGINT NOT NULL DEFAULT 0,
  corrupted    BIGINT NOT NULL DEFAULT 0
);

COMMENT ON COLUMN gitserver_repos_statistics.shard_id IS 'ID of this gitserver shard. If an empty string then the repositories havent been assigned a shard.';
COMMENT ON COLUMN gitserver_repos_statistics.total IS 'Number of repositories in gitserver_repos table on this shard';
COMMENT ON COLUMN gitserver_repos_statistics.not_cloned IS 'Number of repositories in gitserver_repos table on this shard that are not cloned yet';
COMMENT ON COLUMN gitserver_repos_statistics.cloning IS 'Number of repositories in gitserver_repos table on this shard that cloning';
COMMENT ON COLUMN gitserver_repos_statistics.cloned IS 'Number of repositories in gitserver_repos table on this shard that are cloned';
COMMENT ON COLUMN gitserver_repos_statistics.failed_fetch IS 'Number of repositories in gitserver_repos table on this shard where last_error is set';
COMMENT ON COLUMN gitserver_repos_statistics.corrupted IS 'Number of repositories that are NOT soft-deleted and not blocked and have corrupted_at set in gitserver_repos table';

-- Insert initial values into gitserver_repos_statistics
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
  total        = gitserver_repos_statistics.total        + excluded.total,
  not_cloned   = gitserver_repos_statistics.not_cloned   + excluded.not_cloned,
  cloning      = gitserver_repos_statistics.cloning      + excluded.cloning,
  cloned       = gitserver_repos_statistics.cloned       + excluded.cloned,
  failed_fetch = gitserver_repos_statistics.failed_fetch + excluded.failed_fetch
;

-- Now restore triggers

-- UPDATE
CREATE OR REPLACE FUNCTION recalc_gitserver_repos_statistics_on_update() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
      INSERT INTO gitserver_repos_statistics AS grs (shard_id, total, not_cloned, cloning, cloned, failed_fetch, corrupted)
      SELECT
        newtab.shard_id AS shard_id,
        COUNT(*) AS total,
        COUNT(*) FILTER(WHERE clone_status = 'not_cloned')  AS not_cloned,
        COUNT(*) FILTER(WHERE clone_status = 'cloning') AS cloning,
        COUNT(*) FILTER(WHERE clone_status = 'cloned') AS cloned,
        COUNT(*) FILTER(WHERE last_error IS NOT NULL) AS failed_fetch,
        COUNT(*) FILTER(WHERE corrupted_at IS NOT NULL) AS corrupted
      FROM
        newtab
      GROUP BY newtab.shard_id
      ON CONFLICT(shard_id) DO
      UPDATE
      SET
        total        = grs.total        + (excluded.total        - (SELECT COUNT(*)                                              FROM oldtab ot WHERE ot.shard_id = excluded.shard_id)),
        not_cloned   = grs.not_cloned   + (excluded.not_cloned   - (SELECT COUNT(*) FILTER(WHERE ot.clone_status = 'not_cloned') FROM oldtab ot WHERE ot.shard_id = excluded.shard_id)),
        cloning      = grs.cloning      + (excluded.cloning      - (SELECT COUNT(*) FILTER(WHERE ot.clone_status = 'cloning')    FROM oldtab ot WHERE ot.shard_id = excluded.shard_id)),
        cloned       = grs.cloned       + (excluded.cloned       - (SELECT COUNT(*) FILTER(WHERE ot.clone_status = 'cloned')     FROM oldtab ot WHERE ot.shard_id = excluded.shard_id)),
        failed_fetch = grs.failed_fetch + (excluded.failed_fetch - (SELECT COUNT(*) FILTER(WHERE ot.last_error IS NOT NULL)      FROM oldtab ot WHERE ot.shard_id = excluded.shard_id)),
        corrupted    = grs.corrupted    + (excluded.corrupted    - (SELECT COUNT(*) FILTER(WHERE ot.corrupted_at IS NOT NULL)    FROM oldtab ot WHERE ot.shard_id = excluded.shard_id))
      ;

      WITH moved AS (
        SELECT
          oldtab.shard_id AS shard_id,
          COUNT(*) AS total,
          COUNT(*) FILTER(WHERE oldtab.clone_status = 'not_cloned')  AS not_cloned,
          COUNT(*) FILTER(WHERE oldtab.clone_status = 'cloning') AS cloning,
          COUNT(*) FILTER(WHERE oldtab.clone_status = 'cloned') AS cloned,
          COUNT(*) FILTER(WHERE oldtab.last_error IS NOT NULL) AS failed_fetch,
          COUNT(*) FILTER(WHERE oldtab.corrupted_at IS NOT NULL) AS corrupted
        FROM
          oldtab
        JOIN newtab ON newtab.repo_id = oldtab.repo_id
        WHERE
          oldtab.shard_id != newtab.shard_id
        GROUP BY oldtab.shard_id
      )
      UPDATE gitserver_repos_statistics grs
      SET
        total        = grs.total        - moved.total,
        not_cloned   = grs.not_cloned   - moved.not_cloned,
        cloning      = grs.cloning      - moved.cloning,
        cloned       = grs.cloned       - moved.cloned,
        failed_fetch = grs.failed_fetch - moved.failed_fetch,
        corrupted    = grs.corrupted    - moved.corrupted
      FROM moved
      WHERE moved.shard_id = grs.shard_id;

      WITH diff(not_cloned, cloning, cloned, failed_fetch, corrupted) AS (
        VALUES (
          (
            (SELECT COUNT(*) FROM newtab JOIN repo r ON newtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND newtab.clone_status = 'not_cloned')
            -
            (SELECT COUNT(*) FROM oldtab JOIN repo r ON oldtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND oldtab.clone_status = 'not_cloned')
          ),
          (
            (SELECT COUNT(*) FROM newtab JOIN repo r ON newtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND newtab.clone_status = 'cloning')
            -
            (SELECT COUNT(*) FROM oldtab JOIN repo r ON oldtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND oldtab.clone_status = 'cloning')
          ),
          (
            (SELECT COUNT(*) FROM newtab JOIN repo r ON newtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND newtab.clone_status = 'cloned')
            -
            (SELECT COUNT(*) FROM oldtab JOIN repo r ON oldtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND oldtab.clone_status = 'cloned')
          ),
          (
            (SELECT COUNT(*) FROM newtab JOIN repo r ON newtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND newtab.last_error IS NOT NULL)
            -
            (SELECT COUNT(*) FROM oldtab JOIN repo r ON oldtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND oldtab.last_error IS NOT NULL)
          ),
          (
            (SELECT COUNT(*) FROM newtab JOIN repo r ON newtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND newtab.corrupted_at IS NOT NULL)
            -
            (SELECT COUNT(*) FROM oldtab JOIN repo r ON oldtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND oldtab.corrupted_at IS NOT NULL)
          )
        )
      )
      INSERT INTO repo_statistics (not_cloned, cloning, cloned, failed_fetch, corrupted)
      SELECT not_cloned, cloning, cloned, failed_fetch, corrupted
      FROM diff
      WHERE
           not_cloned != 0
        OR cloning != 0
        OR cloned != 0
        OR failed_fetch != 0
        OR corrupted != 0
      ;

      RETURN NULL;
  END
$$;
DROP TRIGGER IF EXISTS trig_recalc_gitserver_repos_statistics_on_update ON gitserver_repos;
CREATE TRIGGER trig_recalc_gitserver_repos_statistics_on_update
AFTER UPDATE ON gitserver_repos
REFERENCING OLD TABLE AS oldtab NEW TABLE AS newtab
FOR EACH STATEMENT EXECUTE FUNCTION recalc_gitserver_repos_statistics_on_update();

-- INSERT
CREATE OR REPLACE FUNCTION recalc_gitserver_repos_statistics_on_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
      INSERT INTO gitserver_repos_statistics AS grs (shard_id, total, not_cloned, cloning, cloned, failed_fetch)
      SELECT
        shard_id,
        COUNT(*) AS total,
        COUNT(*) FILTER(WHERE clone_status = 'not_cloned') AS not_cloned,
        COUNT(*) FILTER(WHERE clone_status = 'cloning') AS cloning,
        COUNT(*) FILTER(WHERE clone_status = 'cloned') AS cloned,
        COUNT(*) FILTER(WHERE last_error IS NOT NULL) AS failed_fetch
      FROM
        newtab
      GROUP BY shard_id
      ON CONFLICT(shard_id)
      DO UPDATE
      SET
        total        = grs.total        + excluded.total,
        not_cloned   = grs.not_cloned   + excluded.not_cloned,
        cloning      = grs.cloning      + excluded.cloning,
        cloned       = grs.cloned       + excluded.cloned,
        failed_fetch = grs.failed_fetch + excluded.failed_fetch
      ;

      RETURN NULL;
  END
$$;
DROP TRIGGER IF EXISTS trig_recalc_gitserver_repos_statistics_on_insert ON gitserver_repos;
CREATE TRIGGER trig_recalc_gitserver_repos_statistics_on_insert
AFTER INSERT ON gitserver_repos
REFERENCING NEW TABLE AS newtab
FOR EACH STATEMENT EXECUTE FUNCTION recalc_gitserver_repos_statistics_on_insert();

-- DELETE
CREATE OR REPLACE FUNCTION recalc_gitserver_repos_statistics_on_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
      UPDATE gitserver_repos_statistics grs
      SET
        total        = grs.total      - (SELECT COUNT(*)                                           FROM oldtab WHERE oldtab.shard_id = grs.shard_id),
        not_cloned   = grs.not_cloned - (SELECT COUNT(*) FILTER(WHERE clone_status = 'not_cloned') FROM oldtab WHERE oldtab.shard_id = grs.shard_id),
        cloning      = grs.cloning    - (SELECT COUNT(*) FILTER(WHERE clone_status = 'cloning')    FROM oldtab WHERE oldtab.shard_id = grs.shard_id),
        cloned       = grs.cloned     - (SELECT COUNT(*) FILTER(WHERE clone_status = 'cloned')     FROM oldtab WHERE oldtab.shard_id = grs.shard_id),
        failed_fetch = grs.cloned     - (SELECT COUNT(*) FILTER(WHERE last_error IS NOT NULL)      FROM oldtab WHERE oldtab.shard_id = grs.shard_id)
      ;

      RETURN NULL;
  END
$$;
DROP TRIGGER IF EXISTS trig_recalc_gitserver_repos_statistics_on_delete ON gitserver_repos;
CREATE TRIGGER trig_recalc_gitserver_repos_statistics_on_delete
AFTER DELETE ON gitserver_repos REFERENCING
OLD TABLE AS oldtab
FOR EACH STATEMENT EXECUTE FUNCTION recalc_gitserver_repos_statistics_on_delete();
