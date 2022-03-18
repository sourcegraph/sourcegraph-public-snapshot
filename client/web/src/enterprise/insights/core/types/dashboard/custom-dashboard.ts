import { InsightsDashboardScope, InsightsDashboardType } from './core'

export interface CustomInsightDashboard {
    id: string
    type: InsightsDashboardType.Custom
    scope: InsightsDashboardScope
    title: string
    insightIds: string[]
    grants: {
        users: string[]
        organizations: string[]
        global: boolean
    }
}
