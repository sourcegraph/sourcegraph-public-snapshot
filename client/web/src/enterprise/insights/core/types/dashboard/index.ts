import { InsightDashboardOwner, InsightsDashboardScope, InsightsDashboardType } from './core'
import { BuiltInInsightDashboard, CustomInsightDashboard, RealInsightDashboard } from './real-dashboard'
import { VirtualInsightsDashboard } from './virtual-dashboard'

export { InsightsDashboardScope, InsightsDashboardType }
export type {
    RealInsightDashboard,
    VirtualInsightsDashboard,
    CustomInsightDashboard,
    BuiltInInsightDashboard,
    InsightDashboardOwner,
}

/** Main insight dashboard definition */
export type InsightDashboard = RealInsightDashboard | VirtualInsightsDashboard

// Type guards for code insights dashboards
export const isVirtualDashboard = (dashboard: InsightDashboard): dashboard is VirtualInsightsDashboard =>
    dashboard.type === InsightsDashboardType.Virtual

export const isRealDashboard = (dashboard: InsightDashboard): dashboard is RealInsightDashboard =>
    dashboard.type === InsightsDashboardType.BuiltIn || dashboard.type === InsightsDashboardType.Custom

// Scope dashboard selectors
export const isOrganizationDashboard = (dashboard: InsightDashboard): boolean =>
    dashboard.scope === InsightsDashboardScope.Organization

export const isPersonalDashboard = (dashboard: InsightDashboard): boolean =>
    dashboard.scope === InsightsDashboardScope.Personal

export const isGlobalDashboard = (dashboard: InsightDashboard): boolean =>
    dashboard.scope === InsightsDashboardScope.Global
