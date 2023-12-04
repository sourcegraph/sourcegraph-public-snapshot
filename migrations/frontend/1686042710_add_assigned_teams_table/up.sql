CREATE TABLE IF NOT EXISTS assigned_teams
(
    id                   SERIAL PRIMARY KEY,
    owner_team_id        INTEGER   NOT NULL REFERENCES teams (id) ON DELETE CASCADE DEFERRABLE,
    file_path_id         INTEGER   NOT NULL REFERENCES repo_paths (id),
    who_assigned_team_id INTEGER   NULL REFERENCES users (id) ON DELETE SET NULL DEFERRABLE,
    assigned_at          TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS assigned_teams_file_path_owner
    ON assigned_teams
        USING btree (file_path_id, owner_team_id);

COMMENT ON TABLE assigned_teams
    IS 'Table for team ownership assignments, one entry contains an assigned team ID, which repo_path is assigned and the date and user who assigned the owner team.';
