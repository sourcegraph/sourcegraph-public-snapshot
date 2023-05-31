BEGIN;

ALTER TABLE public.product_subscriptions
    RENAME COLUMN cody_gateway_enabled TO llm_proxy_enabled;

ALTER TABLE public.product_subscriptions
    RENAME COLUMN cody_gateway_chat_rate_limit TO llm_proxy_chat_rate_limit;

ALTER TABLE public.product_subscriptions
    RENAME COLUMN cody_gateway_chat_rate_interval_seconds TO llm_proxy_chat_rate_interval_seconds;

ALTER TABLE public.product_subscriptions
    RENAME COLUMN cody_gateway_chat_rate_limit_allowed_models TO llm_proxy_chat_rate_limit_allowed_models;

ALTER TABLE public.product_subscriptions
    RENAME COLUMN cody_gateway_code_rate_limit TO llm_proxy_code_rate_limit;

ALTER TABLE public.product_subscriptions
    RENAME COLUMN cody_gateway_code_rate_interval_seconds TO llm_proxy_code_rate_interval_seconds;

ALTER TABLE public.product_subscriptions
    RENAME COLUMN cody_gateway_code_rate_limit_allowed_models TO llm_proxy_code_rate_limit_allowed_models;

COMMIT;
