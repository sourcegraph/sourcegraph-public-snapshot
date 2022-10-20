import { Series } from '@sourcegraph/wildcard'

import { InsightDataSeriesStatus2 } from '../../../../../../graphql-operations';

export interface BackendInsightDTO {
    isInProgress: boolean
    status: InsightDataSeriesStatus2
    data: BackendInsightData
}

export type BackendInsightData = InsightSeriesContent | InsightCategoricalContent

export enum InsightContentType {
    Series,
    Categorical,
}

export interface InsightSeriesContent {
    type: InsightContentType.Series
    series: BackendInsightSeries[]
}

export interface BackendInsightSeries extends Series<BackendInsightDatum> {
    status: InsightDataSeriesStatus2
}

export interface BackendInsightDatum {
    dateTime: Date
    value: number
    link?: string
}

export interface InsightCategoricalContent {
    type: InsightContentType.Categorical
    content: CategoricalChartContent
}

export interface CategoricalChartContent {
    data: CategoricalDatum[]
    getDatumValue: (datum: CategoricalDatum) => number
    getDatumName: (datum: CategoricalDatum) => string
    getDatumColor: (datum: CategoricalDatum) => string | undefined
    getDatumLink?: (datum: CategoricalDatum) => string | undefined
}

export interface CategoricalDatum {
    name: string
    fill: string
    value: number
}
