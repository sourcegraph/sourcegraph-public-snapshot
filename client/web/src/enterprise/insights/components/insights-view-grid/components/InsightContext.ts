import { createContext } from 'react'

import { InsightDashboard } from '../../../core'

export interface DashboardInsightsContextData {
    currentDashboard: InsightDashboard | null
    dashboards: InsightDashboard[]
}

export const InsightContext = createContext<DashboardInsightsContextData>({
    currentDashboard: null,
    dashboards: [],
})
