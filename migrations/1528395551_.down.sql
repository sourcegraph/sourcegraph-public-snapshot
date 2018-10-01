BEGIN;
DROP INDEX "discussion_comments_reports_array_length_idx";
ALTER TABLE discussion_comments DROP reports;
END;
