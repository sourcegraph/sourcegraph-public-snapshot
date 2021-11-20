BEGIN;

CREATE TABLE IF NOT EXISTS batch_changes_lifecycle_hooks (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NULL,
    url TEXT NOT NULL,
    secret TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS
    batch_changes_lifecycle_hooks_expires_at_idx
ON
    batch_changes_lifecycle_hooks (expires_at);

COMMIT;
