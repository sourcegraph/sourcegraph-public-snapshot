CREATE TABLE IF NOT EXISTS codeintel_ranking_graph_keys (
    id SERIAL PRIMARY KEY,
    graph_key TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
