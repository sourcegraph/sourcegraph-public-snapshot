BEGIN;

-- Begin cm_webhooks
CREATE TABLE IF NOT EXISTS cm_webhooks (
	id BIGSERIAL PRIMARY KEY,
	monitor BIGINT NOT NULL REFERENCES cm_monitors(id) ON DELETE CASCADE,
	url TEXT NOT NULL,
	enabled BOOLEAN NOT NULL,
	created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	changed_by INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	changed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS cm_webhooks_monitor ON cm_webhooks (monitor);

COMMENT ON TABLE cm_webhooks IS 'Webhook actions configured on code monitors';
COMMENT ON COLUMN cm_webhooks.monitor IS 'The code monitor that the action is defined on';
COMMENT ON COLUMN cm_webhooks.url IS 'The webhook URL we send the code monitor event to';
COMMENT ON COLUMN cm_webhooks.enabled IS 'Whether this webhook action is enabled. When not enabled, the action will not be run when its code monitor generates events';
-- End cm_webhooks

-- Begin cm_slack_webhooks
CREATE TABLE IF NOT EXISTS cm_slack_webhooks (
	id BIGSERIAL PRIMARY KEY,
	monitor BIGINT NOT NULL REFERENCES cm_monitors(id) ON DELETE CASCADE,
	url TEXT NOT NULL,
	enabled BOOLEAN NOT NULL,
	created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	changed_by INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	changed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS cm_slack_webhooks_monitor ON cm_slack_webhooks (monitor);

COMMENT ON TABLE cm_slack_webhooks IS 'Slack webhook actions configured on code monitors';
COMMENT ON COLUMN cm_slack_webhooks.monitor IS 'The code monitor that the action is defined on';
COMMENT ON COLUMN cm_slack_webhooks.url IS 'The Slack webhook URL we send the code monitor event to';
COMMENT ON COLUMN cm_webhooks.enabled IS 'Whether this Slack webhook action is enabled. When not enabled, the action will not be run when its code monitor generates events';
-- End cm_slack_webhooks

-- Begin add non-email actions to cm_triggers
ALTER TABLE cm_action_jobs
	ALTER COLUMN email DROP NOT NULL, -- make email nullable (drop the not null constraint)
	ADD COLUMN IF NOT EXISTS webhook BIGINT -- create a nullable webhook column
		REFERENCES cm_webhooks(id) ON DELETE CASCADE,
	ADD COLUMN IF NOT EXISTS slack_webhook BIGINT  --create a nullable slack webhook column
		REFERENCES cm_slack_webhooks(id) ON DELETE CASCADE,
	ADD CONSTRAINT cm_action_jobs_only_one_action_type CHECK ( -- constrain that only one of email, webhook, and slack_webhook is non-null
		( 
			CASE WHEN email IS NULL THEN 0 ELSE 1 END
			+ CASE WHEN webhook IS NULL THEN 0 ELSE 1 END 
			+ CASE WHEN slack_webhook IS NULL THEN 0 ELSE 1 END 
		) = 1
	);

COMMENT ON COLUMN cm_action_jobs.email IS 'The ID of the cm_emails action to execute if this is an email job. Mutually exclusive with webhook and slack_webhook';
COMMENT ON COLUMN cm_action_jobs.webhook IS 'The ID of the cm_webhooks action to execute if this is a webhook job. Mutually exclusive with email and slack_webhook';
COMMENT ON COLUMN cm_action_jobs.slack_webhook IS 'The ID of the cm_slack_webhook action to execute if this is a slack webhook job. Mutually exclusive with email and webhook';
COMMENT ON CONSTRAINT cm_action_jobs_only_one_action_type ON cm_action_jobs IS 'Constrains that each queued code monitor action has exactly one action type';
-- End add non-email actions to cm_triggers

COMMIT;
