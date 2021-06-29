import { InsightDashboard as InsightDashboardConfiguration } from '../../../schema/settings.schema'

export type InsightDashboard = InsightBuiltInDashboard | InsightCustomDashboard

/**
 * All insights dashboards are separated on two categories.
 */
export enum InsightsDashboardType {
    /**
     * A dashboard that includes all insights from the personal and organization
     * level dashboards.
     */
    All = 'all',

    /**
     * A dashboard that includes insights from personal settings or from
     * dashboards that are stored in personal settings (personal dashboard)
     */
    Personal = 'personal',

    /**
     * A dashboard that includes insights from organization settings or from
     * dashboards that are stored in organization settings (org dashboard)
     */
    Organization = 'organization',
}

/**
 * An Owner of dashboard. It can be user subject (a personal dashboard), org subject
 * (an org level dashboard)
 */
export interface InsightDashboardOwner {
    id: string | null
    name: string
}

/**
 * An extended custom insights dashboard. A user can have they own dashboards created
 * by insights dashboard creation UI.
 */
export interface InsightCustomDashboard extends InsightDashboardConfiguration {
    /**
     * All dashboards that were created in users or org settings explicitly are
     * custom dashboards.
     */
    type: InsightsDashboardType.Personal | InsightsDashboardType.Organization

    /**
     * Subject that has a particular dashboard, it can be personal setting
     * or organization setting subject.
     */
    owner: InsightDashboardOwner
}

/**
 * A built-in insights dashboard. Currently we have 3 type of built-in dashboards.
 * All users have "All Insights" dashboard and "Personal Insights" dashboard
 * Also users who were included in org have built-in org level dashboard by default.
 */
export interface InsightBuiltInDashboard {
    type: InsightsDashboardType

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
    type: InsightsDashboardType.All,
}

/**
 * The key for accessing  insights dashboards in the subject settings.
 */
export const INSIGHTS_DASHBOARDS_SETTINGS_KEY = 'insights.dashboards'
