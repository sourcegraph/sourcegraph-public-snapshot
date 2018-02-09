CREATE TABLE cert_cache (
	id bigserial NOT NULL PRIMARY KEY,
	cache_key text NOT NULL,
	b64data text NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE UNIQUE INDEX cert_cache_key_idx ON cert_cache USING btree(cache_key);
