import { Remote } from 'comlink'
import { Duration } from 'date-fns'
import { Observable } from 'rxjs'
import { LineChartContent, PieChartContent } from 'sourcegraph'

import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'

import { SearchBasedInsightSeries } from '../types/insight/search-insight'

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

/**
 * Backend insight filters is subset of search based backend filters.
 * We don't have repo list filter support yet. Only regexp filters are
 * supported.
 */
export interface BackendInsightFilters {
    excludeRepoRegexp: string | null
    includeRepoRegexp: string | null
}

export interface BackendInsightInputs {
    id: string
    filters?: BackendInsightFilters
    series?: SearchBasedInsightSeries[]
}

export interface LangStatsInsightsSettings {
    /**
     * URL of git repository from which statistics will be collected
     */
    repository: string

    /**
     * The threshold below which a language is counted as part of 'Other'
     */
    threshold: number
}

export interface ApiService {
    /**
     * Basic method to get backend and extension based insights together.
     * Used by the insights page and other non-insights specific consumers
     * homepage, directory pages.
     *
     * @param getExtensionsInsights - extensions based insights getter via extension API.
     * @param backendInsightsIds - specific dashboard subset of BE-like insight ids.
     */
    getCombinedViews: (
        getExtensionsInsights: () => Observable<ViewProviderResult[]>,
        backendInsightsIds?: string[]
    ) => Observable<ViewInsightProviderResult[]>

    /**
     * Returns backend insight (via gql API handler) by insight id.
     */
    getBackendInsightById: (inputs: BackendInsightInputs) => Observable<BackendInsightData>

    /**
     * Returns resolved extension provider result by extension view id via extension API.
     */
    getExtensionViewById: (
        id: string,
        extensionApi: Promise<Remote<FlatExtensionHostAPI>>
    ) => Observable<ViewInsightProviderResult>

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
    getSearchInsightContent: (insight: SearchInsightSettings) => Promise<LineChartContent<any, string>>

    /**
     * Returns content for the code stats insight live preview chart.
     *
     * @param insight - An insight configuration (title, repos, data series settings)
     */
    getLangStatsInsightContent: (insight: LangStatsInsightsSettings) => Promise<PieChartContent<any>>

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
