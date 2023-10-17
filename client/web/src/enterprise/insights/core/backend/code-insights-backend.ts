import type { Observable } from 'rxjs'

import type { Insight, InsightsDashboardOwner } from '../types'

import type {
    AssignInsightsToDashboardInput,
    DashboardCreateInput,
    DashboardCreateResult,
    DashboardDeleteInput,
    DashboardUpdateInput,
    DashboardUpdateResult,
    InsightCreateInput,
    InsightUpdateInput,
    RemoveInsightFromDashboardInput,
} from './code-insights-backend-types'

/**
 * The main interface for code insights backend. Each backend versions should
 * implement this interface in order to support all functionality that code insights
 * pages and components have.
 */
export interface CodeInsightsBackend {
    /**
     * Returns all possible visibility options for dashboard. Dashboard can be stored
     * as private (user subject), org level (organization subject) or global (site subject)
     */
    getDashboardOwners: () => Observable<InsightsDashboardOwner[]>

    createDashboard: (input: DashboardCreateInput) => Observable<DashboardCreateResult>

    updateDashboard: (input: DashboardUpdateInput) => Observable<DashboardUpdateResult>

    deleteDashboard: (input: DashboardDeleteInput) => Observable<void>

    assignInsightsToDashboard: (input: AssignInsightsToDashboardInput) => Observable<unknown>

    /**
     * Return all accessible for a user insights that are filtered by ids param.
     * If ids is nullable value then returns all insights. Insights in this case
     * present only insight configurations and metadata without actual data about
     * data series or pie chart data.
     *
     * @param ids - list of insight ids
     */
    getInsights: (input: { dashboardId: string; withCompute: boolean }) => Observable<Insight[]>

    /**
     * Return insight (meta and presentation data) by insight id.
     * Note that insight model doesn't contain any data series points.
     */
    getInsightById: (id: string) => Observable<Insight | null>

    getActiveInsightsCount: (insightsCount: number) => Observable<number>

    createInsight: (input: InsightCreateInput) => Observable<unknown>

    updateInsight: (event: InsightUpdateInput) => Observable<unknown>

    deleteInsight: (insightId: string) => Observable<unknown>

    removeInsightFromDashboard: (input: RemoveInsightFromDashboardInput) => Observable<unknown>
}
