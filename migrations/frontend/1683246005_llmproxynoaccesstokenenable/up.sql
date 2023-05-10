ALTER TABLE product_licenses ALTER COLUMN access_token_enabled SET DEFAULT TRUE;

ALTER TABLE product_subscriptions ALTER COLUMN llm_proxy_enabled SET DEFAULT FALSE;

-- One time migration to disable all licenses again, we want to enable them manually.
UPDATE product_subscriptions SET llm_proxy_enabled = FALSE;
