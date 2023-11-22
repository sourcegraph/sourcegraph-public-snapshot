ALTER TABLE product_licenses ALTER COLUMN access_token_enabled SET DEFAULT FALSE;
-- One time migration to disable all licenses again.
UPDATE product_licenses SET access_token_enabled = FALSE;

ALTER TABLE product_subscriptions ALTER COLUMN llm_proxy_enabled SET DEFAULT TRUE;

UPDATE product_subscriptions
SET llm_proxy_enabled = TRUE;
-- Initially, mark any subscription that has no active license as without LLM-proxy access,
-- since there are a lot of old subscriptions out there.
UPDATE product_subscriptions
SET llm_proxy_enabled = false
WHERE id IN (
    SELECT product_subscription_id
    FROM product_licenses
    WHERE license_expires_at > NOW()
    GROUP BY product_subscription_id
    HAVING COUNT(*) = 0
);
