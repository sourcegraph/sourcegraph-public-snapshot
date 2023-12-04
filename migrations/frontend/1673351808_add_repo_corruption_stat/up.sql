ALTER TABLE gitserver_repos_statistics
  ADD COLUMN IF NOT EXISTS corrupted bigint NOT NULL DEFAULT 0;
COMMENT ON COLUMN gitserver_repos_statistics.corrupted IS 'Number of repositories that are NOT soft-deleted and not blocked and have corrupted_at set in gitserver_repos table';

ALTER TABLE repo_statistics
  ADD COLUMN IF NOT EXISTS corrupted bigint NOT NULL DEFAULT 0;
COMMENT ON COLUMN repo_statistics.corrupted IS 'Number of repositories that are NOT soft-deleted and not blocked and have corrupted_at set in gitserver_repos table';

--- repo UPDATE trigger
CREATE OR REPLACE FUNCTION recalc_repo_statistics_on_repo_update() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
      -- Insert diff of changes
      WITH diff(total, soft_deleted, not_cloned, cloning, cloned, failed_fetch, corrupted) AS (
        VALUES (
          (SELECT COUNT(*) FROM newtab WHERE deleted_at IS NULL     AND blocked IS NULL) - (SELECT COUNT(*) FROM oldtab WHERE deleted_at IS NULL     AND blocked IS NULL),
          (SELECT COUNT(*) FROM newtab WHERE deleted_at IS NOT NULL AND blocked IS NULL) - (SELECT COUNT(*) FROM oldtab WHERE deleted_at IS NOT NULL AND blocked IS NULL),
          (
            (SELECT COUNT(*) FROM newtab JOIN gitserver_repos gr ON gr.repo_id = newtab.id WHERE newtab.deleted_at is NULL AND newtab.blocked IS NULL AND gr.clone_status = 'not_cloned')
            -
            (SELECT COUNT(*) FROM oldtab JOIN gitserver_repos gr ON gr.repo_id = oldtab.id WHERE oldtab.deleted_at is NULL AND oldtab.blocked IS NULL AND gr.clone_status = 'not_cloned')
          ),
          (
            (SELECT COUNT(*) FROM newtab JOIN gitserver_repos gr ON gr.repo_id = newtab.id WHERE newtab.deleted_at is NULL AND newtab.blocked IS NULL AND gr.clone_status = 'cloning')
            -
            (SELECT COUNT(*) FROM oldtab JOIN gitserver_repos gr ON gr.repo_id = oldtab.id WHERE oldtab.deleted_at is NULL AND oldtab.blocked IS NULL AND gr.clone_status = 'cloning')
          ),
          (
            (SELECT COUNT(*) FROM newtab JOIN gitserver_repos gr ON gr.repo_id = newtab.id WHERE newtab.deleted_at is NULL AND newtab.blocked IS NULL AND gr.clone_status = 'cloned')
            -
            (SELECT COUNT(*) FROM oldtab JOIN gitserver_repos gr ON gr.repo_id = oldtab.id WHERE oldtab.deleted_at is NULL AND oldtab.blocked IS NULL AND gr.clone_status = 'cloned')
          ),
          (
            (SELECT COUNT(*) FROM newtab JOIN gitserver_repos gr ON gr.repo_id = newtab.id WHERE newtab.deleted_at is NULL AND newtab.blocked IS NULL AND gr.last_error IS NOT NULL)
            -
            (SELECT COUNT(*) FROM oldtab JOIN gitserver_repos gr ON gr.repo_id = oldtab.id WHERE oldtab.deleted_at is NULL AND oldtab.blocked IS NULL AND gr.last_error IS NOT NULL)
          ),
          (
            (SELECT COUNT(*) FROM newtab JOIN gitserver_repos gr ON gr.repo_id = newtab.id WHERE newtab.deleted_at is NULL AND newtab.blocked IS NULL AND gr.corrupted_at IS NOT NULL)
            -
            (SELECT COUNT(*) FROM oldtab JOIN gitserver_repos gr ON gr.repo_id = oldtab.id WHERE oldtab.deleted_at is NULL AND oldtab.blocked IS NULL AND gr.corrupted_at IS NOT NULL)
          )
        )
      )
      INSERT INTO
        repo_statistics (total, soft_deleted, not_cloned, cloning, cloned, failed_fetch, corrupted)
      SELECT total, soft_deleted, not_cloned, cloning, cloned, failed_fetch, corrupted
      FROM diff
      WHERE
           total != 0
        OR soft_deleted != 0
        OR not_cloned != 0
        OR cloning != 0
        OR cloned != 0
        OR failed_fetch != 0
        OR corrupted != 0
      ;
      RETURN NULL;
  END
$$;

CREATE OR REPLACE FUNCTION recalc_gitserver_repos_statistics_on_update() RETURNS trigger
    LANGUAGE plpgsql
    -------------------------------------------------
    -- IMPORTANT: THIS IS CHANGED TO INCLUDE `corrupted`
    -------------------------------------------------
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

      -------------------------------------------------
      -- IMPORTANT: THIS IS CHANGED TO INCLUDE `corrupted`
      -------------------------------------------------
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

      -------------------------------------------------
      -- IMPORTANT: THIS IS CHANGED TO INCLUDE `corrupted`
      -------------------------------------------------
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
