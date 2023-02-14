CREATE TYPE
access_request_status AS ENUM (
    'PENDING',
    'APPROVED',
    'REJECTED'
);

CREATE TABLE
    IF NOT EXISTS access_requests (
        id SERIAL NOT NULL PRIMARY KEY,
        created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
        updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
        deleted_at TIMESTAMP WITH TIME ZONE,
        name TEXT NOT NULL,
        email TEXT NOT NULL UNIQUE,
        additional_info TEXT,
        status access_request_status NOT NULL DEFAULT 'PENDING'
    );
