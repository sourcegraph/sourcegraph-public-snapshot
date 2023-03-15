import fetch, { Response } from 'node-fetch'

import { Embeddings, EmbeddingSearchResults } from '.'

export class EmbeddingsClient implements Embeddings {
    private headers: { authorization: string }

    constructor(private embeddingsUrl: string, private accessToken: string, private codebaseId: string) {
        this.headers = { authorization: `Bearer ${this.accessToken}` }
    }

    public async search(query: string, codeCount: number, markdownCount: number): Promise<EmbeddingSearchResults> {
        const url = `${this.embeddingsUrl}/embeddings/search/${encodeURIComponent(this.codebaseId)}`
        const body = {
            query,
            codeCount,
            markdownCount,
        }
        return fetch(url, {
            method: 'post',
            body: JSON.stringify(body),
            headers: {
                'Content-Type': 'application/json',
                ...this.headers,
            },
        })
            .then(verifyResponseCode)
            .then(response => response.json())
            .then(data => data as EmbeddingSearchResults)
    }

    public async queryNeedsAdditionalContext(query: string): Promise<boolean> {
        const url = `${this.embeddingsUrl}/embeddings/needs-additional-context`
        // TODO: Type, and check, responses from the API endpoint
        return fetch(url, {
            method: 'post',
            body: JSON.stringify({
                query,
            }),
            headers: {
                'Content-Type': 'application/json',
                ...this.headers,
            },
        })
            .then(verifyResponseCode)
            .then(response => response.json())
            .then(json => (json as { needsAdditionalContext: boolean }).needsAdditionalContext)
    }
}

function verifyResponseCode(response: Response): Response {
    if (!response.ok) {
        throw new Error(`HTTP status code: ${response.status}`)
    }
    return response
}
