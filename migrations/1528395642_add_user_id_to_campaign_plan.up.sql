BEGIN;

ALTER TABLE campaign_plans
    ADD COLUMN user_id integer,
    ADD CONSTRAINT campaign_plans_user_id_fkey FOREIGN KEY (user_id)
    REFERENCES users (id) DEFERRABLE INITIALLY IMMEDIATE;

COMMIT;
