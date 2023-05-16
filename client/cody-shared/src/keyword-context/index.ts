export interface KeywordContextFetcherResult {
    repoName?: string
    revision?: string
    fileName: string
    content: string
}

export interface KeywordContextFetcher {
    getContext(query: string, numResults: number): Promise<KeywordContextFetcherResult[]>
    getSearchContext(query: string, numResults: number): Promise<KeywordContextFetcherResult[]>
}
