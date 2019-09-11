BEGIN;

CREATE TABLE event_logs (
    id                  BIGSERIAL NOT NULL PRIMARY KEY,
    name                TEXT NOT NULL,
    url                 TEXT NOT NULL,
    user_id             INTEGER NOT NULL,
    anonymous_user_id   TEXT NOT NULL,
    argument            TEXT NOT NULL,
    version             TEXT NOT NULL,
    timestamp           TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

ALTER TABLE event_logs ADD CONSTRAINT event_logs_check_name_not_empty CHECK (name != '');
ALTER TABLE event_logs ADD CONSTRAINT event_logs_check_has_user CHECK ((user_id = 0 AND anonymous_user_id != '') OR (user_id != 0 AND anonymous_user_id = '') OR (user_id != 0 AND anonymous_user_id != ''));

CREATE INDEX event_logs_name ON event_logs(name);
CREATE INDEX event_logs_user_id ON event_logs(user_id);
CREATE INDEX event_logs_timestamp ON event_logs(timestamp);

COMMIT;
