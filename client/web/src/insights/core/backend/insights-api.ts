import { getCombinedViews, getInsightCombinedViews } from './api/get-combined-views';
import { getSearchInsightContent } from './api/get-search-insight-content';
import { getSubjectSettings, updateSubjectSettings } from './api/subject-settings';
import { ApiService } from './types'

/**
 * Main API service to get data for code insights
 * */
export const createInsightAPI = (): ApiService => ({
    getCombinedViews,
    getInsightCombinedViews,
    getSubjectSettings,
    updateSubjectSettings,
    getSearchInsightContent
})
