import { InsightType } from './common'
import { isLangStatsdInsightId, LangStatsInsight, LangStatsInsightConfiguration } from './lang-stat-insight'
import {
    isSearchBasedInsightId,
    SearchBasedInsight,
    SearchBasedInsightConfiguration,
    SearchBasedExtensionInsightSettings,
    SearchExtensionBasedInsight,
} from './search-insight'

export * from './common'

export const INSIGHTS_ALL_REPOS_SETTINGS_KEY = 'insights.allrepos'

export type Insight = SearchBasedInsight | LangStatsInsight
export type ExtensionInsight = SearchExtensionBasedInsight | LangStatsInsight

export type InsightExtensionBasedConfiguration =
    // Since lang stat insight doesn't have be version
    LangStatsInsightConfiguration | SearchBasedExtensionInsightSettings

export type InsightConfiguration = SearchBasedInsightConfiguration | LangStatsInsightConfiguration
export type { SearchBasedInsight, LangStatsInsight }

// Type and settings insight guards.

export function isInsightSettingKey(key: string): boolean {
    return isSearchBasedInsightId(key) || isLangStatsdInsightId(key)
}

export function isExtensionInsight(insight: Insight): insight is ExtensionInsight {
    return insight.type === InsightType.Extension
}

export function isSearchBasedInsight(possibleInsight: { id: string }): possibleInsight is SearchBasedInsight {
    return isSearchBasedInsightId(possibleInsight.id)
}

export function isLangStatsInsight(insight: Insight): insight is LangStatsInsight {
    return isLangStatsdInsightId(insight.id)
}
