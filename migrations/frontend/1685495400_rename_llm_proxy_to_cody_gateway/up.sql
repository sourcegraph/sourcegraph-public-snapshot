DO $$
BEGIN
    ALTER TABLE public.product_subscriptions
        RENAME COLUMN llm_proxy_enabled TO cody_gateway_enabled;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column llm_proxy_enabled does not exist in table product_subscriptions';
END $$;

COMMENT ON COLUMN public.product_subscriptions.cody_gateway_enabled IS NULL;

DO $$
BEGIN
    ALTER TABLE public.product_subscriptions
        RENAME COLUMN llm_proxy_chat_rate_limit TO cody_gateway_chat_rate_limit;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column llm_proxy_chat_rate_limit does not exist in table product_subscriptions';
END $$;

COMMENT ON COLUMN public.product_subscriptions.cody_gateway_chat_rate_limit IS NULL;

DO $$
BEGIN
    ALTER TABLE public.product_subscriptions
        RENAME COLUMN llm_proxy_chat_rate_interval_seconds TO cody_gateway_chat_rate_interval_seconds;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column llm_proxy_chat_rate_interval_seconds does not exist in table product_subscriptions';
END $$;

COMMENT ON COLUMN public.product_subscriptions.cody_gateway_chat_rate_interval_seconds IS NULL;

DO $$
BEGIN
    ALTER TABLE public.product_subscriptions
        RENAME COLUMN llm_proxy_chat_rate_limit_allowed_models TO cody_gateway_chat_rate_limit_allowed_models;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column llm_proxy_chat_rate_limit_allowed_models does not exist in table product_subscriptions';
END $$;

COMMENT ON COLUMN public.product_subscriptions.cody_gateway_chat_rate_limit_allowed_models IS NULL;

DO $$
BEGIN
    ALTER TABLE public.product_subscriptions
        RENAME COLUMN llm_proxy_code_rate_limit TO cody_gateway_code_rate_limit;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column llm_proxy_code_rate_limit does not exist in table product_subscriptions';
END $$;

COMMENT ON COLUMN public.product_subscriptions.cody_gateway_code_rate_limit IS NULL;

DO $$
BEGIN
    ALTER TABLE public.product_subscriptions
        RENAME COLUMN llm_proxy_code_rate_interval_seconds TO cody_gateway_code_rate_interval_seconds;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column llm_proxy_code_rate_interval_seconds does not exist in table product_subscriptions';
END $$;

COMMENT ON COLUMN public.product_subscriptions.cody_gateway_code_rate_interval_seconds IS NULL;

DO $$
BEGIN
    ALTER TABLE public.product_subscriptions
        RENAME COLUMN llm_proxy_code_rate_limit_allowed_models TO cody_gateway_code_rate_limit_allowed_models;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column llm_proxy_code_rate_limit_allowed_models does not exist in table product_subscriptions';
END $$;

COMMENT ON COLUMN public.product_subscriptions.cody_gateway_code_rate_limit_allowed_models IS NULL;
