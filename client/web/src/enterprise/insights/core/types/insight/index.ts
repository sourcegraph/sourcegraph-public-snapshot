import { InsightType, type InsightFilters, type InsightDashboardReference } from './common'
import type { CaptureGroupInsight } from './types/capture-group-insight'
import type { ComputeInsight } from './types/compute-insight'
import type { LangStatsInsight } from './types/lang-stat-insight'
import type { SearchBasedInsight, SearchBasedInsightSeries } from './types/search-insight'

export { InsightType }

export type {
    InsightDashboardReference,
    SearchBasedInsight,
    SearchBasedInsightSeries,
    LangStatsInsight,
    CaptureGroupInsight,
    ComputeInsight,
    InsightFilters,
}

/**
 * Main insight model. Union of all different insights by execution type (backend, runtime)
 * and insight type (lang-stats, search based, capture group) insights.
 */
export type Insight = SearchBasedInsight | LangStatsInsight | CaptureGroupInsight | ComputeInsight

/**
 * Backend insights - insights that have all data series points already in gql API.
 */
export type BackendInsight = SearchBasedInsight | CaptureGroupInsight | ComputeInsight

export function isBackendInsight(insight: Insight): insight is BackendInsight {
    return isSearchBasedInsight(insight) || isCaptureGroupInsight(insight) || isComputeInsight(insight)
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

export function isComputeInsight(insight: Insight): insight is ComputeInsight {
    return insight.type === InsightType.Compute
}
