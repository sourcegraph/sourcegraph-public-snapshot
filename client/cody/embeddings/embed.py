from typing import List

from tenacity import retry, stop_after_attempt, wait_random_exponential
import openai

CHARS_PER_TOKEN = 4
EMBEDDING_ENGINE = "text-embedding-ada-002"


@retry(wait=wait_random_exponential(min=1, max=20), stop=stop_after_attempt(6))
def get_embeddings(list_of_text: List[str], engine: str) -> List[List[float]]:
    assert len(list_of_text) <= 2048, "The batch size should not be larger than 2048."

    # replace newlines, which can negatively affect performance.
    list_of_text = [text.replace("\n", " ") for text in list_of_text]

    data = openai.Embedding.create(input=list_of_text, engine=engine).data
    data = sorted(data, key=lambda x: x["index"])
    return [d["embedding"] for d in data]
