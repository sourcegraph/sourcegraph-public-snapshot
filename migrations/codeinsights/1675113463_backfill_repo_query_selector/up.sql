UPDATE insight_series
SET repository_criteria = 'repo:.*'
WHERE (CARDINALITY(repositories) = 0 AND generation_method != 'lang_stats')

select * from insight_series where cardinality(repositories) = 0;
