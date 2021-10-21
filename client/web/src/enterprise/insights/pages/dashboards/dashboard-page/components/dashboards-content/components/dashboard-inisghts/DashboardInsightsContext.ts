import { createContext } from 'react'

import { InsightDashboardSettingsApi } from '../../../../../../../core/types'

export interface DashboardInsightsContextData {
    dashboard: InsightDashboardSettingsApi | null
}

export const DEFAULT_DASHBOARD_CONTEXT: DashboardInsightsContextData = {
    dashboard: null,
}

export const DashboardInsightsContext = createContext<DashboardInsightsContextData>(DEFAULT_DASHBOARD_CONTEXT)
