BEGIN;
ALTER TABLE discussion_comments ADD reports text[] NOT NULL DEFAULT '{}';
CREATE INDEX "discussion_comments_reports_array_length_idx" ON discussion_comments (array_length(reports,1));
END;
