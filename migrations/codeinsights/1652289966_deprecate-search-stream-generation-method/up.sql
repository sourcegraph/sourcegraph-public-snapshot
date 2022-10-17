UPDATE insight_series
    SET generation_method = 'search'
    WHERE generation_method = 'search-stream';
