ALTER TABLE cm_emails
    DROP COLUMN IF EXISTS include_results;
ALTER TABLE cm_webhooks
    DROP COLUMN IF EXISTS include_results;
ALTER TABLE cm_slack_webhooks
    DROP COLUMN IF EXISTS include_results;
