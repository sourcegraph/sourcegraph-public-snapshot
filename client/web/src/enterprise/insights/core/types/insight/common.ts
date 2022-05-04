export enum InsightExecutionType {
    /**
     * This type of insights run on FE via search API.
     */
    Runtime = 'runtime',

    /**
     * This type of insights work via our backend and gql API returns this insight with
     * pre-calculated data points.
     */
    Backend = 'backend',
}

export enum InsightType {
    SearchBased = 'SearchBased',
    LangStats = 'LangStats',
    CaptureGroup = 'CaptureGroup',
}

export enum InsightContentType {
    Categorical,
    Series,
}

export interface InsightFilters {
    includeRepoRegexp: string
    excludeRepoRegexp: string
    context: string
    repositories?: string[]
}

export interface BaseInsight {
    id: string
    title: string
    executionType: InsightExecutionType
    type: InsightType
    dashboardReferenceCount: number
    isFrozen: boolean
}
