import { Duration } from 'date-fns'

import { DataSeries } from '../backend/types'

export enum InsightTypePrefix {
    search = 'searchInsights.insight',
    langStats = 'codeStatsInsights.insight',
}

/**
 * Visibility setting which responsible for where insight will appear.
 * possible value 'personal' | '<org id 1> ... | ... <org id N>'
 * */
export type InsightVisibility = string

export type Insight = SearchBasedInsight | LangStatsInsight

/**
 * Extended Search Insight.
 * Some fields and settings (id, visibility) do not exist implicitly in user/org settings but
 * we have to have these to operate with insight properly.
 * */
export interface SearchBasedInsight extends SearchBasedInsightOrigin {
    /**
     * [Synthetic] field needed only for code insight logic but not for extension
     * ID of insight <type of insight>.insight.<name of insight>
     * */
    id: string

    /**
     * [Synthetic] field needed only for code insight logic but not for extension
     * Visibility of insight. Personal or organization setting cascade.
     * */
    visibility: InsightVisibility
}

/**
 * See public API of search insight extension
 * https://github.com/sourcegraph/sourcegraph-search-insights/blob/master/package.json#L26
 */
export interface SearchBasedInsightOrigin {
    title: string
    repositories: string[]
    series: DataSeries[]
    step: Duration
}

/**
 * Extended Lang Stats Insight.
 * Some fields and settings (id, visibility) do not exist implicitly in user/org settings but
 * we have to have these to operate with insight properly.
 * */
export interface LangStatsInsight extends LangStatsInsightOrigin {
    /**
     * [Synthetic] field needed only for code insight logic but not for extension
     * ID of insight <type of insight>.insight.<name of insight>
     * */
    id: string

    /**
     * [Synthetic] field needed only for code insight logic but not for extension
     * Visibility of insight. Personal or organization setting cascade.
     * */
    visibility: InsightVisibility
}

/**
 * Lang stats insight as it is presented in user/org settings cascade.
 * See public API of code stats extension
 * https://github.com/sourcegraph/sourcegraph-code-stats-insights/blob/master/package.json#L27
 */
export interface LangStatsInsightOrigin {
    title: string
    repository: string
    otherThreshold: number
}

export function isSearchBasedInsight(insight: Insight): insight is SearchBasedInsight {
    return insight.id.startsWith(InsightTypePrefix.search)
}

export function isLangStatsInsight(insight: Insight): insight is LangStatsInsight {
    return insight.id.startsWith(InsightTypePrefix.langStats)
}
