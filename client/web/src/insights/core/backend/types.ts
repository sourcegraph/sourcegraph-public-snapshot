import { Duration } from 'date-fns'
import { Observable } from 'rxjs'
import { LineChartContent, PieChartContent } from 'sourcegraph'

import { ViewContexts, ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'

import { ExtensionInsight } from '../types'
import { SearchBackendBasedInsight, SearchBasedInsightSeries } from '../types/insight/search-insight'

import { RepositorySuggestion } from './requests/fetch-repository-suggestions'

export enum ViewInsightProviderSourceType {
    Backend = 'Backend',
    Extension = 'Extension',
}

/**
 * Unified insight data interface.
 */
export interface ViewInsightProviderResult extends ViewProviderResult {
    /**
     * The source of view provider to distinguish between data from extension
     * and data from backend
     */
    source: ViewInsightProviderSourceType
}

/**
 * Backend insight result data interface
 */
export interface BackendInsightData {
    id: string
    view: {
        title: string
        subtitle: string
        content: LineChartContent<any, string>[]
        isFetchingHistoricalData: boolean
    }
}

export interface SubjectSettingsResult {
    id: number | null
    contents: string
}

export interface SearchInsightSettings {
    series: SearchBasedInsightSeries[]
    step: Duration
    repositories: string[]
}

export interface LangStatsInsightsSettings {
    /**
     * URL of git repository from which statistics will be collected
     */
    repository: string

    /**
     * The threshold below which a language is counted as part of 'Other'
     */
    otherThreshold: number
}

export interface ApiService {
    /**
     * Returns backend insight (via gql API handler)
     */
    getBackendInsight: (insight: SearchBackendBasedInsight) => Observable<BackendInsightData>

    /**
     * Returns extension like built-in insight that is fetched via frontend
     * network requests to Sourcegraph search API.
     */
    getBuiltInInsight: <D extends keyof ViewContexts>(
        insight: ExtensionInsight,
        options: { where: D; context: ViewContexts[D] }
    ) => Observable<ViewProviderResult>

    /**
     * Finds and returns the subject settings by the subject id.
     *
     * @param id - subject id
     */
    getSubjectSettings: (id: string) => Observable<SubjectSettingsResult>

    /**
     * Updates the subject settings by the subject id.
     * Rehydrate local settings and call gql mutation
     *
     * @param context - global context object with updateSettings method
     * @param subjectId - subject id
     * @param content - a new settings content
     */
    updateSubjectSettings: (
        context: Pick<PlatformContext, 'updateSettings'>,
        subjectId: string,
        content: string
    ) => Observable<void>

    /**
     * Returns content for the search based insight live preview chart.
     *
     * @param insight - An insight configuration (title, repos, data series settings)
     */
    getSearchInsightContent: <D extends keyof ViewContexts>(
        insight: SearchInsightSettings,
        options: { where: D; context: ViewContexts[D] }
    ) => Promise<LineChartContent<any, string>>

    /**
     * Returns content for the code stats insight live preview chart.
     *
     * @param insight - An insight configuration (title, repos, data series settings)
     */
    getLangStatsInsightContent: <D extends keyof ViewContexts>(
        insight: LangStatsInsightsSettings,
        options: { where: D; context: ViewContexts[D] }
    ) => Promise<PieChartContent<any>>

    /**
     * Returns a list of suggestions for the repositories field in the insight creation UI.
     *
     * @param query - A string with a possible value for the repository name
     */
    getRepositorySuggestions: (query: string) => Promise<RepositorySuggestion[]>

    /**
     * Returns a list of resolved repositories from the search page query via search API.
     * Used by 1-click insight creation flow. Since users can have a repo: filter in their
     * query we have to resolve these filters by our search API.
     *
     * @param query - search page query value
     */
    getResolvedSearchRepositories: (query: string) => Promise<string[]>
}
