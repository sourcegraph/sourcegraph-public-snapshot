import { createContext } from 'react'

import { CodeInsightsFakeBackend } from './create-insights-api'
import { CodeInsightsBackend } from './types'

export const InsightsApiContext = createContext<CodeInsightsBackend>(new CodeInsightsFakeBackend())
