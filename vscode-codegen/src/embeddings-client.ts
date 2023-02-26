import fetch, { Response } from 'node-fetch'

export interface EmbeddingSearchResult {
	filePath: string
	start: number
	end: number
	text: string
}

export interface EmbeddingSearchResults {
	codeResults: EmbeddingSearchResult[]
	markdownResults: EmbeddingSearchResult[]
}

export class EmbeddingsClient {
	headers: { headers: { authorization: string } }

	constructor(private embeddingsUrl: string, private accessToken: string, private codebaseId: string) {
		this.headers = { headers: { authorization: `Bearer ${this.accessToken}` } }
	}

	async search(query: string, codeCount: number, markdownCount: number): Promise<EmbeddingSearchResults> {
		const url = `${this.embeddingsUrl}/embeddings/search/${encodeURIComponent(
			this.codebaseId
		)}?query=${encodeURIComponent(query)}&codeCount=${encodeURIComponent(
			codeCount
		)}&markdownCount=${encodeURIComponent(markdownCount)}`
		return fetch(url, this.headers)
			.then(verifyResponseCode)
			.then(response => response.json())
			.then(data => data as EmbeddingSearchResults)
	}

	async queryNeedsAdditionalContext(query: string): Promise<boolean> {
		const url = `${this.embeddingsUrl}/embeddings/needs-additional-context?query=${encodeURIComponent(query)}`
		return fetch(url, this.headers)
			.then(verifyResponseCode)
			.then(response => response.json())
			.then(data => data.needsAdditionalContext as boolean)
	}
}

function verifyResponseCode(response: Response): Response {
	if (!response.ok) {
		throw new Error('HTTP status code: ' + response.status)
	}
	return response
}
