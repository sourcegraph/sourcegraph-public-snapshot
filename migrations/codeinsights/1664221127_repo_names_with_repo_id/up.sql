ALTER TABLE IF EXISTS repo_names
    ADD COLUMN IF NOT EXISTS repo_id INT NOT NULL DEFAULT 0;

DROP INDEX IF EXISTS repo_names_name_unique_idx;

CREATE UNIQUE INDEX IF NOT EXISTS repo_names_repo_id_name_unique_idx ON repo_names (repo_id, name);

UPDATE repo_names rn
SET repo_id = sub.repo_id
FROM (SELECT sp.repo_id, rni.id AS repo_name_id
      FROM series_points sp
               JOIN repo_names rni ON sp.repo_name_id = rni.id) AS sub
WHERE sub.repo_name_id = rn.id
  AND rn.repo_id = 0;
