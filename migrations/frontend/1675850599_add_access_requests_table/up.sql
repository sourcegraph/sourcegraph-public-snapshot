CREATE TYPE
    access_request_status AS ENUM (
        'pending',
        'approved',
        'rejected'
    );

CREATE TABLE
    IF NOT EXISTS access_requests (
        id SERIAL NOT NULL PRIMARY KEY,
        name TEXT NOT NULL,
        email TEXT NOT NULL UNIQUE,
        additional_info TEXT,
        status access_request_status NOT NULL DEFAULT 'pending',
        requests_count INTEGER NOT NULL DEFAULT 1,
        created_at TIMESTAMP
        WITH
            TIME ZONE NOT NULL DEFAULT now(),
            updated_at TIMESTAMP
        WITH
            TIME ZONE NOT NULL DEFAULT now(),
            -- TODO: For what cases we might need this?
            deleted_at TIMESTAMP
        WITH TIME ZONE
    );
