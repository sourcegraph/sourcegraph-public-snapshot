BEGIN;

CREATE TABLE campaign_plans (
  id bigserial PRIMARY KEY,
  campaign_type text NOT NULL CHECK (campaign_type != ''),
  arguments jsonb NOT NULL DEFAULT '{}'
    CHECK (jsonb_typeof(arguments) = 'object'),
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  updated_at timestamp with time zone NOT NULL DEFAULT now()
);

ALTER TABLE campaigns
ADD COLUMN campaign_plan_id integer REFERENCES campaign_plans(id)
DEFERRABLE INITIALLY IMMEDIATE;

COMMIT;

