import { createContext } from 'react'

import { InsightDashboard } from '../../../core'

export interface DashboardInsightsContextData {
    dashboard: InsightDashboard | null
}

export const InsightContext = createContext<DashboardInsightsContextData>({
    dashboard: null,
})
