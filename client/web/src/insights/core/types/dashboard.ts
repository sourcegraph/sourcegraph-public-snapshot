import { InsightDashboard as InsightDashboardConfiguration } from '../../../schema/settings.schema'

export type InsightDashboard = InsightBuiltInDashboard | InsightCustomDashboard

/**
 * All insights dashboards are separated on two categories.
 */
export enum InsightsDashboardType {
    /**
     * A built in dashboard that all users have as a default and non-removable
     * dashboards like "all insights", "personal", "org level dashboard".
     */
    BuiltIn = 'built-in',

    /**
     * A Custom dashboard that a user created by dashboard creation UI.
     * These dashboard types can be deleted or moved from user personal settings
     * to org setting scope and vice versa.
     */
    Custom = 'custom',
}

/**
 * An extended custom insights dashboard configuration.
 */
export interface InsightCustomDashboard extends InsightDashboardConfiguration {
    /**
     * All dashboards that were created in users settings explicitly are
     * custom dashboards.
     */
    type: InsightsDashboardType.Custom

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

/**
 * A built in insights dashboard.
 */
export interface InsightBuiltInDashboard {
    type: InsightsDashboardType.BuiltIn

    /**
     * Title of insights dashboards ("All insights", "Personal", ...)
     */
    title?: string

    /**
     * Possible owner of dashboard. That might be equal to undefined when
     * this dashboard is "all insights" dashboard.
     */
    owner?: InsightDashboardOwner

    insightIds?: string[]
}

/**
 * A built-in type of dashboard that contains all insights from all settings level
 * like organizations level and personal settings.
 */
export const ALL_INSIGHTS_DASHBOARD: InsightBuiltInDashboard = {
    title: 'All',
    type: InsightsDashboardType.BuiltIn,
}
