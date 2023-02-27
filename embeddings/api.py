import os
import json
import itertools
from typing import List

from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse
from pydantic import BaseModel

from api_auth import authenticate
from embed_context_dataset import (
    query_needs_additional_context,
    get_additional_context_embeddings,
)
from search import get_codebase_search_index

app = FastAPI()

USERS_PATH = os.environ["CODY_USERS_PATH"]
EMBEDDINGS_DIR = os.environ["EMBEDDINGS_DIR"]


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
    codebase_id: get_codebase_search_index(codebase_id, EMBEDDINGS_DIR)
    for codebase_id in get_codebase_ids(USERS_PATH)
}

additional_context_embeddings = get_additional_context_embeddings(EMBEDDINGS_DIR)


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



class SearchJSON(BaseModel):
    query: str
    codeCount: int
    markdownCount: int


APP_PATH = "/embeddings"
SEARCH_ROUTE = APP_PATH + "/search/{codebase_id:path}"


@app.get(SEARCH_ROUTE)
def search_extracted_functions_by_text_get(
    request: Request,
    codebase_id: str = "",
    query: str = "",
    codeCount: int = 5,
    markdownCount: int = 5,
):
    search = SearchJSON()
    search.query = query
    search.codeCount = codeCount
    search.markdownCount = markdownCount
    return search_extracted_functions_by_text(request, codebase_id, search)


@app.post(SEARCH_ROUTE)
def search_extracted_functions_by_text(
        request: Request,
        codebase_id: str = "",
        search: SearchJSON = None,
):
    user = request.state.user
    if codebase_id not in user["accessibleCodebaseIDs"]:
        return JSONResponse({}, status_code=401)

    code_search_index, markdown_search_index = codebase_embeddings[codebase_id]
    code_results = code_search_index.search(search.query, search.codeCount)
    markdown_results = (
        markdown_search_index.search(search.query, search.markdownCount)
        if markdown_search_index
        else []
    )
    return {"codeResults": code_results, "markdownResults": markdown_results}


class AdditionalContextJSON(BaseModel):
    query: str


ADDITIONAL_CONTEXT_ROUTE = APP_PATH + "/needs-additional-context"


@app.get(ADDITIONAL_CONTEXT_ROUTE)
def additional_context_get(query: str = ""):
    body = AdditionalContextJSON()
    body.query = query
    return additional_context(body)


@app.post(ADDITIONAL_CONTEXT_ROUTE)
def additional_context(body: AdditionalContextJSON = None):
    return {
        "needsAdditionalContext": query_needs_additional_context(
            body.query, additional_context_embeddings
        )
    }
