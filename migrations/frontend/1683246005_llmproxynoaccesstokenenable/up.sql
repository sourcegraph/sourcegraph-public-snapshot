ALTER TABLE product_licenses DROP COLUMN IF EXISTS access_token_enabled;

ALTER TABLE product_subscriptions ALTER COLUMN llm_proxy_enabled SET DEFAULT FALSE;

-- One time migration to disable all licenses again, we want to enable them manually.
UPDATE product_subscriptions SET llm_proxy_enabled = FALSE;
