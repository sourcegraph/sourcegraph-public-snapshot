DO $$
BEGIN
    ALTER TABLE public.product_subscriptions
            RENAME COLUMN cody_gateway_enabled TO llm_proxy_enabled;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column cody_gateway_enabled does not exist in table product_subscriptions';
END $$;

DO $$
BEGIN
    ALTER TABLE public.product_subscriptions
        RENAME COLUMN cody_gateway_chat_rate_limit TO llm_proxy_chat_rate_limit;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column cody_gateway_chat_rate_limit does not exist in table product_subscriptions';
END $$;

DO $$
BEGIN
    ALTER TABLE public.product_subscriptions
        RENAME COLUMN cody_gateway_chat_rate_interval_seconds TO llm_proxy_chat_rate_interval_seconds;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column cody_gateway_chat_rate_interval_seconds does not exist in table product_subscriptions';
END $$;

DO $$
BEGIN
    ALTER TABLE public.product_subscriptions
        RENAME COLUMN cody_gateway_chat_rate_limit_allowed_models TO llm_proxy_chat_rate_limit_allowed_models;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column cody_gateway_chat_rate_limit_allowed_models does not exist in table product_subscriptions';
END $$;

DO $$
BEGIN
    ALTER TABLE public.product_subscriptions
        RENAME COLUMN cody_gateway_code_rate_limit TO llm_proxy_code_rate_limit;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column cody_gateway_code_rate_limit does not exist in table product_subscriptions';
END $$;

DO $$
BEGIN
    ALTER TABLE public.product_subscriptions
        RENAME COLUMN cody_gateway_code_rate_interval_seconds TO llm_proxy_code_rate_interval_seconds;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column cody_gateway_code_rate_interval_seconds does not exist in table product_subscriptions';
END $$;

DO $$
BEGIN
    ALTER TABLE public.product_subscriptions
        RENAME COLUMN cody_gateway_code_rate_limit_allowed_models TO llm_proxy_code_rate_limit_allowed_models;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column cody_gateway_code_rate_limit_allowed_models does not exist in table product_subscriptions';
END $$;
