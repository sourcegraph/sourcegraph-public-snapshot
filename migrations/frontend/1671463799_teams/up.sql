CREATE TABLE IF NOT EXISTS teams (
    id SERIAL PRIMARY KEY,
    name citext NOT NULL CONSTRAINT teams_name_max_length CHECK (char_length(name::text) <= 255) CONSTRAINT teams_name_valid_chars CHECK (name ~ '^[a-zA-Z0-9](?:[a-zA-Z0-9]|[-.](?=[a-zA-Z0-9]))*-?$'::citext),
    display_name text CONSTRAINT teams_display_name_max_length CHECK (char_length(display_name) <= 255),
    readonly boolean NOT NULL DEFAULT false,
    parent_team_id integer REFERENCES teams(id) ON DELETE CASCADE,
    creator_id integer NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS teams_name ON teams(name);

CREATE TABLE IF NOT EXISTS team_members (
    team_id integer NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT team_members_team_id_user_id_key UNIQUE (team_id, user_id),
    PRIMARY KEY(team_id, user_id)
);

ALTER TABLE names ADD COLUMN IF NOT EXISTS team_id integer REFERENCES teams(id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE names DROP CONSTRAINT IF EXISTS names_check;
ALTER TABLE names ADD CONSTRAINT names_check CHECK (user_id IS NOT NULL OR org_id IS NOT NULL OR team_id IS NOT NULL);
