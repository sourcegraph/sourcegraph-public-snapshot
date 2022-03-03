import { InsightDashboard, InsightDashboardOwner, InsightsDashboardType } from './core'

/**
 * Derived dashboard from the setting cascade subject.
 */
export interface BuiltInInsightDashboard extends InsightDashboard {
    /**
     * Property to distinguish between real user-created dashboards and
     * built-in dashboards. Currently, we support 3 types of user built-in dashboard.
     *
     * "Personal" - all personal insights from personal settings (all users
     * have it by default)
     *
     * "Organizations level" - all organizations act as an insights dashboard.
     *
     * "Global level" - all insights from site (global) setting subject.
     */
    type: InsightsDashboardType.BuiltIn

    /**
     * Subject that has a particular dashboard, it can be personal setting
     * or organization setting subject.
     */
    owner: InsightDashboardOwner
}

/**
 * Explicitly created in the settings cascade insights dashboard.
 */
export interface CustomInsightDashboard extends InsightDashboard {
    type: InsightsDashboardType.Custom

    /**
     * Subject that has a particular dashboard, it can be personal setting
     * or organization setting subject.
     */
    owner?: InsightDashboardOwner
}

/**
 * Insights dashboards that were created in a user/org settings.
 */
export type RealInsightDashboard = CustomInsightDashboard | BuiltInInsightDashboard

export const isBuiltInInsightDashboard = (dashboard: RealInsightDashboard): dashboard is BuiltInInsightDashboard =>
    dashboard.type === InsightsDashboardType.BuiltIn

export const isCustomInsightDashboard = (dashboard: RealInsightDashboard): dashboard is CustomInsightDashboard =>
    dashboard.type === InsightsDashboardType.Custom
