import React from 'react'

import { createInsightAPI } from './create-insights-api'
import { ApiService } from './types'

export const InsightsApiContext = React.createContext<ApiService>(createInsightAPI())
