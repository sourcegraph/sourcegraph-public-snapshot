import React from 'react'

import { InsightsAPI } from './insights-api'
import { ApiService } from './types'

export const InsightsApiContext = React.createContext<ApiService>(new InsightsAPI())
