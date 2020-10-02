BEGIN;

CREATE TABLE IF NOT EXISTS campaign_user_tokens (
    user_id INTEGER NOT NULL,
    external_service_id BIGINT NOT NULL,
    token TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- deleted_at is intentionally left out: given the secret nature of tokens,
    -- users will likely want to be certain that tokens are deleted when
    -- removed, even with encryption in place.

    -- user_id is first because we'll need to be able to query all the tokens
    -- for a user on a regular basis, and PostgreSQL can use the primary key
    -- index for the first column of a composite key individually.
    PRIMARY KEY (user_id, external_service_id),

    -- Set up the foreign keys.
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE DEFERRABLE,
    FOREIGN KEY (external_service_id) REFERENCES external_services (id) ON DELETE CASCADE DEFERRABLE
);

COMMIT;
