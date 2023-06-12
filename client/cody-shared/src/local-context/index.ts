export interface ContextResult {
    repoName?: string
    revision?: string
    fileName: string
    content: string
}

export interface KeywordContextFetcher {
    getContext(query: string, numResults: number): Promise<ContextResult[]>
    getSearchContext(query: string, numResults: number): Promise<ContextResult[]>
}

export interface FilenameContextFetcher {
    getContext(query: string, numResults: number): Promise<ContextResult[]>
}
