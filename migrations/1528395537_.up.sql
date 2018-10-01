CREATE TABLE "discussion_threads" (
    "id" bigserial NOT NULL PRIMARY KEY,
    "author_user_id" int NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    "title" text,
    "target_repo_id" bigint,
    "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    "archived_at" TIMESTAMP WITH TIME ZONE,
    "updated_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMP WITH TIME ZONE
);

CREATE INDEX ON discussion_threads(id);
CREATE INDEX ON discussion_threads(author_user_id);

CREATE TABLE "discussion_threads_target_repo" (
    "id" bigserial NOT NULL PRIMARY KEY,
    "thread_id" bigint NOT NULL REFERENCES discussion_threads (id) ON DELETE RESTRICT,
    "repo_id" int NOT NULL REFERENCES repo (id) ON DELETE RESTRICT,
    "path" text,
    "branch" text,
    "revision" text,
    "start_line" int,
    "end_line" int,
    "start_character" int,
    "end_character" int,
    "lines_before" text,
    "lines" text,
    "lines_after" text
);

CREATE INDEX ON discussion_threads_target_repo(repo_id, path);
ALTER TABLE discussion_threads ADD CONSTRAINT discussion_threads_target_repo_id_fk FOREIGN KEY (target_repo_id) REFERENCES discussion_threads_target_repo (id) ON DELETE RESTRICT;

CREATE table "discussion_comments" (
    "id" bigserial NOT NULL PRIMARY KEY,
    "thread_id" bigint NOT NULL REFERENCES discussion_threads (id) ON DELETE RESTRICT,
    "author_user_id" int NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    "contents" text NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMP WITH TIME ZONE
);

CREATE INDEX ON discussion_comments(thread_id);
CREATE INDEX ON discussion_comments(author_user_id);
