import { createContext } from 'react'

import { InsightDashboard } from '../../../../core/types'

export interface BackendInsightContextData {
    currentDashboard: InsightDashboard | null
}

const DEFAULT_CONTEXT = {
    currentDashboard: null,
}

export const BackendInsightContext = createContext<BackendInsightContextData>(DEFAULT_CONTEXT)
