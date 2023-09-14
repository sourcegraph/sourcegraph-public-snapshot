ALTER TABLE product_subscriptions
    DROP COLUMN IF EXISTS cody_gateway_embeddings_api_rate_limit,
    DROP COLUMN IF EXISTS cody_gateway_embeddings_api_rate_interval_seconds,
    DROP COLUMN IF EXISTS cody_gateway_embeddings_api_allowed_models;
