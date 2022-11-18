CREATE TABLE IF NOT EXISTS search_context_stars (
    search_context_id bigint NOT NULL,
    user_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

ALTER TABLE ONLY search_context_stars
    ADD CONSTRAINT search_context_stars_pkey PRIMARY KEY (search_context_id, user_id);

COMMENT ON TABLE search_context_stars IS 'When a user stars a search context, a row is inserted into this table. If the user unstars the search context, the row is deleted.';

CREATE TABLE IF NOT EXISTS search_context_default (
    user_id integer NOT NULL,
    search_context_id bigint NOT NULL
);

ALTER TABLE ONLY search_context_default
    ADD CONSTRAINT search_context_default_pkey PRIMARY KEY (user_id);

COMMENT ON TABLE search_context_default IS 'When a user sets a search context as default, a row is inserted into this table. A user can only have one default search context. If the user has not set their default search context, it will fall back to `global`.';
