BEGIN;

CREATE TABLE IF NOT EXISTS search_contexts (
    id BIGSERIAL PRIMARY KEY,
    name citext NOT NULL,
    description text NOT NULL,
    public boolean NOT NULL,
    namespace_user_id integer,
    namespace_org_id integer,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,

    CONSTRAINT search_contexts_has_one_or_no_namespace CHECK (((namespace_user_id IS NULL) OR (namespace_org_id IS NULL))),

    CONSTRAINT search_contexts_namespace_user_id_fk
        FOREIGN KEY (namespace_user_id)
            REFERENCES users (id)
            ON DELETE CASCADE,

    CONSTRAINT search_contexts_namespace_org_id_fk
        FOREIGN KEY (namespace_org_id)
            REFERENCES orgs (id)
            ON DELETE CASCADE
);

CREATE UNIQUE INDEX search_contexts_name_namespace_user_id_unique
    ON search_contexts (name, namespace_user_id)
    WHERE namespace_user_id IS NOT NULL;

CREATE UNIQUE INDEX search_contexts_name_namespace_org_id_unique
    ON search_contexts (name, namespace_org_id)
    WHERE namespace_org_id IS NOT NULL;

CREATE UNIQUE INDEX search_contexts_name_without_namespace_unique
    ON search_contexts (name)
    WHERE namespace_user_id IS NULL AND namespace_org_id IS NULL;

CREATE TABLE IF NOT EXISTS search_context_repos (
    search_context_id bigint NOT NULL,
    repo_id integer NOT NULL,
    revision text NOT NULL,

    CONSTRAINT search_context_repos_search_context_id_fk
        FOREIGN KEY (search_context_id)
            REFERENCES search_contexts (id)
            ON DELETE CASCADE,

    CONSTRAINT search_context_repos_repo_id_fk
        FOREIGN KEY (repo_id)
            REFERENCES repo (id)
            ON DELETE CASCADE,

    CONSTRAINT search_context_repos_search_context_id_repo_id_revision_unique UNIQUE (search_context_id, repo_id, revision)
);

COMMIT;
