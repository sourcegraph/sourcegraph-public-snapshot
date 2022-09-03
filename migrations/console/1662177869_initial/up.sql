CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS console_users (
    id                  SERIAL PRIMARY KEY,
    email               CITEXT NOT NULL
)
