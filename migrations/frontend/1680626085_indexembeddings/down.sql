-- Undo the changes made in the up migration
DROP TABLE IF EXISTS text_embeddings;
DROP TABLE IF EXISTS code_embeddings;
DROP TABLE IF EXISTS embedding_vectors;
