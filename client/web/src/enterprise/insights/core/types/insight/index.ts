import { CaptureGroupInsight } from './capture-group-insight'
import { InsightExecutionType, InsightType } from './common'
import { LangStatsInsight } from './lang-stat-insight'
import {
    SearchBackendBasedInsight,
    SearchBasedInsight,
    SearchBasedInsightSeries,
    SearchRuntimeBasedInsight,
} from './search-insight'

export * from './common'
export type { SearchBasedInsight, SearchBasedInsightSeries, LangStatsInsight, CaptureGroupInsight }

/**
 * Main insight model. Union of all different insights by execution type (backend, runtime)
 * and insight type (lang-stats, search based, capture group) insights.
 */
export type Insight = SearchBasedInsight | LangStatsInsight | CaptureGroupInsight

/**
 * Backend insights - insights that have all data series points already in gql API.
 */
export type BackendInsight = SearchBackendBasedInsight | CaptureGroupInsight

/**
 * Extension insights - insights that are processed in FE runtime via search API.
 */
export type RuntimeInsight = SearchRuntimeBasedInsight | LangStatsInsight

export function isBackendInsight(insight: Insight): insight is BackendInsight {
    return insight.executionType === InsightExecutionType.Backend
}

export function isSearchBasedInsight(insight: Insight): insight is SearchBasedInsight {
    return insight.type === InsightType.SearchBased
}

export function isCaptureGroupInsight(insight: Insight): insight is CaptureGroupInsight {
    return insight.type === InsightType.CaptureGroup
}

export function isLangStatsInsight(insight: Insight): insight is LangStatsInsight {
    return insight.type === InsightType.LangStats
}
