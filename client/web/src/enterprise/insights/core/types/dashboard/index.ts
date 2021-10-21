import { InsightsDashboardType, InsightDashboardOwner } from './core'
import { GraphQlDashboard } from './graphql-dashboard'
import { RealInsightDashboard, SettingsBasedInsightDashboard } from './real-dashboard'
import { VirtualInsightsDashboard } from './virtual-dashboard'

/**
 * Main insight dashboard definition
 */
export type InsightDashboardSettingsApi = RealInsightDashboard | VirtualInsightsDashboard
export type InsightDashboardGraphQlApi = GraphQlDashboard

export { InsightsDashboardType }

export type { RealInsightDashboard, VirtualInsightsDashboard, SettingsBasedInsightDashboard, InsightDashboardOwner }

/**
 * Key for accessing insights dashboards in a subject settings.
 */
export const INSIGHTS_DASHBOARDS_SETTINGS_KEY = 'insights.dashboards'

// Type guards for code insights dashboards
export const isOrganizationDashboard = (
    dashboard: InsightDashboardSettingsApi | undefined
): dashboard is RealInsightDashboard => dashboard?.type === InsightsDashboardType.Organization

export const isPersonalDashboard = (
    dashboard: InsightDashboardSettingsApi | undefined
): dashboard is RealInsightDashboard => dashboard?.type === InsightsDashboardType.Personal

export const isGlobalDashboard = (
    dashboard: InsightDashboardSettingsApi | undefined
): dashboard is RealInsightDashboard => dashboard?.type === InsightsDashboardType.Global

export const isVirtualDashboard = (
    dashboard: InsightDashboardSettingsApi | undefined | null
): dashboard is VirtualInsightsDashboard => dashboard?.type === InsightsDashboardType.All

export const isRealDashboard = (
    dashboard: InsightDashboardSettingsApi | undefined
): dashboard is RealInsightDashboard =>
    isOrganizationDashboard(dashboard) || isPersonalDashboard(dashboard) || isGlobalDashboard(dashboard)
