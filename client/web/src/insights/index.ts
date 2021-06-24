// Pages exports
export { InsightsPage } from './pages/insights/InsightsPage'

// Core insights exports
export { InsightsApiContext } from './core/backend/api-provider'
export * from './core/analytics'

// Public Insights components
export { InsightsViewGrid } from './components'
export { InsightsRouter } from './InsightsRouter'
export type { InsightsViewGridProps } from './components'

// Guard
export { isCodeInsightsEnabled } from './utils/is-code-insights-enabled'
