import { Duration } from 'date-fns'

import { ConfiguredSubject } from '@sourcegraph/shared/src/settings/settings';

import { DataSeries } from '../backend/types'

import { RealInsightDashboard } from './dashboard';

export enum InsightTypePrefix {
    search = 'searchInsights.insight',
    langStats = 'codeStatsInsights.insight',
}

export type Insight = SearchBasedInsight | LangStatsInsight
export type InsightSettings = SearchBasedInsightOrigin | LangStatsInsightOrigin

/**
 * [Synthetic] fields that are needed only for the code insight logic but not for extension API
 */
interface SyntheticInsightFields {
    /**
     * ID of insight <type of insight>.insight.<name of insight>
     * */
    id: string

    /**
     * List of all dashboard that store an insight id.
     * */
    dashboards: RealInsightDashboard[]

    /**
     * Subject that has an insight configuration in its subject settings.
     */
    storeSubject: ConfiguredSubject

}

/**
 * Extended Search Insight.
 * Some fields and settings (id, visibility) do not exist implicitly in user/org settings but
 * we have to have these to operate with insight properly.
 * */
export interface SearchBasedInsight extends SearchBasedInsightOrigin, SyntheticInsightFields {}

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
export interface LangStatsInsight extends LangStatsInsightOrigin, SyntheticInsightFields {}

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

export function isInsightSettingKey(key: string): boolean {
    return key.startsWith(InsightTypePrefix.search) || key.startsWith(InsightTypePrefix.langStats)
}
