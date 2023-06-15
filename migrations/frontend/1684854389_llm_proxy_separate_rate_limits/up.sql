DO $$
BEGIN
    ALTER TABLE product_subscriptions RENAME COLUMN llm_proxy_rate_limit TO llm_proxy_chat_rate_limit;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column llm_proxy_rate_limit does not exist in table product_subscriptions';
END $$;

DO $$
BEGIN
    ALTER TABLE product_subscriptions RENAME COLUMN llm_proxy_rate_interval_seconds TO llm_proxy_chat_rate_interval_seconds;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column llm_proxy_rate_interval_seconds does not exist in table product_subscriptions';
END $$;

ALTER TABLE product_subscriptions
    ADD COLUMN IF NOT EXISTS llm_proxy_chat_rate_limit_allowed_models TEXT[],
    ADD COLUMN IF NOT EXISTS llm_proxy_code_rate_limit INTEGER,
	ADD COLUMN IF NOT EXISTS llm_proxy_code_rate_interval_seconds INTEGER,
	ADD COLUMN IF NOT EXISTS llm_proxy_code_rate_limit_allowed_models TEXT[];
