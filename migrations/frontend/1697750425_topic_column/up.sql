CREATE OR REPLACE FUNCTION extract_topics_from_metadata(external_service_type text, metadata jsonb)
    RETURNS text[]
    IMMUTABLE
AS $$
BEGIN
    RETURN CASE external_service_type
    WHEN 'github' THEN
        ARRAY(SELECT * FROM jsonb_array_elements_text(jsonb_path_query_array(metadata, '$.RepositoryTopics.Nodes[*].Topic.Name')))
    WHEN 'gitlab' THEN
        ARRAY(SELECT * FROM jsonb_array_elements_text(metadata->'topics'))
    ELSE
        '{}'::text[]
    END;
EXCEPTION WHEN others THEN
    -- Catch exceptions in the case that metadata is not shaped like we expect
    RETURN '{}'::text[];
END;
$$ LANGUAGE plpgsql;


ALTER TABLE IF EXISTS repo
ADD COLUMN IF NOT EXISTS topics text[] GENERATED ALWAYS AS (extract_topics_from_metadata(external_service_type, metadata)) STORED;
