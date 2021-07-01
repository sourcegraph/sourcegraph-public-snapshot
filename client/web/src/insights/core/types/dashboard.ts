/**
 * Visibility setting which responsible for where dashboard will appear.
 * possible value 'personal' | '<org id 1> ... | ... <org id N>'
 */
export type InsightDashboardVisibility = string

export interface InsightDashboard {
    id: string
    visibility: InsightDashboardVisibility
    insightsIds: string[]
}
