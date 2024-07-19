CREATE TABLE IF NOT EXISTS prompts (
    id serial PRIMARY KEY,

    name citext NOT NULL CONSTRAINT prompts_name_max_length CHECK (char_length(name::text) <= 255) CONSTRAINT prompts_name_valid_chars CHECK (name ~ '^[a-zA-Z0-9](?:[a-zA-Z0-9]|[-.](?=[a-zA-Z0-9]))*-?$'::citext),
    description text NOT NULL CONSTRAINT prompts_description_max_length CHECK (char_length(description) <= 1024*50),
    definition_text text NOT NULL CONSTRAINT prompts_definition_text_max_length CHECK (char_length(definition_text) <= 1024*100),
    draft boolean NOT NULL DEFAULT false,
    visibility_secret boolean NOT NULL DEFAULT true,

    owner_user_id integer REFERENCES users(id) ON DELETE CASCADE,
    owner_org_id integer REFERENCES orgs(id) ON DELETE CASCADE,

    created_by integer NULL REFERENCES users (id) ON DELETE SET NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_by integer NULL REFERENCES users (id) ON DELETE SET NULL,
    updated_at timestamp with time zone NOT NULL DEFAULT now(),

    CONSTRAINT prompts_has_valid_owner CHECK ((((owner_user_id IS NOT NULL) AND (owner_org_id IS NULL)) OR ((owner_org_id IS NOT NULL) AND (owner_user_id IS NULL))))
);

CREATE UNIQUE INDEX IF NOT EXISTS prompts_name_is_unique_in_owner_user ON prompts (owner_user_id, name) WHERE owner_user_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS prompts_name_is_unique_in_owner_org ON prompts (owner_org_id, name) WHERE owner_org_id IS NOT NULL;

CREATE OR REPLACE VIEW prompts_view AS
    SELECT
        prompts.*,
        COALESCE(users.username, orgs.name) || '/' || prompts.name AS name_with_owner
        FROM prompts
        LEFT JOIN users ON users.id = prompts.owner_user_id
        LEFT JOIN orgs ON orgs.id = prompts.owner_org_id;
