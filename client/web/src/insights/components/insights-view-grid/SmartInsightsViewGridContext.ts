import { createContext } from 'react';

import { InsightDashboard } from '../../core/types';

export interface SmartInsightsViewGridContextData {
    currentDashboard: InsightDashboard | null
}

const DEFAULT_CONTEXT = {
    currentDashboard: null
}

export const SmartInsightsViewGridContext = createContext<SmartInsightsViewGridContextData>(DEFAULT_CONTEXT);
