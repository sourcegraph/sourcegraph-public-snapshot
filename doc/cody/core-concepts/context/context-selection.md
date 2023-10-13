# Cody Context Selection

<p class="subtitle">Learn about where does context come from and how Cody ensures it's good.</p>

The quality of the context is critical, even with increasing context window sizes. Cody obtains context in two primary ways:

1. Embeddings
2. Keyword Search

## Embeddings

Embedding search, on the other hand, utilizes text embeddings, which encode words and sentences as numeric vectors. These vectors are designed to capture the meaning of text, enabling Cody to match code based on semantic similarity rather than exact terms. Embeddings are the preferred method, as they provide:

- Access to the entire code in specified repositories, not limited to the local workspace
- Matches based on the meaning of the query, not just the exact terms

You can learn more about embeddings [here →](./../embeddings.md).

## Keyword Search

Keyword search is a traditional text search approach. It involves splitting content into terms and mapping terms to documents. At query time, terms from the query are matched to candidate documents. While powerful, this approach may be used as a fallback when a codebase lacks embeddings.
