import { createContext } from 'react'

import { InsightDashboard } from '../../../../../../../core/types'

export interface DashboardInsightsContextData {
    dashboard: InsightDashboard | null
}

export const DEFAULT_DASHBOARD_CONTEXT: DashboardInsightsContextData = {
    dashboard: null,
}

export const DashboardInsightsContext = createContext<DashboardInsightsContextData>(DEFAULT_DASHBOARD_CONTEXT)
