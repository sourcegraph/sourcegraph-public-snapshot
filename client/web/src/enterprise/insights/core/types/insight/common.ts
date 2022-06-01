import { InsightViewNode, SeriesDisplayOptionsInput, SeriesSortOptionsInput } from '../../../../../graphql-operations'

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

export type SeriesDisplayOptions = InsightViewNode['appliedSeriesDisplayOptions'] &
    InsightViewNode['defaultSeriesDisplayOptions']

export interface BaseInsight {
    id: string
    title: string
    executionType: InsightExecutionType
    type: InsightType
    dashboardReferenceCount: number
    dashboards: InsightDashboardReference[]
    isFrozen: boolean

    seriesDisplayOptions?: SeriesDisplayOptionsInput
    appliedSeriesDisplayOptions?: SeriesDisplayOptions
    defaultSeriesDisplayOptions?: SeriesDisplayOptions
}

// This type simply resets limit and sortOptions to required.
// This makes reasoning about the code simpler.
export interface SeriesDisplayOptionsInputRequired extends SeriesDisplayOptionsInput {
    limit: number
    sortOptions: SeriesSortOptionsInput
}

export interface InsightDashboardReference {
    id: string
    title: string
}
