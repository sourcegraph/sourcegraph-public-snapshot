import { Series } from '@sourcegraph/wildcard'

export interface BackendInsightDTO {
    isInProgress: boolean
    data: BackendInsightData
}

export type BackendInsightData = InsightSeriesContent<any> | InsightCategoricalContent<any>

export enum InsightContentType {
    Series,
    Categorical,
}

export interface InsightSeriesContent<Datum> {
    type: InsightContentType.Series
    series: Series<Datum>[]
}

export interface InsightCategoricalContent<Datum> {
    type: InsightContentType.Categorical
    content: CategoricalChartContent<Datum>
}

export interface CategoricalChartContent<Datum> {
    data: Datum[]
    getDatumValue: (datum: Datum) => number
    getDatumName: (datum: Datum) => string
    getDatumColor: (datum: Datum) => string | undefined
    getDatumLink?: (datum: Datum) => string | undefined
}
