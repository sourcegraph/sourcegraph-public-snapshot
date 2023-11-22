ALTER TABLE product_subscriptions
ALTER COLUMN cody_gateway_chat_rate_limit TYPE integer,
    ALTER COLUMN cody_gateway_code_rate_limit TYPE integer,
    ALTER COLUMN cody_gateway_embeddings_api_rate_limit TYPE integer;
