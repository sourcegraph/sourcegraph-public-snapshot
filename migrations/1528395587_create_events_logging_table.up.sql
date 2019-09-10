BEGIN;

CREATE TABLE event_logs (
    id                  BIGSERIAL NOT NULL PRIMARY KEY,
    name                TEXT NOT NULL,
    argument            TEXT NOT NULL,
    url                 TEXT NOT NULL,
    user_id             INTEGER NOT NULL,
    anonymous_user_id   TEXT NOT NULL,
	timestamp           TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE INDEX event_logs_name ON event_logs(name);
CREATE INDEX event_logs_user_id ON event_logs(user_id);
CREATE INDEX event_logs_timestamp ON event_logs(timestamp);

COMMIT;
