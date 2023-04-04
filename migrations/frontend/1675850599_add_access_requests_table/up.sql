CREATE TABLE IF NOT EXISTS access_requests (
    id SERIAL NOT NULL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    additional_info TEXT,
    status TEXT NOT NULL
);
