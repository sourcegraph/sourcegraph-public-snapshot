import type { GroupByField } from '../../../../../../../graphql-operations'

export enum InsightType {
    Detect = 'Detect',
    DetectAndTrack = 'Detect and Track',
    Compute = 'Compute',
    LanguageStats = 'Language stats',
}

export interface DashboardInsight {
    id: string
    title: string
}

export interface DetectInsight extends DashboardInsight {
    type: InsightType.Detect
    queries: string[]
}

export interface DetectAndTrackInsight extends DashboardInsight {
    type: InsightType.DetectAndTrack
    query: string
}

export interface ComputeInsight extends DashboardInsight {
    type: InsightType.Compute
    query: string
    groupBy: GroupByField
}

export interface LanguageStatsInsight extends DashboardInsight {
    type: InsightType.LanguageStats
}

export type InsightSuggestion = DetectInsight | DetectAndTrackInsight | ComputeInsight | LanguageStatsInsight

export const getInsightId = (insight: DashboardInsight): string => insight.id
export const getInsightTitle = (insight: DashboardInsight): string => insight.title
