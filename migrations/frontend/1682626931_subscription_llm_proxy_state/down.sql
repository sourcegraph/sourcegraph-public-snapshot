ALTER TABLE product_subscriptions
DROP COLUMN IF EXISTS llm_proxy_enabled,
DROP COLUMN IF EXISTS llm_proxy_rate_limit,
DROP COLUMN IF EXISTS llm_proxy_rate_interval_seconds;
