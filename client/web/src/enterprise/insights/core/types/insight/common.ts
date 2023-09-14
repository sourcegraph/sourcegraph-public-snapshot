import type { InsightDataNode, SeriesSortDirection, SeriesSortMode } from '../../../../../graphql-operations'

export interface BaseInsight {
    id: string
    title: string
    type: InsightType
    dashboardReferenceCount: number
    dashboards: InsightDashboardReference[]
    isFrozen: boolean
}

export type IncompleteDatapointAlert = InsightDataNode['dataSeries'][number]['status']['incompleteDatapoints'][number]

export enum InsightType {
    SearchBased = 'SearchBased',
    LangStats = 'LangStats',
    CaptureGroup = 'CaptureGroup',
    Compute = 'Compute',
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
    seriesDisplayOptions: InsightSeriesDisplayOptions
}

export interface InsightSeriesDisplayOptions {
    numSamples: number | null
    limit: number | null
    sortOptions: {
        mode: SeriesSortMode
        direction: SeriesSortDirection
    }
}

export interface InsightDashboardReference {
    id: string
    title: string
}
