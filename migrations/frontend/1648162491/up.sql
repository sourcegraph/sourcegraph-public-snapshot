CREATE TABLE IF NOT EXISTS code_monitors_batch_changes (
    id SERIAL NOT NULL,
    monitor bigint NOT NULL,
    batch_change_id integer NOT NULL,
    enabled boolean NOT NULL DEFAULT false,
    created_by integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    changed_by integer NOT NULL,
    changed_at timestamp with time zone DEFAULT now() NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (monitor) REFERENCES cm_monitors(id) ON DELETE CASCADE,
    FOREIGN KEY (batch_change_id) REFERENCES batch_changes(id)
);

ALTER TABLE batch_specs ADD COLUMN IF NOT EXISTS auto_apply boolean NOT NULL DEFAULT false;
ALTER TABLE batch_specs ADD COLUMN IF NOT EXISTS auto_execute boolean NOT NULL DEFAULT false;

ALTER TABLE cm_action_jobs ADD COLUMN IF NOT EXISTS batch_change integer;

ALTER TABLE cm_action_jobs DROP CONSTRAINT cm_action_jobs_only_one_action_type;
ALTER TABLE cm_action_jobs ADD CONSTRAINT cm_action_jobs_only_one_action_type CHECK (((((
CASE
    WHEN (email IS NULL) THEN 0
    ELSE 1
END +
CASE
    WHEN (webhook IS NULL) THEN 0
    ELSE 1
END) +
CASE
    WHEN (slack_webhook IS NULL) THEN 0
    ELSE 1
END) +
CASE
    WHEN (batch_change IS NULL) THEN 0
    ELSE 1
END) = 1));
