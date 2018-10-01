CREATE TABLE "discussion_mail_reply_tokens" (
    "token" text NOT NULL PRIMARY KEY,
    "user_id" int NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    "thread_id" bigint NOT NULL REFERENCES discussion_threads (id) ON DELETE RESTRICT,
    "deleted_at" TIMESTAMP WITH TIME ZONE
);
CREATE INDEX ON discussion_mail_reply_tokens(token);
CREATE INDEX ON discussion_mail_reply_tokens(user_id, thread_id);
