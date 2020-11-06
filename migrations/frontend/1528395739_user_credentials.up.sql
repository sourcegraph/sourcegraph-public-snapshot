BEGIN;

CREATE TABLE IF NOT EXISTS user_credentials (
    id BIGSERIAL PRIMARY KEY,
    domain TEXT NOT NULL,
    user_id INTEGER NOT NULL,
    external_service_type TEXT NOT NULL,
    external_service_id TEXT NOT NULL,
    credential TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- deleted_at is intentionally left out: given the secret nature of tokens,
    -- users will likely want to be certain that tokens are deleted when
    -- removed, even with encryption in place.

    -- Set up the foreign key.
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE DEFERRABLE,

    -- Set up a unique constraint across the fields that are, in fact, unique.
    UNIQUE (domain, user_id, external_service_type, external_service_id)
);

COMMIT;
