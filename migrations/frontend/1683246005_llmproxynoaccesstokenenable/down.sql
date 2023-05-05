ALTER TABLE product_licenses
ADD COLUMN IF NOT EXISTS access_token_enabled BOOLEAN NOT NULL DEFAULT false;

COMMENT ON COLUMN product_licenses.access_token_enabled
IS 'Whether this license key can be used as an access token to authenticate API requests';

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
