CREATE TABLE IF NOT EXISTS free_license (
    id uuid NOT NULL,
    license_key text NOT NULL,
    license_version int NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);
