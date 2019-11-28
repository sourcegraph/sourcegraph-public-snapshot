BEGIN;

DROP INDEX campaign_jobs_campaign_plan_id;
DROP INDEX discussion_mail_reply_tokens_token_idx;
DROP INDEX discussion_threads_id_idx;

COMMIT;
