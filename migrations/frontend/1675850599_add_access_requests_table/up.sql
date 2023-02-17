-- Active: 1674811951294@@127.0.0.1@5432@sourcegraph
DO $$ BEGIN IF NOT EXISTS (
    SELECT 1
    FROM pg_type
    WHERE typname = 'access_request_status'
) THEN -- create type
CREATE TYPE access_request_status AS ENUM ('PENDING', 'APPROVED', 'REJECTED');
END IF;
END $$;

CREATE TABLE IF NOT EXISTS access_requests (
    id SERIAL NOT NULL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    additional_info TEXT,
    status access_request_status NOT NULL DEFAULT 'PENDING'
);
