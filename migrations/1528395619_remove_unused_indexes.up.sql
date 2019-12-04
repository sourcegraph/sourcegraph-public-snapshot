BEGIN;

DROP INDEX IF EXISTS campaign_jobs_campaign_plan_id;
DROP INDEX IF EXISTS discussion_mail_reply_tokens_token_idx;
DROP INDEX IF EXISTS discussion_threads_id_idx;

COMMIT;
