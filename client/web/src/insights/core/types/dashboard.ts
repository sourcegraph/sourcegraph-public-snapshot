import { InsightDashboard as InsightDashboardConfiguration } from '../../../schema/settings.schema'

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
 * An extended custom insights dashboard. A user can have they own dashboards created
 * by insights dashboard creation UI.
 */
export interface InsightDashboard extends InsightDashboardConfiguration {
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

    /**
     * Property to distinguish between real user-created dashboard and virtual
     * built-in dashboard. Currently we support 3 types of built-in dashboard.
     *
     * "All" - all insights that dashboard have all users by default
     *
     * "Personal" - all personal insights from personal settings (also all users
     * it have by default)
     *
     * "Organizations level" - all organizations act as insights dashboard
     */
    builtIn?: boolean
}

/**
 * An Owner of dashboard. It can be user subject (a personal dashboard), org subject
 * (an org level dashboard)
 */
export interface InsightDashboardOwner {
    id: string
    name: string
}

/**
 * The key for accessing  insights dashboards in the subject settings.
 */
export const INSIGHTS_DASHBOARDS_SETTINGS_KEY = 'insights.dashboards'
