BEGIN;

CREATE INDEX campaign_jobs_campaign_plan_id ON campaign_jobs(campaign_plan_id);
CREATE INDEX campaign_jobs_started_at ON campaign_jobs(started_at);
CREATE INDEX campaign_jobs_finished_at ON campaign_jobs(finished_at);

-- When we create a `changeset` with a non null campaign_plan_id we must
-- validate that that all of that campaign_plan's jobs are finished.

CREATE OR REPLACE FUNCTION validate_campaign_plan_is_finished() RETURNS TRIGGER AS
$validate_campaign_plan_is_finished$
DECLARE
  running int;
BEGIN
  running := (
    SELECT COUNT(*) FROM campaign_jobs
    WHERE campaign_plan_id = NEW.campaign_plan_id
    AND finished_at IS NULL
  );

  IF (running != 0) THEN
    RAISE EXCEPTION 'CampaignPlan{ID: %} has % unfinished jobs',
      NEW.campaign_plan_id, running;
  END IF;

  RETURN NEW;
END;
$validate_campaign_plan_is_finished$
LANGUAGE plpgsql;

CREATE TRIGGER trig_validate_campaign_plan_is_finished
BEFORE INSERT OR UPDATE ON campaigns
FOR EACH ROW EXECUTE PROCEDURE validate_campaign_plan_is_finished();

COMMIT;
