CREATE TABLE IF NOT EXISTS code_monitors_batch_changes (
    id SERIAL NOT NULL,
    code_monitor_id integer NOT NULL,
    batch_change_id integer NOT NULL,
    user_author_id integer NOT NULL,
    enabled boolean NOT NULL DEFAULT false,
    PRIMARY KEY (id),
    FOREIGN KEY (code_monitor_id) REFERENCES cm_monitors(id) ON DELETE CASCADE,
    FOREIGN KEY (batch_change_id) REFERENCES batch_changes(id),
    FOREIGN KEY (user_author_id) REFERENCES users(id)
);

ALTER TABLE batch_specs ADD COLUMN IF NOT EXISTS auto_apply boolean NOT NULL DEFAULT false;

ALTER TABLE cm_action_jobs ADD COLUMN IF NOT EXISTS batch_change integer;
