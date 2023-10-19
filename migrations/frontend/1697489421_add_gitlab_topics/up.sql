/* ALTER TABLE repo ADD COLUMN topics text[]
GENERATED ALWAYS AS (CASE WHEN external_service_type='github'
THEN jsonb_path_query_array(metadata, '$.RepositoryTopics.Nodes[*].Topic.Name')
WHEN external_service_type='gitlab'
THEN metadata->'topics'
ELSE ARRAY[]::text[]
END CASE) STORED; */

CREATE FUNCTION get_topics(external_service_type text, metadata jsonb)
RETURNS text[] AS $$
  CASE external_service_type
    WHEN 'github' THEN
      jsonb_array_elements_text(jsonb_path_query_array(metadata, '$.RepositoryTopics.Nodes[*].Topic.Name'))
    WHEN 'gitlab' THEN
      metadata->'topics'
    ELSE
      ARRAY[]::text[]
  END;
$$ LANGUAGE sql IMMUTABLE;

ALTER TABLE repo ADD COLUMN topics text[]
GENERATED ALWAYS AS (get_topics(external_service_type, metadata)) STORED;
