import type { Series } from '@sourcegraph/wildcard'

import type {
    CaptureGroupInsight,
    LangStatsInsight,
    InsightsDashboardOwner,
    SearchBasedInsight,
    ComputeInsight,
} from '../types'
import type { InsightContentType, IncompleteDatapointAlert } from '../types/insight/common'

export interface CategoricalChartContent<Datum> {
    data: Datum[]
    getDatumValue: (datum: Datum) => number
    getDatumName: (datum: Datum) => string
    getDatumColor: (datum: Datum) => string | undefined
    getDatumLink?: (datum: Datum) => string | undefined
    getCategory?: (datum: Datum) => string | undefined
}

export interface BackendInsightSeries<Datum> extends Series<Datum> {
    alerts: IncompleteDatapointAlert[]
}

export interface SeriesChartContent<Datum> {
    series: BackendInsightSeries<Datum>[]
}

export interface InsightCategoricalContent<Datum> {
    type: InsightContentType.Categorical
    content: CategoricalChartContent<Datum>
}

export interface InsightSeriesContent<Datum> {
    type: InsightContentType.Series
    series: BackendInsightSeries<Datum>[]
}

export type InsightContent<Datum> = InsightSeriesContent<Datum> | InsightCategoricalContent<Datum>

export interface DashboardCreateInput {
    name: string
    owners: InsightsDashboardOwner[]
}

export interface DashboardCreateResult {
    id: string
}

export interface DashboardUpdateInput {
    nextDashboardInput: DashboardCreateInput
    id: string
}

export interface AssignInsightsToDashboardInput {
    id: string
    prevInsightIds: string[]
    nextInsightIds: string[]
}

export interface DashboardUpdateResult {
    id: string
}

export interface DashboardDeleteInput {
    id: string
}

export type MinimalSearchBasedInsightData = Omit<SearchBasedInsight, 'id' | 'dashboardReferenceCount' | 'isFrozen'>
export type MinimalCaptureGroupInsightData = Omit<CaptureGroupInsight, 'id' | 'dashboardReferenceCount' | 'isFrozen'>
export type MinimalLangStatsInsightData = Omit<LangStatsInsight, 'id' | 'dashboardReferenceCount' | 'isFrozen'>
export type MinimalComputeInsightData = Omit<ComputeInsight, 'id' | 'dashboardReferenceCount' | 'isFrozen'>

export type CreationInsightInput =
    | MinimalSearchBasedInsightData
    | MinimalCaptureGroupInsightData
    | MinimalLangStatsInsightData
    | MinimalComputeInsightData

export interface InsightCreateInput {
    insight: CreationInsightInput
    dashboardId: string | null
}

export interface InsightUpdateInput {
    insightId: string
    nextInsightData: CreationInsightInput
}

export interface RemoveInsightFromDashboardInput {
    insightId: string
    dashboardId: string
}

export interface BackendInsightDatum {
    dateTime: Date
    value: number
    link?: string
}

export interface BackendInsightData {
    data: InsightContent<any>
    isFetchingHistoricalData: boolean
    incompleteAlert: IncompleteDatapointAlert | null
}
