import { Observable } from 'rxjs'
import { LineChartContent, PieChartContent } from 'sourcegraph'

import { ViewContexts, ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'

import { Insight, InsightDashboard } from '../types'
import { SearchBackendBasedInsight } from '../types/insight/search-insight'
import { SupportedInsightSubject } from '../types/subjects'

import {
    BackendInsightData,
    DashboardCreateInput,
    DashboardDeleteInput,
    DashboardUpdateInput,
    FindInsightByNameInput,
    GetBuiltInsightInput,
    GetLangStatsInsightContentInput,
    GetSearchInsightContentInput,
    InsightCreateInput,
    InsightUpdateInput,
    ReachableInsight,
    RepositorySuggestionData,
} from './code-insights-backend-types'

/**
 * The main interface for code insights backend. Each backend versions should
 * implement this interface in order to support all functionality that code insights
 * pages and components have.
 */
export interface CodeInsightsBackend {
    /**
     * Returns all accessible code insights dashboards for the current user.
     * This includes virtual (like "all insights") and real dashboards.
     */
    getDashboards: () => Observable<InsightDashboard[]>

    getDashboardById: (dashboardId?: string) => Observable<InsightDashboard | null>

    findDashboardByName: (name: string) => Observable<InsightDashboard | null>

    createDashboard: (input: DashboardCreateInput) => Observable<void>

    updateDashboard: (input: DashboardUpdateInput) => Observable<void>

    deleteDashboard: (input: DashboardDeleteInput) => Observable<void>

    /**
     * Return all accessible for a user insights that are filtered by ids param.
     * If ids is nullable value then returns all insights.
     *
     * @param ids - list of insight ids
     */
    getInsights: (ids?: string[]) => Observable<Insight[]>

    /**
     * Returns all reachable subject's insights from subject with subjectId.
     *
     * User subject has access to all insights from all organizations and global site settings.
     * Organization subject has access to only its insights.
     */
    getReachableInsights: (subjectId: string) => Observable<ReachableInsight[]>

    getInsightById: (id: string) => Observable<Insight | null>

    findInsightByName: (input: FindInsightByNameInput) => Observable<Insight | null>

    createInsight: (input: InsightCreateInput) => Observable<void>

    updateInsight: (event: InsightUpdateInput) => Observable<void[]>

    deleteInsight: (insightId: string) => Observable<void[]>

    /**
     * Returns all available for users subjects (sharing levels, historically it was introduced
     * from the setting cascade subject levels - global, org levels, personal)
     */
    getInsightSubjects: () => Observable<SupportedInsightSubject[]>

    /**
     * Returns backend insight (via gql API handler)
     */
    getBackendInsightData: (insight: SearchBackendBasedInsight) => Observable<BackendInsightData>

    /**
     * Returns extension like built-in insight that is fetched via frontend
     * network requests to Sourcegraph search API.
     */
    getBuiltInInsightData: <D extends keyof ViewContexts>(
        input: GetBuiltInsightInput<D>
    ) => Observable<ViewProviderResult>

    /**
     * Returns content for the search based insight live preview chart.
     */
    getSearchInsightContent: <D extends keyof ViewContexts>(
        input: GetSearchInsightContentInput<D>
    ) => Promise<LineChartContent<any, string>>

    /**
     * Returns content for the code stats insight live preview chart.
     */
    getLangStatsInsightContent: <D extends keyof ViewContexts>(
        input: GetLangStatsInsightContentInput<D>
    ) => Promise<PieChartContent<any>>

    /**
     * Returns a list of suggestions for the repositories field in the insight creation UI.
     *
     * @param query - A string with a possible value for the repository name
     */
    getRepositorySuggestions: (query: string) => Promise<RepositorySuggestionData[]>

    /**
     * Returns a list of resolved repositories from the search page query via search API.
     * Used by 1-click insight creation flow. Since users can have a repo: filter in their
     * query we have to resolve these filters by our search API.
     *
     * @param query - search page query value
     */
    getResolvedSearchRepositories: (query: string) => Promise<string[]>
}
