export type InsightDashboard = VirtualInsightsDashboard | CustomInsightDashboard

export enum InsightsDashboardType {
    Virtual = 'virtual',
    Custom = 'custom',
}

/**
 * Special 'virtual' dashboard that includes all insights from the personal and organization and global
 * level dashboards. Virtual dashboard doesn't exist in settings but lives only in a runtime.
 */
export interface VirtualInsightsDashboard {
    id: string
    type: InsightsDashboardType.Virtual
    title: string
}

export interface CustomInsightDashboard {
    id: string
    type: InsightsDashboardType.Custom
    title: string
    owners: InsightsDashboardOwner[]
}

export enum InsightsDashboardOwnerType {
    Personal = 'personal',
    Organization = 'organization',
    Global = 'global',
}

export interface InsightsDashboardOwner {
    type: InsightsDashboardOwnerType
    id: string
    title: string
}

// Type guards for code insights dashboards
export const isVirtualDashboard = (dashboard: InsightDashboard): dashboard is VirtualInsightsDashboard =>
    dashboard.type === InsightsDashboardType.Virtual

export const isCustomDashboard = (dashboard: InsightDashboard): dashboard is CustomInsightDashboard =>
    dashboard.type === InsightsDashboardType.Custom

// Scope dashboard selectors
export const isPersonalDashboard = (dashboard: CustomInsightDashboard): boolean =>
    dashboard.owners.some(isPersonalOwner)

export const isOrganizationDashboard = (dashboard: CustomInsightDashboard): boolean =>
    dashboard.owners.some(isOrganizationOwner)

export const isGlobalDashboard = (dashboard: CustomInsightDashboard): boolean => dashboard.owners.some(isGlobalOwner)

export const isPersonalOwner = (owner: InsightsDashboardOwner): boolean =>
    owner.type === InsightsDashboardOwnerType.Personal

export const isOrganizationOwner = (owner: InsightsDashboardOwner): boolean =>
    owner.type === InsightsDashboardOwnerType.Organization

export const isGlobalOwner = (owner: InsightsDashboardOwner): boolean =>
    owner.type === InsightsDashboardOwnerType.Global
