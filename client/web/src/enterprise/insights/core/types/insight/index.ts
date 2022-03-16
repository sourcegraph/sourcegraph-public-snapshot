import { CaptureGroupInsight } from './capture-group-insight'
import { InsightExecutionType, InsightType } from './common'
import { isLangStatsdInsightId, LangStatsInsight, LangStatsInsightConfiguration } from './lang-stat-insight'
import {
    isSearchBasedInsightId,
    SearchBackendBasedInsight,
    SearchBasedExtensionInsightSettings,
    SearchBasedInsight,
    SearchBasedInsightSeries,
    SearchBasedInsightConfiguration,
    SearchExtensionBasedInsight,
} from './search-insight'

export * from './common'
export type { SearchBasedInsight, SearchBasedInsightSeries, LangStatsInsight, CaptureGroupInsight }

/**
 * Main insight model. Union of all different insights by execution type (backend, runtime)
 * and insight type (lang-stats, search based, capture group) insights.
 */
export type Insight = SearchBasedInsight | LangStatsInsight | CaptureGroupInsight

/**
 * Extension insights - insights that are processed in FE runtime via search API.
 * TODO: Move all insights processing to the BE gql handler to simplify FE runtime.
 */
export type ExtensionInsight = SearchExtensionBasedInsight | LangStatsInsight

/**
 * Backend insights - insights that have all data series points already in gql API.
 */
export type BackendInsight = SearchBackendBasedInsight | CaptureGroupInsight

export function isBackendInsight(insight: Insight): insight is BackendInsight {
    return insight.type === InsightExecutionType.Backend
}

export function isExtensionInsight(insight: Insight): insight is ExtensionInsight {
    return insight.type === InsightExecutionType.Runtime
}

export function isSearchBasedInsight(insight: Insight): insight is SearchBasedInsight {
    return insight.viewType === InsightType.SearchBased
}

export function isCaptureGroupInsight(insight: Insight): insight is CaptureGroupInsight {
    return insight.viewType === InsightType.CaptureGroup
}

export function isLangStatsInsight(insight: Insight): insight is LangStatsInsight {
    return insight.viewType === InsightType.LangStats
}

// Setting-based api specific types and models.
// TODO: Remove these types when setting based api will be deprecated

export const INSIGHTS_ALL_REPOS_SETTINGS_KEY = 'insights.allrepos'

export type InsightExtensionBasedConfiguration = LangStatsInsightConfiguration | SearchBasedExtensionInsightSettings

export type InsightConfiguration = SearchBasedInsightConfiguration | LangStatsInsightConfiguration

/**
 * This function returns insight type based on insight settings id naming convention.
 * In the setting based API we store insights in setting cascade (which is jsonc file)
 * with keys (<specialInsightTypePrefix>.<camelCasedTitle>)
 */
export function parseInsightTypeFromSettingId(insightId: string): InsightType | null {
    if (isSearchBasedInsightId(insightId)) {
        return InsightType.SearchBased
    }

    if (isLangStatsdInsightId(insightId)) {
        return InsightType.LangStats
    }

    return null
}
