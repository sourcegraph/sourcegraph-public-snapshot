/**
 * Visibility setting which responsible for where dashboard will appear.
 * possible value 'personal' | '<org id 1> ... | ... <org id N>'
 */
export type InsightDashboardVisibility = string

export interface InsightDashboard {
    id: string
    title: string
    visibility: InsightDashboardVisibility
    insightsIds: string[]
}

export interface InsightDashboardConfiguration {
    title: string
    insightsIds: string[]
}

export const INSIGHT_DASHBOARD_PREFIX = 'insights.dashboard'
