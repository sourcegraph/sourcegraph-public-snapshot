CREATE OR REPLACE FUNCTION get_topics(external_service_type text, metadata jsonb)
    RETURNS text[]
    LANGUAGE SQL
    IMMUTABLE
    RETURN CASE external_service_type
    WHEN 'github' THEN
        ARRAY(SELECT * FROM jsonb_array_elements_text(jsonb_path_query_array(metadata, '$.RepositoryTopics.Nodes[*].Topic.Name')))
    WHEN 'gitlab' THEN
        ARRAY(SELECT * FROM jsonb_array_elements_text(metadata->'topics'))
    ELSE
        '{}'::text[]
    END;


ALTER TABLE IF EXISTS repo
ADD COLUMN IF NOT EXISTS topics text[] GENERATED ALWAYS AS (get_topics(external_service_type, metadata)) STORED;
