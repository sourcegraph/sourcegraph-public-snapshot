CREATE TABLE IF NOT EXISTS own_signal_configurations
(
    id                     SERIAL PRIMARY KEY,
    name                   TEXT    NOT NULL,
    description            TEXT    NOT NULL DEFAULT '',
    excluded_repo_patterns TEXT[]  NULL,
    enabled                BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE UNIQUE INDEX IF NOT EXISTS own_signal_configurations_name_uidx ON own_signal_configurations(name);

INSERT INTO own_signal_configurations (id, name, enabled, description)
VALUES (1, 'recent-contributors', FALSE, 'Indexes contributors in each file using repository history.')
ON CONFLICT DO NOTHING;
INSERT INTO own_signal_configurations (id, name, enabled, description)
VALUES (2, 'recent-views', FALSE, 'Indexes users that recently viewed files in Sourcegraph.')
ON CONFLICT DO NOTHING;

