import { LineChartContent as LineChartContentType, LineChartSeries } from 'sourcegraph'

interface SeriesWithQuery extends LineChartSeries<any> {
    name: string
    query: string
}

export interface InsightExampleCommonContent {
    title: string
    repositories: string
}

export interface SearchInsightExampleContent
    extends Omit<LineChartContentType<any, string>, 'chart' | 'series'>,
        InsightExampleCommonContent {
    series: SeriesWithQuery[]
}

export interface CaptureGroupExampleContent
    extends Omit<LineChartContentType<any, string>, 'chart'>,
        InsightExampleCommonContent {
    groupSearch: string
}
