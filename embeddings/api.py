import os
import re
import json
import itertools
from typing import Tuple, Optional, List

import numpy as np
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse

from embed import get_embeddings, EMBEDDING_ENGINE
from api_auth import authenticate
from search import EmbeddingsSearchIndex

app = FastAPI()

USERS_PATH = os.environ["CODY_USERS_PATH"]
EMBEDDINGS_DIR = os.environ["EMBEDDINGS_DIR"]

MARKDOWN_EXTENSIONS = set(["md", "markdown"])


def is_markdown_file(file_path: str) -> bool:
    _, ext = os.path.splitext(file_path)
    ext_lower = ext[1:].lower()
    return ext_lower in MARKDOWN_EXTENSIONS


def get_codebase_search_index(
    codebase_id: str,
) -> Tuple[EmbeddingsSearchIndex, Optional[EmbeddingsSearchIndex]]:
    embeddings_metadata_path = os.path.join(
        EMBEDDINGS_DIR, f"{codebase_id}_embeddings_metadata.json"
    )
    with open(embeddings_metadata_path, encoding="utf-8") as f:
        embeddings_metadata = json.load(f)

    embeddings_path = os.path.join(EMBEDDINGS_DIR, f"{codebase_id}_embeddings.npy")
    embeddings = np.load(embeddings_path).astype(np.float32)

    # Split the embeddings into text (markdown) and code indices. In a combined index, Markdown results tended to
    # always feature at the top, and didn't leave room for the code. To avoid that we query the markdown and code
    # indices separately.

    code_embeddings, code_metadata = [], []
    markdown_embeddings, markdown_metadata = [], []

    for idx, row in enumerate(embeddings_metadata):
        if is_markdown_file(row["filePath"]):
            markdown_metadata.append(row)
            markdown_embeddings.append(embeddings[idx, :])
        else:
            code_metadata.append(row)
            code_embeddings.append(embeddings[idx, :])

    code_search_index = EmbeddingsSearchIndex(np.vstack(code_embeddings), code_metadata)

    markdown_search_index = (
        EmbeddingsSearchIndex(np.vstack(markdown_embeddings), markdown_metadata)
        if len(markdown_embeddings) > 0
        else None
    )

    return code_search_index, markdown_search_index


def get_users(users_path: str):
    with open(users_path, encoding="utf-8") as f:
        return json.load(f)


def get_codebase_ids(users_path: str) -> List[str]:
    return list(
        set(
            itertools.chain.from_iterable(
                [user["accessibleCodebaseIDs"] for user in get_users(users_path)]
            )
        )
    )


codebase_embeddings = {
    codebase_id: get_codebase_search_index(codebase_id)
    for codebase_id in get_codebase_ids(USERS_PATH)
}

additional_context_embeddings = {
    "yes": np.load(
        os.path.join(EMBEDDINGS_DIR, "need_additional_context_messages_embeddings.npy")
    ),
    "no": np.load(
        os.path.join(EMBEDDINGS_DIR, "no_additional_context_messages_embeddings.npy")
    ),
}


needs_no_context_message_regexps = [
    re.compile(r"(previous|above)\s+(message|code|text)", re.IGNORECASE),
    re.compile(
        r"(translate|convert|change|for|make|refactor|rewrite|ignore|explain|fix|try|show)\s+(that|this|above|previous|it|again)",
        re.IGNORECASE,
    ),
    re.compile(
        r"(this|that).*?\s+(is|seems|looks)\s+(wrong|incorrect|bad|good)", re.IGNORECASE
    ),
    re.compile(r"^(yes|no|correct|wrong|nope|yep|now|cool)(\s|.|,)", re.IGNORECASE),
]


def get_mean_similarity(embeddings: np.ndarray, query_embedding: np.ndarray):
    return np.matmul(embeddings, query_embedding.T).mean()


def is_query_similar_to_no_context_messages(query: str, delta=0.02) -> bool:
    query_embedding = np.array(get_embeddings([query], engine=EMBEDDING_ENGINE))
    need_context_messages_similarity = get_mean_similarity(
        additional_context_embeddings["yes"], query_embedding
    )
    no_context_messages_similarity = get_mean_similarity(
        additional_context_embeddings["no"], query_embedding
    )

    # We have to be really sure that the query requires no context. So we check if the query
    # is at least `delta` more similar to no context messages compared to messages that need additional context.
    return (no_context_messages_similarity - need_context_messages_similarity) > delta


def query_needs_additional_context(query: str) -> bool:
    query = query.strip()

    # Allow the user to ask general questions (not related to the codebase) by prefixing the query with "general:"
    if query.lower().startswith("general:"):
        return False

    if len(query) < 15:
        return False

    # User provided their own code context in the form of a Markdown code block.
    if "```\n" in query:
        return False

    for regexp in needs_no_context_message_regexps:
        if regexp.search(query):
            return False

    if is_query_similar_to_no_context_messages(query):
        return False

    return True


@app.middleware("http")
async def auth_middleware(request: Request, call_next):
    authorization_header = request.headers.get("authorization")
    if not authorization_header:
        return JSONResponse({}, status_code=400)

    user = authenticate(authorization_header, get_users(USERS_PATH))
    if not user:
        return JSONResponse({}, status_code=401)

    request.state.user = user
    response = await call_next(request)
    return response


# TODO: {codebase_id:path}
@app.get("/embeddings/search/{codebase_id}")
def search_extracted_functions_by_text(
    request: Request,
    codebase_id: str = "",
    query: str = "",
    codeCount: int = 5,
    markdownCount: int = 5,
):
    user = request.state.user
    if codebase_id not in user["accessibleCodebaseIDs"]:
        return JSONResponse({}, status_code=401)

    code_search_index, markdown_search_index = codebase_embeddings[codebase_id]
    code_results = code_search_index.search(query, codeCount)
    markdown_results = (
        markdown_search_index.search(query, markdownCount)
        if markdown_search_index
        else []
    )
    return {"codeResults": code_results, "markdownResults": markdown_results}


@app.get("/embeddings/needs-additional-context")
def additional_context(query: str = ""):
    return {"needsAdditionalContext": query_needs_additional_context(query)}
