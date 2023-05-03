CREATE TABLE IF NOT EXISTS repo_life_cycle (
    repo_id integer NOT NULL,
    logs jsonb NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS repo_life_cycle_repo_id_unique ON repo_life_cycle USING btree (repo_id);
