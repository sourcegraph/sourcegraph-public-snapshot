# Embeddings

Provides natural language code search capabilities using embeddings.

## Dependencies

```
pip3 install "fastapi[all]" faiss-cpu tenacity numpy openai
```

## Environment variables

To access the OpenAI embeddings API:

- `export OPENAI_API_KEY=sk-`
- `export EMBEDDINGS_DIR=/path/to/embeddings`

## Embedding a codebase

```
python3 embed.py --codebase-id=codebase-id-1 --codebase-path=/path/to/repo --output-dir=/path/to/embeddings
```

## Embedding context dataset

```
python3 embed_context_dataset.py --output-dir=/path/to/embeddings
```

## Running the API (in development)

```
uvicorn api:app --reload
```
