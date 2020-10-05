BEGIN;

CREATE TABLE graphs (
    id bigserial PRIMARY KEY,
    owner_user_id integer REFERENCES users(id) ON DELETE CASCADE DEFERRABLE,
    owner_org_id integer REFERENCES orgs(id) ON DELETE CASCADE DEFERRABLE,
    name text NOT NULL,
    description text,
    spec text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT graphs_has_1_owner CHECK ((owner_user_id IS NULL) <> (owner_org_id IS NULL)),
    CONSTRAINT graphs_name_valid_chars CHECK (name ~ '^[a-zA-Z0-9](?:[a-zA-Z0-9]|[-.](?=[a-zA-Z0-9]))*-?$'),
    CONSTRAINT graphs_name_max_length CHECK (char_length(name) <= 255),
    CONSTRAINT graphs_name_not_blank CHECK (name <> '')
);

COMMIT;
