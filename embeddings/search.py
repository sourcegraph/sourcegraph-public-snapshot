import os
import argparse
import json
from functools import lru_cache
from typing import List, Dict, Any

import numpy as np
import faiss

from embed import EMBEDDING_ENGINE, get_embeddings


@lru_cache
def embed_query(query: str) -> np.ndarray:
    return np.array(get_embeddings([query], engine=EMBEDDING_ENGINE), dtype=np.float32)


class EmbeddingsSearchIndex:
    def __init__(self, embeddings: np.ndarray, metadata: List[Dict[str, Any]]):
        self.embeddings = embeddings
        self.metadata = metadata

        dimension = self.embeddings.shape[1]
        self.index = faiss.IndexFlatIP(dimension)
        self.index.add(self.embeddings)

    def search(self, query: str, n_results: int) -> List[Dict[str, Any]]:
        if n_results == 0:
            return []

        query_embedding = embed_query(query)
        _, indices = self.index.search(query_embedding, n_results)
        return [self.metadata[i] for i in indices[0]]


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--codebase-id", dest="codebase_id")
    parser.add_argument("--embeddings-dir", dest="embeddings_dir")
    parser.add_argument("--query", dest="query")
    args = parser.parse_args()

    embeddings_metadata_path = os.path.join(
        args.embeddings_dir, f"{args.codebase_id}_embeddings_metadata.json"
    )
    with open(embeddings_metadata_path, encoding="utf-8") as f:
        embeddings_metadata = json.load(f)

    embeddings_path = os.path.join(
        args.embeddings_dir, f"{args.codebase_id}_embeddings.npy"
    )
    embeddings = np.load(embeddings_path).astype(np.float32)

    search_index = EmbeddingsSearchIndex(embeddings, embeddings_metadata)

    for chunk in search_index.search(args.query, 5):
        print("===", chunk["filePath"], "===")
        print(chunk["text"])
        print()
