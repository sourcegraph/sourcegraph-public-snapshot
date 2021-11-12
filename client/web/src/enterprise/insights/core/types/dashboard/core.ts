/**
 * All insights dashboards are separated on three categories.
 */
export enum InsightsDashboardScope {
    /**
     * This type of dashboard includes insights from the personal settings or from
     * dashboards that are stored in the personal settings (personal dashboard)
     */
    Personal = 'personal',

    /**
     * This type of dashboard includes insights from the organization settings or from
     * dashboards that are stored in the organization settings (org dashboard)
     */
    Organization = 'organization',

    /**
     * This type of dashboard includes all insights from the site (global settings) or
     * from dashboards that are stored in global (site) settings subject.
     */
    Global = 'global',
}

export enum InsightsDashboardType {
    Virtual = 'virtual',
    BuiltIn = 'builtIn',
    Custom = 'custom',
}

/**
 * Information about dashboard owner. It can be a user-type subject (personal dashboard), org subject
 * (org level dashboard)
 */
export interface InsightDashboardOwner {
    id: string
    name: string
}

export interface InsightDashboard {
    id: string
    type: InsightsDashboardType
    scope: InsightsDashboardScope
    title: string
    insightIds?: string[]

    /**
     * Subject that has a particular dashboard, it can be personal setting
     * or organization setting subject.
     */
    owner?: InsightDashboardOwner

    grants?: {
        users?: string[]
        organizations?: string[]
        global?: boolean
    }
}
