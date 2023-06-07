export interface UnifiedContextFetcherResult {
    filePath: string
    content: string
    startLine: number
    endLine: number
    repoName: string
    revision: string
}

export interface UnifiedContextFetcher {
    getContext(
        query: string,
        codeResultsCount: number,
        textResultsCount: number
    ): Promise<UnifiedContextFetcherResult[] | Error>
}
