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

export interface Embeddings {
    search(query: string, codeCount: number, markdownCount: number): Promise<EmbeddingSearchResults>
    queryNeedsAdditionalContext(query: string): Promise<boolean>
}
