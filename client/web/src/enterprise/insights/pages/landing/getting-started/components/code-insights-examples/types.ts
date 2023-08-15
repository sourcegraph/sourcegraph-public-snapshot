import type { Series, SeriesLikeChart } from '@sourcegraph/wildcard'

interface SeriesWithQuery<T> extends Series<T> {
    query: string
}

export interface InsightExampleCommonContent {
    title: string
    repositories: string
}

export interface SearchInsightExampleContent<T> extends SeriesLikeChart<T> {
    series: SeriesWithQuery<T>[]
    title: string
    repositories: string
}

export interface CaptureGroupExampleContent<T> extends SeriesLikeChart<T> {
    groupSearch: string
    title: string
    repositories: string
}
