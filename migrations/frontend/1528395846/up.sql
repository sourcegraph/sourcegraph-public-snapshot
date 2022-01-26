-- +++
-- parent: 1528395845
-- +++

BEGIN;

ALTER TABLE lsif_uploads_visible_at_tip ADD COLUMN branch_or_tag_name text NOT NULL DEFAULT '';
ALTER TABLE lsif_uploads_visible_at_tip ADD COLUMN is_default_branch boolean NOT NULL DEFAULT false;

COMMENT ON COLUMN lsif_uploads_visible_at_tip.upload_id IS 'The identifier of the upload visible from the tip of the specified branch or tag.';
COMMENT ON COLUMN lsif_uploads_visible_at_tip.branch_or_tag_name IS 'The name of the branch or tag.';
COMMENT ON COLUMN lsif_uploads_visible_at_tip.is_default_branch IS  'Whether the specified branch is the default of the repository. Always false for tags.';

-- Update all existing visible uploads to be the default branch, which is true until
-- we start recalcaulting the commit graph with tags and non-default branches.
UPDATE lsif_uploads_visible_at_tip SET is_default_branch = true;

-- Mark every graph as dirty so we recalculate retention correctly once the instance
-- boots up.
UPDATE lsif_dirty_repositories SET dirty_token = dirty_token + 1;

COMMIT;
