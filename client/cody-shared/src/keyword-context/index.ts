export interface KeywordContextFetcherResult {
    fileName: string
    content: string
}

export interface KeywordContextFetcher {
    getContext(query: string, numResults: number): Promise<KeywordContextFetcherResult[]>
    getSearchContext(query: string, numResults: number): Promise<KeywordContextFetcherResult[]>
}
