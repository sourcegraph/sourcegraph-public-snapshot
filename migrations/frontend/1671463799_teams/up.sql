CREATE TABLE IF NOT EXISTS teams (
    id SERIAL PRIMARY KEY,
    name citext NOT NULL CHECK (char_length(name::text) <= 255) CHECK (name ~ '^[a-zA-Z0-9](?:[a-zA-Z0-9]|[-.](?=[a-zA-Z0-9]))*-?$'::citext), -- this is copied from user and org, let's see if we need this.
    display_name text CHECK (char_length(display_name) <= 255),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    creator_id integer NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    readonly boolean NOT NULL DEFAULT false,
    parent_team integer REFERENCES teams(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS teams_name ON teams(name);

CREATE TABLE IF NOT EXISTS team_members (
    id SERIAL PRIMARY KEY,
    team_id integer NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    user_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT team_members_team_id_user_id_key UNIQUE (team_id, user_id)
);
