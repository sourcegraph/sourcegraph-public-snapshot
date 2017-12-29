BEGIN;
ALTER TABLE threads RENAME COLUMN author_user_id TO author_user_id_new;
ALTER TABLE threads ADD COLUMN author_user_id text;
UPDATE threads SET author_user_id=(SELECT users.auth_id FROM users WHERE users.id=threads.author_user_id_new);
ALTER TABLE threads ALTER COLUMN author_user_id SET NOT NULL;
ALTER TABLE threads DROP COLUMN author_user_id_new;
INSERT INTO threads(id, org_repo_id, repo_revision_path, repo_revision, start_line, end_line, start_character, end_character, created_at, archived_at, updated_at, deleted_at, range_length, branch, author_user_id, html_lines_before, html_lines, html_lines_after, text_lines_before, text_lines, text_lines_after, text_lines_selection_range_start, text_lines_selection_range_length, lines_revision, lines_revision_path)
	SELECT id, org_repo_id, repo_revision_path, repo_revision, start_line, end_line, start_character, end_character, created_at, archived_at, updated_at, deleted_at, range_length, branch, author_user_id_old, html_lines_before, html_lines, html_lines_after, text_lines_before, text_lines, text_lines_after, text_lines_selection_range_start, text_lines_selection_range_length, lines_revision, lines_revision_path FROM threads_bkup_1514544774;
DROP TABLE threads_bkup_1514544774;
COMMIT;
