import { InsightDashboard as InsightDashboardConfiguration } from '../../../../schema/settings.schema'

/**
 * All insights dashboards are separated on three categories.
 */
export enum InsightsDashboardType {
    /**
     * The dashboard includes all insights from the personal and organization
     * settings.
     */
    All = 'all',

    /**
     * The dashboard includes insights from the personal settings or from
     * dashboards that are stored in the personal settings (personal dashboard)
     */
    Personal = 'personal',

    /**
     * The dashboard includes insights from the organization settings or from
     * dashboards that are stored in the organization settings (org dashboard)
     */
    Organization = 'organization',
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
 * An extended custom insights dashboard by payload info about type and owner of dashboard.
 */
export interface ExtendedInsightDashboard extends InsightDashboardConfiguration {
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
