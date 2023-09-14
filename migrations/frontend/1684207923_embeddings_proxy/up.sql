ALTER TABLE product_subscriptions
    ADD COLUMN IF NOT EXISTS cody_gateway_embeddings_api_rate_limit INTEGER,
    ADD COLUMN IF NOT EXISTS cody_gateway_embeddings_api_rate_interval_seconds INTEGER,
    ADD COLUMN IF NOT EXISTS cody_gateway_embeddings_api_allowed_models TEXT[];

COMMENT ON COLUMN product_subscriptions.cody_gateway_embeddings_api_rate_limit IS 'Custom requests per time interval allowed for embeddings';
COMMENT ON COLUMN product_subscriptions.cody_gateway_embeddings_api_rate_interval_seconds IS 'Custom time interval over which the embeddings rate limit is applied';
COMMENT ON COLUMN product_subscriptions.cody_gateway_embeddings_api_allowed_models IS 'Custom override for the set of models allowed for embedding';
