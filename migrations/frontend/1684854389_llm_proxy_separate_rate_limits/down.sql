DO $$
BEGIN
    ALTER TABLE product_subscriptions RENAME COLUMN llm_proxy_chat_rate_limit TO llm_proxy_rate_limit;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column llm_proxy_chat_rate_limit does not exist in table product_subscriptions';
END $$;

DO $$
BEGIN
    ALTER TABLE product_subscriptions RENAME COLUMN llm_proxy_chat_rate_interval_seconds TO llm_proxy_rate_interval_seconds;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column llm_proxy_rate_interval_seconds does not exist in table product_subscriptions';
END $$;

ALTER TABLE product_subscriptions
    DROP COLUMN IF EXISTS llm_proxy_chat_rate_limit_allowed_models,
    DROP COLUMN IF EXISTS llm_proxy_code_rate_limit,
	DROP COLUMN IF EXISTS llm_proxy_code_rate_interval_seconds,
	DROP COLUMN IF EXISTS llm_proxy_code_rate_limit_allowed_models;
