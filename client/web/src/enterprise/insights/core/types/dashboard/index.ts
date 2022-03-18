import { InsightsDashboardScope, InsightsDashboardType } from './core'
import { CustomInsightDashboard } from './custom-dashboard'
import { VirtualInsightsDashboard } from './virtual-dashboard'

export { InsightsDashboardScope, InsightsDashboardType }

export type {
    VirtualInsightsDashboard,
    CustomInsightDashboard,
}

/** Main insight dashboard definition */
export type InsightDashboard = CustomInsightDashboard | VirtualInsightsDashboard

// Type guards for code insights dashboards
export const isVirtualDashboard = (dashboard: InsightDashboard): dashboard is VirtualInsightsDashboard =>
    dashboard.type === InsightsDashboardType.Virtual

export const isCustomDashboard = (dashboard: InsightDashboard): dashboard is CustomInsightDashboard =>
    dashboard.type === InsightsDashboardType.Custom

// Scope dashboard selectors
export const isOrganizationDashboard = (dashboard: InsightDashboard): boolean =>
    dashboard.scope === InsightsDashboardScope.Organization

export const isPersonalDashboard = (dashboard: InsightDashboard): boolean =>
    dashboard.scope === InsightsDashboardScope.Personal

export const isGlobalDashboard = (dashboard: InsightDashboard): boolean =>
    dashboard.scope === InsightsDashboardScope.Global

