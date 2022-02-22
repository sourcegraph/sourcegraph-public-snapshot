CREATE TABLE IF NOT EXISTS notebooks (
    id BIGSERIAL PRIMARY KEY,
    title CITEXT NOT NULL,
    blocks JSONB DEFAULT '[]'::JSONB NOT NULL,
    public BOOLEAN NOT NULL,
    creator_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL DEFERRABLE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT blocks_is_array CHECK (jsonb_typeof(blocks) = 'array')
);
