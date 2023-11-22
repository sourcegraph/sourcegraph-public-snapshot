import { createContext } from 'react'

import type { InsightDashboard } from '../../../core'

export interface DashboardInsightsContextData {
    currentDashboard: InsightDashboard | null
}

export const InsightContext = createContext<DashboardInsightsContextData>({
    currentDashboard: null,
})
