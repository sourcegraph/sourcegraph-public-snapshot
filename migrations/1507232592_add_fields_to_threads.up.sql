BEGIN;
ALTER TABLE threads ADD COLUMN "author_user_id" text;
ALTER TABLE threads ADD COLUMN "html_lines_before" text;
ALTER TABLE threads ADD COLUMN "html_lines" text;
ALTER TABLE threads ADD COLUMN "html_lines_after" text;
ALTER TABLE threads ADD COLUMN "text_lines_before" text;
ALTER TABLE threads ADD COLUMN "text_lines" text;
ALTER TABLE threads ADD COLUMN "text_lines_after" text;
ALTER TABLE threads ADD COLUMN "text_lines_selection_range_start" integer NOT NULL DEFAULT '0';
ALTER TABLE threads ADD COLUMN "text_lines_selection_range_length" integer NOT NULL DEFAULT '0';

-- Set thread author_user_id to the first comment in the thread, since before
-- this migration that was the thread author.
UPDATE threads
	SET author_user_id = subquery.author_user_id
	FROM (
		SELECT author_user_id, thread_id
		FROM comments
		ORDER BY id ASC
	) as subquery
	WHERE subquery.thread_id = threads.id;

ALTER TABLE threads ALTER COLUMN author_user_id SET NOT NULL;
COMMIT;
