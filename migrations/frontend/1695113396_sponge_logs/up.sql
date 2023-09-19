CREATE TABLE IF NOT EXISTS sponge_log_interpreters (id SERIAL PRIMARY KEY, name TEXT NOT NULL);

COMMENT ON TABLE sponge_log_interpreters IS 'References to pieces of UI logic that can be used to display test logs in a way that is easier to interpret.';

CREATE TABLE IF NOT EXISTS sponge_logs (
    id UUID PRIMARY KEY,
    log TEXT NOT NULL,
    interpreter_id INTEGER NULL REFERENCES sponge_log_interpreters(id) ON DELETE
    SET NULL
);

COMMENT ON TABLE sponge_logs IS 'Logs from tests that are uploaded and stored within a Sourcegraph instance, so that it is easy to share, compare, and collaborate on.';