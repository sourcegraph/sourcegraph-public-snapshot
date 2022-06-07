import { InsightExecutionType, InsightType, InsightFilters, InsightDashboardReference } from './common'
import { CaptureGroupInsight } from './types/capture-group-insight'
import { LangStatsInsight } from './types/lang-stat-insight'
import { SearchBasedInsight, SearchBasedInsightSeries } from './types/search-insight'

export { InsightType, InsightExecutionType }

export type {
    InsightDashboardReference,
    SearchBasedInsight,
    SearchBasedInsightSeries,
    LangStatsInsight,
    CaptureGroupInsight,
    InsightFilters,
}

/**
 * Main insight model. Union of all different insights by execution type (backend, runtime)
 * and insight type (lang-stats, search based, capture group) insights.
 */
export type Insight = SearchBasedInsight | LangStatsInsight | CaptureGroupInsight

/**
 * Backend insights - insights that have all data series points already in gql API.
 */
export type BackendInsight = SearchBasedInsight | CaptureGroupInsight

/**
 * Extension insights - insights that are processed in FE runtime via search API.
 */
export type RuntimeInsight = LangStatsInsight

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
