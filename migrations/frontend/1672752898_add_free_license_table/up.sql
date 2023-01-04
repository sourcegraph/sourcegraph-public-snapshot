CREATE TABLE IF NOT EXISTS free_license (
    id uuid NOT NULL,
    license_key text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    license_version integer,
    license_tags text[],
    license_user_count integer,
    license_expires_at timestamp with time zone
);
