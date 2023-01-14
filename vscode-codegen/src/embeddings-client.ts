import axios from "axios";

export interface EmbeddingSearchResult {
	filePath: string;
	start: number;
	end: number;
	text: string;
}

export interface EmbeddingSearchResults {
	codeResults: EmbeddingSearchResult[];
	markdownResults: EmbeddingSearchResult[];
}

export class EmbeddingsClient {
	private embeddingsURL: string;

	constructor(private embeddingsAddr: string, private codebaseID: string) {
		this.embeddingsURL = `http://${this.embeddingsAddr}`;
	}

	async search(
		query: string,
		codeCount: number,
		markdownCount: number
	): Promise<EmbeddingSearchResults> {
		const url = `${this.embeddingsURL}/search/${encodeURIComponent(
			this.codebaseID
		)}?query=${encodeURIComponent(query)}&codeCount=${encodeURIComponent(
			codeCount
		)}&markdownCount=${encodeURIComponent(markdownCount)}`;
		return axios
			.get(url)
			.then((response) => response.data as EmbeddingSearchResults);
	}

	async queryNeedsAdditionalContext(query: string): Promise<boolean> {
		const url = `${
			this.embeddingsURL
		}/needs-additional-context?query=${encodeURIComponent(query)}`;
		return axios
			.get(url)
			.then((response) => response.data.needsAdditionalContext as boolean);
	}
}
