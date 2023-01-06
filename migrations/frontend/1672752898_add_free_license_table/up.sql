CREATE TABLE IF NOT EXISTS free_license (
    id int PRIMARY KEY DEFAULT 1,
    license_key text NOT NULL,
    license_version int NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT single_entry CHECK (id = 1)
);
