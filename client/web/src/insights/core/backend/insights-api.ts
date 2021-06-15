import { of, throwError } from 'rxjs'

import { getCombinedViews, getInsightCombinedViews } from './api/get-combined-views'
import { getLangStatsInsightContent } from './api/get-lang-stats-insight-content'
import { getRepositorySuggestions } from './api/get-repository-suggestions'
import { getSearchInsightContent } from './api/get-search-insight-content/get-search-insight-content'
import { getSubjectSettings, updateSubjectSettings } from './api/subject-settings'
import { ApiService } from './types'

/**
 * Main API service to get data for code insights
 *
 * See {@link ApiService} for full description of each method.
 * */
export const createInsightAPI = (): ApiService => ({
    getCombinedViews,
    getInsightCombinedViews,
    getSubjectSettings,
    updateSubjectSettings,
    getSearchInsightContent,
    getLangStatsInsightContent,
    getRepositorySuggestions,
})

/**
 * Mock API service. Used to mock part or some specific api requests in demo and
 * storybook stories.
 * */
export const createMockInsightAPI = (overrideRequests: Partial<ApiService>): ApiService => ({
    getCombinedViews: () => of([]),
    getInsightCombinedViews: () => of([]),
    getSubjectSettings: () => throwError(new Error('Implement getSubjectSettings handler first')),
    updateSubjectSettings: () => throwError(new Error('Implement getSubjectSettings handler first')),
    getSearchInsightContent: () => Promise.reject(new Error('Implement getSubjectSettings handler first')),
    getLangStatsInsightContent: () => Promise.reject(new Error('Implement getLangStatsInsightContent handler first')),
    getRepositorySuggestions: () => Promise.reject(new Error('Implement getRepositorySuggestions handler first')),
    ...overrideRequests,
})
