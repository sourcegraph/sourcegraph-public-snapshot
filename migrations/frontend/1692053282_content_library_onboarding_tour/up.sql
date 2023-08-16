CREATE TABLE IF NOT EXISTS user_onboarding_tour
(
    id         SERIAL PRIMARY KEY NOT NULL,
    raw_json   TEXT               NOT NULL,
    created_at TIMESTAMP          NOT NULL DEFAULT NOW(),
    updated_by INT,
    CONSTRAINT user_onboarding_tour_users_fk FOREIGN KEY(updated_by) REFERENCES users(id)
)
