ALTER TABLE product_subscriptions
ADD COLUMN IF NOT EXISTS llm_proxy_enabled BOOLEAN NOT NULL DEFAULT TRUE,
ADD COLUMN IF NOT EXISTS llm_proxy_rate_limit INTEGER,
ADD COLUMN IF NOT EXISTS llm_proxy_rate_interval_seconds INTEGER;

COMMENT ON COLUMN product_subscriptions.llm_proxy_enabled IS 'Whether or not this subscription has access to LLM-proxy';
COMMENT ON COLUMN product_subscriptions.llm_proxy_rate_limit IS 'Custom requests per time interval allowed for LLM-proxy';
COMMENT ON COLUMN product_subscriptions.llm_proxy_rate_interval_seconds IS 'Custom time interval over which the for LLM-proxy rate limit is applied';

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
