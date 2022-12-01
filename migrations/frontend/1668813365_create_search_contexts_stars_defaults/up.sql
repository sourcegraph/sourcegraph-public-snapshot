CREATE TABLE IF NOT EXISTS search_context_stars (
    search_context_id bigint REFERENCES search_contexts(id) ON DELETE CASCADE DEFERRABLE,
    user_id integer REFERENCES users(id) ON DELETE CASCADE DEFERRABLE,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    PRIMARY KEY (search_context_id, user_id)
);

COMMENT ON TABLE search_context_stars IS 'When a user stars a search context, a row is inserted into this table. If the user unstars the search context, the row is deleted. The global context is not in the database, and therefore cannot be starred.';

CREATE TABLE IF NOT EXISTS search_context_default (
    user_id integer PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE DEFERRABLE,
    search_context_id bigint NOT NULL REFERENCES search_contexts(id) ON DELETE CASCADE DEFERRABLE
);

COMMENT ON TABLE search_context_default IS 'When a user sets a search context as default, a row is inserted into this table. A user can only have one default search context. If the user has not set their default search context, it will fall back to `global`.';
