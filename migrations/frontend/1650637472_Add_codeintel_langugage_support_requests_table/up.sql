CREATE TABLE IF NOT EXISTS codeintel_langugage_support_requests (
    id SERIAL,
    user_id integer NOT NULL,
    language_id text NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS codeintel_langugage_support_requests_user_id_language ON codeintel_langugage_support_requests(user_id, language_id);
