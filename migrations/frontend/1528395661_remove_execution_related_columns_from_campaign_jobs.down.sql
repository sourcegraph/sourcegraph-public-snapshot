BEGIN;

ALTER TABLE campaign_jobs ADD COLUMN description text NOT NULL;
ALTER TABLE campaign_jobs ADD COLUMN started_at timestamp with time zone;
ALTER TABLE campaign_jobs ADD COLUMN finished_at timestamp with time zone;
ALTER TABLE campaign_jobs ADD COLUMN error text NOT NULL;

DROP TRIGGER IF EXISTS trig_validate_campaign_plan_is_finished ON campaigns;
DROP FUNCTION IF EXISTS validate_campaign_plan_is_finished();

CREATE FUNCTION validate_campaign_plan_is_finished() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
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
$$;

CREATE TRIGGER trig_validate_campaign_plan_is_finished BEFORE INSERT OR UPDATE ON campaigns FOR EACH ROW EXECUTE PROCEDURE validate_campaign_plan_is_finished();

COMMIT;
