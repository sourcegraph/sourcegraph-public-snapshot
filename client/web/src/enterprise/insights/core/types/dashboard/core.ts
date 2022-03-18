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
    Custom = 'custom',
}

