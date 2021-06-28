import { InsightDashboard as InsightDashboardConfiguration } from '../../../schema/settings.schema'

/**
 * Visibility setting which responsible for where dashboard will appear.
 * possible value 'personal' | '<org id 1> ... | ... <org id N>'
 */
export type InsightDashboardVisibility = string

export interface InsightDashboard extends InsightDashboardConfiguration {
    /**
     * Subject that has a particular dashboard, it can be personal setting
     * or organization setting subject.
     */
    owner: InsightDashboardOwner
}

export interface InsightDashboardOwner {
    id: string | null
    name: string
}
