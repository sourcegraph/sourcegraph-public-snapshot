import { InsightsDashboardType, type VirtualInsightsDashboard } from './core'

/**
 * Special virtual dashboard - "All Insights". This dashboard doesn't
 * exist in settings or in BE database.
 */
export const ALL_INSIGHTS_DASHBOARD: VirtualInsightsDashboard = {
    id: 'all',
    type: InsightsDashboardType.Virtual,
    title: 'All Insights',
}

export const MAX_NUMBER_OF_SERIES = 20
export const MAX_NUMBER_OF_SAMPLES = 90

export const DATA_SERIES_COLORS = {
    RED: 'var(--oc-red-7)',
    PINK: 'var(--oc-pink-7)',
    GRAPE: 'var(--oc-grape-7)',
    VIOLET: 'var(--oc-violet-7)',
    INDIGO: 'var(--oc-indigo-7)',
    BLUE: 'var(--oc-blue-7)',
    CYAN: 'var(--oc-cyan-7)',
    TEAL: 'var(--oc-teal-7)',
    GREEN: 'var(--oc-green-7)',
    LIME: 'var(--oc-lime-7)',
    YELLOW: 'var(--oc-yellow-7)',
    ORANGE: 'var(--oc-orange-7)',
}

export const DATA_SERIES_COLORS_LIST = Object.values(DATA_SERIES_COLORS)
export const DEFAULT_DATA_SERIES_COLOR = DATA_SERIES_COLORS.GRAPE
