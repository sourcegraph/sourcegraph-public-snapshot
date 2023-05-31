-- According to Cody, ALTER TABLE is idempotent

BEGIN;

ALTER TABLE public.product_subscriptions
    RENAME COLUMN llm_proxy_enabled TO cody_gateway_enabled;

COMMENT ON COLUMN public.product_subscriptions.cody_gateway_enabled IS NULL;

ALTER TABLE public.product_subscriptions
    RENAME COLUMN llm_proxy_chat_rate_limit TO cody_gateway_chat_rate_limit;

COMMENT ON COLUMN public.product_subscriptions.cody_gateway_chat_rate_limit IS NULL;

ALTER TABLE public.product_subscriptions
    RENAME COLUMN llm_proxy_chat_rate_interval_seconds TO cody_gateway_chat_rate_interval_seconds;

COMMENT ON COLUMN public.product_subscriptions.cody_gateway_chat_rate_interval_seconds IS NULL;

ALTER TABLE public.product_subscriptions
    RENAME COLUMN llm_proxy_chat_rate_limit_allowed_models TO cody_gateway_chat_rate_limit_allowed_models;

COMMENT ON COLUMN public.product_subscriptions.cody_gateway_chat_rate_limit_allowed_models IS NULL;

ALTER TABLE public.product_subscriptions
    RENAME COLUMN llm_proxy_code_rate_limit TO cody_gateway_code_rate_limit;

COMMENT ON COLUMN public.product_subscriptions.cody_gateway_code_rate_limit IS NULL;

ALTER TABLE public.product_subscriptions
    RENAME COLUMN llm_proxy_code_rate_interval_seconds TO cody_gateway_code_rate_interval_seconds;

COMMENT ON COLUMN public.product_subscriptions.cody_gateway_code_rate_interval_seconds IS NULL;

ALTER TABLE public.product_subscriptions
    RENAME COLUMN llm_proxy_code_rate_limit_allowed_models TO cody_gateway_code_rate_limit_allowed_models;

COMMENT ON COLUMN public.product_subscriptions.cody_gateway_code_rate_limit_allowed_models IS NULL;

COMMIT;
