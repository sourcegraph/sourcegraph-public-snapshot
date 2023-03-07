# Embeddings

Provides natural-language code search capabilities using embeddings.

## Development

### Setup

See the [top-level README.md](../README.md#setup):

- You must install Python and the Python library dependencies.
- You must set the following environment variables: `OPENAI_API_KEY`, `EMBEDDINGS_DIR`, and `CODY_USERS_PATH`.

### Usage

#### Embedding a codebase on disk

```shell
python3 embed_codebase.py --codebase-id=CODEBASE-ID-1 --codebase-path=PATH-TO-REPO --output-dir=$EMBEDDINGS_DIR
```

- `CODEBASE-ID-1`: Some identifier for the codebase, such as `github.com/sourcegraph/conc`.
- `PATH-TO-REPO`: The file path to an existing codebase (e.g., a Git checkout) on disk, such as `$HOME/src/github.com/sourcegraph/conc`.

#### Embedding a codebase by Git repository URL

```shell
python3 embed_repos.py --repos GIT-CLONE-URL --output-dir=$EMBEDDINGS_DIR
```

- `GIT-CLONE-URL`: One or more Git clone URLs (separated by whitespace), such as `https://github.com/sourcegraph/conc`.

#### Embedding context dataset

```shell
python3 embed_context_dataset.py --output-dir=$EMBEDDINGS_DIR
```

#### Running the API (in development)

```shell
asdf env python uvicorn api:app --reload --port 9301
```
