BEGIN;

ALTER TABLE cm_action_jobs
	DROP CONSTRAINT IF EXISTS cm_action_jobs_only_one_action_type,
	DROP COLUMN IF EXISTS slack_webhook,
	DROP COLUMN IF EXISTS webhook,
	ALTER COLUMN email SET NOT NULL;

DROP TABLE IF EXISTS cm_slack_webhooks;
DROP TABLE IF EXISTS cm_webhooks;

COMMIT;
