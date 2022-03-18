import { CustomInsightDashboard } from '../../../../../core/types';

export interface DashboardOrganizationGroup {
    id: string
    name: string
    dashboards: CustomDashboardWithOwner[]
}

export interface CustomDashboardWithOwner extends CustomInsightDashboard{
    owner: InsightDashboardOwner
}

export interface InsightDashboardOwner {
    id: string
    name: string
}
